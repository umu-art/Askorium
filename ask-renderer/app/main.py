import asyncio
import logging
import signal

from app.config import config
from app.browser_manager import BrowserManager
from app.health import HealthServer
from app.rabbitmq.broker import Broker
from app.rabbitmq.consumer import RabbitMQConsumer
from app.rabbitmq.handler import MessageHandler

logging.basicConfig(
    level=getattr(logging, config.LOG_LEVEL, logging.INFO),
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
)
logger = logging.getLogger("main")


async def main() -> None:
    browser_manager = BrowserManager(
        max_concurrent_pages=config.MAX_CONCURRENT_PAGES,
        default_timeout_ms=config.DEFAULT_TIMEOUT_MS,
    )
    broker = Broker(config.AMQP_URL, config.RECONNECT_INTERVAL_S)
    consumer: RabbitMQConsumer | None = None

    main_task = asyncio.current_task()
    loop = asyncio.get_running_loop()
    shutdown_event = asyncio.Event()

    def _shutdown():
        shutdown_event.set()
        main_task.cancel()

    for sig in (signal.SIGINT, signal.SIGTERM):
        loop.add_signal_handler(sig, _shutdown)

    try:
        await browser_manager.get_browser()
        logger.info("browser pre-warmed")

        await broker.connect(prefetch_count=config.MAX_CONCURRENT_PAGES)

        input_queue = await broker.declare_queue_with_dlq(
            config.INPUT_QUEUE, config.MESSAGE_TTL_MS,
        )
        await broker.declare_queue_with_dlq(config.OUTPUT_QUEUE)

        handler = MessageHandler(browser_manager, broker, config.OUTPUT_QUEUE)
        consumer = RabbitMQConsumer(input_queue, handler)

        health = HealthServer()
        health.register("browser", browser_manager.health_check)
        health.register("rabbitmq", broker.health_check)
        await health.start(config.HEALTH_PORT)

        logger.info("ask-renderer starting")
        await consumer.run(shutdown_event)
    except asyncio.CancelledError:
        logger.info("interrupted by signal")
    finally:
        if consumer:
            await consumer.stop()
        await broker.close()
        await browser_manager.shutdown()
        logger.info("ask-renderer stopped")


if __name__ == "__main__":
    asyncio.run(main())