package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	DefaultCacheTTL        = 24 * time.Hour
	DefaultHTTPPort        = ":8080"
	SchemaMaxDepth         = 2
	DetailedSchemaMaxDepth = 4
)

type OpenAPIServer struct {
	spec       *openapi3.T
	specSource string // URL or file path
	cache      *Cache
}

func NewOpenAPIServer(specSource string, cacheDir string) *OpenAPIServer {
	return &OpenAPIServer{
		specSource: specSource,
		cache:      NewCache(cacheDir, DefaultCacheTTL),
	}
}

func (oas *OpenAPIServer) LoadSpec() error {
	var data []byte
	var err error

	// Check if source is a URL
	if strings.HasPrefix(oas.specSource, "http://") || strings.HasPrefix(oas.specSource, "https://") {
		data, err = oas.cache.LoadFromURL(oas.specSource)
	} else {
		// Load from file
		data, err = os.ReadFile(oas.specSource)
	}

	if err != nil {
		return fmt.Errorf("failed to load spec: %w", err)
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	spec, err := loader.LoadFromData(data)
	if err != nil {
		return fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	oas.spec = spec
	return nil
}

func GetCacheDir() string {
	cacheDir := os.Getenv("OPENAPI_CACHE_DIR")
	if cacheDir == "" {
		homeDir, _ := os.UserHomeDir()
		cacheDir = filepath.Join(homeDir, ".openapi-mcp-cache")
	}
	return cacheDir
}

// CreateMCPServerWithTools creates an MCP server instance with all tools registered
func CreateMCPServerWithTools(oas *OpenAPIServer) *server.MCPServer {
	s := server.NewMCPServer(
		"OpenAPI MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Register all tools
	listCategoriesTool := mcp.NewTool("list_categories",
		mcp.WithDescription("List all categories based on the first path segment of endpoints. Always call this before querying deeper!"),
	)
	s.AddTool(listCategoriesTool, oas.listCategoriesHandler)

	listEndpointsTool := mcp.NewTool("list_endpoints",
		mcp.WithDescription("List endpoints, filtered by category (based on first path segment). Always check the list of categories first!"),
		mcp.WithString("category",
			mcp.Description("The category (first path segment) to filter endpoints by."),
		),
	)
	s.AddTool(listEndpointsTool, oas.listEndpointsHandler)

	showEndpointTool := mcp.NewTool("show_endpoint",
		mcp.WithDescription("Show detailed information about a specific endpoint including types"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("The path of the endpoint (e.g., /users/{id})"),
		),
		mcp.WithString("method",
			mcp.Required(),
			mcp.Description("The HTTP method (GET, POST, PUT, DELETE, etc.)"),
		),
	)
	s.AddTool(showEndpointTool, oas.showEndpointHandler)

	getSpecInfoTool := mcp.NewTool("get_spec_info",
		mcp.WithDescription("Get general information about the OpenAPI specification"),
	)
	s.AddTool(getSpecInfoTool, oas.getSpecInfoHandler)

	showSchemaTool := mcp.NewTool("show_schema",
		mcp.WithDescription("Show details of a specific schema component by reference"),
		mcp.WithString("ref",
			mcp.Required(),
			mcp.Description("The schema reference (e.g., #/components/schemas/User)"),
		),
	)
	s.AddTool(showSchemaTool, oas.showSchemaHandler)

	return s
}
