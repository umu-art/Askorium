import logging
import sys
from contextlib import asynccontextmanager

import uvicorn
from fastapi import FastAPI
from python_encoder_api.apis.encoder_api import router
from starlette.responses import PlainTextResponse

from app.config import settings
from app.encoder_impl import EncoderImpl  # noqa: F401 — registers BaseEncoderApi subclass
from app.model_loader import get_embedder, get_reranker, models_ready

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")

if "--fetch" in sys.argv:
    get_embedder(settings)
    get_reranker(settings)
    sys.exit(0)


@asynccontextmanager
async def lifespan(_: FastAPI):
    get_embedder(settings)
    get_reranker(settings)
    yield


app = FastAPI(title="ask-encoder", version="1.0.0", lifespan=lifespan)
app.include_router(router)


@app.get("/health")
async def health():
    if models_ready():
        return "OK"
    return PlainTextResponse(status_code=503, content="Models are not ready yet")


if __name__ == "__main__":
    uvicorn.run("app.__main__:app", host=settings.host, port=settings.port)
