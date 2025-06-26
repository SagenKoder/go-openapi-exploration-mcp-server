package internal

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func RunInteractiveMode(oas *OpenAPIServer) {
	scanner := bufio.NewScanner(os.Stdin)
	ctx := context.Background()

	for {
		displayMenu()

		if !scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(scanner.Text())
		if choice == "6" {
			fmt.Println("Exiting...")
			return
		}

		handleUserChoice(ctx, choice, oas, scanner)
	}
}

func displayMenu() {
	fmt.Println("\n=== OpenAPI Tools ===")
	fmt.Println("1. List Categories")
	fmt.Println("2. List Endpoints")
	fmt.Println("3. Show Endpoint Details")
	fmt.Println("4. Get Spec Info")
	fmt.Println("5. Show Schema Details")
	fmt.Println("6. Exit")
	fmt.Print("\nSelect a tool (1-6): ")
}

func handleUserChoice(ctx context.Context, choice string, oas *OpenAPIServer, scanner *bufio.Scanner) {
	switch choice {
	case "1":
		result, err := oas.listCategoriesHandler(ctx, mcp.CallToolRequest{})
		printResult(result, err)

	case "2":
		fmt.Print("Enter category (or press Enter for all): ")
		scanner.Scan()
		category := strings.TrimSpace(scanner.Text())

		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "list_endpoints",
			},
		}
		if category != "" {
			req.Params.Arguments = map[string]interface{}{"category": category}
		}

		result, err := oas.listEndpointsHandler(ctx, req)
		printResult(result, err)

	case "3":
		fmt.Print("Enter path (e.g., /users/{id}): ")
		scanner.Scan()
		path := strings.TrimSpace(scanner.Text())

		fmt.Print("Enter method (GET, POST, PUT, DELETE, etc.): ")
		scanner.Scan()
		method := strings.TrimSpace(scanner.Text())

		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "show_endpoint",
				Arguments: map[string]interface{}{
					"path":   path,
					"method": method,
				},
			},
		}

		result, err := oas.showEndpointHandler(ctx, req)
		printResult(result, err)

	case "4":
		result, err := oas.getSpecInfoHandler(ctx, mcp.CallToolRequest{})
		printResult(result, err)

	case "5":
		fmt.Print("Enter schema reference (e.g., #/components/schemas/User): ")
		scanner.Scan()
		ref := strings.TrimSpace(scanner.Text())

		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "show_schema",
				Arguments: map[string]interface{}{
					"ref": ref,
				},
			},
		}

		result, err := oas.showSchemaHandler(ctx, req)
		printResult(result, err)

	default:
		fmt.Println("Invalid choice. Please select 1-6.")
	}
}

func printResult(result *mcp.CallToolResult, err error) {
	if err != nil {
		fmt.Printf("\nError: %v\n", err)
		return
	}

	if result == nil {
		fmt.Println("\nNo result returned")
		return
	}

	for _, content := range result.Content {
		if textContent, ok := mcp.AsTextContent(content); ok {
			// Try to pretty print JSON
			var data interface{}
			if err := json.Unmarshal([]byte(textContent.Text), &data); err == nil {
				prettyJSON, _ := json.MarshalIndent(data, "", "  ")
				fmt.Printf("\n%s\n", string(prettyJSON))
			} else {
				fmt.Printf("\n%s\n", textContent.Text)
			}
		} else {
			fmt.Printf("\nUnknown content type\n")
		}
	}
}
