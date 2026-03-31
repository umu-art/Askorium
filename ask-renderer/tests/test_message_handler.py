"""
Unit tests for MessageHandler.

All external dependencies (BrowserManager, DocumentRenderer, Broker, AMQP message)
are replaced with simple fakes — no real network or broker needed.
"""
import json
import pytest
from unittest.mock import AsyncMock, MagicMock, patch

from app.browser_manager import ErrorCode, RenderError
from app.rabbitmq.handler import MessageHandler


# ---------------------------------------------------------------------------
# Fakes
# ---------------------------------------------------------------------------

def _make_message(body: dict | None, *, valid: bool = True) -> MagicMock:
    """Return a fake aio_pika message."""
    msg = MagicMock()
    if valid and body is not None:
        msg.body = json.dumps(body).encode()
    else:
        msg.body = b"not-json"
    msg.ack = AsyncMock()
    msg.nack = AsyncMock()
    return msg


def _render_input_body(
    task_id: str = "t1",
    url: str = "https://example.com",
    timeout_ms: int = 5000,
    metadata: dict | None = None,
) -> dict:
    return {
        "task_id": task_id,
        "url": url,
        "timeout_ms": timeout_ms,
        "metadata": metadata,
    }


def _make_handler(
    render_html_result=None,
    render_html_side_effect=None,
    render_doc_result=None,
    publish_side_effect=None,
) -> tuple[MessageHandler, MagicMock, MagicMock]:
    browser = MagicMock()
    doc_renderer = MagicMock()
    broker = MagicMock()
    broker.publish = AsyncMock(side_effect=publish_side_effect)

    handler = MessageHandler(
        browser_manager=browser,
        document_renderer=doc_renderer,
        broker=broker,
        output_queue="askorium.render.output",
    )

    # Patch internal render methods for higher-level tests
    if render_html_result is not None or render_html_side_effect is not None:
        handler._render_html = AsyncMock(
            return_value=render_html_result,
            side_effect=render_html_side_effect,
        )
    if render_doc_result is not None:
        handler._render_document = AsyncMock(return_value=render_doc_result)

    return handler, broker, browser


def _make_render_output(success: bool = True, task_id: str = "t1", url: str = "https://example.com"):
    from python_renderer_api.models.render_output import RenderOutput
    from python_renderer_api.models.render_error import RenderError as ApiError
    from python_renderer_api.models.render_error_code import RenderErrorCode

    if success:
        return RenderOutput(task_id=task_id, url=url, success=True, html="<html/>", elapsed_ms=100)
    else:
        return RenderOutput(
            task_id=task_id, url=url, success=False,
            error=ApiError(code=RenderErrorCode.NETWORK_ERROR, message="failed"),
        )


# ---------------------------------------------------------------------------
# Tests: handle() — routing and ack/nack
# ---------------------------------------------------------------------------

class TestHandleRouting:
    @pytest.mark.asyncio
    async def test_invalid_json_nacks_without_requeue(self):
        handler, broker, _ = _make_handler()
        msg = _make_message(None, valid=False)

        await handler.handle(msg)

        msg.nack.assert_called_once_with(requeue=False)
        msg.ack.assert_not_called()
        broker.publish.assert_not_called()

    @pytest.mark.asyncio
    async def test_html_hint_routes_to_render_html(self):
        output = _make_render_output(success=True)
        handler, broker, _ = _make_handler(render_html_result=output)

        msg = _make_message(_render_input_body(metadata={"content_type_hint": "html"}))
        await handler.handle(msg)

        handler._render_html.assert_called_once()
        msg.ack.assert_called_once()

    @pytest.mark.asyncio
    async def test_no_hint_routes_to_render_html(self):
        output = _make_render_output(success=True)
        handler, broker, _ = _make_handler(render_html_result=output)

        msg = _make_message(_render_input_body(metadata=None))
        await handler.handle(msg)

        handler._render_html.assert_called_once()

    @pytest.mark.asyncio
    async def test_pdf_hint_routes_to_render_document(self):
        output = _make_render_output(success=True)
        handler, broker, _ = _make_handler(render_doc_result=output)
        # also patch _render_html so we can verify it's NOT called
        handler._render_html = AsyncMock()

        msg = _make_message(_render_input_body(metadata={"content_type_hint": "pdf"}))
        await handler.handle(msg)

        handler._render_document.assert_called_once()
        handler._render_html.assert_not_called()
        msg.ack.assert_called_once()

    @pytest.mark.asyncio
    async def test_docx_hint_routes_to_render_document(self):
        output = _make_render_output(success=True)
        handler, broker, _ = _make_handler(render_doc_result=output)
        handler._render_html = AsyncMock()

        msg = _make_message(_render_input_body(metadata={"content_type_hint": "docx"}))
        await handler.handle(msg)

        handler._render_document.assert_called_once()
        handler._render_html.assert_not_called()

    @pytest.mark.asyncio
    async def test_publish_failure_nacks_with_requeue(self):
        output = _make_render_output(success=True)
        handler, broker, _ = _make_handler(
            render_html_result=output,
            publish_side_effect=Exception("broker down"),
        )

        msg = _make_message(_render_input_body())
        await handler.handle(msg)

        msg.nack.assert_called_once_with(requeue=True)
        msg.ack.assert_not_called()

    @pytest.mark.asyncio
    async def test_successful_render_acks_message(self):
        output = _make_render_output(success=True)
        handler, broker, _ = _make_handler(render_html_result=output)

        msg = _make_message(_render_input_body())
        await handler.handle(msg)

        msg.ack.assert_called_once()
        msg.nack.assert_not_called()


# ---------------------------------------------------------------------------
# Tests: _render_html — retry logic
# ---------------------------------------------------------------------------

