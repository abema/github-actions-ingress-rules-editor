NAME := ingress-rules-editor
VERSION := $(shell git describe --tags --abbrev=0)
REVISION := $(shell git rev-parse --short HEAD)
LDFLAGS := -X 'main.version=$(VERSION)' \
           -X 'main.revision=$(REVISION)'

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
	go build -ldflags "$(LDFLAGS)" -o ingress_rules_editor ./main.go

.PHONY: cross
cross: main.go  ## Build binaries for cross platform.
	mkdir -p builds
	@for arch in "amd64" "386"; do \
		GOOS=darwin GOARCH=$${arch} make build; \
		zip builds/ingress_rules_editor-$(VERSION)_darwin_$${arch}.zip ingress_rules_editor; \
	done;
	@for arch in "amd64" "386" "arm64"; do \
		GOOS=linux GOARCH=$${arch} make build; \
		zip builds/ingress_rules_editor-$(VERSION)_linux_$${arch}.zip ingress_rules_editor; \
	done;

.PHONY: docker-build
docker-build: ## Build docker image.
	docker build -t $(IMAGE_NAME) .

.PHONY: help
help: ## Show help text
	@echo "Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'
