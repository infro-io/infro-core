default: help

.PHONY: help
help: # Show help for each of the Makefile recipes.
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done

.PHONY: build-docker
build-docker: # Build Docker image
	docker build -t infro-core -f build/Dockerfile .

.PHONY: build-go
build-go: # Build go binary
	go build -o bin/infro ./cmd/main.go

.PHONY: format
format: # Format code based on linter configuration
	golangci-lint run --fix -v ./...

.PHONY: lint
lint: # Run linters
	golangci-lint run

.PHONY: test-integration
test-integration: # Run integration tests
	go test ./test/...

.PHONY: test-unit
test-unit: # Run unit tests
	go test `go list ./... | grep -v test`
