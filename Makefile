.DEFAULT_GOAL := default

COWBOY_IMAGE_NAME := ghcr.io/puskunalis/cowboy
COWBOY_IMAGE_VERSION := 1.0.0

COWBOY_CONTROLLER_IMAGE_NAME := ghcr.io/puskunalis/cowboy-controller
COWBOY_CONTROLLER_IMAGE_VERSION := 1.0.0

.PHONY: default
default: create-cowboy-image create-cowboy-controller-image ## Run default target

.PHONY: protogen
protogen: ## Generate gRPC files
ifeq ($(shell which protoc || echo false),false)
	@echo Error: protoc not found in \$$PATH
	@exit 1
endif
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/proto/damage/damage.proto api/proto/shootout/shootout.proto

.PHONY: create-cowboy-image
create-cowboy-image: ## Build and push cowboy container image
ifeq ($(shell which docker || echo false),false)
	@echo Error: docker not found in \$$PATH
	@exit 1
endif
	docker buildx build --push --platform linux/amd64 -t $(COWBOY_IMAGE_NAME):$(COWBOY_IMAGE_VERSION) -f cowboy.Dockerfile .

.PHONY: create-cowboy-controller-image
create-cowboy-controller-image: ## Build and push cowboy-controller container image
ifeq ($(shell which docker || echo false),false)
	@echo Error: docker not found in \$$PATH
	@exit 1
endif
	docker buildx build --push --platform linux/amd64 -t $(COWBOY_CONTROLLER_IMAGE_NAME):$(COWBOY_CONTROLLER_IMAGE_VERSION) -f cowboy-controller.Dockerfile .

.PHONY: lint
lint: ## Run linter
ifeq ($(shell which golangci-lint || echo false),false)
	@echo Error: golangci-lint not found in \$$PATH
	@exit 1
endif
	golangci-lint run --enable-all ./...

.PHONY: test
test: ## Run tests
ifeq ($(shell which go || echo false),false)
	@echo Error: go not found in \$$PATH
	@exit 1
endif
	go test -race -cover ./...

.PHONY: help
help: ## Makefile help
	@grep -P '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
