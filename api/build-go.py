#!/usr/bin/env python3
"""
Полная сборка Go-кода из OpenAPI-спецификаций.

Скрипт автоматически:
  - Проверяет наличие Go
  - Добавляет $GOPATH/bin в PATH
  - Устанавливает oapi-codegen если отсутствует
  - Генерирует типы и клиенты
  - Инициализирует go.mod для каждого модуля
  - Запускает go mod tidy

Использование:
    python build-go.py          # сборка
    python build-go.py --clean  # очистка + сборка
"""

import os
import subprocess
import shutil
import sys
from pathlib import Path

# ===========================================================================
# Конфигурация
# ===========================================================================

TARGETS = [
    {
        "spec": "src/scrapper-api.yaml",
        "module": "scrapper-model",
        "package": "scrapper_model",
        "generate": ["types"],
    },
    {
        "spec": "src/renderer-api.yaml",
        "module": "renderer-model",
        "package": "renderer_model",
        "generate": ["types", "client"],
    },
]

# Тип генерации → имя выходного файла
FILE_NAMES = {
    "types": "model.go",
    "client": "client.go",
}

GO_VERSION = "1.23"
OAPI_CODEGEN_PKG = "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest"


# ===========================================================================
# Вспомогательные функции
# ===========================================================================

def run(cmd, **kwargs):
    """Запуск команды, возвращает результат."""
    return subprocess.run(cmd, capture_output=True, text=True, **kwargs)


def run_or_die(cmd, msg="Ошибка выполнения команды", **kwargs):
    """Запуск команды, при ошибке — выход."""
    result = run(cmd, **kwargs)
    if result.returncode != 0:
        print(f"[ОШИБКА] {msg}")
        print(f"  Команда: {' '.join(cmd) if isinstance(cmd, list) else cmd}")
        if result.stderr:
            print(f"  Вывод:   {result.stderr.strip()}")
        sys.exit(1)
    return result


# ===========================================================================
# Проверка и подготовка окружения
# ===========================================================================

def ensure_go():
    """Проверяем, что Go установлен."""
    result = run(["go", "version"])
    if result.returncode != 0:
        print("[ОШИБКА] Go не установлен. Скачайте: https://go.dev/dl/")
        sys.exit(1)
    print(f"  Go: {result.stdout.strip()}")


def ensure_gopath_in_env():
    """Добавляем $GOPATH/bin в PATH, чтобы находить установленные утилиты."""
    result = run_or_die(["go", "env", "GOPATH"], msg="Не удалось получить GOPATH")
    gopath = result.stdout.strip()
    go_bin = str(Path(gopath) / "bin")

    if go_bin not in os.environ.get("PATH", ""):
        os.environ["PATH"] = go_bin + os.pathsep + os.environ.get("PATH", "")
        print(f"  Добавлен {go_bin} в PATH")

    return go_bin


def ensure_oapi_codegen():
    """Проверяем oapi-codegen; если нет — устанавливаем автоматически."""
    if shutil.which("oapi-codegen"):
        print("  oapi-codegen: найден")
        return

    print("  oapi-codegen: не найден, устанавливаю...")
    run_or_die(
        ["go", "install", OAPI_CODEGEN_PKG],
        msg="Не удалось установить oapi-codegen",
    )
    # После установки проверяем ещё раз
    if not shutil.which("oapi-codegen"):
        print("[ОШИБКА] oapi-codegen установлен, но не найден в PATH")
        sys.exit(1)
    print("  oapi-codegen: установлен")


def check_environment():
    """Полная проверка окружения."""
    print("\n[1/3] Проверка окружения")
    ensure_go()
    ensure_gopath_in_env()
    ensure_oapi_codegen()


# ===========================================================================
# Генерация кода
# ===========================================================================

def generate_file(spec_path: Path, output_file: Path, package: str, gen_type: str):
    """Генерация одного файла из спецификации."""
    result = run_or_die(
        ["oapi-codegen", "-generate", gen_type, "-package", package, str(spec_path)],
        msg=f"Ошибка генерации [{gen_type}] из {spec_path.name}",
    )
    output_file.write_text(result.stdout, encoding="utf-8")
    print(f"    [{gen_type}] -> {output_file.name}")


def init_go_mod(module_dir: Path, module_name: str):
    """Создаём go.mod и подтягиваем зависимости."""
    go_mod = module_dir / "go.mod"

    # Минимальный go.mod — go mod tidy сам подтянет нужные зависимости
    content = f"module {module_name}\n\ngo {GO_VERSION}\n"
    go_mod.write_text(content, encoding="utf-8")

    # go mod tidy скачает все зависимости и заполнит go.sum
    result = run(["go", "mod", "tidy"], cwd=module_dir)
    if result.returncode != 0:
        print(f"    [ПРЕДУПРЕЖДЕНИЕ] go mod tidy: {result.stderr.strip()}")
    else:
        print(f"    go.mod инициализирован")


def generate_module(base_dir: Path, build_dir: Path, target: dict):
    """Генерация одного модуля: файлы + go.mod."""
    spec_path = base_dir / target["spec"]
    module_name = target["module"]
    package = target["package"]
    gen_types = target["generate"]

    if not spec_path.exists():
        print(f"[ОШИБКА] Спецификация не найдена: {spec_path}")
        sys.exit(1)

    module_dir = build_dir / module_name
    module_dir.mkdir(parents=True, exist_ok=True)

    print(f"\n  {module_name}/")

    # Генерируем каждый файл
    for gen_type in gen_types:
        filename = FILE_NAMES[gen_type]
        output_file = module_dir / filename
        generate_file(spec_path, output_file, package, gen_type)

    # Инициализируем go.mod
    init_go_mod(module_dir, module_name)


def generate_all(base_dir: Path, build_dir: Path):
    """Генерация всех модулей."""
    print("\n[2/3] Генерация кода")
    for target in TARGETS:
        generate_module(base_dir, build_dir, target)


# ===========================================================================
# Вывод инструкции по подключению
# ===========================================================================

def print_usage(build_dir: Path):
    """Показываем, как подключить сгенерированные модули."""
    print("\n[3/3] Готово!")
    print(f"\n  Результаты: {build_dir}/")
    for target in TARGETS:
        module_dir = build_dir / target["module"]
        files = [FILE_NAMES[g] for g in target["generate"]]
        print(f"    {target['module']}/  ({', '.join(files)})")

    print("\n  Подключение в ask-scrapper/go.mod:")
    print()
    print("    require (")
    for target in TARGETS:
        print(f"        {target['module']} v0.0.0")
    print("    )")
    print()
    print("    replace (")
    for target in TARGETS:
        # Относительный путь от ask-scrapper/ до api/build/module
        print(f"        {target['module']} => ../api/build/{target['module']}")
    print("    )")


# ===========================================================================
# Очистка
# ===========================================================================

def clean_build_dir(build_dir: Path):
    """Полная очистка директории сборки."""
    if build_dir.exists():
        shutil.rmtree(build_dir)
        print(f"  Очищено: {build_dir}")
    build_dir.mkdir(parents=True)


# ===========================================================================
# Точка входа
# ===========================================================================

def main():
    base_dir = Path(__file__).parent
    build_dir = base_dir / "build"

    # Флаг --clean
    if "--clean" in sys.argv:
        print("\n[0] Очистка")
        clean_build_dir(build_dir)
    else:
        build_dir.mkdir(parents=True, exist_ok=True)

    check_environment()
    generate_all(base_dir, build_dir)
    print_usage(build_dir)


if __name__ == "__main__":
    main()