COMPOSE = docker compose -f docker-compose.local.yml
COMPOSE_DEV = docker compose -f docker-compose.dev.yml

.PHONY: docker/build docker/up docker/down docker/clean docker/logs docker/logs/mysql docker/logs/user docker/logs/snippet docker/logs/frontend docker/ps docker/dev/up docker/dev/down docker/dev/clean

docker/build:
	$(COMPOSE) build

docker/up:
	$(COMPOSE) up -d --build

docker/down:
	$(COMPOSE) down

docker/clean:
	$(COMPOSE) down -v

docker/logs:
	$(COMPOSE) logs -f

docker/logs/mysql:
	$(COMPOSE) logs -f snippetbox-mysql

docker/logs/user:
	$(COMPOSE) logs -f snippetbox-backend-user

docker/logs/snippet:
	$(COMPOSE) logs -f snippetbox-backend-snippet

docker/logs/frontend:
	$(COMPOSE) logs -f snippetbox-frontend

docker/ps:
	$(COMPOSE) ps

docker/dev/up:
	$(COMPOSE_DEV) up -d

docker/dev/down:
	$(COMPOSE_DEV) down

docker/dev/clean:
	$(COMPOSE_DEV) down -v
