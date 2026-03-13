import json
import logging

import aio_pika

from app.parser import parse
from python_parser_api.models.error_code import ErrorCode
from python_parser_api.models.scrape_error import ScrapeError
from python_parser_api.models.scrape_response import ScrapeResponse
from python_renderer_api.models.render_output import RenderOutput
from app.rabbitmq.broker import Broker

logger = logging.getLogger("app.handler")


class MessageHandler:
    def __init__(self, broker: Broker, output_queue: str):
        self.broker = broker
        self.output_queue = output_queue

    async def handle(self, message: aio_pika.abc.AbstractIncomingMessage) -> None:
        try:
            body = json.loads(message.body)
            render_output = RenderOutput.from_dict(body)
        except Exception as e:
            logger.error("invalid message, sending to DLQ: %s", e)
            await message.nack(requeue=False)
            return

        task_id = render_output.task_id
        metadata = render_output.metadata

        if not render_output.success:
            error_code = render_output.error.code.value if render_output.error else "UNKNOWN"
            error_message = render_output.error.message if render_output.error else "unknown render error"
            response = ScrapeResponse(
                task_id=task_id,
                success=False,
                error=ScrapeError(
                    code=ErrorCode(error_code),
                    message=error_message,
                ),
                metadata=metadata,
            )
        else:
            try:
                page = parse(render_output.html, render_output.url)
                response = ScrapeResponse(
                    task_id=task_id,
                    success=True,
                    page=page,
                    metadata=metadata,
                )
            except Exception as e:
                logger.error("task=%s parse error: %s", task_id, e)
                response = ScrapeResponse(
                    task_id=task_id,
                    success=False,
                    error=ScrapeError(
                        code=ErrorCode.PARSE_ERROR,
                        message=str(e),
                    ),
                    metadata=metadata,
                )

        try:
            out_body = json.dumps(response.to_dict(), ensure_ascii=False).encode()
            await self.broker.publish(self.output_queue, out_body)
        except Exception as e:
            logger.error("task=%s publish failed, requeuing: %s", task_id, e)
            await message.nack(requeue=True)
            return

        await message.ack()
        logger.info("task=%s done success=%s", task_id, response.success)
