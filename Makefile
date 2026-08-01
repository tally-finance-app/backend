CONTAINER_ENGINE := $(shell command -v docker 2>/dev/null || command -v podman 2>/dev/null)

ifeq ($(CONTAINER_ENGINE),)
$(error No container engine found — install Docker or Podman)
endif

# More targets (build, test, lint) land in later tickets.

.PHONY: db-up db-down migrate-up migrate-down generate

db-up:
	$(CONTAINER_ENGINE) compose up -d

db-down:
	$(CONTAINER_ENGINE) compose down -v

migrate-up:
	if [ -f .env ]; then set -a && . ./.env && set +a; fi; go run ./cmd/migrate up

migrate-down:
	if [ -f .env ]; then set -a && . ./.env && set +a; fi; go run ./cmd/migrate down

generate:
	sqlc generate
