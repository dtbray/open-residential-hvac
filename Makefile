GO ?= go

.PHONY: check test build frontend desktop
check:
	test -z "$$(gofmt -l pkg app/*.go cmd internal)"
	$(GO) vet ./...
	$(GO) test -race ./...
test:
	$(GO) test ./...
build:
	mkdir -p bin
	$(GO) build -o bin/hvac ./cmd/hvac
	$(GO) build -o bin/hvac-compare ./cmd/hvac-compare
frontend:
	cd app/frontend && npm ci && npm run check && npm run build
desktop: frontend
	mkdir -p bin
	$(GO) build -tags "desktop,webkit2_41" -o bin/open-residential-hvac ./cmd/desktop
