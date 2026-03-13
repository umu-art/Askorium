import logging

import aio_pika
from aio_pika import ExchangeType, DeliveryMode, Message

logger = logging.getLogger("rabbitmq.broker")


class Broker:
    def __init__(self, url: str, reconnect_interval: int = 5):
        self._url = url
        self._reconnect_interval = reconnect_interval
        self.connection: aio_pika.abc.AbstractRobustConnection | None = None
        self.channel: aio_pika.abc.AbstractChannel | None = None

    async def connect(self, prefetch_count: int = 1):
        self.connection = await aio_pika.connect_robust(
            self._url,
            reconnect_interval=self._reconnect_interval,
            fail_fast=False,
        )
        logger.info("connected to RabbitMQ")
        self.channel = await self.connection.channel()
        await self.channel.set_qos(prefetch_count=prefetch_count)

    async def declare_queue_with_dlq(
            self,
            queue_name: str,
            message_ttl_ms: int | None = None,
    ) -> aio_pika.abc.AbstractQueue:
        dlx_name = f"{queue_name}.dlx"
        dlq_name = f"{queue_name}.dlq"

        dlx = await self.channel.declare_exchange(
            dlx_name, ExchangeType.FANOUT, durable=True,
        )
        dlq = await self.channel.declare_queue(dlq_name, durable=True)
        await dlq.bind(dlx)

        arguments: dict[str, str | int] = {"x-dead-letter-exchange": dlx_name}
        if message_ttl_ms is not None:
            arguments["x-message-ttl"] = message_ttl_ms

        queue = await self.channel.declare_queue(
            queue_name, durable=True, arguments=arguments,
        )
        logger.info("declared queue %s with DLQ %s", queue_name, dlq_name)
        return queue

    async def declare_queue(self, queue_name: str) -> aio_pika.abc.AbstractQueue:
        queue = await self.channel.declare_queue(queue_name, durable=True)
        logger.info("declared queue %s", queue_name)
        return queue

    async def publish(self, routing_key: str, body: bytes):
        await self.channel.default_exchange.publish(
            Message(
                body=body,
                delivery_mode=DeliveryMode.PERSISTENT,
                content_type="application/json",
            ),
            routing_key=routing_key,
        )

    async def health_check(self) -> bool:
        return self.connection is not None and not self.connection.is_closed

    async def close(self):
        if self.channel:
            await self.channel.close()
        if self.connection:
            await self.connection.close()
        logger.info("RabbitMQ connection closed")
