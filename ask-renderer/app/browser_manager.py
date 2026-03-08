import asyncio
import logging
from enum import Enum
from dataclasses import dataclass
from typing import Literal

from playwright.async_api import (
    async_playwright,
    Playwright,
    Browser,
    Error as PlaywrightError,
    TimeoutError as PlaywrightTimeoutError,
)

logger = logging.getLogger("browser_manager")


class ErrorCode(str, Enum):
    TIMEOUT = "TIMEOUT"
    NETWORK_ERROR = "NETWORK_ERROR"
    PAGE_NOT_FOUND = "PAGE_NOT_FOUND"
    ACCESS_DENIED = "ACCESS_DENIED"
    INVALID_URL = "INVALID_URL"
    PARSE_ERROR = "PARSE_ERROR"


@dataclass
class RenderError(Exception):
    code: ErrorCode
    message: str


def _classify_error(err: PlaywrightError, url: str) -> RenderError:
    msg = str(err).lower()

    if isinstance(err, PlaywrightTimeoutError):
        return RenderError(ErrorCode.TIMEOUT, f"page load timed out: {url}")

    if "net::err_name_not_resolved" in msg or "net::err_connection" in msg:
        return RenderError(ErrorCode.NETWORK_ERROR, f"cannot reach: {url}")

    if "net::err_aborted" in msg and "404" in msg:
        return RenderError(ErrorCode.PAGE_NOT_FOUND, f"not found: {url}")

    if "403" in msg or "401" in msg:
        return RenderError(ErrorCode.ACCESS_DENIED, f"access denied: {url}")

    if "net::err_invalid_url" in msg or "invalid url" in msg:
        return RenderError(ErrorCode.INVALID_URL, f"invalid url: {url}")

    return RenderError(ErrorCode.NETWORK_ERROR, f"playwright error: {err}")


class BrowserManager:
    def __init__(self, max_concurrent_pages: int, default_timeout_ms: int):
        self.max_concurrent_pages = max_concurrent_pages
        self.default_timeout_ms = default_timeout_ms
        self._playwright: Playwright | None = None
        self._browser: Browser | None = None
        self._lock = asyncio.Lock()
        logger.info("BrowserManager initialized")

    async def get_browser(self) -> Browser:
        async with self._lock:
            if self._browser is None or not self._browser.is_connected():
                logger.info(
                    "launching browser (max_concurrent_pages=%d)",
                    self.max_concurrent_pages,
                )
                self._playwright = await async_playwright().start()
                self._browser = await self._playwright.chromium.launch(
                    headless=True,
                    args=[
                        "--disable-gpu",
                        "--disable-dev-shm-usage",
                        "--no-sandbox",
                    ],
                )
        return self._browser

    async def health_check(self) -> bool:
        return self._browser is not None and self._browser.is_connected()

    async def shutdown(self):
        if self._browser is not None:
            try:
                await self._browser.close()
            except Exception as e:
                logger.warning("error closing browser: %s", e)
            self._browser = None
            logger.info("browser closed")
        if self._playwright is not None:
            try:
                await self._playwright.stop()
            except Exception as e:
                logger.warning("error stopping playwright: %s", e)
            self._playwright = None

    async def render_page(
            self,
            url: str,
            timeout_ms: int | None = None,
            wait_until: Literal["commit", "domcontentloaded", "load", "networkidle"] = "load",
    ) -> str:
        if timeout_ms is None:
            timeout_ms = self.default_timeout_ms

        browser = await self.get_browser()
        page = await browser.new_page()
        try:
            logger.debug("rendering %s (timeout=%dms)", url, timeout_ms)

            response = await page.goto(url, timeout=timeout_ms, wait_until=wait_until)

            if response is None:
                raise RenderError(ErrorCode.NETWORK_ERROR, f"no response: {url}")

            status = response.status
            if status == 404:
                raise RenderError(ErrorCode.PAGE_NOT_FOUND, f"404: {url}")
            if status in (401, 403):
                raise RenderError(ErrorCode.ACCESS_DENIED, f"{status}: {url}")
            if status >= 400:
                raise RenderError(ErrorCode.NETWORK_ERROR, f"HTTP {status}: {url}")

            html = await page.content()
            logger.info("rendered %s — %d chars, status %d", url, len(html), status)
            return html

        except RenderError:
            raise
        except PlaywrightError as e:
            raise _classify_error(e, url) from e
        except Exception as e:
            raise RenderError(
                ErrorCode.PARSE_ERROR, f"unexpected error: {e}"
            ) from e
        finally:
            await page.close()