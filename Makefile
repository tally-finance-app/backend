CONTAINER_ENGINE := $(shell command -v docker 2>/dev/null || command -v podman 2>/dev/null)

ifeq ($(CONTAINER_ENGINE),)
$(error No container engine found — install Docker or Podman)
endif

# More targets (generate, build, test, lint) land in later tickets.

.PHONY: db-up db-down migrate-up migrate-down

db-up:
	$(CONTAINER_ENGINE) compose up -d

db-down:
	$(CONTAINER_ENGINE) compose down -v

migrate-up:
	set -a && . ./.env && set +a && go run ./cmd/migrate up

migrate-down:
	set -a && . ./.env && set +a && go run ./cmd/migrate down
