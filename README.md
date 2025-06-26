# OpenAPI MCP Server

A Model Context Protocol (MCP) server that provides tools for exploring and analyzing OpenAPI specifications. This server enables LLMs to intelligently navigate and understand API documentation through structured tools.

## Features

- **Dynamic OpenAPI Loading**: Load specifications from local files or remote URLs with automatic caching
- **Intelligent Navigation**: Hierarchical exploration of API endpoints by categories
- **Comprehensive Analysis**: Detailed inspection of endpoints, parameters, request/response schemas
- **Multiple Run Modes**: Stdio (default), HTTP server, and interactive CLI modes
- **Schema Inspection**: Deep dive into data models and schema definitions

## MCP Tools

The server exposes the following tools through the MCP protocol:

### 1. `list_categories`
Lists all API categories based on the first path segment of endpoints. This provides a high-level overview of the API structure.

**Parameters**: None

**Example Response**:
```json
[
  {
    "name": "users",
    "endpoint_count": 5
  },
  {
    "name": "products", 
    "endpoint_count": 8
  }
]
```

### 2. `list_endpoints`
Lists all endpoints, optionally filtered by category. Provides summary information for each endpoint.

**Parameters**:
- `category` (optional): Filter endpoints by category name

**Example Response**:
```json
[
  {
    "path": "/users/{id}",
    "method": "GET",
    "summary": "Get user by ID",
    "operationId": "getUserById"
  }
]
```

### 3. `show_endpoint`
Shows detailed information about a specific endpoint including parameters, request body, and responses.

**Parameters**:
- `path` (required): The endpoint path (e.g., `/users/{id}`)
- `method` (required): HTTP method (GET, POST, PUT, DELETE, etc.)

**Returns**: Complete endpoint details including:
- Path and query parameters with types
- Request body schema (limited expansion depth)
- Response schemas for different status codes
- Operation metadata (tags, summary, description)

### 4. `get_spec_info`
Retrieves general information about the OpenAPI specification.

**Parameters**: None

**Returns**: 
- API title, version, description
- Contact and license information
- Server URLs
- Statistics (total paths, operations, tags)

### 5. `show_schema`
Shows detailed information about a specific schema component.

**Parameters**:
- `ref` (required): Schema reference (e.g., `#/components/schemas/User`)

**Returns**: Complete schema definition with properties, types, and constraints (up to 4 levels deep).

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/go_openapi_mcp.git
cd go_openapi_mcp

# Build the binary
go build -o go_openapi_mcp .
```

## Run Modes

### 1. Stdio Mode (Default)
Standard input/output mode for MCP communication. This is the default mode used by MCP clients like Claude Desktop.

```bash
# Run with default OpenAPI spec (data/openapi.json)
./go_openapi_mcp

# Run with custom local spec
OPENAPI_SPEC_URL=./my-api-spec.json ./go_openapi_mcp

# Run with remote spec
OPENAPI_SPEC_URL=https://api.example.com/openapi.json ./go_openapi_mcp
```

### 2. HTTP Server Mode
Runs as an HTTP server with Server-Sent Events (SSE) support for MCP communication over HTTP.

```bash
# Run on default port 8080
./go_openapi_mcp -http

# Run on custom port
./go_openapi_mcp -http -addr :3000

# Run on specific interface
./go_openapi_mcp -http -addr 192.168.1.100:8080
```

The MCP endpoint will be available at: `http://[addr]/mcp`

This mode implements the MCP Streamable HTTP transport specification:
- POST `/mcp` - Send JSON-RPC requests
- GET `/mcp` - Establish SSE connection for streaming
- DELETE `/mcp` - Terminate session

### 3. Interactive CLI Mode
Interactive command-line interface for testing and exploring the API manually.

```bash
./go_openapi_mcp -interactive
```

This mode provides a menu-driven interface for:
1. Listing categories
2. Browsing endpoints (with optional category filter)
3. Viewing endpoint details
4. Getting spec information
5. Inspecting schema definitions

## Command-Line Flags

| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `-http` | Run as HTTP server | false | `./go_openapi_mcp -http` |
| `-addr` | HTTP server address (only with -http) | `:8080` | `./go_openapi_mcp -http -addr :3000` |
| `-interactive` | Run in interactive CLI mode | false | `./go_openapi_mcp -interactive` |

