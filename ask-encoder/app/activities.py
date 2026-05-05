import asyncio

from temporalio import activity

from app.config import settings
from app.model_loader import get_embedder, get_reranker

_sem = asyncio.Semaphore(settings.max_concurrent_requests)


class EncoderActivities:

    @activity.defn(name="generateEmbeddings")
    async def generate_embeddings(self, texts: list[str]) -> list[list[float]]:
        async with _sem:
            embedder = get_embedder(settings)
            loop = asyncio.get_event_loop()
            vectors = await loop.run_in_executor(
                None, lambda: embedder.encode(texts, normalize_embeddings=True)
            )
            return vectors.tolist()

    @activity.defn(name="rerank")
    async def rerank(self, query: str, blocks: list[dict]) -> list[dict]:
        async with _sem:
            reranker = get_reranker(settings)
            pairs = [(query, b["text"]) for b in blocks]
            loop = asyncio.get_event_loop()
            scores = await loop.run_in_executor(None, lambda: reranker.predict(pairs))
            return sorted(
                [{"id": b["id"], "score": float(s)} for b, s in zip(blocks, scores)],
                key=lambda r: r["score"],
                reverse=True,
            )
