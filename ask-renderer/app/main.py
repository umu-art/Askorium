import asyncio
import logging
import signal

from app.config import config
from app.browser_manager import BrowserManager
from app.health import HealthServer
from app.rabbitmq_consumer import RabbitMQConsumer

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
    consumer = RabbitMQConsumer(browser_manager)

    await browser_manager.get_browser()
    logger.info("browser pre-warmed")

    health = HealthServer()
    health.register("browser", browser_manager.health_check)
    health.register("rabbitmq", consumer.health_check)
    await health.start(config.HEALTH_PORT)

    loop = asyncio.get_running_loop()
    for sig in (signal.SIGINT, signal.SIGTERM):
        loop.add_signal_handler(sig, consumer.shutdown_event.set)

    logger.info("ask-renderer starting")
    try:
        await consumer.start()
    finally:
        await health.stop()
        await consumer.stop()
        await browser_manager.shutdown()
        logger.info("ask-renderer stopped")


if __name__ == "__main__":
    asyncio.run(main())