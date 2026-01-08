# Chucky SDK for Go

The official Go SDK for interacting with Chucky (Claude Code sandbox). This SDK provides a simple, idiomatic Go API for creating sessions with Claude, sending messages, and receiving responses including tool calls.

## Installation

```bash
go get github.com/chucky-cloud/chucky-sdk-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    chucky "github.com/anthropics/chucky-sdk-go"
)

func main() {
    // Create a budget token
    token, _ := chucky.CreateToken(chucky.CreateTokenOptions{
        UserID:    "user-123",
        ProjectID: "your-project-id",
        Secret:    "your-secret",
        Budget: chucky.CreateBudget(chucky.CreateBudgetOptions{
            AIDollars:    10.0,
            ComputeHours: 1.0,
            Window:       chucky.BudgetWindowDay,
            WindowStart:  time.Now(),
        }),
    })

    // Create client
    client := chucky.NewClient(chucky.ClientOptions{
        Token: token,
    })
    defer client.Close()

    // Send a prompt
    ctx := context.Background()
    result, err := client.Prompt(ctx, "What is 2 + 2?", nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result.Result)
}
```

## Features

- **Full API compatibility** with TypeScript and Python SDKs
- **MCP tools support** - Define tools that execute in your application
- **Multi-turn conversations** - Maintain session state across messages
- **Token-based authentication** - JWT tokens with budget control
- **WebSocket transport** - Efficient real-time communication

## API Reference

### Client

```go
// Create a new client
client := chucky.NewClient(chucky.ClientOptions{
    BaseURL:           "wss://conjure.chucky.cloud/ws", // Default
    Token:             token,
    Debug:             false,
    Timeout:           60 * time.Second,
    KeepAliveInterval: 5 * time.Minute,
})

// One-shot prompt
result, err := client.Prompt(ctx, "Hello!", &chucky.SessionOptions{
    BaseOptions: chucky.BaseOptions{
        Model: chucky.ModelClaudeSonnet,
    },
})

// Create a session for multi-turn
session := client.CreateSession(&chucky.SessionOptions{
    BaseOptions: chucky.BaseOptions{
        Model: chucky.ModelClaudeSonnet,
    },
})

// Resume an existing session
session := client.ResumeSession("session-id", nil)

// Close client
client.Close()
```

### Session

```go
// Send a message
err := session.Send(ctx, "Hello, Claude!")

// Stream responses
for msg := range session.Stream(ctx) {
    switch m := msg.(type) {
    case *chucky.SDKAssistantMessage:
        fmt.Println(chucky.GetAssistantText(m))
    case *chucky.SDKResultMessage:
        fmt.Println(m.Result)
    case *chucky.SDKSystemMessage:
        fmt.Println("System:", m.Subtype)
    }
}

// Close session
session.Close()
```

### Tools

```go
// Create a tool with schema builder
greetTool := chucky.Tool(
    "greet",
    "Greet someone by name",
    chucky.NewSchema().
        String("name", "The name of the person").
        Enum("style", "Greeting style", "formal", "casual").
        Required("name").
        Build(),
    chucky.SimpleHandler(func(input map[string]any) (string, error) {
        name := input["name"].(string)
        return fmt.Sprintf("Hello, %s!", name), nil
    }),
)

// Use with session
session := client.CreateSession(&chucky.SessionOptions{
    BaseOptions: chucky.BaseOptions{
        McpServers: []chucky.McpServerDefinition{
            chucky.CreateSdkMcpServer("my-tools", greetTool),
        },
    },
})

// Or use the MCP server builder
server := chucky.NewMcpServer("my-server").
    Version("1.0.0").
    Add(greetTool).
    AddTool(chucky.CreateToolOptions{
        Name:        "another_tool",
        Description: "Another tool",
        InputSchema: chucky.NewSchema().String("arg", "An argument").Build(),
        Handler:     myHandler,
    }).
    Build()
```

### Tool Results

```go
// Text result
return chucky.TextResult("Success!")

// Error result
return chucky.ErrorResult("Something went wrong")

// Image result
return chucky.ImageResult(base64Data, "image/png")

// Resource result
return chucky.ResourceResult("file:///path/to/file",
    chucky.WithMimeType("text/plain"),
    chucky.WithText("file contents"),
)
```

