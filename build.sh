#!/bin/bash

# Build all three executables
echo "Building openapi-mcp-http..."
go build -o openapi-mcp-http ./cmd/openapi-mcp-http

echo "Building openapi-mcp-interactive..."
go build -o openapi-mcp-interactive ./cmd/openapi-mcp-interactive

echo "Building openapi-mcp-stdio..."
go build -o openapi-mcp-stdio ./cmd/openapi-mcp-stdio

echo "All builds completed!"

# Optional: build Docker images
if [ "$1" = "docker" ]; then
    echo "Building Docker images..."
    
    # Build stdio version (default)
    echo "Building stdio version..."
    docker build --build-arg MODE=stdio -t openapi-mcp:stdio .
    docker tag openapi-mcp:stdio openapi-mcp:latest
    
    # Build HTTP version
    echo "Building HTTP version..."
    docker build --build-arg MODE=http -t openapi-mcp:http .
    
    # Build interactive version
    echo "Building interactive version..."
    docker build --build-arg MODE=interactive -t openapi-mcp:interactive .
    
    echo "Docker builds completed!"
    echo "Available images:"
    echo "  - openapi-mcp:stdio (also tagged as openapi-mcp:latest)"
    echo "  - openapi-mcp:http"
    echo "  - openapi-mcp:interactive"
fi