import asyncio
import json
import logging
import time

import aio_pika
from aio_pika import DeliveryMode, Message

from app.browser_manager import BrowserManager, ErrorCode, RenderError
from app.config import config
from app.metrics import (
    renderer_active_tasks,
    renderer_duration_seconds,
    renderer_errors_total,
    renderer_messages_total,
    renderer_retries_total,
)
from python_renderer_api.models.render_error import RenderError as ApiRenderError
from python_renderer_api.models.render_error_code import RenderErrorCode
from python_renderer_api.models.render_input import RenderInput
from python_renderer_api.models.render_output import RenderOutput

logger = logging.getLogger("message_handler")

_RETRYABLE_CODES = {ErrorCode.TIMEOUT, ErrorCode.NETWORK_ERROR}


class MessageHandler:
    def __init__(self, browser_manager: BrowserManager, channel: aio_pika.abc.AbstractChannel, output_queue: str):
        self.browser_manager = browser_manager
        self.channel = channel
        self.output_queue = output_queue

    async def handle(self, message: aio_pika.abc.AbstractIncomingMessage) -> None:
        try:
            body = json.loads(message.body)
            render_input = RenderInput.from_dict(body)
        except Exception as e:
            logger.error("invalid message, sending to DLQ: %s", e)
            renderer_messages_total.labels(status="dlq").inc()
            await message.nack(requeue=False)
            return

        task_id = render_input.task_id
        url = render_input.url
        timeout_ms = render_input.timeout_ms
        metadata = render_input.metadata

        render_output: RenderOutput | None = None
        max_attempts = config.MAX_RETRIES + 1

        renderer_active_tasks.inc()
        try:
            for attempt in range(1, max_attempts + 1):
                start = time.monotonic()
                try:
                    html = await self.browser_manager.render_page(url, timeout_ms)
                    elapsed_s = time.monotonic() - start
                    elapsed_ms = int(elapsed_s * 1000)

                    renderer_duration_seconds.observe(elapsed_s)
                    renderer_messages_total.labels(status="success").inc()

                    render_output = RenderOutput(
                        task_id=task_id,
                        url=url,
                        success=True,
                        html=html,
                        elapsed_ms=elapsed_ms,
                        metadata=metadata,
                    )
                    break

                except RenderError as e:
                    if e.code in _RETRYABLE_CODES and attempt < max_attempts:
                        logger.warning(
                            "task=%s attempt=%d/%d retryable error [%s]: %s",
                            task_id, attempt, max_attempts, e.code.value, e.message,
                        )
                        renderer_retries_total.inc()
                        continue

                    logger.warning(
                        "task=%s render failed [%s]: %s", task_id, e.code.value, e.message,
                    )
                    renderer_messages_total.labels(status="error").inc()
                    renderer_errors_total.labels(code=e.code.value).inc()
                    render_output = RenderOutput(
                        task_id=task_id,
                        url=url,
                        success=False,
                        error=ApiRenderError(
                            code=RenderErrorCode(e.code.value),
                            message=e.message,
                        ),
                        metadata=metadata,
                    )
                    break
        finally:
            renderer_active_tasks.dec()

        try:
            out_body = json.dumps(render_output.to_dict()).encode()
            await self.channel.default_exchange.publish(
                Message(
                    body=out_body,
                    delivery_mode=DeliveryMode.PERSISTENT,
                    content_type="application/json",
                ),
                routing_key=self.output_queue,
            )
        except Exception as e:
            logger.error("task=%s publish failed, requeuing: %s", task_id, e)
            await message.nack(requeue=True)
            return

        await message.ack()
        logger.info("task=%s done success=%s", task_id, render_output.success)