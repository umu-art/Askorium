import asyncio
import logging

import aio_pika
from aio_pika import ExchangeType

from app.config import config
from app.browser_manager import BrowserManager
from app.message_handler import MessageHandler

logger = logging.getLogger("rabbitmq_consumer")


class RabbitMQConsumer:
    def __init__(self, browser_manager: BrowserManager):
        self.browser_manager = browser_manager
        self.connection = None
        self.channel = None
        self.input_queue = None
        self.consumer_tag = None
        self.handler: MessageHandler | None = None
        self.active_tasks: set[asyncio.Task] = set()
        self.shutdown_event = asyncio.Event()

    async def health_check(self) -> bool:
        return self.connection is not None and not self.connection.is_closed

    async def connect(self):
        self.connection = await aio_pika.connect_robust(
            config.AMQP_URL,
            reconnect_interval=config.RECONNECT_INTERVAL_S,
        )
        logger.info("connected to RabbitMQ")
        self.channel = await self.connection.channel()
        await self.channel.set_qos(prefetch_count=config.MAX_CONCURRENT_PAGES)

    async def setup_queues(self):
        dlx = await self.channel.declare_exchange(
            "render.dlx", ExchangeType.FANOUT, durable=True
        )
        dlq = await self.channel.declare_queue("render.dlq", durable=True)
        await dlq.bind(dlx)

        self.input_queue = await self.channel.declare_queue(
            config.INPUT_QUEUE,
            durable=True,
            arguments={
                "x-dead-letter-exchange": "render.dlx",
                "x-message-ttl": config.MESSAGE_TTL_MS,
            },
        )

        await self.channel.declare_queue(config.OUTPUT_QUEUE, durable=True)
        logger.info("queues declared")

    def _task_done(self, task: asyncio.Task) -> None:
        self.active_tasks.discard(task)
        if not task.cancelled() and task.exception():
            logger.error("unhandled error in message task", exc_info=task.exception())

    async def _on_message(self, message: aio_pika.abc.AbstractIncomingMessage) -> None:
        task = asyncio.create_task(self.handler.handle(message))
        self.active_tasks.add(task)
        task.add_done_callback(self._task_done)

    async def start(self):
        await self.connect()
        await self.setup_queues()
        self.handler = MessageHandler(self.browser_manager, self.channel, config.OUTPUT_QUEUE)
        self.consumer_tag = await self.input_queue.consume(self._on_message)
        logger.info(
            "consuming from %s (prefetch=%d)",
            config.INPUT_QUEUE,
            config.MAX_CONCURRENT_PAGES,
        )

        await self.shutdown_event.wait()
        logger.info("shutting down consumer...")

    async def stop(self):
        if self.consumer_tag and self.input_queue:
            await self.input_queue.cancel(self.consumer_tag)
        if self.active_tasks:
            logger.info("waiting for %d active tasks…", len(self.active_tasks))
            await asyncio.gather(*self.active_tasks, return_exceptions=True)
        if self.channel:
            await self.channel.close()
        if self.connection:
            await self.connection.close()
        logger.info("consumer stopped")