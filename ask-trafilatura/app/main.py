import asyncio
import logging
import signal

from app.config import config
from app.handler import MessageHandler
from app.rabbitmq.broker import Broker
from app.rabbitmq.consumer import RabbitMQConsumer

logger = logging.getLogger("app.main")


async def main():
    logging.basicConfig(
        level=getattr(logging, config.LOG_LEVEL, logging.INFO),
        format="%(asctime)s %(levelname)s %(name)s: %(message)s",
    )

    broker = Broker(config.AMQP_URL, reconnect_interval=config.RECONNECT_INTERVAL_S)
    await broker.connect(prefetch_count=config.PREFETCH_COUNT)

    input_queue = await broker.declare_queue_with_dlq(config.INPUT_QUEUE)
    await broker.declare_queue_with_dlq(config.OUTPUT_QUEUE)

    handler = MessageHandler(broker, config.OUTPUT_QUEUE)
    consumer = RabbitMQConsumer(input_queue, handler)

    shutdown_event = asyncio.Event()

    loop = asyncio.get_running_loop()
    for sig in (signal.SIGINT, signal.SIGTERM):
        loop.add_signal_handler(sig, shutdown_event.set)

    try:
        await consumer.run(shutdown_event)
    finally:
        await consumer.stop()
        await broker.close()
        logger.info("shutdown complete")


if __name__ == "__main__":
    asyncio.run(main())
