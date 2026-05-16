.PHONY: help up down reset run-net run-django run-go test-net test-django test-go e2e parity css

help: ## Show available targets and descriptions
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make <target>\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*##/ { printf "  %-12s %s\n", $$1, $$2 }' Makefile

up: ## Start PostgreSQL 17 via Docker Compose
	docker compose up -d

down: ## Stop PostgreSQL and remove containers
	docker compose down

reset: ## Destroy the DB volume and restart — re-runs all init scripts
	docker compose down -v && docker compose up -d

run-net: ## Run the .NET stack on :5000
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
