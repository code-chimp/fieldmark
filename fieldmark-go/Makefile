# POSIX Make (Git Bash, WSL, macOS, Linux). On Windows without make, see CLAUDE.md for raw go commands.
.PHONY: fmt fmt-check vet staticcheck lint check check-all test

GO ?= go
GOIMPORTS = $(GO) tool golang.org/x/tools/cmd/goimports
STATICCHECK = $(GO) tool honnef.co/go/tools/cmd/staticcheck

# Write gofmt + goimports fixes in place
fmt:
	gofmt -w .
	$(GOIMPORTS) -w .

# Fail if formatting would change files (CI-friendly)
fmt-check:
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi
	@out=$$($(GO) tool golang.org/x/tools/cmd/goimports -l .); if [ -n "$$out" ]; then echo "goimports needed:"; echo "$$out"; exit 1; fi

vet:
	$(GO) vet ./...

staticcheck:
	$(STATICCHECK) ./...

# Requires golangci-lint v2 on PATH (uses .golangci.yml)
lint:
	golangci-lint run ./...

test:
	$(GO) test ./...

# Quality gate without external golangci-lint binary
check: fmt-check vet staticcheck test

# Includes golangci-lint (aggregator matching QA checklist)
check-all: check lint