class TestRenderHtmlRetry:
    @pytest.mark.asyncio
    async def test_success_on_first_attempt(self):
        browser = MagicMock()
        browser.render_page = AsyncMock(return_value="<html/>")
        broker = MagicMock()
        broker.publish = AsyncMock()
        handler = MessageHandler(browser, MagicMock(), broker, "q")

        with patch("app.rabbitmq.handler.config") as cfg:
            cfg.MAX_RETRIES = 3
            result = await handler._render_html("t1", "https://example.com", 5000, None)

        assert result.success is True
        assert browser.render_page.call_count == 1

    @pytest.mark.asyncio
    async def test_retries_on_timeout(self):
        browser = MagicMock()
        browser.render_page = AsyncMock(side_effect=[
            RenderError(ErrorCode.TIMEOUT, "timed out"),
            RenderError(ErrorCode.TIMEOUT, "timed out"),
            "<html>ok</html>",
        ])
        handler = MessageHandler(browser, MagicMock(), MagicMock(), "q")

        with patch("app.rabbitmq.handler.config") as cfg:
            cfg.MAX_RETRIES = 3
            result = await handler._render_html("t1", "https://example.com", 5000, None)

        assert result.success is True
        assert browser.render_page.call_count == 3

    @pytest.mark.asyncio
    async def test_retries_on_network_error(self):
        browser = MagicMock()
        browser.render_page = AsyncMock(side_effect=[
            RenderError(ErrorCode.NETWORK_ERROR, "connection refused"),
            "<html>ok</html>",
        ])
        handler = MessageHandler(browser, MagicMock(), MagicMock(), "q")

        with patch("app.rabbitmq.handler.config") as cfg:
            cfg.MAX_RETRIES = 3
            result = await handler._render_html("t1", "https://example.com", 5000, None)

        assert result.success is True
        assert browser.render_page.call_count == 2

    @pytest.mark.asyncio
    async def test_no_retry_on_page_not_found(self):
        browser = MagicMock()
        browser.render_page = AsyncMock(
            side_effect=RenderError(ErrorCode.PAGE_NOT_FOUND, "404")
        )
        handler = MessageHandler(browser, MagicMock(), MagicMock(), "q")

        with patch("app.rabbitmq.handler.config") as cfg:
            cfg.MAX_RETRIES = 3
            result = await handler._render_html("t1", "https://example.com/missing", 5000, None)

        assert result.success is False
        assert browser.render_page.call_count == 1

    @pytest.mark.asyncio
    async def test_no_retry_on_access_denied(self):
        browser = MagicMock()
        browser.render_page = AsyncMock(
            side_effect=RenderError(ErrorCode.ACCESS_DENIED, "403")
        )
        handler = MessageHandler(browser, MagicMock(), MagicMock(), "q")

        with patch("app.rabbitmq.handler.config") as cfg:
            cfg.MAX_RETRIES = 3
            result = await handler._render_html("t1", "https://example.com/private", 5000, None)

        assert result.success is False
        assert browser.render_page.call_count == 1

    @pytest.mark.asyncio
    async def test_exhausted_retries_returns_failure(self):
        browser = MagicMock()
        browser.render_page = AsyncMock(
            side_effect=RenderError(ErrorCode.TIMEOUT, "always times out")
        )
        handler = MessageHandler(browser, MagicMock(), MagicMock(), "q")

        with patch("app.rabbitmq.handler.config") as cfg:
            cfg.MAX_RETRIES = 2
            result = await handler._render_html("t1", "https://example.com", 5000, None)

        assert result.success is False
        # 1 initial + 2 retries = 3 total
        assert browser.render_page.call_count == 3

    @pytest.mark.asyncio
    async def test_metadata_passed_through_to_output(self):
        browser = MagicMock()
        browser.render_page = AsyncMock(return_value="<html/>")
        handler = MessageHandler(browser, MagicMock(), MagicMock(), "q")
        meta = {"source": "test", "content_type_hint": "html"}

        with patch("app.rabbitmq.handler.config") as cfg:
            cfg.MAX_RETRIES = 0
            result = await handler._render_html("t1", "https://example.com", None, meta)

        assert result.metadata == meta


# ---------------------------------------------------------------------------
# Tests: _render_document
# ---------------------------------------------------------------------------

class TestRenderDocument:
    @pytest.mark.asyncio
    async def test_success(self):
        doc_renderer = MagicMock()
        doc_renderer.render = AsyncMock(return_value="<html><body>doc</body></html>")
        handler = MessageHandler(MagicMock(), doc_renderer, MagicMock(), "q")

        result = await handler._render_document("t1", "https://example.com/doc.pdf", "pdf", None)

        assert result.success is True
        assert result.html == "<html><body>doc</body></html>"
        doc_renderer.render.assert_called_once_with("https://example.com/doc.pdf", "pdf")

    @pytest.mark.asyncio
    async def test_failure_returns_parse_error(self):
        doc_renderer = MagicMock()
        doc_renderer.render = AsyncMock(side_effect=ValueError("unsupported document type: pptx"))
        handler = MessageHandler(MagicMock(), doc_renderer, MagicMock(), "q")

        result = await handler._render_document("t1", "https://example.com/file.pptx", "pptx", None)

        assert result.success is False
        assert result.error is not None
        assert "pptx" in result.error.message

    @pytest.mark.asyncio
    async def test_metadata_passed_through(self):
        doc_renderer = MagicMock()
        doc_renderer.render = AsyncMock(return_value="<html/>")
        handler = MessageHandler(MagicMock(), doc_renderer, MagicMock(), "q")
        meta = {"x": 1}

        result = await handler._render_document("t1", "https://example.com/doc.pdf", "pdf", meta)

        assert result.metadata == meta
