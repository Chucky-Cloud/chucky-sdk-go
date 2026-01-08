// Test Chucky Go SDK with tools
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	chucky "github.com/chucky-cloud/chucky-sdk-go"
)

func main() {
	projectID := "jd7d8hs32wvhr6m4hk31m1z4j17ypj64"
	secret := "hk_live_1e56ea40844e4a4b842f81b893247de9"

	// Create a budget token
	token, err := chucky.CreateToken(chucky.CreateTokenOptions{
		UserID:    "test-user",
		ProjectID: projectID,
		Secret:    secret,
		Budget: chucky.CreateBudget(chucky.CreateBudgetOptions{
			AIDollars:    1.0,
			ComputeHours: 0.5,
			Window:       chucky.BudgetWindowHour,
			WindowStart:  time.Now(),
		}),
	})
	if err != nil {
		log.Fatalf("Failed to create token: %v", err)
	}

	fmt.Println("Token created successfully!")

	// Create a simple calculator tool using the schema builder
	calculatorSchema := chucky.NewSchema().
		Enum("operation", "The operation to perform", "add", "subtract", "multiply", "divide").
		Number("a", "First operand").
		Number("b", "Second operand").
		Required("operation", "a", "b").
		Build()

	calculatorTool := chucky.Tool(
		"calculator",
		"Perform basic arithmetic calculations",
		calculatorSchema,
		func(ctx context.Context, input map[string]interface{}) (*chucky.ToolResult, error) {
			operation := input["operation"].(string)
			a := input["a"].(float64)
			b := input["b"].(float64)

			var result float64
			switch operation {
			case "add":
				result = a + b
			case "subtract":
				result = a - b
			case "multiply":
				result = a * b
			case "divide":
				if b == 0 {
					return chucky.ErrorResult("Cannot divide by zero"), nil
				}
				result = a / b
			default:
				return chucky.ErrorResult("Unknown operation: " + operation), nil
			}

			fmt.Printf("[Tool Called] calculator(%s, %.0f, %.0f) = %.2f\n", operation, a, b, result)
			return chucky.TextResult(fmt.Sprintf("Result: %.2f", result)), nil
		},
	)

	// Create MCP server with the tool
	mcpServer := chucky.NewMcpServer("calculator-server").
		Version("1.0.0").
		Add(calculatorTool).
		Build()

	// Create client
	client := chucky.NewClient(chucky.ClientOptions{
		Token: token,
		Debug: true,
	})
	defer client.Close()

	// Create session with tools
	session := client.CreateSession(&chucky.SessionOptions{
		BaseOptions: chucky.BaseOptions{
			Model:      chucky.ModelClaudeSonnet,
			MaxTurns:   5,
			McpServers: []chucky.McpServerDefinition{mcpServer},
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Send a message that requires using the tool
	fmt.Println("\nSending message to Claude...")
	if err := session.Send(ctx, "Use the calculator tool to compute 15 multiplied by 7. Reply with just the result."); err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	// Stream responses
	fmt.Println("Waiting for response...")
	for msg := range session.Stream(ctx) {
		switch m := msg.(type) {
		case *chucky.SDKAssistantMessage:
			text := chucky.GetAssistantText(m)
			if text != "" {
				fmt.Printf("[Assistant] %s\n", text)
			}
		case *chucky.SDKResultMessage:
			fmt.Printf("\n=== Result ===\n")
			fmt.Printf("Answer: %s\n", m.Result)
			fmt.Printf("Cost: $%.6f\n", m.TotalCostUsd)
			fmt.Printf("Turns: %d\n", m.NumTurns)
			fmt.Printf("Duration: %dms\n", m.DurationMs)
			if m.IsError {
				fmt.Printf("Errors: %v\n", m.Errors)
			}
		case *chucky.SDKSystemMessage:
			fmt.Printf("[System] %s\n", m.Subtype)
		case *chucky.ErrorEnvelope:
			fmt.Printf("[Error] %s (code: %s)\n", m.Payload.Message, m.Payload.Code)
		default:
			fmt.Printf("[Unknown] %T\n", m)
		}
	}

	fmt.Println("\nTest with tools complete!")
}
