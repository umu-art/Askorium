import json
import logging
import time

import aio_pika

from app.browser_manager import BrowserManager, ErrorCode, RenderError
from app.document_renderer import DocumentRenderer
from app.config import config
from app.metrics import (
    renderer_active_tasks,
    renderer_duration_seconds,
    renderer_errors_total,
    renderer_messages_total,
    renderer_retries_total,
)
from app.rabbitmq.broker import Broker
from python_renderer_api.models.render_error import RenderError as ApiRenderError
from python_renderer_api.models.render_error_code import RenderErrorCode
from python_renderer_api.models.render_input import RenderInput
from python_renderer_api.models.render_output import RenderOutput

logger = logging.getLogger("rabbitmq.handler")

_RETRYABLE_CODES = {ErrorCode.TIMEOUT, ErrorCode.NETWORK_ERROR}


class MessageHandler:
    def __init__(self, browser_manager: BrowserManager, document_renderer: DocumentRenderer, broker: Broker, output_queue: str):
        self.browser_manager = browser_manager
        self.document_renderer = document_renderer
        self.broker = broker
        self.output_queue = output_queue

    async def _render_html(self, task_id: str, url: str, timeout_ms: int | None, metadata: dict | None) -> RenderOutput:
        max_attempts = config.MAX_RETRIES + 1
        last_error: RenderError | None = None

        for attempt in range(1, max_attempts + 1):
            start = time.monotonic()
            try:
                html = await self.browser_manager.render_page(url, timeout_ms)
                elapsed_s = time.monotonic() - start

                renderer_duration_seconds.observe(elapsed_s)
                renderer_messages_total.labels(status="success").inc()

                return RenderOutput(
                    task_id=task_id, url=url, success=True,
                    html=html, elapsed_ms=int(elapsed_s * 1000),
                    metadata=metadata,
                )
            except RenderError as e:
                last_error = e
                if e.code in _RETRYABLE_CODES and attempt < max_attempts:
                    logger.warning(
                        "task=%s attempt=%d/%d [%s]: %s",
                        task_id, attempt, max_attempts, e.code.value, e.message,
                    )
                    renderer_retries_total.inc()
                    continue
                break

        logger.warning("task=%s render failed [%s]: %s", task_id, last_error.code.value, last_error.message)
        renderer_messages_total.labels(status="error").inc()
        renderer_errors_total.labels(code=last_error.code.value).inc()

        return RenderOutput(
            task_id=task_id, url=url, success=False,
            error=ApiRenderError(
                code=RenderErrorCode(last_error.code.value),
                message=last_error.message,
            ),
            metadata=metadata,
        )

    async def _render_document(self, task_id: str, url: str, hint: str, metadata: dict | None) -> RenderOutput:
        start = time.monotonic()
        try:
            html = await self.document_renderer.render(url, hint)
            elapsed_s = time.monotonic() - start
            renderer_duration_seconds.observe(elapsed_s)
            renderer_messages_total.labels(status="success").inc()
            return RenderOutput(
                task_id=task_id, url=url, success=True,
                html=html, elapsed_ms=int(elapsed_s * 1000),
                metadata=metadata,
            )
        except Exception as e:
            logger.warning("task=%s document render failed: %s", task_id, e)
            renderer_messages_total.labels(status="error").inc()
            renderer_errors_total.labels(code="PARSE_ERROR").inc()
            return RenderOutput(
                task_id=task_id, url=url, success=False,
                error=ApiRenderError(code=RenderErrorCode.PARSE_ERROR, message=str(e)),
                metadata=metadata,
            )

    async def handle(self, message: aio_pika.abc.AbstractIncomingMessage) -> None:
        try:
            body = json.loads(message.body)
            render_input = RenderInput.from_dict(body)
        except Exception as e:
            logger.error("invalid message, sending to DLQ: %s", e)
            renderer_messages_total.labels(status="dlq").inc()
            await message.nack(requeue=False)
            return

        renderer_active_tasks.inc()
        metadata = render_input.metadata
        hint = (metadata or {}).get("content_type_hint", "")

        try:
            if hint and hint != "html":
                render_output = await self._render_document(
                    render_input.task_id, render_input.url, hint, metadata,
                )
            else:
                render_output = await self._render_html(
                    render_input.task_id, render_input.url,
                    render_input.timeout_ms, metadata,
                )
        finally:
            renderer_active_tasks.dec()

        try:
            out_body = json.dumps(render_output.to_dict()).encode()
            await self.broker.publish(self.output_queue, out_body)
        except Exception as e:
            logger.error("task=%s publish failed, requeuing: %s", render_input.task_id, e)
            await message.nack(requeue=True)
            return

        await message.ack()
        logger.info("task=%s done success=%s", render_input.task_id, render_output.success)
