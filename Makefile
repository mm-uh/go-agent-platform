DEFAULT_TARGET := run

.PHONY: build run test-panels test

DEFAULT_APP_NAME ?= "platform"
DEFAULT_APP_IP ?= "127.0.0.1"
DEFAULT_APP_PORT ?= "8080"

build: ## Build the node
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o $(DEFAULT_APP_NAME) .

run: build ## Run the node
	@./$(DEFAULT_APP_NAME) $(DEFAULT_APP_IP) $(DEFAULT_APP_PORT)

test: ## Test all app
	go test ./... -v
