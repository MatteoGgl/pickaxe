## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS = -X github.com/matteoggl/pickaxe/cmd.version=$(VERSION)

## build: build binary to dist/pickaxe
.PHONY: build
build:
	@go build -ldflags "$(LDFLAGS)" -o dist/pickaxe .

## run: run the application
.PHONY: run
run:
	@go run ./
	
.PHONY: test
test:
	@go test ./...

## audit: tidy deps and format, vet and test all code
.PHONY: audit
audit:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...