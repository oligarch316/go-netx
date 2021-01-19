#!/usr/bin/make -f

# Disable all default suffixes
.SUFFIXES:

go := $(shell command -v go || echo "go")
golint := $(shell command -v golint || echo "golint")

# ----- Default
.PHONY: default

default: fmt tidy test ## fmt + tidy + test

# ----- Tools
.PHONY: fmt lint tidy
fmt: ## Format
	$(info Foramatting)
	@$(go) fmt ./...

lint: ## Lint
	$(info Linting)
	@$(golint) ./...

tidy: ## Tidy go modules
	$(info Tidying)
	@$(go) mod tidy

# ----- Test
test_flags :=
test_coverprofile_target := coverage.out

.PHONY: test test.v test.race test.race.v test.cover coverage clean.test.cover

test: go := $(shell command -v richgo || echo "$(go)")
test: ## Run tests
	$(info Testing)
	@$(go) test $(test_flags) ./...

test.v: test_flags += -v
test.v: test ## Run tests with verbose output

test.race: test_flags += -race -run ^TestConcurrent
test.race: test ## Run tests with race detection

test.race.v: test_flags += -v
test.race.v: test.race ## Run tests with race detectiona and verbose output

test.cover: test_flags += -coverprofile=$(test_coverprofile_target)
test.cover: test ## Generate test coverage report

coverage: test.cover ## Generate and open test coverage report in browser
	@$(go) tool cover -html $(test_coverprofile_target)

clean.test.cover: ## Clean test coverage artifacts
	$(info Cleaning test coverage)
	@rm $(test_coverprofile_target) 2> /dev/null || true

# ----- Clean
.PHONY: clean

clean: clean.test.cover ## Clean all

# ----- HELP
.PHONY: help

print.%: ; @echo "$($*)"

# TODO: help
help: ## Show help information
	@awk -F ':|##' '/^[^\t].+?:.*?##/ {printf "\033[36m%-30s\033[0m %s\n", $$1, $$NF }' $(MAKEFILE_LIST);
