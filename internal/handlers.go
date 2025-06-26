package internal

import (
	"context"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mark3labs/mcp-go/mcp"
)

func (oas *OpenAPIServer) listCategoriesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract unique first path segments as categories
	categoriesMap := make(map[string]int)

	for path := range oas.spec.Paths.Map() {
		// Remove leading slash and get first segment
		trimmedPath := strings.TrimPrefix(path, "/")
		segments := strings.Split(trimmedPath, "/")
		if len(segments) > 0 && segments[0] != "" {
			// Extract the base category (remove any path parameters)
			category := strings.Split(segments[0], "{")[0]
			categoriesMap[category]++
		}
	}

	if len(categoriesMap) == 0 {
		return mcp.NewToolResultText("No categories found in the OpenAPI specification"), nil
	}

	// Convert to sorted slice
	categories := make([]map[string]interface{}, 0, len(categoriesMap))
	for name, count := range categoriesMap {
		categories = append(categories, map[string]interface{}{
			"name":           name,
			"endpoint_count": count,
		})
	}

	// Sort categories by name
	SortMapsByName(categories)

	return JSONResponse(categories)
}

func (oas *OpenAPIServer) listEndpointsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Category is optional
	categoryFilter := request.GetString("category", "")

	endpoints := []map[string]interface{}{}

	for path, pathItem := range oas.spec.Paths.Map() {
		// Check if we should include this path
		includeEndpoint := false

		if categoryFilter == "" {
			// No filter, include all endpoints
			includeEndpoint = true
		} else {
			// Check if path starts with the category
			trimmedPath := strings.TrimPrefix(path, "/")
			segments := strings.Split(trimmedPath, "/")

			if len(segments) > 0 && segments[0] != "" {
				// Extract the base category (remove any path parameters)
				pathCategory := strings.Split(segments[0], "{")[0]

				// Check if this path belongs to the requested category
				if strings.EqualFold(pathCategory, categoryFilter) {
					includeEndpoint = true
				}
			}
		}

		if includeEndpoint {
			for method, operation := range pathItem.Operations() {
				endpoint := map[string]interface{}{
					"path":    path,
					"method":  method,
					"summary": operation.Summary,
				}
				if operation.Description != "" {
					endpoint["description"] = operation.Description
				}
				if operation.OperationID != "" {
					endpoint["operationId"] = operation.OperationID
				}
				endpoints = append(endpoints, endpoint)
			}
		}
	}

	if len(endpoints) == 0 {
		if categoryFilter != "" {
			return mcp.NewToolResultText(fmt.Sprintf("No endpoints found for category: %s", categoryFilter)), nil
		}
		return mcp.NewToolResultText("No endpoints found in the OpenAPI specification"), nil
	}

	return JSONResponse(endpoints)
}

func (oas *OpenAPIServer) showEndpointHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := request.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	method, err := request.RequireString("method")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	pathItem := oas.spec.Paths.Find(path)
	if pathItem == nil {
		return mcp.NewToolResultError(fmt.Sprintf("Path not found: %s", path)), nil
	}

	operation := pathItem.GetOperation(strings.ToUpper(method))
	if operation == nil {
		return mcp.NewToolResultError(fmt.Sprintf("Method %s not found for path: %s", method, path)), nil
	}

	result := map[string]interface{}{
		"path":        path,
		"method":      strings.ToUpper(method),
		"summary":     operation.Summary,
		"description": operation.Description,
		"operationId": operation.OperationID,
		"tags":        operation.Tags,
	}

	// Add parameters
	if operation.Parameters != nil && len(operation.Parameters) > 0 {
		params := []map[string]interface{}{}
		for _, paramRef := range operation.Parameters {
			param := paramRef.Value
			paramInfo := map[string]interface{}{
				"name":        param.Name,
				"in":          param.In,
				"required":    param.Required,
				"description": param.Description,
			}
			if param.Schema != nil && param.Schema.Value != nil {
				paramInfo["type"] = param.Schema.Value.Type
				if param.Schema.Value.Format != "" {
					paramInfo["format"] = param.Schema.Value.Format
				}
			}
			params = append(params, paramInfo)
		}
		result["parameters"] = params
	}

	// Add request body
	if operation.RequestBody != nil && operation.RequestBody.Value != nil {
		reqBody := map[string]interface{}{
			"description": operation.RequestBody.Value.Description,
			"required":    operation.RequestBody.Value.Required,
		}
		if operation.RequestBody.Value.Content != nil {
			content := map[string]interface{}{}
			for mediaType, mediaTypeObj := range operation.RequestBody.Value.Content {
				if mediaTypeObj.Schema != nil {
					content[mediaType] = oas.schemaToMap(mediaTypeObj.Schema)
				}
			}
			reqBody["content"] = content
		}
		result["requestBody"] = reqBody
	}

	// Add responses
	if operation.Responses != nil {
		responses := map[string]interface{}{}
		for statusCode, responseRef := range operation.Responses.Map() {
			response := responseRef.Value
			respInfo := map[string]interface{}{
				"description": response.Description,
			}
			if response.Content != nil {
				content := map[string]interface{}{}
				for mediaType, mediaTypeObj := range response.Content {
					if mediaTypeObj.Schema != nil {
						content[mediaType] = oas.schemaToMap(mediaTypeObj.Schema)
					}
				}
				respInfo["content"] = content
			}
			responses[statusCode] = respInfo
		}
		result["responses"] = responses
	}

	return JSONResponse(result)
}

