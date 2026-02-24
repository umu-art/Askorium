import logging
from sentence_transformers import SentenceTransformer, CrossEncoder
from app.config import Settings

log = logging.getLogger(__name__)

_embedder: SentenceTransformer | None = None
_reranker: CrossEncoder | None = None


def get_embedder(settings: Settings) -> SentenceTransformer:
    global _embedder
    if _embedder is None:
        log.info("Loading embedding model: %s", settings.embedding_model)
        _embedder = SentenceTransformer(settings.embedding_model)
    return _embedder


def get_reranker(settings: Settings) -> CrossEncoder:
    global _reranker
    if _reranker is None:
        log.info("Loading reranker model: %s", settings.reranker_model)
        _reranker = CrossEncoder(settings.reranker_model)
    return _reranker


def models_ready() -> bool:
    return _embedder is not None and _reranker is not None
