COMPOSE = docker compose -f docker-compose.local.yml
COMPOSE_DEV = docker compose -f docker-compose.dev.yml

.PHONY: docker/build docker/up docker/down docker/clean docker/logs/mysql docker/logs/backend docker/logs/frontend docker/ps docker/dev/up docker/dev/down

docker/build:
	$(COMPOSE) build

docker/up:
	$(COMPOSE) up -d --build

docker/down:
	$(COMPOSE) down

docker/clean:
	$(COMPOSE) down -v

docker/logs/mysql:
	$(COMPOSE) logs -f snippetbox-mysql

docker/logs/backend:
	$(COMPOSE) logs -f snippetbox-backend

docker/logs/frontend:
	$(COMPOSE) logs -f snippetbox-frontend

docker/ps:
	$(COMPOSE) ps

docker/dev/up:
	$(COMPOSE_DEV) up -d

docker/dev/down:
	$(COMPOSE_DEV) down
