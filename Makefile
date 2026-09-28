.PHONY: help install frontend backend test-frontend test-backend build-frontend build-backend check-installer
help:
	@echo "make install         Start the Docker Swarm setup wizard"
	@echo "make frontend        Start Next.js development (configure client/.env first)"
	@echo "make backend         Start the Go API (configure server/.env and dependencies first)"
	@echo "make test-frontend   Run TypeScript and Jest checks"
	@echo "make test-backend    Run Go tests"
	@echo "make build-frontend Build the Next.js app"
	@echo "make build-backend  Build the Go commands"
	@echo "make check-installer Check installer and release helpers"
install:
	python3 devops/swarm/install.py
frontend:
	cd client && bun run dev
backend:
	cd server && go run ./cmd
test-frontend:
	cd client && bunx tsc --noEmit && bunx jest --runInBand
test-backend:
	cd server && go test ./...
build-frontend:
	cd client && bun run build
build-backend:
	cd server && go build ./cmd/...
check-installer:
	bash -n scripts/install.sh
	shellcheck scripts/install.sh
	python3 -m unittest discover -s devops/swarm -p 'test_*.py'
	python3 -m unittest discover -s scripts -p 'test_*.py'
