# OpenAPI MCP Server

A Model Context Protocol (MCP) server that enables LLMs to explore and understand OpenAPI specifications through structured tools.

## Features

- 🔍 **Smart API Exploration** - Navigate APIs by categories, endpoints, and schemas
- 🚀 **Multiple Modes** - Run as stdio (for Claude Desktop), HTTP server, or interactive CLI
- 💾 **Intelligent Caching** - Caches remote OpenAPI specs for faster access
- 🏗️ **Multi-Architecture** - Supports Linux AMD64 and ARM64

## Quick Start

### Using Claude Desktop

Add to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "openapi": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "-e", "OPENAPI_SPEC_URL=https://api.example.com/openapi.json", "ghcr.io/sagenkoder/go-openapi-exploration-mcp-server:latest"]
    }
  }
}
```

Or use a local binary:

```json
{
  "mcpServers": {
    "openapi": {
      "command": "/path/to/openapi-mcp-stdio",
      "env": {
        "OPENAPI_SPEC_URL": "https://api.example.com/openapi.json"
      }
    }
  }
}
```

### Installation

#### Docker (Recommended)

```bash
# Latest (stdio mode)
docker pull ghcr.io/sagenkoder/go-openapi-exploration-mcp-server:latest

# Specific modes
docker pull ghcr.io/sagenkoder/go-openapi-exploration-mcp-server:http
docker pull ghcr.io/sagenkoder/go-openapi-exploration-mcp-server:interactive
```

#### Download Binaries

Download from [releases](https://github.com/SagenKoder/go-openapi-exploration-mcp-server/releases):
- `openapi-mcp-stdio-linux-amd64` - MCP stdio mode
- `openapi-mcp-http-linux-amd64` - HTTP server mode
- `openapi-mcp-interactive-linux-amd64` - Interactive CLI mode

#### Build from Source

```bash
# Clone
git clone https://github.com/SagenKoder/go-openapi-exploration-mcp-server.git
cd go-openapi-exploration-mcp-server

# Build all modes
./build.sh

# Or build specific mode
go build -o openapi-mcp-stdio ./cmd/openapi-mcp-stdio
```

## Usage

### Environment Variables

- `OPENAPI_SPEC_URL` (required) - URL or file path to OpenAPI spec
- `OPENAPI_CACHE_DIR` (optional) - Cache directory (default: `~/.openapi-mcp-cache`)

### Stdio Mode (for MCP clients)

```bash
# Docker
docker run -i --rm \
  -e OPENAPI_SPEC_URL=https://petstore3.swagger.io/api/v3/openapi.json \
  ghcr.io/sagenkoder/go-openapi-exploration-mcp-server:latest

# Binary
OPENAPI_SPEC_URL=https://petstore3.swagger.io/api/v3/openapi.json ./openapi-mcp-stdio
```

### HTTP Mode

```bash
# Docker
docker run -p 8080:8080 \
  -e OPENAPI_SPEC_URL=https://petstore3.swagger.io/api/v3/openapi.json \
  ghcr.io/sagenkoder/go-openapi-exploration-mcp-server:http

# Binary
OPENAPI_SPEC_URL=https://petstore3.swagger.io/api/v3/openapi.json ./openapi-mcp-http -addr :8080
```

### Interactive Mode

```bash
# Docker
docker run -it --rm \
  -e OPENAPI_SPEC_URL=https://petstore3.swagger.io/api/v3/openapi.json \
  ghcr.io/sagenkoder/go-openapi-exploration-mcp-server:interactive

# Binary
OPENAPI_SPEC_URL=https://petstore3.swagger.io/api/v3/openapi.json ./openapi-mcp-interactive
```

## Available Tools

The server provides these tools to LLMs:

1. **list_categories** - List API categories based on path segments
2. **list_endpoints** - List endpoints, optionally filtered by category
3. **show_endpoint** - Show detailed endpoint information including parameters and schemas
4. **get_spec_info** - Get general information about the API
5. **show_schema** - Inspect specific schema components

## Examples

### Local File

```bash
OPENAPI_SPEC_URL=/path/to/openapi.yaml ./openapi-mcp-stdio
```

### With Custom Cache

```bash
OPENAPI_CACHE_DIR=/tmp/api-cache \
OPENAPI_SPEC_URL=https://api.example.com/openapi.json \
./openapi-mcp-stdio
```

### Docker with Volume Mount

```bash
docker run -i --rm \
  -v $(pwd)/openapi.yaml:/openapi.yaml:ro \
  -e OPENAPI_SPEC_URL=/openapi.yaml \
  ghcr.io/sagenkoder/go-openapi-exploration-mcp-server:latest
```

## Development

### Project Structure

```
cmd/
├── openapi-mcp-stdio/       # MCP stdio mode
├── openapi-mcp-http/        # HTTP server mode
└── openapi-mcp-interactive/ # Interactive CLI mode

internal/
├── cache.go      # Caching logic
├── handlers.go   # MCP tool handlers
├── server.go     # Core server logic
└── utils.go      # Utilities
```

### Building Docker Images

```bash
# Build specific mode
docker build --build-arg MODE=stdio -t my-openapi-mcp:stdio .

# Build all modes
./build.sh docker
```

## License

MIT License - see [LICENSE](LICENSE) file for details.