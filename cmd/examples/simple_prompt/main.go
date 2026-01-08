// Package main demonstrates a simple one-shot prompt.
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
			AIDollars:    1.0,
			ComputeHours: 0.5,
			Window:       chucky.BudgetWindowHour,
			WindowStart:  time.Now(),
		}),
	})
	if err != nil {
		log.Fatalf("Failed to create token: %v", err)
	}

	// Create client
	client := chucky.NewClient(chucky.ClientOptions{
		Token: token,
	})
	defer client.Close()

	// Send a one-shot prompt
	ctx := context.Background()
	result, err := client.Prompt(ctx, "What is 2 + 2? Reply with just the number.", &chucky.SessionOptions{
		BaseOptions: chucky.BaseOptions{
			Model:    chucky.ModelClaudeSonnet,
			MaxTurns: 1,
		},
	})

	if err != nil {
		log.Fatalf("Prompt failed: %v", err)
	}

	fmt.Printf("Result: %s\n", result.Result)
	fmt.Printf("Cost: $%.6f\n", result.TotalCostUsd)
}
