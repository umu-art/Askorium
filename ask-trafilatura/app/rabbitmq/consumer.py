import asyncio
import logging

import aio_pika

from app.handler import MessageHandler

logger = logging.getLogger("rabbitmq.consumer")


class RabbitMQConsumer:
    def __init__(
        self,
        input_queue: aio_pika.abc.AbstractQueue,
        handler: MessageHandler,
    ):
        self.input_queue = input_queue
        self.handler = handler
        self.consumer_tag: str | None = None
        self.active_tasks: set[asyncio.Task] = set()

    def _task_done(self, task: asyncio.Task) -> None:
        self.active_tasks.discard(task)
        if not task.cancelled() and task.exception():
            logger.error("unhandled error in message task", exc_info=task.exception())

    async def _on_message(self, message: aio_pika.abc.AbstractIncomingMessage) -> None:
        task = asyncio.create_task(self.handler.handle(message))
        self.active_tasks.add(task)
        task.add_done_callback(self._task_done)

    async def run(self, shutdown_event: asyncio.Event):
        self.consumer_tag = await self.input_queue.consume(self._on_message)
        logger.info("consuming from %s", self.input_queue.name)
        await shutdown_event.wait()
        logger.info("shutting down consumer...")

    async def stop(self):
        if self.consumer_tag and self.input_queue:
            await self.input_queue.cancel(self.consumer_tag)
        if self.active_tasks:
            logger.info("waiting for %d active tasks…", len(self.active_tasks))
            await asyncio.gather(*self.active_tasks, return_exceptions=True)
        logger.info("consumer stopped")