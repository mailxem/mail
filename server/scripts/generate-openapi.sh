#!/bin/bash

# OpenAPI 3.0 Generation Script
# This script generates Swagger 2.0 first, then converts it to OpenAPI 3.0

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

# Function to check if required tools are installed
check_dependencies() {
    gobin=$(go env GOPATH)/bin
    export PATH=$PATH:$gobin
    # Check for swag
    if ! command -v swag &> /dev/null; then
        print_error "swag CLI tool not found. Installing..."
        go install github.com/swaggo/swag/cmd/swag@latest
        export PATH=$PATH:$gobin
    fi
    
    # Check for redocly
    if ! command -v redocly &> /dev/null; then
        print_error "redocly CLI tool not found. Installing..."
        npm install -g @redocly/cli
    fi
}

# Function to generate OpenAPI 3.0 specification
generate_openapi() {
    print_header "Generating OpenAPI 3.0 Specification"
    check_dependencies
    
    print_status "Cleaning old documentation..."
    rm -rf docs/swagger/*
    
    print_status "Generating Swagger 2.0 specification..."
    export PATH=$PATH:$gobin
    swag init -g cmd/main.go -o docs/swagger --parseDependency --parseInternal
    
    print_status "Converting Swagger 2.0 to OpenAPI 3.0..."
    
    # Create a temporary conversion script
    cat > /tmp/convert_to_openapi.js << 'EOF'
const fs = require('fs');
const path = require('path');

// Get the current working directory
const currentDir = process.cwd();

// Read the swagger.json file
const swaggerPath = path.join(currentDir, 'docs/swagger/swagger.json');
const swaggerContent = fs.readFileSync(swaggerPath, 'utf8');
const swagger = JSON.parse(swaggerContent);

// Convert to OpenAPI 3.0 format
let openapi = {
    openapi: "3.0.3",
    info: {
        title: swagger.info.title || "Xem API",
        version: swagger.info.version || "1.0",
        description: swagger.info.description || "Comprehensive API server for Xem email marketing platform",
        contact: swagger.info.contact || {},
        license: swagger.info.license || {
            name: "MIT",
            url: "https://opensource.org/licenses/MIT"
        }
    },
    servers: [
        {
            url: "https://api.xem.email/api/v1",
            description: "Production server"
        },
        {
            url: "http://localhost:8080/api/v1",
            description: "Development server"
        }
    ],
    paths: {},
    components: {
        securitySchemes: {
            BearerAuth: {
                type: "http",
                scheme: "bearer",
                bearerFormat: "JWT",
                description: "Enter the token with the `Bearer: ` prefix, e.g. \"Bearer abcde12345\"."
            },
            ApiKeyAuth: {
                type: "apiKey",
                in: "header",
                name: "X-API-KEY",
                description: "API key for programmatic access"
            }
        },
        schemas: {}
    },
    tags: [
        { name: "Authentication", description: "User authentication and authorization endpoints" },
        { name: "Teams", description: "Team management and collaboration features" },
        { name: "Campaigns", description: "Email campaign creation, management, and tracking" },
        { name: "Analytics", description: "Campaign analytics, audience insights, and performance metrics" },
        { name: "Contacts", description: "Contact management and mailing list operations" },
        { name: "Templates", description: "Email template management and customization" },
        { name: "Automations", description: "Email automation workflows and triggers" },
        { name: "SMTP", description: "SMTP configuration and email delivery settings" },
        { name: "IMAP", description: "IMAP configuration for email inbox management" },
        { name: "Webhooks", description: "Webhook management for real-time event notifications" },
        { name: "Files", description: "File upload and management for attachments and media" },
        { name: "Domains", description: "Domain management for email authentication" },
        { name: "API Keys", description: "API key management for programmatic access" }
    ]
};

// Convert paths
for (const [path, methods] of Object.entries(swagger.paths)) {
    openapi.paths[path] = {};
    
    for (const [method, operation] of Object.entries(methods)) {
        const openapiOperation = {
            summary: operation.summary || "",
            description: operation.description || "",
            tags: operation.tags || [],
            parameters: [],
            responses: {},
            security: []
        };
        
        // Convert parameters
        if (operation.parameters) {
            const formDataParams = [];
            const bodyParams = [];
            
            for (const param of operation.parameters) {
                // Handle formData parameters (convert to requestBody for OpenAPI 3.0)
                if (param.in === 'formData') {
                    formDataParams.push(param);
                    continue;
                }
                
                // Handle body parameters (convert to requestBody for OpenAPI 3.0)
                if (param.in === 'body') {
                    bodyParams.push(param);
                    continue;
                }
                
                const openapiParam = {
                    name: param.name,
                    in: param.in,
                    description: param.description || "",
                    required: param.required || false
                };
                
                if (param.schema) {
                    openapiParam.schema = param.schema;
                } else if (param.type) {
                    openapiParam.schema = { type: param.type };
                }
                
                openapiOperation.parameters.push(openapiParam);
            }
            
            // Convert formData parameters to requestBody
            if (formDataParams.length > 0) {
                const requestBody = {
                    required: formDataParams.some(p => p.required),
                    content: {
                        "multipart/form-data": {
                            schema: {
                                type: "object",
                                properties: {},
                                required: []
                            }
                        }
                    }
                };
                
                for (const param of formDataParams) {
                    requestBody.content["multipart/form-data"].schema.properties[param.name] = {
                        type: param.type === "file" ? "string" : "string",
                        format: param.type === "file" ? "binary" : undefined,
                        description: param.description || ""
                    };
                    
                    if (param.required) {
                        requestBody.content["multipart/form-data"].schema.required.push(param.name);
                    }
                }
                
                openapiOperation.requestBody = requestBody;
            }
            
            // Convert body parameters to requestBody
            if (bodyParams.length > 0) {
                const bodyParam = bodyParams[0]; // Usually only one body parameter
                const requestBody = {
                    required: bodyParam.required || false,
                    content: {
                        "application/json": {
                            schema: bodyParam.schema || {}
                        }
                    }
                };
                
                openapiOperation.requestBody = requestBody;
            }
        }
        
        // Convert responses
        for (const [code, response] of Object.entries(operation.responses)) {
            openapiOperation.responses[code] = {
                description: response.description || "",
                content: {
                    "application/json": {
                        schema: response.schema || {}
                    }
                }
            };
        }
        
        // Add security if needed
        if (operation.security) {
            openapiOperation.security = operation.security;
        }
        
        openapi.paths[path][method] = openapiOperation;
    }
}

// Convert definitions to components.schemas
if (swagger.definitions) {
    for (const [name, schema] of Object.entries(swagger.definitions)) {
        openapi.components.schemas[name] = schema;
    }
}

// Fix reference paths from #/definitions/ to #/components/schemas/
function fixReferences(obj) {
    if (typeof obj !== 'object' || obj === null) {
        return obj;
    }
    
    if (Array.isArray(obj)) {
        return obj.map(fixReferences);
    }
    
    const result = {};
    for (const [key, value] of Object.entries(obj)) {
        if (key === '$ref' && typeof value === 'string' && value.startsWith('#/definitions/')) {
            result[key] = value.replace('#/definitions/', '#/components/schemas/');
        } else {
            result[key] = fixReferences(value);
        }
    }
    return result;
}

// Apply reference fixes to the entire OpenAPI object
openapi = fixReferences(openapi);

// Write the OpenAPI 3.0 specification
const openapiPath = path.join(currentDir, 'openapi.json');
fs.writeFileSync(openapiPath, JSON.stringify(openapi, null, 2));

console.log('OpenAPI 3.0 specification generated successfully!');
EOF

    # Run the conversion script
    node /tmp/convert_to_openapi.js
    
    # Clean up
    rm /tmp/convert_to_openapi.js
    
    print_status "OpenAPI 3.0 specification generated as openapi.json!"
    print_status "Files created:"
    ls -la openapi.json
    ls -la docs/swagger/
}

# Function to show help
show_help() {
    echo "OpenAPI 3.0 Generation Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  generate  - Generate OpenAPI 3.0 specification as openapi.json"
    echo "  clean     - Clean generated documentation"
    echo "  help      - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 generate  # Generate openapi.json"
}

# Main script logic
case "${1:-help}" in
    "generate"|"gen")
        generate_openapi
        ;;
    "clean")
        print_header "Cleaning Documentation"
        rm -rf docs/swagger/*
        rm -f openapi.json
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
