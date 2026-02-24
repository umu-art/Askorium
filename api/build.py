#!/usr/bin/env python3

import argparse
import logging
import shutil
import subprocess
import sys
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

import yaml

logging.basicConfig(level=logging.INFO, format='%(asctime)s %(levelname)s %(message)s')
log = logging.getLogger(__name__)


def extract_paths(yaml_file):
    """Return paths dict from an OpenAPI YAML file."""
    try:
        with open(yaml_file, 'r', encoding='utf-8') as f:
            spec = yaml.safe_load(f)
        return spec.get('paths', {})
    except Exception as e:
        log.error("Error reading %s: %s", yaml_file, e)
        return {}


def create_combined_spec(yaml_files, output_file):
    """Write a combined OpenAPI spec that $ref-references paths from the given files."""
    log.info("Creating combined spec → %s", output_file)
    combined = {
        'openapi': '3.1.0',
        'info': {'title': 'Combined API', 'version': '1.0.0'},
        'paths': {}
    }
    for yaml_file in yaml_files:
        rel = Path(yaml_file).relative_to(output_file.parent.parent.parent).as_posix()
        for path_url in extract_paths(yaml_file):
            encoded = path_url.replace('~', '~0').replace('/', '~1')
            combined['paths'][path_url] = {'$ref': f'../../{rel}#/paths/{encoded}'}

    output_file.parent.mkdir(parents=True, exist_ok=True)
    with open(output_file, 'w', encoding='utf-8') as f:
        yaml.dump(combined, f, allow_unicode=True, sort_keys=False)


def bundle_with_redocly(input_file, output_file):
    """Bundle an OpenAPI spec with Redocly; exits on failure."""
    log.info("Bundling %s → %s", input_file, output_file)
    output_file.parent.mkdir(parents=True, exist_ok=True)
    result = subprocess.run(
        ['redocly', 'bundle', str(input_file), '-o', str(output_file)],
        capture_output=True, text=True
    )
    if result.returncode != 0:
        log.error("redocly error: %s", result.stderr)
        sys.exit(1)


def download_openapi_generator(build_dir):
    """Download the OpenAPI Generator CLI JAR if not already present."""
    jar_file = build_dir / "openapi-generator-cli.jar"
    if jar_file.exists():
        log.info("OpenAPI Generator CLI already present")
        return jar_file
    log.info("Downloading OpenAPI Generator CLI...")
    url = "https://repo1.maven.org/maven2/org/openapitools/openapi-generator-cli/7.14.0/openapi-generator-cli-7.14.0.jar"
    result = subprocess.run(['wget', url, '-O', str(jar_file)], capture_output=True)
    if result.returncode != 0:
        log.error("Failed to download OpenAPI Generator CLI")
        sys.exit(1)
    return jar_file


def generate_api(jar_file, yaml_file, config_file, output_dir):
    """Run OpenAPI Generator for the given spec and config."""
    log.info("Generating API: %s", config_file.name)
    result = subprocess.run(
        ['java', '-jar', str(jar_file), 'generate',
         '-i', str(yaml_file), '-o', str(output_dir), '-c', str(config_file)],
        capture_output=True, text=True
    )
    if result.returncode != 0:
        log.error("Generation failed for %s: %s", config_file, result.stderr)


def install_generated(build_dir, langs=None):
    """Install all generated *-api artifacts in parallel."""

    def install_dir(target_dir):
        """Install a single generated artifact directory by language."""
        lang = target_dir.name.split('-')[0]
        log.info("Installing %s (%s)", target_dir.name, lang)
        if lang == "java":
            subprocess.run(['mvn', 'clean', 'install'], cwd=target_dir)
        elif lang == "ts":
            subprocess.run(['npm', 'install'], cwd=target_dir)
            subprocess.run(['npm', 'run', 'build'], cwd=target_dir)
            subprocess.run(['npm', 'link'], cwd=target_dir)
        elif lang == "python":
            subprocess.run([sys.executable, '-m', 'build', '--wheel'], cwd=target_dir)
            whl_files = list((target_dir / 'dist').glob('*.whl'))
            subprocess.run([sys.executable, '-m', 'pip', 'install', str(whl_files[0])], cwd=target_dir)
            subprocess.run(
                [sys.executable, '-m', 'pip', 'install', '--force-reinstall', '--no-deps', str(whl_files[0])],
                cwd=target_dir)

    dirs = [d for d in build_dir.glob("*-api")
            if langs is None or d.name.split('-')[0] in langs]

    with ThreadPoolExecutor(max_workers=8) as executor:
        futures = [executor.submit(install_dir, d) for d in dirs]
        for f in as_completed(futures):
            f.result()


def process_config(config_file, jar_file, build_dir, src_dir, force=False):
    """Process one generator config file: combine, bundle, and generate."""
    log.info("Processing config: %s", config_file.name)
    with open(config_file, 'r', encoding='utf-8') as f:
        config = yaml.safe_load(f)

    sources = config['sources']
    if sources is None:
        sources_files = list(src_dir.glob("**/*.yaml"))
    elif isinstance(sources, list):
        sources_files = list(set(file for pattern in sources for file in src_dir.glob(pattern)))
    else:
        sources_files = list(src_dir.glob(sources))

    target_dir = build_dir / config['target']
    if force and target_dir.exists():
        log.info("Removing %s (--force)", target_dir)
        shutil.rmtree(target_dir)
    target_dir.mkdir(parents=True, exist_ok=True)

    combined_yaml = target_dir / "combined.yaml"
    create_combined_spec(sources_files, combined_yaml)

    processed_yaml = target_dir / "processed.yaml"
    bundle_with_redocly(combined_yaml, processed_yaml)
    generate_api(jar_file, processed_yaml, config_file, target_dir)


def main():
    """Entry point: download generator, process all configs, install artifacts."""
    parser = argparse.ArgumentParser(description='Build OpenAPI generated clients/servers.')
    parser.add_argument('--langs', nargs='+', metavar='LANG',
                        help='Languages to build (e.g. java python go). Default all langs.')
    parser.add_argument('--force', action='store_true',
                        help='Remove output dirs before building.')
    args = parser.parse_args()

    langs = set(args.langs) if args.langs else None

    root_dir = Path(__file__).parent.absolute()
    build_dir = root_dir / "build"
    build_dir.mkdir(exist_ok=True)

    jar_file = download_openapi_generator(build_dir)

    configs = [cfg for cfg in (root_dir / "config").glob("*config.yaml")
               if langs is None or cfg.name.split('-')[0] in langs]

    with ThreadPoolExecutor(max_workers=8) as executor:
        futures = [
            executor.submit(process_config, cfg, jar_file, build_dir, root_dir / "src", args.force)
            for cfg in configs
        ]
        for f in as_completed(futures):
            f.result()

    install_generated(build_dir, langs)
    log.info("Build completed")


if __name__ == "__main__":
    main()
