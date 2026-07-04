# mapdb-golang developer tasks. Mirrors the CI pipeline so contributors can run
# the same gates locally. Zero third-party tools required (gofmt/go vet ship
# with the toolchain); the optional `lint` target uses golangci-lint if present.

.PHONY: all ci build vet test test-race fmt-check codegen-check lint tidy

# `make` runs the full local CI gate.
all: ci

ci: build vet fmt-check test test-race codegen-check

build:
	go build ./...

vet:
	go vet ./...

test:
	go test ./...

# The Synchronized wrappers' whole value proposition is lock safety; exercise it.
test-race:
	go test -race ./...

# gofmt check without rewriting: fail if any file is unformatted.
fmt-check:
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt needed on:"; echo "$$unformatted"; exit 1; \
	fi; \
	echo "gofmt: OK"

# Delete-then-regenerate drift gate (see scripts/check-codegen.sh).
codegen-check:
	./scripts/check-codegen.sh

# Optional: only runs if golangci-lint is installed.
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed; skipping"; \
	fi

tidy:
	go mod tidy
