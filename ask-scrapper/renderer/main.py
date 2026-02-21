from fastapi import FastAPI
from pydantic import BaseModel
from renderer import render_page

app = FastAPI()

class RenderRequest(BaseModel):
    url: str
    timeout_ms: int = 15000

class RenderResponse(BaseModel):
    html: str

@app.post("/render")
async def render(req: RenderRequest) -> RenderResponse:
    html = await render_page(req.url, req.timeout_ms)
    return RenderResponse(html=html)
