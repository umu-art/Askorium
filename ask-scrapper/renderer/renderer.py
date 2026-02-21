from playwright.async_api import async_playwright

_playwright = None
_browser = None

async def get_browser():
    global _playwright, _browser
    if _browser is None:
        _playwright = await async_playwright().start()
        _browser = await _playwright.chromium.launch(headless=True)
    return _browser

async def render_page(url: str, timeout_ms: int) -> str:
    browser = await get_browser()
    page = await browser.new_page()
    try:
        await page.goto(url, timeout=timeout_ms)
        await page.wait_for_load_state("load")
        return await page.content()
    finally:
        await page.close()
