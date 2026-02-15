#!/usr/bin/env python3

import argparse
import subprocess
import sys
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

import yaml


def collect_yaml_files(src_dir):
    """Собрать все YAML файлы из директории src"""
    yaml_files = []
    for file in src_dir.rglob("*.yaml"):
        if file.name != "common.yaml":
            yaml_files.append(file)
    return yaml_files


def extract_paths(yaml_file):
    """Извлечь все paths из OpenAPI спецификации"""
    try:
        with open(yaml_file, 'r', encoding='utf-8') as f:
            spec = yaml.safe_load(f)
        return spec.get('paths', {})
    except Exception as e:
        print(f"Ошибка при чтении {yaml_file}: {e}")
        return {}


def create_combined_spec(yaml_files, output_file):
    combined = {
        'openapi': '3.1.0',
        'info': {
            'title': 'Combined API',
            'version': '1.0.0'
        },
        'paths': {}
    }

    for yaml_file in yaml_files:
        paths = extract_paths(yaml_file)
        relative_path = Path(yaml_file).relative_to(output_file.parent.parent)
        relative_path_str = relative_path.as_posix()

        for path_url in paths.keys():
            encoded_path = path_url.replace('~', '~0').replace('/', '~1')
            combined['paths'][path_url] = {
                '$ref': f'../{relative_path_str}#/paths/{encoded_path}'
            }

    output_file.parent.mkdir(parents=True, exist_ok=True)
    with open(output_file, 'w', encoding='utf-8') as f:
        yaml.dump(combined, f, allow_unicode=True, sort_keys=False)

    print(f"Создан объединенный файл: {output_file}")
    print(f"  Всего paths: {len(combined['paths'])}")


def bundle_with_redocly(input_file, output_file):
    """Прогнать через redocly bundle"""
    print(f"Bundling {input_file} -> {output_file}")
    output_file.parent.mkdir(parents=True, exist_ok=True)

    result = subprocess.run(
        ['redocly', 'bundle', str(input_file), '-o', str(output_file)],
        capture_output=True,
        text=True
    )

    if result.returncode != 0:
        print(f"Ошибка redocly: {result.stderr}")
        sys.exit(1)

    print(f"Bundling завершен: {output_file}")


def download_openapi_generator(build_dir):
    """Скачать OpenAPI Generator CLI если его нет"""
    jar_file = build_dir / "openapi-generator-cli.jar"

    if jar_file.exists():
        print("OpenAPI Generator CLI уже загружен")
        return jar_file

    print("Скачивание OpenAPI Generator CLI...")
    url = "https://repo1.maven.org/maven2/org/openapitools/openapi-generator-cli/7.14.0/openapi-generator-cli-7.14.0.jar"

    result = subprocess.run(['wget', url, '-O', str(jar_file)], capture_output=True)

    if result.returncode != 0:
        print("Ошибка при скачивании OpenAPI Generator CLI")
        sys.exit(1)

    return jar_file


def generate_api(jar_file, yaml_file, lang, mode, config_dir, build_dir):
    """Генерация API для указанного языка и режима"""
    output_dir = build_dir / f"{lang}-{mode}-api"
    config_file = config_dir / f"{lang}-{mode}-config.yaml"

    if not config_file.exists():
        print(f"Конфиг не найден: {config_file}, пропускаем")
        return

    print(f"Генерация {lang} {mode} API")

    cmd = [
        'java', '-jar', str(jar_file),
        'generate',
        '-i', str(yaml_file),
        '-o', str(output_dir),
        '-c', str(config_file)
    ]

    result = subprocess.run(cmd, capture_output=True, text=True)

    if result.returncode != 0:
        print(f"Ошибка генерации {lang} {mode}: {result.stderr}")
    else:
        print(f"Генерация {lang} {mode} завершена")


def install_generated(build_dir, lang_filter="all"):
    """Установка сгенерированных библиотек"""
    api_dirs = list(build_dir.glob("*-api"))

    def install_dir(target_dir):
        lang = target_dir.name.split('-')[-3]

        if lang_filter != "all" and lang_filter != lang:
            return

        print(f"Установка {target_dir}")

        if lang == "java":
            subprocess.run(['mvn', 'clean', 'install'], cwd=target_dir)
        elif lang == "ts":
            subprocess.run(['npm', 'install'], cwd=target_dir)
            subprocess.run(['npm', 'run', 'build'], cwd=target_dir)
            subprocess.run(['npm', 'link'], cwd=target_dir)
        else:
            print(f"Неизвестный язык: {lang}")

    with ThreadPoolExecutor(max_workers=4) as executor:
        futures = [executor.submit(install_dir, d) for d in api_dirs]
        for future in as_completed(futures):
            try:
                future.result()
            except Exception as e:
                print(f"Ошибка установки: {e}")


def parse_args():
    """Парсинг аргументов командной строки"""
    parser = argparse.ArgumentParser(
        description='Сборка OpenAPI спецификаций и генерация клиентов/серверов',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Примеры использования:
  python build.py                      # Полная сборка
  python build.py --force              # Очистить build и пересобрать
  python build.py --lang java          # Собрать только Java
  python build.py --lang ts --force    # Пересобрать только TypeScript
        """
    )

    parser.add_argument(
        '--force',
        action='store_true',
        help='Удалить папку build перед началом сборки'
    )

    parser.add_argument(
        '--lang',
        type=str,
        choices=['java', 'ts', 'all'],
        default='all',
        help='Ограничить сборку определенным языком (по умолчанию: all)'
    )

    return parser.parse_args()


def main():
    # Парсинг аргументов
    args = parse_args()

    # Определяем директории
    root_dir = Path(__file__).parent.absolute()
    src_dir = root_dir / "src"
    build_dir = root_dir / "build"
    config_dir = root_dir / "config"

    # Очистка build директории если --force
    if args.force:
        print("Forcing the build")
        if build_dir.exists():
            import shutil
            shutil.rmtree(build_dir)

    # Создание build директории
    build_dir.mkdir(exist_ok=True)

    # 1. Собираем все YAML файлы
    print("Сбор YAML файлов...")
    yaml_files = collect_yaml_files(src_dir)
    print(f"Найдено {len(yaml_files)} YAML файлов")

    # 2-3. Создаем объединенную спецификацию
    all_yaml = build_dir / "all.yaml"
    create_combined_spec(yaml_files, all_yaml)

    # 4. Bundling через redocly
    bundled_yaml = build_dir / "bundled.yaml"
    bundle_with_redocly(all_yaml, bundled_yaml)

    # Скачиваем OpenAPI Generator
    jar_file = download_openapi_generator(build_dir)

    # Генерация для разных языков и режимов
    generators = [
        ("java", "server"),
        ("java", "client")
    ]

    with ThreadPoolExecutor(max_workers=4) as executor:
        futures = []
        for lang, mode in generators:
            if args.lang == "all" or args.lang == lang:
                future = executor.submit(
                    generate_api, jar_file, bundled_yaml,
                    lang, mode, config_dir, build_dir
                )
                futures.append(future)

        for future in as_completed(futures):
            try:
                future.result()
            except Exception as e:
                print(f"Ошибка генерации: {e}")

    # Установка сгенерированных библиотек
    install_generated(build_dir, args.lang)

    print("Build completed")


if __name__ == "__main__":
    main()
