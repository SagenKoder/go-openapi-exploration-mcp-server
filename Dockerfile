# Build stage
FROM golang:1.23-alpine AS builder

# Build argument to specify which mode to build (stdio, http, or interactive)
ARG MODE=stdio

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application based on MODE argument
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/openapi-mcp-${MODE}

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -g '' appuser

# Copy the binary from builder
COPY --from=builder /app/app /usr/local/bin/openapi-mcp

# Create necessary directories
RUN mkdir -p /home/appuser/.openapi-mcp-cache && \
    chown -R appuser:appuser /home/appuser

# Switch to non-root user
USER appuser

# Default environment variables
ENV OPENAPI_CACHE_DIR=/home/appuser/.openapi-mcp-cache
# Default URL - override this with your own OpenAPI spec URL or file path
ENV OPENAPI_SPEC_URL=https://tripletex.no/v2/openapi.json

# Run the application
CMD ["/usr/local/bin/openapi-mcp"]