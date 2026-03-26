import logging
import re

import config
from openai import OpenAI
from rouge_score import rouge_scorer

logger = logging.getLogger(__name__)

_llm_client = OpenAI(
    api_key=config.OPENROUTER_API_KEY,
    base_url=config.OPENROUTER_BASE_URL,
)


class CyrillicTokenizer:
    def tokenize(self, text):
        return re.split(r'\s+', text.lower().strip())


_rouge_scorer = rouge_scorer.RougeScorer(["rougeL"], tokenizer=CyrillicTokenizer())


def rouge_l(answer: str, ground_truth: str) -> float:
    scores = _rouge_scorer.score(ground_truth, answer)
    return scores["rougeL"].fmeasure


def crag_score(answer: str, ground_truth: str) -> int:
    prompt = (
        "Оцени фактическую корректность предсказанного ответа относительно эталонного.\n"
        "Верни ТОЛЬКО одно целое число без пояснений:\n"
        "\n"
        " 1 — ответ фактически верный.\n"
        "     Засчитывается даже если ответ неполный, перефразирован или содержит\n"
        "     дополнительные детали — главное, что ключевой смысл совпадает с эталоном\n"
        "     и не содержит ложных утверждений.\n"
        "\n"
        " 0 — ответ отсутствует или выражает незнание.\n"
        "     Примеры: пустая строка, «не знаю», «нет информации», «не могу ответить».\n"
        "\n"
        "-1 — ответ фактически неверный.\n"
        "     Только если ответ содержит утверждения, прямо противоречащие эталону\n"
        "     или вводящие в заблуждение (неверные факты, числа, даты, имена).\n"
        "     Неполнота сама по себе не является ошибкой.\n"
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
        "Оцени фактическую верность ответа по отношению к предоставленным источникам.\n"
        "Верни ТОЛЬКО число от 0.0 до 1.0 без пояснений.\n"
        "\n"
        "Правила оценки:\n"
        " 1.0 — все утверждения в ответе подтверждаются источниками;\n"
        "       допустимо перефразирование и обобщение.\n"
        " 0.5 — часть утверждений подтверждается источниками,\n"
        "       часть отсутствует в них, но не противоречит.\n"
        " 0.0 — ответ содержит утверждения, прямо противоречащие источникам,\n"
        "       или источники не содержат никакой релевантной информации.\n"
        "\n"
        "Важно: если ответ корректно говорит «нет информации» при её реальном\n"
        "отсутствии в источниках — это 1.0, а не 0.0.\n"
        "Дополнительные детали сверх источников снижают оценку только если\n"
        "они противоречат источникам.\n"
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
        value = float(response.choices[0].message.content.splitlines()[0].strip())
        return max(0.0, min(1.0, value))
    except (ValueError, TypeError, IndexError) as e:
        logger.warning("faithfulness parse error: %s", e)
        return 0.0
    except Exception as e:
        logger.warning("faithfulness API error: %s", e)
        return 0.0