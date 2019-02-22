.DEFAULT_GOAL := help

PKGS := $(shell go list ./...)
SOURCES := $(shell find . -path ./vendor -prune -o -name "*.go" -not -name '*_test.go' -print)
IMAGE_NAME := "github-actions-ingress-rules-editor"

.PHONY: fmt
fmt: $(SOURCES) ## Formatting source codes.
	@goimports -w $^

.PHONY: lint
lint: ## Run golint and go vet.
	@golint -set_exit_status=1 $(PKGS)
	@go vet $(PKGS)

.PHONY: build
build: ## Build ingress_rules_editor.
	go build -o ingress_rules_editor ./main.go

.PHONY: docker-build
docker-build: ## Build docker image.
	docker build -t $(IMAGE_NAME) .

.PHONY: help
help: ## Show help text
	@echo "Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'
