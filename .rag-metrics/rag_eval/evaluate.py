import asyncio
import csv
import json
import logging
import os
from collections import defaultdict
from datetime import datetime
from statistics import mean

from tqdm import tqdm

import config
from metrics.generation import crag_score, faithfulness, rouge_l
from metrics.retrieval import hit_rate, ndcg_at_k, precision_at_k, recall_at_k
from pipeline import EvalPipeline

logging.basicConfig(level=logging.INFO, format="%(levelname)s: %(message)s")

ALL_METRICS = [
    "recall_at_k",
    "precision_at_k",
    "ndcg_at_k",
    "hit_rate",
    "rouge_l",
    "crag_score",
    "faithfulness",
]
SLICE_METRICS = ["recall_at_k", "ndcg_at_k", "crag_score"]


def load_query_cache(path: str) -> dict[str, str]:
    """Load question→query_id mapping from a CSV file (no header, two columns)."""
    if not path:
        return {}
    cache: dict[str, str] = {}
    with open(path, encoding="utf-8", newline="") as f:
        for row in csv.reader(f, delimiter=';'):
            if len(row) >= 2:
                question, query_id = row[0].strip(), row[1].strip()
                if question and query_id:
                    cache[question] = query_id
    logging.info("Loaded %d cached query IDs from %s", len(cache), path)
    return cache


def load_dataset(path: str) -> list[dict]:
    with open(path, encoding="utf-8") as f:
        return [json.loads(line) for line in f if line.strip()]


def compute_metrics(item: dict) -> dict:
    sources = item.get("sources", [])
    contexts = item.get("contexts", [])
    answer = item.get("answer", "")
    ground_truth = item.get("ground_truth", "")

    return {
        "recall_at_k": recall_at_k(sources, contexts, config.RETRIEVAL_K, config.RETRIEVAL_THRESHOLD),
        "precision_at_k": precision_at_k(sources, contexts, config.RETRIEVAL_K, config.RETRIEVAL_THRESHOLD),
        "ndcg_at_k": ndcg_at_k(sources, contexts, config.RETRIEVAL_K),
        "hit_rate": hit_rate(sources, contexts, config.RETRIEVAL_THRESHOLD),
        "rouge_l": rouge_l(answer, ground_truth),
        "crag_score": crag_score(answer, ground_truth),
        "faithfulness": faithfulness(answer, sources),
    }


def aggregate(samples: list[dict], metric_names: list[str]) -> dict:
    result = {}
    for name in metric_names:
        values = [s["metrics"][name] for s in samples]
        result[name] = round(mean(values), 4) if values else 0.0
    return result


def aggregate_by_group(samples: list[dict], group_key: str, metric_names: list[str]) -> dict:
    groups: dict[str, list[dict]] = defaultdict(list)
    for s in samples:
        groups[s[group_key]].append(s)
    return {k: aggregate(v, metric_names) for k, v in sorted(groups.items())}


def save_results(result: dict, timestamp: datetime) -> str:
    experiments_dir = os.path.join(os.path.dirname(os.path.abspath(__file__)), "experiments")
    os.makedirs(experiments_dir, exist_ok=True)

    filename = timestamp.strftime("%Y%m%d-%H%M%S") + "-exp-result.json"
    filepath = os.path.join(experiments_dir, filename)

    with open(filepath, "w", encoding="utf-8") as f:
        json.dump(result, f, ensure_ascii=False, indent=2)

    return filepath


def main() -> None:
    dataset = load_dataset(config.DATASET_PATH)
    logging.info("Loaded %d samples from %s", len(dataset), config.DATASET_PATH)

    query_cache = load_query_cache(config.QUERY_CACHE_CSV)
    if query_cache:
        for item in dataset:
            cached_id = query_cache.get(item["question"])
            if cached_id:
                item["_cached_query_id"] = cached_id

    pipeline = EvalPipeline(
        base_url=config.RAG_BASE_URL,
        source_id=config.RAG_SOURCE_ID,
        mode=config.RAG_MODE,
        poll_interval=config.RAG_POLL_INTERVAL,
        timeout=config.RAG_TIMEOUT,
    )
    dataset = asyncio.run(pipeline.run_dataset(dataset, config.RAG_CONCURRENCY))

    failed = []
    for item in tqdm(dataset, desc="Computing metrics"):
        try:
            item["metrics"] = compute_metrics(item)
        except Exception as exc:
            logging.warning("Skipping item %r: %s", item.get("question", "?")[:60], exc)
            failed.append(item)

    for item in failed:
        dataset.remove(item)

    if failed:
        logging.warning("Skipped %d/%d items due to errors", len(failed), len(dataset) + len(failed))

    per_sample = [
        {
            "question": item["question"],
            "ground_truth": item["ground_truth"],
            "answer": item.get("answer", ""),
            "difficulty": item.get("difficulty", "unknown"),
            "question_type": item.get("question_type", "unknown"),
            "metrics": item["metrics"],
        }
        for item in dataset
    ]

    overall = aggregate(per_sample, ALL_METRICS)
    by_difficulty = aggregate_by_group(per_sample, "difficulty", SLICE_METRICS)
    by_question_type = aggregate_by_group(per_sample, "question_type", SLICE_METRICS)

    timestamp = datetime.now()
    result = {
        "meta": {
            "timestamp": timestamp.isoformat(timespec="seconds"),
            "dataset_path": config.DATASET_PATH,
            "dataset_size": len(per_sample),
            "judge_model": config.JUDGE_MODEL,
            "rag_mode": config.RAG_MODE,
            "retrieval_k": config.RETRIEVAL_K,
        },
        "overall": overall,
        "by_difficulty": by_difficulty,
        "by_question_type": by_question_type,
        "per_sample": per_sample,
    }

    filepath = save_results(result, timestamp)

    print(f"\nResults saved: {filepath}")


if __name__ == "__main__":
    logging.getLogger("httpx").setLevel(logging.WARNING)
    main()
