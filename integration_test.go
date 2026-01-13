// Integration tests for Chucky SDK
//
// Requires environment variables:
// - CHUCKY_PROJECT_ID: Project ID from Chucky portal
// - CHUCKY_HMAC_KEY: HMAC key for the project

package chuckysdk_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	chucky "github.com/chucky-cloud/chucky-sdk-go"
)

func getTestToken(t *testing.T) string {
	projectID := os.Getenv("CHUCKY_PROJECT_ID")
	hmacKey := os.Getenv("CHUCKY_HMAC_KEY")

	if projectID == "" || hmacKey == "" {
		t.Skip("Missing CHUCKY_PROJECT_ID or CHUCKY_HMAC_KEY")
	}

	token, err := chucky.CreateToken(chucky.CreateTokenOptions{
		UserID:    "test-user",
		ProjectID: projectID,
		Secret:    hmacKey,
		ExpiresIn: time.Hour,
		Budget: chucky.CreateBudget(chucky.CreateBudgetOptions{
			AIDollars:    1.0,
			ComputeHours: 1.0,
			Window:       chucky.BudgetWindowDay,
			WindowStart:  time.Now(),
		}),
	})
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	return token
}

func TestTokenCreation(t *testing.T) {
	projectID := os.Getenv("CHUCKY_PROJECT_ID")
	hmacKey := os.Getenv("CHUCKY_HMAC_KEY")

	if projectID == "" || hmacKey == "" {
		t.Skip("Missing CHUCKY_PROJECT_ID or CHUCKY_HMAC_KEY")
	}

	token, err := chucky.CreateToken(chucky.CreateTokenOptions{
		UserID:    "test-user",
		ProjectID: projectID,
		Secret:    hmacKey,
		ExpiresIn: time.Hour,
		Budget: chucky.CreateBudget(chucky.CreateBudgetOptions{
			AIDollars:    1.0,
			ComputeHours: 1.0,
			Window:       chucky.BudgetWindowDay,
			WindowStart:  time.Now(),
		}),
	})

	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// JWT should have 3 parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("Expected JWT with 3 parts, got %d", len(parts))
	}
}

func TestSimplePrompt(t *testing.T) {
	token := getTestToken(t)

	t.Logf("Token (first 50 chars): %s...", token[:min(50, len(token))])

	client := chucky.NewClient(chucky.ClientOptions{
		Token: token,
		Debug: true,
	})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	result, err := client.Prompt(ctx, `Say "hello test" and nothing else.`, &chucky.SessionOptions{
		BaseOptions: chucky.BaseOptions{
			Model: chucky.ModelClaudeSonnet,
		},
	})

	if err != nil {
		t.Fatalf("Prompt failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	resultText := strings.ToLower(result.Result)
	if !strings.Contains(resultText, "hello") {
		t.Errorf("Expected result to contain 'hello', got: %s", result.Result)
	}

	t.Logf("Result: %s", result.Result)
	t.Logf("Cost: $%.6f", result.TotalCostUsd)
}

func TestStructuredOutput(t *testing.T) {
	// Wait for previous test's session to be released
	time.Sleep(10 * time.Second)

	token := getTestToken(t)

	client := chucky.NewClient(chucky.ClientOptions{
		Token: token,
	})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	result, err := client.Prompt(ctx, `What is 2 + 2? Answer with just the number.`, &chucky.SessionOptions{
		BaseOptions: chucky.BaseOptions{
			Model: chucky.ModelClaudeSonnet,
		},
	})

	if err != nil {
		t.Fatalf("Prompt failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if !strings.Contains(result.Result, "4") {
		t.Errorf("Expected result to contain '4', got: %s", result.Result)
	}

	t.Logf("Result: %s", result.Result)
}

func TestMcpToolExecution(t *testing.T) {
	// Wait for previous test's session to be released
	time.Sleep(10 * time.Second)

	token := getTestToken(t)

	// Track tool calls
	toolWasCalled := false
	var toolInputA, toolInputB int

	// Create add tool with the correct handler signature
	addTool := chucky.Tool("add", "Add two numbers together",
		chucky.NewSchema().
			Integer("a", "First number").
			Integer("b", "Second number").
			Required("a", "b").
			Build(),
		func(ctx context.Context, input map[string]any) (*chucky.ToolResult, error) {
			toolWasCalled = true
			a := int(input["a"].(float64))
			b := int(input["b"].(float64))
			toolInputA = a
			toolInputB = b
			sum := a + b
			return chucky.TextResult(fmt.Sprintf("The sum of %d and %d is %d", a, b, sum)), nil
		},
	)

	// Create MCP server and add the tool
	mcpServer := chucky.NewMcpServer("calculator").
		Add(addTool).
		Build()

	client := chucky.NewClient(chucky.ClientOptions{
		Token: token,
	})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	result, err := client.Prompt(ctx, "Use the add tool to calculate 7 + 15. Report the result.", &chucky.SessionOptions{
		BaseOptions: chucky.BaseOptions{
			Model:      chucky.ModelClaudeSonnet,
			McpServers: []chucky.McpServerDefinition{mcpServer},
		},
	})

	if err != nil {
		t.Fatalf("Prompt failed: %v", err)
	}

	if !toolWasCalled {
		t.Error("Tool was not called")
	}

	if toolInputA != 7 || toolInputB != 15 {
		t.Errorf("Expected tool inputs a=7, b=15, got a=%d, b=%d", toolInputA, toolInputB)
	}

	if !strings.Contains(result.Result, "22") {
		t.Errorf("Expected result to contain '22', got: %s", result.Result)
	}

	t.Logf("Tool called with a=%d, b=%d", toolInputA, toolInputB)
	t.Logf("Result: %s", result.Result)
}
