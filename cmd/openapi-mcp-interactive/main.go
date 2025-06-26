package main

import (
	"log"
	"os"

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

	// Run interactive mode
	internal.RunInteractiveMode(oas)
}