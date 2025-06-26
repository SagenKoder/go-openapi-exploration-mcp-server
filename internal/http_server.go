package internal

import (
	"log"

	"github.com/mark3labs/mcp-go/server"
)

// StartHTTPServer starts the HTTP server using the built-in StreamableHTTPServer
func StartHTTPServer(oas *OpenAPIServer, addr string) error {
	// Create MCP server instance with the same tools as stdio
	mcpServer := CreateMCPServerWithTools(oas)

	// Create HTTP server with StreamableHTTPServer
	httpServer := server.NewStreamableHTTPServer(mcpServer)

	log.Printf("MCP HTTP server starting on %s", addr)

	return httpServer.Start(addr)
}