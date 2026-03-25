import asyncio
import logging
import time

import httpx
from tqdm.asyncio import tqdm

logger = logging.getLogger(__name__)


class EvalPipeline:
    def __init__(
        self,
        base_url: str,
        source_id: str,
        mode: str,
        poll_interval: float,
        timeout: float,
    ):
        self.base_url = base_url.rstrip("/")
        self.source_id = source_id
        self.mode = mode
        self.poll_interval = poll_interval
        self.timeout = timeout

    async def run_single(self, client: httpx.AsyncClient, question: str) -> dict:
        empty = {"answer": "", "sources": []}

        try:
            resp = await client.post(
                f"{self.base_url}/ask/query",
                json={
                    "query": question,
                    "sourceId": self.source_id,
                    "mode": self.mode,
                },
            )
            if resp.status_code == 429:
                retry_after = float(resp.headers.get("Retry-After", self.poll_interval))
                logger.warning("Rate limited on POST, waiting %.1fs", retry_after)
                await asyncio.sleep(retry_after)
                resp = await client.post(
                    f"{self.base_url}/ask/query",
                    json={
                        "query": question,
                        "sourceId": self.source_id,
                        "mode": self.mode,
                    },
                )
            resp.raise_for_status()
            query_id = resp.json()["queryId"]
        except Exception as e:
            logger.warning("POST /ask/query failed for '%s': %s", question[:50], e)
            return empty

        start = time.monotonic()
        while time.monotonic() - start < self.timeout:
            await asyncio.sleep(self.poll_interval)
            try:
                resp = await client.get(f"{self.base_url}/ask/query/{query_id}")
                if resp.status_code == 429:
                    retry_after = float(resp.headers.get("Retry-After", self.poll_interval))
                    await asyncio.sleep(retry_after)
                    continue
                resp.raise_for_status()
                data = resp.json()
            except Exception as e:
                logger.warning("GET /ask/query/%s failed: %s", query_id, e)
                return empty

            status = data.get("status")
            if status == "DONE":
                return {
                    "answer": data.get("answer", ""),
                    "sources": [s["text"] for s in data.get("sources", [])],
                }
            if status == "FAILED":
                logger.warning("Query FAILED for '%s'", question[:50])
                return empty

        logger.warning("Timeout polling query for '%s'", question[:50])
        return empty

    async def run_dataset(self, dataset: list[dict], concurrency: int) -> list[dict]:
        semaphore = asyncio.Semaphore(concurrency)

        async def _process(client: httpx.AsyncClient, item: dict) -> dict:
            async with semaphore:
                result = await self.run_single(client, item["question"])
                item["answer"] = result["answer"]
                item["sources"] = result["sources"]
                return item

        async with httpx.AsyncClient(timeout=httpx.Timeout(30.0)) as client:
            tasks = [_process(client, item) for item in dataset]
            results = await tqdm.gather(*tasks, desc="Querying API")

        return list(results)
