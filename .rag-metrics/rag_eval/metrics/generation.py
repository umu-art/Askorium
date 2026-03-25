import logging

import bert_score
import config
from openai import OpenAI
from rouge_score import rouge_scorer

logger = logging.getLogger(__name__)

_llm_client = OpenAI(
    api_key=config.OPENROUTER_API_KEY,
    base_url=config.OPENROUTER_BASE_URL,
)

_rouge_scorer = rouge_scorer.RougeScorer(["rougeL"], use_stemmer=False)


def bert_score_f1(answer: str, ground_truth: str) -> float:
    _, _, f1 = bert_score.score([answer], [ground_truth], lang="ru", verbose=False)
    return float(f1[0])


def rouge_l(answer: str, ground_truth: str) -> float:
    scores = _rouge_scorer.score(ground_truth, answer)
    return scores["rougeL"].fmeasure


def crag_score(answer: str, ground_truth: str) -> int:
    prompt = (
        "Сравни предсказанный ответ с эталонным.\n"
        "Верни ТОЛЬКО одно целое число без пояснений:\n"
        " 1 — ответ правильный (смысл совпадает с эталоном)\n"
        " 0 — ответ отсутствует, пустой или \"не знаю\"\n"
        "-1 — ответ неверный или противоречит эталону\n"
        "\n"
        f"Эталонный ответ: {ground_truth}\n"
        f"Предсказанный ответ: {answer}"
    )
    try:
        response = _llm_client.chat.completions.create(
            model=config.JUDGE_MODEL,
            temperature=0,
            messages=[{"role": "user", "content": prompt}],
        )
        return int(response.choices[0].message.content.strip())
    except (ValueError, TypeError, IndexError) as e:
        logger.warning("crag_score parse error: %s", e)
        return 0
    except Exception as e:
        logger.warning("crag_score API error: %s", e)
        return 0


def faithfulness(answer: str, sources: list[str]) -> float:
    sources_text = "\n".join(f"[{i + 1}] {s}" for i, s in enumerate(sources))
    prompt = (
        "Оцени, насколько ответ опирается исключительно на предоставленные источники.\n"
        "Верни ТОЛЬКО число от 0.0 до 1.0 без пояснений:\n"
        " 1.0 — ответ полностью основан на источниках, нет информации вне них\n"
        " 0.0 — ответ полностью игнорирует источники или противоречит им\n"
        "\n"
        f"Источники:\n{sources_text}\n"
        "\n"
        f"Ответ:\n{answer}"
    )
    try:
        response = _llm_client.chat.completions.create(
            model=config.JUDGE_MODEL,
            temperature=0,
            messages=[{"role": "user", "content": prompt}],
        )
        value = float(response.choices[0].message.content.strip())
        return max(0.0, min(1.0, value))
    except (ValueError, TypeError, IndexError) as e:
        logger.warning("faithfulness parse error: %s", e)
        return 0.0
    except Exception as e:
        logger.warning("faithfulness API error: %s", e)
        return 0.0
