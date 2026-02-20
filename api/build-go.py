#!/usr/bin/env python3

import os
import subprocess
import shutil
import sys
from pathlib import Path


TARGETS = [
    {
        "spec": "src/scrapper-api.yaml",
        "output_dir": "model/scrapper-model",
        "output_file": "model.go",
        "package": "scrapper_model",
    },
]


def ensure_gopath_in_env():
    """Добавляем $GOPATH/bin в PATH, чтобы найти установленные Go-утилиты."""
    result = subprocess.run(["go", "env", "GOPATH"], capture_output=True, text=True)
    if result.returncode != 0:
        print("Ошибка: Go не установлен или не найден в PATH.")
        sys.exit(1)

    gopath = result.stdout.strip()
    go_bin = str(Path(gopath) / "bin")

    if go_bin not in os.environ.get("PATH", ""):
        os.environ["PATH"] = go_bin + os.pathsep + os.environ.get("PATH", "")
        print(f"Добавлен {go_bin} в PATH")


def check_oapi_codegen():
    """Проверяем, установлен ли oapi-codegen."""
    ensure_gopath_in_env()
    if shutil.which("oapi-codegen") is None:
        print("Ошибка: oapi-codegen не найден.")
        print("Установите его командой:")
        print("  go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest")
        sys.exit(1)


def generate(target: dict):
    """Генерация Go-типов из одной OpenAPI-спецификации."""
    base_dir = Path(__file__).parent
    spec_path = base_dir / target["spec"]
    output_dir = base_dir / target["output_dir"]
    output_file = output_dir / target["output_file"]
    package = target["package"]

    if not spec_path.exists():
        print(f"Ошибка: файл спецификации не найден: {spec_path}")
        sys.exit(1)

    output_dir.mkdir(parents=True, exist_ok=True)

    print(f"Генерация типов из {spec_path} ...")
    result = subprocess.run(
        ["oapi-codegen", "-generate", "types", "-package", package, str(spec_path)],
        capture_output=True,
        text=True,
    )

    if result.returncode != 0:
        print(f"Ошибка при генерации: {result.stderr}")
        sys.exit(1)

    output_file.write_text(result.stdout, encoding="utf-8")
    print(f"Успешно! Сгенерированный файл: {output_file}")


def main():
    check_oapi_codegen()
    for target in TARGETS:
        generate(target)
    print("\nГенерация завершена.")


if __name__ == "__main__":
    main()