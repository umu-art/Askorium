import asyncio
import logging
import sys

from temporalio.client import Client
from temporalio.worker import Worker

from app.activities import EncoderActivities
from app.config import settings
from app.model_loader import get_embedder, get_reranker

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")


async def main():
    if "--fetch" in sys.argv:
        get_embedder(settings)
        get_reranker(settings)
        return

    get_embedder(settings)
    get_reranker(settings)

    client = await Client.connect(settings.temporal_url, tls=True)
    activities = EncoderActivities()

    worker = Worker(
        client,
        task_queue=settings.task_queue,
        activities=[activities.generate_embeddings, activities.rerank],
    )
    await worker.run()


if __name__ == "__main__":
    asyncio.run(main())
