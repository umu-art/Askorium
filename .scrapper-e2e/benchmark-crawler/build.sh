#!/usr/bin/env bash
# Сборка Docker-образов для бенчмарка (только Go-сервисы).
# ask-renderer собирать не нужно — образ уже есть в реестре / локально.
#
# Запускать из корня репозитория ИЛИ из .crawler-benchmark/:
#   ./build.sh            — собрать crawler + parser
#   ./build.sh crawler    — только ask-crawler
#   ./build.sh parser     — только ask-parser

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

TARGET="${1:-all}"

build_crawler() {
    echo "==> Building ask-crawler Go binary..."
    (cd ask-crawler && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ask-crawler .)

    echo "==> Building Docker image askorium/ask-crawler:latest..."
    docker build -f iac/images/ask-crawler/Dockerfile -t askorium/ask-crawler:latest .
    echo "    done: askorium/ask-crawler:latest"
}

build_parser() {
    echo "==> Building ask-parser Go binary..."
    (cd ask-parser && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ask-parser .)

    echo "==> Building Docker image askorium/ask-parser:latest..."
    docker build -f iac/images/ask-parser/Dockerfile -t askorium/ask-parser:latest .
    echo "    done: askorium/ask-parser:latest"
}

case "$TARGET" in
    crawler) build_crawler ;;
    parser)  build_parser ;;
    all)
        build_crawler
        build_parser
        echo ""
        echo "Images ready:"
        docker images | grep askorium
        ;;
    *)
        echo "Usage: $0 [all|crawler|parser]"
        exit 1
        ;;
esac
