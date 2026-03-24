REPO_ROOT     := $(shell cd .. && pwd)
COMPOSE_INFRA := docker compose -f $(REPO_ROOT)/docker-compose.yaml
COMPOSE_E2E   := docker compose -f docker-compose.e2e.yaml

.PHONY: up infra services send restart-parser logs status down

## Поднять всё: инфраструктура + e2e-сервисы
up: infra services
	@echo ""
	@$(MAKE) status

## Только инфраструктура (RabbitMQ, Postgres, Redis, ...)
infra:
	@echo "→ Инфраструктура..."
	@$(COMPOSE_INFRA) up -d
	@echo -n "  Ждём RabbitMQ"
	@until docker exec $$(docker ps -qf name=rabbitmq) rabbitmq-diagnostics ping -q 2>/dev/null; \
		do echo -n "."; sleep 2; done
	@echo " готово"

## Собрать и запустить renderer, parser, listener
services:
	@echo "→ Собираем и запускаем сервисы..."
	@$(COMPOSE_E2E) up -d --build

## Отправить задачи из send.py
send:
	python send.py

## Пересобрать и перезапустить только парсер
restart-parser:
	@$(COMPOSE_E2E) up -d --build parser

## Следить за логами (Ctrl+C для выхода)
logs:
	@$(COMPOSE_E2E) logs -f

## Остановить e2e-сервисы и инфраструктуру
down:
	@$(COMPOSE_E2E) down
	@$(COMPOSE_INFRA) down

## Статус всех контейнеров
status:
	@echo "--- Инфраструктура ---"
	@$(COMPOSE_INFRA) ps --format "table {{.Name}}\t{{.Status}}"
	@echo ""
	@echo "--- E2E сервисы ---"
	@$(COMPOSE_E2E) ps --format "table {{.Name}}\t{{.Status}}"
