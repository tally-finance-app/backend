CONTAINER_ENGINE := $(shell command -v docker 2>/dev/null || command -v podman 2>/dev/null)

# Checked inside the recipes that need it, not at parse time. A parse-time
# $(error) aborts *every* make invocation on a machine without Docker —
# including build, vet, lint and test, which need no container at all.
define require_container_engine
	@[ -n "$(CONTAINER_ENGINE)" ] || { echo "No container engine found — install Docker or Podman"; exit 1; }
endef

# Loads .env when present. Kept conditional so the same target works in CI,
# where there is no .env and DATABASE_URL comes from the environment.
DOTENV := if [ -f .env ]; then set -a && . ./.env && set +a; fi;

# -count=1 disables the test result cache, which otherwise reports a stale pass
# for a DB-backed test whose only input change was the database itself.
GO_TEST_FLAGS := -race -count=1

# --- Development tools -------------------------------------------------------
#
# Tools are pinned in tools/go.mod, a SEPARATE module, so their dependency graphs
# can never influence the versions the application builds against. (Before the
# split, sqlc's requirement silently chose our pgx version.)
#
# They're built into ./bin because `go -C tools tool sqlc ...` would run the tool
# with tools/ as its working directory, where sqlc.yaml and migrations/ don't
# exist. Each binary is rebuilt only when tools/go.mod or tools/go.sum changes.
BIN := $(CURDIR)/bin

SQLC          := $(BIN)/sqlc
GOLANGCI_LINT := $(BIN)/golangci-lint
LEFTHOOK      := $(BIN)/lefthook

TOOLS_DEPS := tools/go.mod tools/go.sum

$(SQLC): $(TOOLS_DEPS)
	@mkdir -p $(BIN)
	GOBIN=$(BIN) go -C tools install github.com/sqlc-dev/sqlc/cmd/sqlc

$(GOLANGCI_LINT): $(TOOLS_DEPS)
	@mkdir -p $(BIN)
	GOBIN=$(BIN) go -C tools install github.com/golangci/golangci-lint/v2/cmd/golangci-lint

$(LEFTHOOK): $(TOOLS_DEPS)
	@mkdir -p $(BIN)
	GOBIN=$(BIN) go -C tools install github.com/evilmartians/lefthook

.PHONY: tools tools-tidy tools-clean db-up db-down migrate-up migrate-down migrate-verify \
        generate verify-generate build vet lint lint-fix test test-integration ci \
        hooks hooks-uninstall

tools: $(SQLC) $(GOLANGCI_LINT) $(LEFTHOOK)

tools-tidy:
	go -C tools mod tidy

tools-clean:
	rm -rf $(BIN)

# --- Git hooks ---------------------------------------------------------------

# Run once per clone; .git/hooks is not version-controlled, so nothing happens
# until someone runs this. Builds every tool the hooks invoke up front, so the
# first commit after cloning isn't a surprise 30-second compile.
hooks: $(LEFTHOOK) $(GOLANGCI_LINT) $(SQLC)
	$(LEFTHOOK) install

hooks-uninstall: $(LEFTHOOK)
	$(LEFTHOOK) uninstall

# --- Database ----------------------------------------------------------------

db-up:
	$(require_container_engine)
	$(CONTAINER_ENGINE) compose up -d

db-down:
	$(require_container_engine)
	$(CONTAINER_ENGINE) compose down -v

migrate-up:
	$(DOTENV) go run ./cmd/migrate up

migrate-down:
	$(DOTENV) go run ./cmd/migrate down

# Proves the .down.sql files actually work — nothing else exercises them.
migrate-verify:
	$(DOTENV) go run ./cmd/migrate up && go run ./cmd/migrate down && go run ./cmd/migrate up

# --- Codegen -----------------------------------------------------------------

generate: $(SQLC)
	$(SQLC) generate

# Fails if the committed generated code is stale relative to migrations/ and the
# queries files. This is the guard that lets the generated code be committed.
verify-generate: $(SQLC)
	$(SQLC) diff

# --- Build, lint, test -------------------------------------------------------

build:
	go build ./...

vet:
	go vet ./...

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run

lint-fix: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run --fix

# Unit tests only. -short makes DB-backed tests skip, so this needs no Postgres.
test:
	go test $(GO_TEST_FLAGS) -short ./...

# Everything, including repository-layer tests against real Postgres.
# Run `make db-up && make migrate-up` first.
#
# -v is deliberate: DB-backed tests skip themselves when DATABASE_URL is unset,
# and a skipped package still reports "ok". Without per-test output there is no
# way to tell a passing integration test from one that silently skipped — which
# would make this whole job green while proving nothing.
test-integration:
	$(DOTENV) go test $(GO_TEST_FLAGS) -v ./...

# What CI runs, so a green `make ci` locally means a green pipeline.
ci: verify-generate build vet lint test