### Token Management

```go
// Create a token
token, err := chucky.CreateToken(chucky.CreateTokenOptions{
    UserID:    "user-123",
    ProjectID: "project-456",
    Secret:    "your-secret-key",
    Budget: chucky.TokenBudget{
        AI:          chucky.MicroDollars(10.0),  // $10 in microdollars
        Compute:     chucky.ComputeSeconds(1.0), // 1 hour in seconds
        Window:      chucky.BudgetWindowDay,
        WindowStart: time.Now().Format(time.RFC3339),
    },
    ExpiresIn: time.Hour,
})

// Or use the budget helper
token, err := chucky.CreateToken(chucky.CreateTokenOptions{
    UserID:    "user-123",
    ProjectID: "project-456",
    Secret:    "your-secret-key",
    Budget: chucky.CreateBudget(chucky.CreateBudgetOptions{
        AIDollars:    10.0,
        ComputeHours: 1.0,
        Window:       chucky.BudgetWindowDay,
        WindowStart:  time.Now(),
    }),
})

// Decode a token (without verification)
decoded, err := chucky.DecodeToken(token)
fmt.Println(decoded.Payload.Subject) // user-123

// Verify a token signature
valid, err := chucky.VerifyToken(token, "your-secret-key")

// Check if token is expired
expired, err := chucky.IsTokenExpired(token)
```

### Session Options

```go
&chucky.SessionOptions{
    BaseOptions: chucky.BaseOptions{
        // Model selection
        Model:         chucky.ModelClaudeSonnet,
        FallbackModel: "claude-sonnet-4-5-20250929",

        // Prompting
        SystemPrompt:      "You are a helpful assistant",
        MaxTurns:          10,
        MaxBudgetUsd:      5.0,
        MaxThinkingTokens: 1024,

        // Tools
        McpServers: []chucky.McpServerDefinition{...},

        // Behavior
        PermissionMode:         chucky.PermissionModeDefault,
        IncludePartialMessages: true,
        Env:                    map[string]string{"KEY": "value"},
    },

    // Session-specific
    SessionID:       "custom-session-id",
    ForkSession:     false,
    ResumeSessionAt: "",
    Continue:        false,
}
```

### MCP Server Types

```go
// Client-side tools (handlers run in your app)
server := chucky.CreateSdkMcpServer("my-tools", tool1, tool2)

// Stdio server (command execution)
server := chucky.StdioServer("my-server", "python", "-m", "my_mcp_server")

// Stdio server with environment
server := chucky.StdioServerWithEnv("my-server", "node", []string{"server.js"}, map[string]string{
    "API_KEY": "secret",
})

// SSE server
server := chucky.SSEServer("my-server", "https://api.example.com/sse", map[string]string{
    "Authorization": "Bearer token",
})

// HTTP server
server := chucky.HTTPServer("my-server", "https://api.example.com/mcp", map[string]string{
    "Authorization": "Bearer token",
})
```

### Error Handling

```go
result, err := client.Prompt(ctx, "Hello!", nil)
if err != nil {
    if chuckyErr, ok := err.(*chucky.ChuckyError); ok {
        switch chuckyErr.Code {
        case chucky.ErrCodeBudgetExceeded:
            // Handle budget exceeded
        case chucky.ErrCodeTimeout:
            // Handle timeout
        case chucky.ErrCodeAuthentication:
            // Handle auth error
        }
    }
}
```

## Available Models

```go
chucky.ModelClaudeSonnet // "claude-sonnet-4-5-20250929"
chucky.ModelClaudeOpus   // "claude-opus-4-5-20251101"
```

## Budget Windows

```go
chucky.BudgetWindowHour  // "hour"
chucky.BudgetWindowDay   // "day"
chucky.BudgetWindowWeek  // "week"
chucky.BudgetWindowMonth // "month"
```

## Permission Modes

```go
chucky.PermissionModeDefault           // Normal permission checks
chucky.PermissionModePlan              // Planning mode
chucky.PermissionModeBypassPermissions // Bypass all permission checks
```

## Examples

See the `cmd/examples` directory for complete examples:

- `basic/` - Full example with tools
- `simple_prompt/` - One-shot prompt
- `multi_turn/` - Multi-turn conversation

## License

MIT
