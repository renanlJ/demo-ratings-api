APP := demo-ratings-api
IMAGE ?= $(APP):0.1.0
PORT ?= 8083

.PHONY: install lint format format-write test test-integration coverage build run docker-build docker-scan k8s-validate clean

install: ## Download modules (none today - stdlib only)
	go mod download

lint: ## go vet
	go vet ./...

format: ## Check gofmt
	@test -z "$$(gofmt -l .)" || (echo "unformatted files:"; gofmt -l .; exit 1)

format-write: ## Apply gofmt
	gofmt -w .

test: ## Unit + handler tests
	go test ./... -v

test-integration: test

coverage: ## Coverage report (coverage.out)
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out | tail -1

build: ## Compile binary
	CGO_ENABLED=0 go build -trimpath -o $(APP) ./cmd/ratings-api

run: build ## Run locally
	PORT=$(PORT) ./$(APP)

docker-build:
	docker build -t $(IMAGE) .

docker-scan:
	trivy image $(IMAGE)

k8s-validate:
	kubectl kustomize deploy/kubernetes | kubectl apply --dry-run=client -f -

clean:
	rm -f $(APP) coverage.out coverage.html
