#!/bin/bash

# OpenAPI/Swagger Documentation Management Script
# Usage: ./scripts/swagger.sh [command]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

# Function to check if swag is installed
check_swag() {
    if ! command -v swag &> /dev/null; then
        print_error "swag CLI tool not found. Installing..."
        go install github.com/swaggo/swag/cmd/swag@latest
        export PATH=$PATH:$(go env GOPATH)/bin
    fi
}

# Function to generate documentation
generate_docs() {
    print_header "Generating OpenAPI/Swagger Documentation"
    check_swag
    
    print_status "Cleaning old documentation..."
    rm -rf docs/swagger/*
    
    print_status "Generating new documentation..."
    export PATH=$PATH:$(go env GOPATH)/bin
    swag init -g cmd/main.go -o docs/swagger --parseDependency --parseInternal
    
    print_status "Documentation generated successfully!"
    print_status "Files created:"
    ls -la docs/swagger/
}

# Function to serve documentation
serve_docs() {
    print_header "Starting Swagger UI Server"
    check_swag
    
    if [ ! -f "docs/swagger/swagger.json" ]; then
        print_warning "No swagger.json found. Generating documentation first..."
        generate_docs
    fi
    
    print_status "Starting Swagger UI server on http://localhost:8080/swagger/"
    export PATH=$PATH:$(go env GOPATH)/bin
    swag serve -F=swagger docs/swagger/swagger.json
}

# Function to show help
show_help() {
    echo "OpenAPI/Swagger Documentation Management Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  generate  - Generate OpenAPI/Swagger documentation"
    echo "  serve     - Start Swagger UI server"
    echo "  clean     - Clean generated documentation"
    echo "  help      - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 generate  # Generate documentation"
    echo "  $0 serve     # Start Swagger UI server"
}

# Main script logic
case "${1:-help}" in
    "generate"|"gen")
        generate_docs
        ;;
    "serve"|"server")
        serve_docs
        ;;
    "clean")
        print_header "Cleaning Documentation"
        rm -rf docs/swagger/*
        print_status "Documentation cleaned!"
        ;;
    "help"|"-h"|"--help"|"")
        show_help
        ;;
    *)
        print_error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac
