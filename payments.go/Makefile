sync-docs:
	$(shell go env GOPATH)/bin/swag init -g cmd/main.go -d . -o docs

build-push:
	docker build -t ghcr.io/synehq/payments.go:latest . --platform linux/amd64
	docker push ghcr.io/synehq/payments.go:latest