import numpy as np
from openai import OpenAI
from sklearn.metrics import ndcg_score
from sklearn.metrics.pairwise import cosine_similarity

import config

_client = OpenAI(
    api_key=config.OPENROUTER_API_KEY,
    base_url=config.OPENROUTER_BASE_URL,
)


def _embed(texts: list[str]) -> np.ndarray:
    response = _client.embeddings.create(
        model=config.EMBEDDING_MODEL,
        input=texts,
    )
    return np.array([d.embedding for d in response.data])


def _sim_matrix(sources: list[str], contexts: list[str]) -> np.ndarray:
    src_emb = _embed(sources)
    ref_emb = _embed(contexts)
    return cosine_similarity(src_emb, ref_emb)


def _relevance_scores(returned_sources: list[str], reference_contexts: list[str]) -> np.ndarray:
    return _sim_matrix(returned_sources, reference_contexts).max(axis=1)


def recall_at_k(
    returned_sources: list[str],
    reference_contexts: list[str],
    k: int,
    threshold: float,
) -> float:
    if not returned_sources or not reference_contexts:
        return 0.0

    top_k = returned_sources[:k]
    sim = _sim_matrix(top_k, reference_contexts)
    matched = (sim >= threshold).any(axis=0).sum()
    return float(matched / len(reference_contexts))


def precision_at_k(
    returned_sources: list[str],
    reference_contexts: list[str],
    k: int,
    threshold: float,
) -> float:
    if not returned_sources or not reference_contexts:
        return 0.0

    top_k = returned_sources[:k]
    scores = _relevance_scores(top_k, reference_contexts)
    relevant = (scores >= threshold).sum()
    return float(relevant / len(top_k))


def ndcg_at_k(
    returned_sources: list[str],
    reference_contexts: list[str],
    k: int,
) -> float:
    if not returned_sources or not reference_contexts:
        return 0.0

    top_k = returned_sources[:k]
    scores = _relevance_scores(top_k, reference_contexts)

    ideal = np.sort(scores)[::-1].reshape(1, -1)
    actual = scores.reshape(1, -1)
    return float(ndcg_score(ideal, actual))


def hit_rate(
    returned_sources: list[str],
    reference_contexts: list[str],
    threshold: float,
) -> float:
    if not returned_sources or not reference_contexts:
        return 0.0

    scores = _relevance_scores(returned_sources, reference_contexts)
    return 1.0 if (scores >= threshold).any() else 0.0
