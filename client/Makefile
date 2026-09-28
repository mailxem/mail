.PHONY: help dev build test lint format db-up db-down migrate superuser clean install logs

# Colors for terminal output
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_CYAN=\033[36m

help: ## Show this help message
	@echo '${COLOR_BOLD}Available commands:${COLOR_RESET}'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  ${COLOR_CYAN}%-20s${COLOR_RESET} %s\n", $$1, $$2}'

install: ## Install dependencies
	@echo "${COLOR_GREEN}Installing dependencies...${COLOR_RESET}"
	bun install

dev: ## Start development server
	@echo "${COLOR_GREEN}Starting development server...${COLOR_RESET}"
	bun run dev

build: ## Build for production
	@echo "${COLOR_GREEN}Building for production...${COLOR_RESET}"
	bun run build

test: ## Run tests
	@echo "${COLOR_GREEN}Running tests...${COLOR_RESET}"
	bun run test

lint: ## Run linter
	@echo "${COLOR_GREEN}Running linter...${COLOR_RESET}"
	bun run lint

db-up: ## Start database containers
	@echo "${COLOR_GREEN}Starting database containers...${COLOR_RESET}"
	docker-compose up -d postgres redis

db-down: ## Stop database containers
	@echo "${COLOR_GREEN}Stopping database containers...${COLOR_RESET}"
	docker-compose down

migrate: ## Run database migrations
	@echo "${COLOR_GREEN}Running database migrations...${COLOR_RESET}"
	bunx prisma migrate dev

superuser: ## Create a superuser
	@echo "${COLOR_GREEN}Creating superuser...${COLOR_RESET}"
	@read -p "Email: " email; \
	read -p "Password: " password; \
	read -p "Name: " name; \
	curl -X POST http://localhost:3000/api/auth/signup \
		-H "Content-Type: application/json" \
		-d "{\"email\":\"$$email\",\"password\":\"$$password\",\"name\":\"$$name\",\"role\":\"ADMIN\"}"

clean: ## Clean build artifacts and cache
	@echo "${COLOR_GREEN}Cleaning build artifacts and cache...${COLOR_RESET}"
	rm -rf .next
	rm -rf node_modules
	rm -rf .turbo
	rm -rf .cache

logs: ## View application logs
	@echo "${COLOR_GREEN}Viewing application logs...${COLOR_RESET}"
	tail -f logs/combined.log

db-shell: ## Open database shell
	@echo "${COLOR_GREEN}Opening database shell...${COLOR_RESET}"
	docker-compose exec postgres psql -U postgres -d emailengine

redis-shell: ## Open Redis shell
	@echo "${COLOR_GREEN}Opening Redis shell...${COLOR_RESET}"
	docker-compose exec redis redis-cli

seed: ## Seed the database with sample data
	@echo "${COLOR_GREEN}Seeding database...${COLOR_RESET}"
	bunx prisma db seed

backup: ## Backup the database
	@echo "${COLOR_GREEN}Backing up database...${COLOR_RESET}"
	docker-compose exec postgres pg_dump -U postgres emailengine > backup_$(shell date +%Y%m%d_%H%M%S).sql

restore: ## Restore database from backup
	@echo "${COLOR_GREEN}Restoring database from backup...${COLOR_RESET}"
	@read -p "Backup file: " file; \
	docker-compose exec -T postgres psql -U postgres emailengine < $$file

init: ## Initialize the project (install, migrate, seed)
	@echo "${COLOR_GREEN}Initializing project...${COLOR_RESET}"
	make install
	make db-up
	sleep 5
	make migrate
	make seed

template-list: ## List all email templates
	@echo "${COLOR_GREEN}Listing email templates...${COLOR_RESET}"
	curl -s http://localhost:3000/api/templates | jq

campaign-status: ## Check campaign status
	@echo "${COLOR_GREEN}Checking campaign status...${COLOR_RESET}"
	@read -p "Campaign ID: " id; \
	curl -s http://localhost:3000/api/campaigns/$$id | jq 