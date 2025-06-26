package main

import (
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"go_openapi_mcp/internal"
)

func main() {
	specSource := os.Getenv("OPENAPI_SPEC_URL")
	if specSource == "" {
		log.Fatalf("OPENAPI_SPEC_URL is not specified. Set it to a file path or http/https url")
	}

	// Create OpenAPI server
	oas := internal.NewOpenAPIServer(specSource, internal.GetCacheDir())

	// Load the spec
	if err := oas.LoadSpec(); err != nil {
		log.Fatalf("Failed to load OpenAPI spec: %v", err)
	}

	// Create and run MCP server
	s := internal.CreateMCPServerWithTools(oas)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}