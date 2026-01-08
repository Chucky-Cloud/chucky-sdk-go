// Package main demonstrates multi-turn conversations.
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
			AIDollars:    5.0,
			ComputeHours: 1.0,
			Window:       chucky.BudgetWindowDay,
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

	// Create a session
	session := client.CreateSession(&chucky.SessionOptions{
		BaseOptions: chucky.BaseOptions{
			Model: chucky.ModelClaudeSonnet,
		},
	})

	ctx := context.Background()

	// First turn
	fmt.Println("=== Turn 1 ===")
	if err := session.Send(ctx, "I'm thinking of a number between 1 and 10. It's 7."); err != nil {
		log.Fatalf("Failed to send: %v", err)
	}

	for msg := range session.Stream(ctx) {
		if result, ok := msg.(*chucky.SDKResultMessage); ok {
			fmt.Printf("Response: %s\n", result.Result)
		}
	}

	// Second turn
	fmt.Println("\n=== Turn 2 ===")
	if err := session.Send(ctx, "What number was I thinking of?"); err != nil {
		log.Fatalf("Failed to send: %v", err)
	}

	for msg := range session.Stream(ctx) {
		if result, ok := msg.(*chucky.SDKResultMessage); ok {
			fmt.Printf("Response: %s\n", result.Result)
		}
	}

	// Third turn
	fmt.Println("\n=== Turn 3 ===")
	if err := session.Send(ctx, "Multiply that number by 3"); err != nil {
		log.Fatalf("Failed to send: %v", err)
	}

	for msg := range session.Stream(ctx) {
		if result, ok := msg.(*chucky.SDKResultMessage); ok {
			fmt.Printf("Response: %s\n", result.Result)
			fmt.Printf("Total cost: $%.6f\n", result.TotalCostUsd)
		}
	}

	session.Close()
	fmt.Println("\nSession complete!")
}
