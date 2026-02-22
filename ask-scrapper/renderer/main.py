from contextlib import asynccontextmanager
from fastapi import FastAPI
from pydantic import BaseModel, HttpUrl
from renderer import render_page, shutdown


@asynccontextmanager
async def lifespan(app: FastAPI):
    yield
    await shutdown()

app = FastAPI(lifespan=lifespan)


class RenderRequest(BaseModel):
    url: HttpUrl
    timeout_ms: int = 15000


class RenderResponse(BaseModel):
    html: str


@app.post("/render")
async def render(req: RenderRequest) -> RenderResponse:
    html = await render_page(req.url, req.timeout_ms)
    return RenderResponse(html=html)