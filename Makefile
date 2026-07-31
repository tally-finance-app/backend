CONTAINER_ENGINE := $(shell command -v docker 2>/dev/null || command -v podman 2>/dev/null)

ifeq ($(CONTAINER_ENGINE),)
$(error No container engine found — install Docker or Podman)
endif

# More targets (migrate-up, migrate-down, generate, build, test, lint) land in later tickets.

.PHONY: db-up db-down

db-up:
	$(CONTAINER_ENGINE) compose up -d

db-down:
	$(CONTAINER_ENGINE) compose down -v
