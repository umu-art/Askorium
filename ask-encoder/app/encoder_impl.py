import asyncio

from python_encoder_api.apis.encoder_api_base import BaseEncoderApi
from python_encoder_api.models.embedding_request import EmbeddingRequest
from python_encoder_api.models.embedding_response import EmbeddingResponse
from python_encoder_api.models.rerank_request import RerankRequest
from python_encoder_api.models.rerank_response import RerankResponse
from python_encoder_api.models.rerank_result import RerankResult
from app.config import settings
from app.model_loader import get_embedder, get_reranker

_sem = asyncio.Semaphore(settings.max_concurrent_requests)


class EncoderImpl(BaseEncoderApi):
    async def generate_embeddings(self, request: EmbeddingRequest) -> EmbeddingResponse:
        async with _sem:
            embedder = get_embedder(settings)
            vectors = embedder.encode(request.texts, normalize_embeddings=True)
            return EmbeddingResponse(embeddings=vectors.tolist())

    async def rerank(self, request: RerankRequest) -> RerankResponse:
        async with _sem:
            reranker = get_reranker(settings)
            pairs = [(request.query, block.text) for block in request.blocks]
            scores = reranker.predict(pairs)
            results = sorted(
                (RerankResult(id=block.id, score=float(score))
                 for block, score in zip(request.blocks, scores)),
                key=lambda r: r.score,
                reverse=True,
            )
            return RerankResponse(results=results)
