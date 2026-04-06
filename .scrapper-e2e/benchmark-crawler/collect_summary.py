#!/usr/bin/env python3
"""
Собирает глобальный results/summary.csv из всех локальных summary.csv сессий.

Использование:
    python collect_summary.py
    python collect_summary.py --results results   # явный путь к папке
    python collect_summary.py --out summary.csv   # куда писать глобальный
"""
import argparse
import csv
from pathlib import Path


def main():
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--results", default="results", help="Папка с сессиями")
    ap.add_argument("--out",     default=None,      help="Выходной файл (default: <results>/summary.csv)")
    args = ap.parse_args()

    results_dir = Path(args.results)
    out_path    = Path(args.out) if args.out else results_dir / "summary.csv"

    # Собираем все локальные summary.csv, кроме самого глобального
    local_summaries = sorted(
        p for p in results_dir.glob("*/summary.csv")
        if p.parent != results_dir  # исключаем корневой, если он там есть
    )

    if not local_summaries:
        print(f"Нет локальных summary.csv в {results_dir}/*/")
        return

    all_rows: list[dict] = []
    header: list[str] | None = None

    for csv_path in local_summaries:
        try:
            with csv_path.open(newline="") as f:
                reader = csv.DictReader(f)
                rows = list(reader)
            if not rows:
                continue
            if header is None:
                header = list(rows[0].keys())
            # Добавляем колонку session_dir для трассируемости
            sess_dir = csv_path.parent.name
            for row in rows:
                row["session_dir"] = sess_dir
            all_rows.extend(rows)
        except Exception as e:
            print(f"  WARN: не удалось прочитать {csv_path}: {e}")

    if not all_rows or header is None:
        print("Нет данных для сборки")
        return

    # session_dir — последняя колонка
    if "session_dir" not in header:
        header = header + ["session_dir"]

    with out_path.open("w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=header, extrasaction="ignore")
        writer.writeheader()
        writer.writerows(all_rows)

    print(f"Собрано {len(all_rows)} прогонов из {len(local_summaries)} сессий")
    print(f"Глобальный summary → {out_path}")


if __name__ == "__main__":
    main()