## Environment Variables

| Variable            | Description                        | Default                |
|---------------------|------------------------------------|------------------------|
| `OPENAPI_SPEC_URL`  | URL or path to OpenAPI spec        | `data/openapi.json`    |
| `OPENAPI_CACHE_DIR` | Directory for caching remote specs | `~/.openapi-mcp-cache` |

## Usage Examples

### With Claude Desktop

Add to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "openapi": {
      "command": "/path/to/go_openapi_mcp",
      "env": {
        "OPENAPI_SPEC_URL": "https://api.example.com/openapi.json"
      }
    }
  }
}
```

### With MCP Inspector

```bash
# Test with local spec
npx @modelcontextprotocol/inspector ./go_openapi_mcp

# Test with remote spec
OPENAPI_SPEC_URL=https://petstore.swagger.io/v2/swagger.json npx @modelcontextprotocol/inspector ./go_openapi_mcp
```

### HTTP Mode for Web Integration

```bash
# Start server
./go_openapi_mcp -http -addr :8080

# In your web application, connect to:
# http://localhost:8080/mcp
```

Example client code:
```javascript
// Initialize session
const response = await fetch('http://localhost:8080/mcp', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    jsonrpc: '2.0',
    method: 'initialize',
    params: { 
      clientInfo: { name: 'web-client', version: '1.0' }
    },
    id: 1
  })
});

const sessionId = response.headers.get('Mcp-Session-Id');

// Call a tool
const toolResponse = await fetch('http://localhost:8080/mcp', {
  method: 'POST',
  headers: { 
    'Content-Type': 'application/json',
    'Mcp-Session-Id': sessionId
  },
  body: JSON.stringify({
    jsonrpc: '2.0',
    method: 'tools/call',
    params: {
      name: 'list_categories',
      arguments: {}
    },
    id: 2
  })
});
```

## Caching

Remote OpenAPI specifications are automatically cached for 24 hours to improve performance:
- Cache location: `~/.openapi-mcp-cache/` (customizable via `OPENAPI_CACHE_DIR`)
- Cache files use SHA256 hash of URL as filename
- Metadata tracks expiration time
- Stale cache is automatically refreshed

## Docker Support

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o go_openapi_mcp .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/go_openapi_mcp .
CMD ["./go_openapi_mcp"]
```

Run with Docker:
```bash
# Build
docker build -t openapi-mcp .

# Run with URL (cache persisted)
docker run -v openapi-cache:/root/.openapi-mcp-cache \
  -e OPENAPI_SPEC_URL=https://api.example.com/openapi.json \
  openapi-mcp

# Run HTTP mode
docker run -p 8080:8080 openapi-mcp -http
```

## Common Use Cases

### API Discovery Workflow

1. **Start with categories** to understand API structure:
   ```
   Tool: list_categories
   ```

2. **Explore a specific category**:
   ```
   Tool: list_endpoints
   Arguments: {"category": "users"}
   ```

3. **Examine endpoint details**:
   ```
   Tool: show_endpoint
   Arguments: {"path": "/users/{id}", "method": "GET"}
   ```

4. **Understand data models**:
   ```
   Tool: show_schema
   Arguments: {"ref": "#/components/schemas/User"}
   ```

### Integration Testing

Use the interactive mode to quickly test API understanding:
```bash
./go_openapi_mcp -interactive
```

Then use the numbered menu to explore different aspects of your API.

## Development

### Project Structure

```
go_openapi_mcp/
├── main.go              # Main app, tool handlers, core logic
├── http_server.go       # HTTP/SSE server implementation
├── go.mod              # Go module definition
├── data/               # Default spec location
│   └── openapi.json
└── examples/           # Example files
    └── http-client.html
```

### Building from Source

```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Build for current platform
go build -o go_openapi_mcp .

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o go_openapi_mcp-linux
GOOS=darwin GOARCH=amd64 go build -o go_openapi_mcp-darwin
GOOS=windows GOARCH=amd64 go build -o go_openapi_mcp.exe
```

## Requirements

- Go 1.23 or later
- OpenAPI 3.0+ specification in JSON format

## License

MIT License

## Acknowledgments

- Built with [MCP Go SDK](https://github.com/mark3labs/mcp-go)
- Uses [kin-openapi](https://github.com/getkin/kin-openapi) for OpenAPI parsing