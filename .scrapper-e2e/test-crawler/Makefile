REPO_ROOT     := $(shell cd .. && pwd)
COMPOSE_INFRA := docker compose -f $(REPO_ROOT)/docker-compose.yaml
COMPOSE_E2E   := docker compose -f docker-compose.e2e.yaml

.PHONY: up infra services send restart restart-crawler logs status down

## Поднять всё: инфраструктура + e2e-сервисы
up: infra services
	@echo ""
	@$(MAKE) status

## Только инфраструктура
infra:
	@echo "→ Инфраструктура..."
	@$(COMPOSE_INFRA) up -d
	@echo -n "  Ждём RabbitMQ"
	@until docker exec $$(docker ps -qf name=rabbitmq) rabbitmq-diagnostics ping -q 2>/dev/null; \
		do echo -n "."; sleep 2; done
	@echo " готово"

## Собрать и запустить все e2e-сервисы
services:
	@echo "→ Собираем и запускаем сервисы..."
	@$(COMPOSE_E2E) up -d --build

## Отправить задачу краулеру
send:
	python send.py

## Полный рестарт: down + up
restart: down up

## Пересобрать и перезапустить только краулер
restart-crawler:
	@$(COMPOSE_E2E) up -d --build crawler

## Следить за логами (Ctrl+C для выхода)
logs:
	@$(COMPOSE_E2E) logs -f

## Статус всех контейнеров
status:
	@echo "--- Инфраструктура ---"
	@$(COMPOSE_INFRA) ps --format "table {{.Name}}\t{{.Status}}"
	@echo ""
	@echo "--- E2E сервисы ---"
	@$(COMPOSE_E2E) ps --format "table {{.Name}}\t{{.Status}}"

## Остановить все сервисы
down:
	@$(COMPOSE_E2E) down
	@$(COMPOSE_INFRA) down
