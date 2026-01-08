.PHONY: all build test clean fmt lint deps examples

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Directories
PKG_DIR=./pkg/...
CMD_DIR=./cmd/...
EXAMPLES_DIR=./cmd/examples

all: deps build test

deps:
	$(GOMOD) download
	$(GOMOD) tidy

build:
	$(GOBUILD) -v $(PKG_DIR)

test:
	$(GOTEST) -v $(PKG_DIR)

clean:
	$(GOCLEAN)
	rm -f coverage.out

fmt:
	$(GOFMT) -s -w .

lint:
	golangci-lint run

# Build examples
examples:
	$(GOBUILD) -o bin/basic $(EXAMPLES_DIR)/basic
	$(GOBUILD) -o bin/simple_prompt $(EXAMPLES_DIR)/simple_prompt
	$(GOBUILD) -o bin/multi_turn $(EXAMPLES_DIR)/multi_turn

# Run examples (requires CHUCKY_PROJECT_ID and CHUCKY_SECRET env vars)
run-basic:
	$(GOCMD) run $(EXAMPLES_DIR)/basic/main.go

run-simple:
	$(GOCMD) run $(EXAMPLES_DIR)/simple_prompt/main.go

run-multi:
	$(GOCMD) run $(EXAMPLES_DIR)/multi_turn/main.go

# Coverage
coverage:
	$(GOTEST) -coverprofile=coverage.out $(PKG_DIR)
	$(GOCMD) tool cover -html=coverage.out

# Documentation
doc:
	godoc -http=:6060

help:
	@echo "Available targets:"
	@echo "  all       - Download deps, build, and test"
	@echo "  deps      - Download and tidy dependencies"
	@echo "  build     - Build the packages"
	@echo "  test      - Run tests"
	@echo "  clean     - Clean build artifacts"
	@echo "  fmt       - Format code"
	@echo "  lint      - Run linter"
	@echo "  examples  - Build example binaries"
	@echo "  run-basic - Run basic example"
	@echo "  run-simple- Run simple prompt example"
	@echo "  run-multi - Run multi-turn example"
	@echo "  coverage  - Generate coverage report"
	@echo "  doc       - Start godoc server"
