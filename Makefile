.PHONY: help up down reset run-net run-django run-go test-net test-django test-go test-integration test-net-integration test-django-integration test-go-integration e2e parity css

help: ## Show available targets and descriptions
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make <target>\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*##/ { printf "  %-12s %s\n", $$1, $$2 }' Makefile

up: ## Start PostgreSQL 17 via Docker Compose
	docker compose up -d

down: ## Stop PostgreSQL and remove containers
	docker compose down

reset: ## Destroy the DB volume and restart — re-runs all init scripts
	docker compose down -v && docker compose up -d

run-net: ## Run the .NET stack on :4000
	cd FieldMark && dotnet run --project FieldMark.Web

run-django: ## Run the Django stack on :8000
	cd fieldmark_py && uv run python manage.py runserver

run-go: ## Run the Go/Fiber stack on :3000
	cd fieldmark-go && go run ./cmd/web

test-net: ## Run .NET tests (xUnit)
	cd FieldMark && dotnet test

test-django: ## Run Django tests (pytest)
	cd fieldmark_py && uv run pytest

test-go: ## Run Go tests
	cd fieldmark-go && go test ./...

# ── Real-DB integration tests (Epic 1 retro action item A3) ────────────────
# Each stack stands alone: .NET spins its own Postgres via Testcontainers and
# needs no external state; Django and Go reuse the running `make up` container
# (FIELDMARK_DATABASE_URL overrides the localhost default). Per the Cross-Stack
# Architecture Principle, the three harnesses share no test base class — each
# is idiomatic to its stack.

test-net-integration: ## Run .NET integration tests (Testcontainers spins its own Postgres)
	cd FieldMark && dotnet test FieldMark.Tests.Integration/FieldMark.Tests.Integration.csproj

test-django-integration: ## Run Django integration tests against the running `make up` Postgres
	cd fieldmark_py && uv run pytest -m integration

test-go-integration: ## Run Go integration tests against the running `make up` Postgres
	cd fieldmark-go && go test -tags=integration ./internal/data/postgres/...

test-integration: test-net-integration test-django-integration test-go-integration ## Run integration tests for all three stacks
	@echo "✓ Integration tests passed across .NET, Django, Go"

e2e: ## Run Playwright end-to-end tests (skips if e2e/ not scaffolded or deps not installed)
	@if [ -f e2e/node_modules/.bin/playwright ]; then \
		cd e2e && pnpm run test:e2e; \
	else \
		echo "(skip) Playwright not installed — run: cd e2e && pnpm install && pnpm exec playwright install"; \
	fi

parity: ## Run cross-stack parity diff scripts (skips if tools/parity/ not yet scaffolded)
	@if [ -f tools/parity/diff-routes.sh ] && [ -f tools/parity/diff-pg-indexes.sh ]; then \
		tools/parity/diff-routes.sh && tools/parity/diff-pg-indexes.sh; \
	elif [ -f tools/parity/diff-routes.sh ] || [ -f tools/parity/diff-pg-indexes.sh ]; then \
		echo "ERROR: tools/parity/ is partially installed — both diff-routes.sh and diff-pg-indexes.sh are required" >&2; \
		exit 1; \
	else \
		echo "(skip) tools/parity/ not yet scaffolded — lands in Story 1.3"; \
	fi

css: ## Build shared Tailwind CSS (skips if fieldmark_shared/ deps not installed)
	@if [ -d fieldmark_shared/node_modules ]; then \
		cd fieldmark_shared && pnpm run build; \
	else \
		echo "(skip) fieldmark_shared deps not installed — run: cd fieldmark_shared && pnpm install"; \
	fi

.PHONY: seed seed-net seed-django seed-go

seed: seed-net seed-django seed-go ## Seed dev users into all three stacks' auth schemas
	@echo "✓ All three stacks seeded from docker/postgres/init/seed-uuids/dev-users.json"

seed-net: ## Seed dev users into dotnet_auth (runs roles seeder then users seeder)
	cd FieldMark && dotnet run --project FieldMark.Web -- --seed-dev-users

seed-django: ## Seed dev users into django_auth
	cd fieldmark_py && uv run python manage.py seed_dev_users

seed-go: ## Seed dev users into fiber_auth
	cd fieldmark-go && go run ./cmd/seed
