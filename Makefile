.DEFAULT_GOAL := default

COWBOY_IMAGE := ghcr.io/puskunalis/cowboy:2.0.0
COWBOY_CONTROLLER_IMAGE := ghcr.io/puskunalis/cowboy-controller:2.0.0

KIND_CLUSTER_NAME=wildwest

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
	sudo docker buildx build --push --platform linux/amd64 -t $(COWBOY_IMAGE) -f cowboy.Dockerfile .

.PHONY: create-cowboy-controller-image
create-cowboy-controller-image: ## Build and push cowboy-controller container image
ifeq ($(shell which docker || echo false),false)
	@echo Error: docker not found in \$$PATH
	@exit 1
endif
	sudo docker buildx build --push --platform linux/amd64 -t $(COWBOY_CONTROLLER_IMAGE) -f cowboy-controller.Dockerfile .

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
	go test -race -cover -count=1 ./...

.PHONY: test-cover
test-cover: ## Run tests and display coverage info in HTML viewer
ifeq ($(shell which go || echo false),false)
	@echo Error: go not found in \$$PATH
	@exit 1
endif
	t="/tmp/go-cover.\$\$\.tmp"; \
	go test -coverprofile=$$t ./... && go tool cover -html=$$t && unlink $$t

.PHONY: kind-up
kind-up: ## Creates a kind-wildwest cluster if it doesn't exist
ifeq ($(shell which kind || echo false),false)
	@echo Error: kind not found in \$$PATH
	@exit 1
endif
	@if [ -z "`kind get clusters | grep $(KIND_CLUSTER_NAME)`" ]; then \
		kind create cluster --name $(KIND_CLUSTER_NAME); \
	else \
		echo "Kind cluster '$(KIND_CLUSTER_NAME)' already exists."; \
	fi

.PHONY: kind-down
kind-down: ## Deletes the kind-wildwest cluster
ifeq ($(shell which kind || echo false),false)
	@echo Error: kind not found in \$$PATH
	@exit 1
endif
	@if [ ! -z "`kind get clusters | grep $(KIND_CLUSTER_NAME)`" ]; then \
		kind delete cluster --name $(KIND_CLUSTER_NAME); \
	else \
		echo "Kind cluster '$(KIND_CLUSTER_NAME)' does not exist."; \
	fi

.PHONY: helm-install
helm-install: kind-up ## Install Helm chart
ifeq ($(shell which helm || echo false),false)
	@echo Error: helm not found in \$$PATH
	@exit 1
endif
	helm install wildwest helm/ -n wildwest --create-namespace

.PHONY: helm-uninstall
helm-uninstall: ## Uninstall Helm chart
ifeq ($(shell which helm || echo false),false)
	@echo Error: helm not found in \$$PATH
	@exit 1
endif
	helm uninstall wildwest -n wildwest
	@echo "To delete the kind cluster, run 'make kind-down'"

.PHONY: logs
logs: ## Get logs of the cowboy pods
ifeq ($(shell which kubectl || echo false),false)
	@echo Error: kubectl not found in \$$PATH
	@exit 1
endif
	kubectl logs -n wildwest --tail=-1 -l 'app in (cowboy, cowboy-controller)' --all-containers --ignore-errors | grep -v "DEBUG" | sort | less

.PHONY: help
help: ## Makefile help
	@grep -P '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
