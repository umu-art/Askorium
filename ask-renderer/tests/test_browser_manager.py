"""
Unit tests for browser_manager module.

Tests focus on the pure _classify_error function and BrowserManager.health_check,
which can be tested without launching a real browser.
"""
import pytest

from playwright.async_api import (
    Error as PlaywrightError,
    TimeoutError as PlaywrightTimeoutError,
)

from app.browser_manager import _classify_error, ErrorCode, RenderError, BrowserManager


def make_pw_error(msg: str) -> PlaywrightError:
    """Create a PlaywrightError with a given message string."""
    err = PlaywrightError(msg)
    return err


def make_timeout_error(msg: str = "Timeout exceeded") -> PlaywrightTimeoutError:
    err = PlaywrightTimeoutError(msg)
    return err


class TestClassifyError:
    def test_timeout_error_returns_timeout_code(self):
        err = make_timeout_error()
        result = _classify_error(err, "https://example.com")
        assert result.code == ErrorCode.TIMEOUT
        assert "example.com" in result.message

    def test_name_not_resolved_returns_network_error(self):
        err = make_pw_error("net::ERR_NAME_NOT_RESOLVED")
        result = _classify_error(err, "https://example.com")
        assert result.code == ErrorCode.NETWORK_ERROR

    def test_connection_refused_returns_network_error(self):
        err = make_pw_error("net::ERR_CONNECTION_REFUSED")
        result = _classify_error(err, "https://example.com")
        assert result.code == ErrorCode.NETWORK_ERROR

    def test_aborted_with_404_returns_page_not_found(self):
        err = make_pw_error("net::ERR_ABORTED 404")
        result = _classify_error(err, "https://example.com/missing")
        assert result.code == ErrorCode.PAGE_NOT_FOUND
        assert "missing" in result.message

    def test_403_returns_access_denied(self):
        err = make_pw_error("403 Forbidden")
        result = _classify_error(err, "https://example.com/private")
        assert result.code == ErrorCode.ACCESS_DENIED

    def test_401_returns_access_denied(self):
        err = make_pw_error("401 Unauthorized")
        result = _classify_error(err, "https://example.com/secure")
        assert result.code == ErrorCode.ACCESS_DENIED

    def test_invalid_url_returns_invalid_url(self):
        err = make_pw_error("net::ERR_INVALID_URL")
        result = _classify_error(err, "not-a-url")
        assert result.code == ErrorCode.INVALID_URL

    def test_invalid_url_text_returns_invalid_url(self):
        err = make_pw_error("invalid url provided")
        result = _classify_error(err, "not-a-url")
        assert result.code == ErrorCode.INVALID_URL

    def test_unknown_error_falls_back_to_network_error(self):
        err = make_pw_error("something completely unexpected")
        result = _classify_error(err, "https://example.com")
        assert result.code == ErrorCode.NETWORK_ERROR

    def test_returns_render_error_instance(self):
        err = make_pw_error("net::ERR_NAME_NOT_RESOLVED")
        result = _classify_error(err, "https://example.com")
        assert isinstance(result, RenderError)

    def test_timeout_takes_priority_over_other_matches(self):
        # TimeoutError message could contain "403" or other patterns
        err = make_timeout_error("Timeout 403 exceeded")
        result = _classify_error(err, "https://example.com")
        assert result.code == ErrorCode.TIMEOUT


class TestBrowserManagerHealthCheck:
    @pytest.mark.asyncio
    async def test_health_check_false_before_launch(self):
        manager = BrowserManager(max_concurrent_pages=2, default_timeout_ms=5000)
        assert await manager.health_check() is False
