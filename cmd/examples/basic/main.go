// Package main demonstrates basic usage of the Chucky SDK.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	chucky "github.com/chucky-cloud/chucky-sdk-go"
)

func main() {
	// Get configuration from environment
	url := os.Getenv("CHUCKY_URL")
	if url == "" {
		url = "wss://conjure.chucky.cloud/ws"
	}

	projectID := os.Getenv("CHUCKY_PROJECT_ID")
	secret := os.Getenv("CHUCKY_SECRET")

	if projectID == "" || secret == "" {
		log.Fatal("CHUCKY_PROJECT_ID and CHUCKY_SECRET environment variables are required")
	}

	// Create a budget token
	token, err := chucky.CreateToken(chucky.CreateTokenOptions{
		UserID:    "user-123",
		ProjectID: projectID,
		Secret:    secret,
		Budget: chucky.CreateBudget(chucky.CreateBudgetOptions{
			AIDollars:    10.0,
			ComputeHours: 1.0,
			Window:       chucky.BudgetWindowDay,
			WindowStart:  time.Now(),
		}),
	})
	if err != nil {
		log.Fatalf("Failed to create token: %v", err)
	}

	// Create a tool
	greetTool := chucky.Tool(
		"greet",
		"Greet someone by name",
		chucky.NewSchema().
			String("name", "The name of the person to greet").
			Enum("style", "The greeting style", "formal", "casual").
			Required("name").
			Build(),
		chucky.SimpleHandler(func(input map[string]any) (string, error) {
			name, _ := input["name"].(string)
			style, _ := input["style"].(string)
			if style == "" {
				style = "casual"
			}

			if style == "formal" {
				return fmt.Sprintf("Good day, %s. It is a pleasure to meet you.", name), nil
			}
			return fmt.Sprintf("Hey %s! What's up?", name), nil
		}),
	)

	// Create client
	client := chucky.NewClient(chucky.ClientOptions{
		BaseURL: url,
		Token:   token,
		Debug:   true,
	})
	defer client.Close()

	// Create session with tool
	session := client.CreateSession(&chucky.SessionOptions{
		BaseOptions: chucky.BaseOptions{
			Model: chucky.ModelClaudeSonnet,
			McpServers: []chucky.McpServerDefinition{
				chucky.CreateSdkMcpServer("my-tools", greetTool),
			},
		},
	})

	ctx := context.Background()

	// Send a message
	fmt.Println("Sending message...")
	if err := session.Send(ctx, "Please use the greet tool to greet Alice with a formal style"); err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	// Stream responses
	fmt.Println("Streaming responses...")
	for msg := range session.Stream(ctx) {
		switch m := msg.(type) {
		case *chucky.SDKAssistantMessage:
			fmt.Printf("[Assistant] %s\n", chucky.GetAssistantText(m))
		case *chucky.SDKResultMessage:
			fmt.Printf("[Result] %s\n", m.Result)
			fmt.Printf("  Cost: $%.6f\n", m.TotalCostUsd)
			fmt.Printf("  Turns: %d\n", m.NumTurns)
			fmt.Printf("  Duration: %dms\n", m.DurationMs)
		case *chucky.SDKSystemMessage:
			fmt.Printf("[System] Subtype: %s\n", m.Subtype)
		}
	}

	fmt.Println("Done!")
}