func (oas *OpenAPIServer) getSpecInfoHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	info := map[string]interface{}{
		"title":       oas.spec.Info.Title,
		"version":     oas.spec.Info.Version,
		"description": oas.spec.Info.Description,
	}

	if oas.spec.Info.Contact != nil {
		info["contact"] = map[string]string{
			"name":  oas.spec.Info.Contact.Name,
			"email": oas.spec.Info.Contact.Email,
			"url":   oas.spec.Info.Contact.URL,
		}
	}

	if oas.spec.Info.License != nil {
		info["license"] = map[string]string{
			"name": oas.spec.Info.License.Name,
			"url":  oas.spec.Info.License.URL,
		}
	}

	// Calculate statistics
	pathCount := 0
	operationCount := 0
	for _, pathItem := range oas.spec.Paths.Map() {
		pathCount++
		operationCount += len(pathItem.Operations())
	}

	info["stats"] = map[string]int{
		"paths":      pathCount,
		"operations": operationCount,
		"tags":       len(oas.spec.Tags),
	}

	if oas.spec.Servers != nil && len(oas.spec.Servers) > 0 {
		servers := []map[string]string{}
		for _, server := range oas.spec.Servers {
			servers = append(servers, map[string]string{
				"url":         server.URL,
				"description": server.Description,
			})
		}
		info["servers"] = servers
	}

	return JSONResponse(info)
}

func (oas *OpenAPIServer) showSchemaHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ref, err := request.RequireString("ref")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Parse the reference to extract the schema name
	// Expected format: #/components/schemas/SchemaName
	if !strings.HasPrefix(ref, "#/components/schemas/") {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid schema reference format. Expected: #/components/schemas/SchemaName, got: %s", ref)), nil
	}

	schemaName := strings.TrimPrefix(ref, "#/components/schemas/")

	// Look up the schema in the components
	if oas.spec.Components == nil || oas.spec.Components.Schemas == nil {
		return mcp.NewToolResultError("No schemas found in the OpenAPI specification"), nil
	}

	schemaRef, exists := oas.spec.Components.Schemas[schemaName]
	if !exists {
		return mcp.NewToolResultError(fmt.Sprintf("Schema not found: %s", schemaName)), nil
	}

	// Convert schema to detailed map
	result := map[string]interface{}{
		"name": schemaName,
		"ref":  ref,
	}

	if schemaRef.Ref != "" {
		result["schema"] = map[string]interface{}{
			"$ref": schemaRef.Ref,
		}
	} else if schemaRef.Value != nil {
		// For individual schema inspection, allow deeper expansion
		result["schema"] = oas.schemaToMapWithDepth(schemaRef, 0, DetailedSchemaMaxDepth)
	}

	return JSONResponse(result)
}

// Helper methods for schema conversion
func (oas *OpenAPIServer) schemaToMap(schemaRef *openapi3.SchemaRef) map[string]interface{} {
	return oas.schemaToMapWithDepth(schemaRef, 0, SchemaMaxDepth)
}

func (oas *OpenAPIServer) schemaToMapWithDepth(schemaRef *openapi3.SchemaRef, currentDepth, maxDepth int) map[string]interface{} {
	if schemaRef == nil {
		return nil
	}

	// If we have a reference, return it as-is instead of expanding
	if schemaRef.Ref != "" {
		return map[string]interface{}{
			"$ref": schemaRef.Ref,
		}
	}

	if schemaRef.Value == nil {
		return nil
	}

	schema := schemaRef.Value
	result := map[string]interface{}{}

	// Handle type which can be a slice
	if schema.Type != nil && len(schema.Type.Slice()) > 0 {
		if len(schema.Type.Slice()) == 1 {
			result["type"] = schema.Type.Slice()[0]
		} else {
			result["type"] = schema.Type.Slice()
		}
	}

	if schema.Format != "" {
		result["format"] = schema.Format
	}

	if schema.Description != "" {
		result["description"] = schema.Description
	}

	if len(schema.Required) > 0 {
		result["required"] = schema.Required
	}

	// Only expand properties if we haven't reached max depth
	if currentDepth < maxDepth {
		if schema.Properties != nil && len(schema.Properties) > 0 {
			properties := map[string]interface{}{}
			for propName, propSchema := range schema.Properties {
				properties[propName] = oas.schemaToMapWithDepth(propSchema, currentDepth+1, maxDepth)
			}
			result["properties"] = properties
		}

		if schema.Items != nil {
			result["items"] = oas.schemaToMapWithDepth(schema.Items, currentDepth+1, maxDepth)
		}
	} else {
		// At max depth, just indicate there's more
		if schema.Properties != nil && len(schema.Properties) > 0 {
			result["properties"] = fmt.Sprintf("[%d properties not expanded]", len(schema.Properties))
		}
		if schema.Items != nil {
			result["items"] = "[not expanded]"
		}
	}

	if len(schema.Enum) > 0 {
		result["enum"] = schema.Enum
	}

	return result
}
