from playwright.async_api import async_playwright, Playwright, Browser

_playwright: Playwright | None = None
_browser: Browser | None = None


async def get_browser():
    global _playwright, _browser
    if _browser is None:
        _playwright = await async_playwright().start()
        _browser = await _playwright.chromium.launch(headless=True)
    return _browser


async def shutdown():
    global _playwright, _browser
    if _browser is not None:
        await _browser.close()
        _browser = None
    if _playwright is not None:
        await _playwright.stop()
        _playwright = None


async def render_page(url: str, timeout_ms: int) -> str:
    browser = await get_browser()
    page = await browser.new_page()
    try:
        await page.goto(url, timeout=timeout_ms)
        await page.wait_for_load_state("load")
        return await page.content()
    finally:
        await page.close()