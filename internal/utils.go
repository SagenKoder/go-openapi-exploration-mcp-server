package internal

import (
	"encoding/json"
	"sort"

	"github.com/mark3labs/mcp-go/mcp"
)

// SortMapsByName sorts a slice of maps by their "name" field
func SortMapsByName(items []map[string]interface{}) {
	sort.Slice(items, func(i, j int) bool {
		nameI, _ := items[i]["name"].(string)
		nameJ, _ := items[j]["name"].(string)
		return nameI < nameJ
	})
}

// JSONResponse creates a standard MCP JSON response
func JSONResponse(data interface{}) (*mcp.CallToolResult, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("Failed to marshal response"), nil
	}
	return mcp.NewToolResultText(string(jsonBytes)), nil
}
