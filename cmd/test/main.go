// Quick test for the Chucky Go SDK
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
	fmt.Printf("Token (first 50 chars): %s...\n", token[:50])

	// Create client
	client := chucky.NewClient(chucky.ClientOptions{
		Token: token,
		Debug: true,
	})
	defer client.Close()

	// Create session - simple, no tools
	session := client.CreateSession(&chucky.SessionOptions{
		BaseOptions: chucky.BaseOptions{
			Model:    chucky.ModelClaudeSonnet,
			MaxTurns: 3,
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Send a simple message
	fmt.Println("\nSending message to Claude...")
	if err := session.Send(ctx, "What is 2 + 2? Reply with just the number."); err != nil {
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

	fmt.Println("\nTest complete!")
}
