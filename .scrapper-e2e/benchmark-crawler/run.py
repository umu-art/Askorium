#!/usr/bin/env python3
"""
Запуск прогонов бенчмарка.

Один таргет:
    python run.py docs.python.org
    python run.py docs.python.org --profile deep
    python run.py docs.python.org --label "r2-redis"
    python run.py docs.python.org --priority            # Priority frontier (Best-First)
    python run.py docs.python.org --no-priority         # FIFO frontier (по умолчанию)

Все таргеты из targets.json:
    python run.py --all
    python run.py --all --profile deep --label "r5-redis"
    python run.py --all --priority --label "r5-prio"

    Теги и группы:
    python run.py --all --tags docs
    python run.py --all --tags ru,university
    python run.py --all --skip hse.ru msu.ru

Результаты:
    results/<timestamp>_<label>/          ← папка сессии
        run_config.json                   ← конфиг всей сессии
        summary.csv                       ← локальное summary
        <target_name>/                    ← папка одного таргета
            timing.json
            task.completed.json
            batch_NNNN.json
"""
import argparse
import json
import os
import sys
import time
import uuid
from datetime import datetime, timezone
from pathlib import Path

import pika

RABBITMQ_URL = os.environ.get("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
INPUT_QUEUE  = "askorium.crawler.input"
OUTPUT_QUEUE = "askorium.crawler.output"
RESULTS_DIR  = Path("results")
TIMEOUT      = 600  # секунд на один таргет

SUMMARY_HEADER = (
    "session,target,profile,frontier,priority,replicas,"
    "elapsed_sec,throughput_ppm,pages_scraped,pages_failed,"
    "completion_reason,n_batches\n"
)


# ── Точка входа ───────────────────────────────────────────────────────────────

def main():
    ap = argparse.ArgumentParser(formatter_class=argparse.RawDescriptionHelpFormatter,
                                 description=__doc__)
    ap.add_argument("target", nargs="?", help="Имя сайта из targets.json")
    ap.add_argument("--all",     action="store_true", help="Прогнать все таргеты")
    ap.add_argument("--profile", default=None,        help="Профиль: shallow / deep / full")
    ap.add_argument("--label",   default="",          help="Метка сессии")
    ap.add_argument("--tags",    default="",          help="AND-фильтр по тегам: doc,ru")
    ap.add_argument("--skip",    nargs="*", default=[], metavar="NAME",
                    help="Пропустить эти таргеты (только с --all)")
    prio = ap.add_mutually_exclusive_group()
    prio.add_argument("--priority",    dest="priority", action="store_true",  default=None,
                      help="Priority frontier (Best-First Search по score)")
    prio.add_argument("--no-priority", dest="priority", action="store_false",
                      help="FIFO frontier (по умолчанию)")
    args = ap.parse_args()

    if not args.all and not args.target:
        ap.error("Укажи имя таргета или --all")
    if args.all and args.target:
        ap.error("--all и явный таргет несовместимы")

    targets_cfg = json.loads(Path("targets.json").read_text())
    compose_txt = _read_compose()
    replicas    = _parse_replicas(compose_txt)
    redis_url   = _parse_redis(compose_txt)
    frontier    = "redis" if redis_url else "memory"

    if args.all:
        sites = _filter_sites(targets_cfg["sites"], args.tags, args.skip)
        if not sites:
            sys.exit("Нет таргетов после фильтрации")
    else:
        site = next((s for s in targets_cfg["sites"] if s.get("name") == args.target), None)
        if not site:
            sys.exit(f"Таргет '{args.target}' не найден в targets.json")
        sites = [site]

    # Определяем профиль сессии (один для всех таргетов)
    profile_name = args.profile or "shallow"

    if profile_name not in targets_cfg["profiles"]:
        sys.exit(f"Профиль '{profile_name}' не найден в targets.json")
    profile = targets_cfg["profiles"][profile_name]

    # Создаём папку сессии
    ts       = datetime.now(timezone.utc).strftime("%Y%m%d_%H%M%S")
    label    = args.label or (args.target.replace(".", "_") if args.target else "run")
    sess_dir = RESULTS_DIR / f"{ts}_{label}"
    sess_dir.mkdir(parents=True)

    priority = bool(args.priority)  # None → False (FIFO по умолчанию)

    run_config = {
        "label":      label,
        "profile":    profile_name,
        "frontier":   frontier,
        "priority":   priority,
        "replicas":   replicas,
        "max_pages":  profile["max_pages"],
        "max_depth":  profile["max_depth"],
        "concurrency": replicas,
        "targets":    [s["name"] for s in sites],
        "started_at": ts,
    }
    (sess_dir / "run_config.json").write_text(json.dumps(run_config, indent=2))

    frontier_str = f"{frontier}+{'priority' if priority else 'fifo'}"
    print(f"▶▶ Сессия: {sess_dir.name}")
    print(f"   profile={profile_name}  frontier={frontier_str}  replicas={replicas}  "
          f"max_pages={profile['max_pages']}  таргетов={len(sites)}\n")

    for site in sites:
        _run_one(site, profile_name, profile, replicas, frontier, priority, sess_dir)
        print()

    print(f"Сессия завершена → {sess_dir}")
    print(f"Локальный summary → {sess_dir / 'summary.csv'}")


# ── Один таргет ───────────────────────────────────────────────────────────────

def _run_one(site: dict, profile_name: str, profile: dict,
             replicas: int, frontier: str, priority: bool, sess_dir: Path):
    target_dir = sess_dir / site["name"].replace(".", "_")
    target_dir.mkdir(parents=True)

    session_label = sess_dir.name.split("_", 2)[2] if sess_dir.name.count("_") >= 2 else sess_dir.name
    frontier_str  = f"{frontier}+{'priority' if priority else 'fifo'}"

    print(f"▶  {site['name']}  profile={profile_name}  "
          f"frontier={frontier_str}  replicas={replicas}")

    task_id = str(uuid.uuid4())
    msg = {
        "task_id":   task_id,
        "domain":    site["domain"],
        "seed_urls": site["seed_urls"],
        "options":   {**profile, "concurrency": replicas, "priority": priority},
        "metadata":  {"session": session_label, "target": site["name"]},
    }

    conn = pika.BlockingConnection(pika.URLParameters(RABBITMQ_URL))
    ch   = conn.channel()
    ch.basic_publish(
        exchange="",
        routing_key=INPUT_QUEUE,
        body=json.dumps(msg),
        properties=pika.BasicProperties(content_type="application/json", delivery_mode=2),
    )
    print(f"   task_id={task_id}")

    t0 = time.monotonic()
    batches, pages_total, done, completion = [], 0, False, {}

    def on_msg(ch, method, props, body):
        nonlocal pages_total, done, completion
        ev = json.loads(body)
        if ev.get("task_id") != task_id:
            ch.basic_ack(method.delivery_tag)
            return
        etype = ev.get("type")
        if etype == "page.batch":
            seq = ev.get("batch_seq", len(batches) + 1)
            n   = len(ev.get("pages", []))
            pages_total += n
            (target_dir / f"batch_{seq:04d}.json").write_text(
                json.dumps(ev, ensure_ascii=False, indent=2))
            elapsed = round(time.monotonic() - t0, 1)
            batches.append({"seq": seq, "pages": n, "elapsed_sec": elapsed})
            print(f"  batch #{seq:02d}  {n:3d} страниц  +{elapsed}s")
        elif etype in ("task.completed", "task.failed"):
            completion = ev
            done = True
            ch.stop_consuming()
        ch.basic_ack(method.delivery_tag)

    ch.queue_declare(queue=OUTPUT_QUEUE, durable=True)
    ch.basic_consume(queue=OUTPUT_QUEUE, on_message_callback=on_msg)
    deadline = time.monotonic() + TIMEOUT
    while not done and time.monotonic() < deadline:
        conn.process_data_events(time_limit=5)
    elapsed = round(time.monotonic() - t0, 1)
    conn.close()

    stats  = completion.get("stats", {})
    reason = completion.get("completion_reason", "TIMEOUT")
    timing = {
        "elapsed_sec":    elapsed,
        "throughput_ppm": round(pages_total / elapsed * 60, 1) if elapsed else 0,
        "pages_total":    pages_total,
        "n_batches":      len(batches),
        "timed_out":      not done,
        "batches":        batches,
    }
    (target_dir / "timing.json").write_text(json.dumps(timing, indent=2))
    (target_dir / "task.completed.json").write_text(
        json.dumps(completion, ensure_ascii=False, indent=2))

    row = [
        session_label,
        site["name"],
        profile_name,
        frontier,
        "priority" if priority else "fifo",
        replicas,
        elapsed,
        timing["throughput_ppm"],
        stats.get("pages_scraped", pages_total),
        stats.get("pages_failed", 0),
        reason,
        len(batches),
    ]
    _append_csv(sess_dir / "summary.csv", SUMMARY_HEADER, row)

    print(f"{'OK' if done else 'TIMEOUT'}  {elapsed}s  "
          f"{timing['throughput_ppm']} pages/min  {reason}")
    print(f"   → {target_dir}")


# ── Вспомогательные ───────────────────────────────────────────────────────────

def _append_csv(path: Path, header: str, row: list):
    write_header = not path.exists()
    with path.open("a") as f:
        if write_header:
            f.write(header)
        f.write(",".join(str(x) for x in row) + "\n")


def _filter_sites(sites: list, tags_arg: str, skip: list) -> list:
    required_tags = [t.strip() for t in tags_arg.split(",") if t.strip()]
    result = []
    for s in sites:
        if "name" not in s:
            continue
        if s["name"] in skip:
            continue
        if required_tags and not all(t in s.get("tags", []) for t in required_tags):
            continue
        result.append(s)
    return result


def _read_compose() -> str:
    for name in ("docker-compose.yml", "docker-compose.yaml"):
        p = Path(name)
        if p.exists():
            return p.read_text()
    return ""


def _parse_replicas(txt: str) -> int:
    for line in txt.splitlines():
        stripped = line.strip()
        if stripped.startswith("#"):
            continue
        if stripped.startswith("replicas:"):
            val = stripped.split(":", 1)[1].split("#")[0].strip()
            try:
                return int(val)
            except ValueError:
                pass
    return 1


def _parse_redis(txt: str) -> str:
    for line in txt.splitlines():
        stripped = line.strip()
        if stripped.startswith("#"):
            continue
        if "REDIS_URL:" not in line:
            continue
        val = line.split("REDIS_URL:")[1]
        val = val.split("#")[0].strip().strip('"').strip("'")
        return val
    return ""


if __name__ == "__main__":
    main()
