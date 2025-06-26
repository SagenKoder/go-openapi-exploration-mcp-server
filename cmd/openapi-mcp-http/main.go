package main

import (
	"flag"
	"log"
	"os"

	"go_openapi_mcp/internal"
)

func main() {
	var addr string
	flag.StringVar(&addr, "addr", internal.DefaultHTTPPort, "HTTP server address")
	flag.Parse()

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

	// Start HTTP server
	if err := internal.StartHTTPServer(oas, addr); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}