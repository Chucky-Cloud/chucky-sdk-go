// Package tools provides helpers for creating tools and MCP servers.
package tools

import (
	"context"

	"github.com/chucky-cloud/chucky-sdk-go/pkg/types"
)

// CreateToolOptions contains options for creating a tool.
type CreateToolOptions struct {
	Name        string
	Description string
	InputSchema types.ToolInputSchema
	ExecuteIn   types.ExecuteLocation
	Handler     types.ToolHandler
}

// CreateTool creates a new tool definition.
func CreateTool(opts CreateToolOptions) types.ToolDefinition {
	executeIn := opts.ExecuteIn
	if executeIn == "" {
		executeIn = types.ExecuteInServer
	}

	return types.ToolDefinition{
		Name:        opts.Name,
		Description: opts.Description,
		InputSchema: opts.InputSchema,
		ExecuteIn:   executeIn,
		Handler:     opts.Handler,
	}
}

// Tool is a shorthand for creating a tool with common options.
func Tool(name, description string, schema types.ToolInputSchema, handler types.ToolHandler) types.ToolDefinition {
	return CreateTool(CreateToolOptions{
		Name:        name,
		Description: description,
		InputSchema: schema,
		Handler:     handler,
	})
}

// BrowserTool creates a tool that executes in the browser (client-side).
func BrowserTool(name, description string, schema types.ToolInputSchema, handler types.ToolHandler) types.ToolDefinition {
	return CreateTool(CreateToolOptions{
		Name:        name,
		Description: description,
		InputSchema: schema,
		ExecuteIn:   types.ExecuteInBrowser,
		Handler:     handler,
	})
}

// ServerTool creates a tool that executes on the server.
func ServerTool(name, description string, schema types.ToolInputSchema, handler types.ToolHandler) types.ToolDefinition {
	return CreateTool(CreateToolOptions{
		Name:        name,
		Description: description,
		InputSchema: schema,
		ExecuteIn:   types.ExecuteInServer,
		Handler:     handler,
	})
}

// TextResult creates a successful text result.
func TextResult(text string) *types.ToolResult {
	return &types.ToolResult{
		Content: []any{
			types.TextToolContent{
				Type: "text",
				Text: text,
			},
		},
	}
}

// ErrorResult creates an error result.
func ErrorResult(message string) *types.ToolResult {
	return &types.ToolResult{
		Content: []any{
			types.TextToolContent{
				Type: "text",
				Text: message,
			},
		},
		IsError: true,
	}
}

// ImageResult creates an image result.
func ImageResult(base64Data, mimeType string) *types.ToolResult {
	return &types.ToolResult{
		Content: []any{
			types.ImageToolContent{
				Type:     "image",
				Data:     base64Data,
				MimeType: mimeType,
			},
		},
	}
}

// ResourceResult creates a resource result.
func ResourceResult(uri string, opts ...ResourceOption) *types.ToolResult {
	content := types.ResourceToolContent{
		Type: "resource",
		URI:  uri,
	}
	for _, opt := range opts {
		opt(&content)
	}
	return &types.ToolResult{
		Content: []any{content},
	}
}

// ResourceOption is a functional option for resource results.
type ResourceOption func(*types.ResourceToolContent)

// WithMimeType sets the MIME type for a resource.
func WithMimeType(mimeType string) ResourceOption {
	return func(c *types.ResourceToolContent) {
		c.MimeType = mimeType
	}
}

// WithText sets the text content for a resource.
func WithText(text string) ResourceOption {
	return func(c *types.ResourceToolContent) {
		c.Text = text
	}
}

// WithBlob sets the blob content for a resource.
func WithBlob(blob string) ResourceOption {
	return func(c *types.ResourceToolContent) {
		c.Blob = blob
	}
}

// SchemaBuilder helps build JSON schemas for tool inputs.
type SchemaBuilder struct {
	schema types.ToolInputSchema
}

// NewSchema creates a new schema builder.
func NewSchema() *SchemaBuilder {
	return &SchemaBuilder{
		schema: types.ToolInputSchema{
			Type:       "object",
			Properties: make(map[string]types.JsonSchemaProperty),
		},
	}
}

// Property adds a property to the schema.
func (b *SchemaBuilder) Property(name string, prop types.JsonSchemaProperty) *SchemaBuilder {
	b.schema.Properties[name] = prop
	return b
}

// String adds a string property.
func (b *SchemaBuilder) String(name, description string) *SchemaBuilder {
	return b.Property(name, types.JsonSchemaProperty{
		Type:        "string",
		Description: description,
	})
}

// Integer adds an integer property.
func (b *SchemaBuilder) Integer(name, description string) *SchemaBuilder {
	return b.Property(name, types.JsonSchemaProperty{
		Type:        "integer",
		Description: description,
	})
}

// Number adds a number property.
func (b *SchemaBuilder) Number(name, description string) *SchemaBuilder {
	return b.Property(name, types.JsonSchemaProperty{
		Type:        "number",
		Description: description,
	})
}

// Boolean adds a boolean property.
func (b *SchemaBuilder) Boolean(name, description string) *SchemaBuilder {
	return b.Property(name, types.JsonSchemaProperty{
		Type:        "boolean",
		Description: description,
	})
}

// Enum adds an enum property.
func (b *SchemaBuilder) Enum(name, description string, values ...any) *SchemaBuilder {
	return b.Property(name, types.JsonSchemaProperty{
		Type:        "string",
		Description: description,
		Enum:        values,
	})
}

// Array adds an array property.
func (b *SchemaBuilder) Array(name, description string, items types.JsonSchemaProperty) *SchemaBuilder {
	return b.Property(name, types.JsonSchemaProperty{
		Type:        "array",
		Description: description,
		Items:       &items,
	})
}

// Required marks properties as required.
func (b *SchemaBuilder) Required(names ...string) *SchemaBuilder {
	b.schema.Required = append(b.schema.Required, names...)
	return b
}

// Build returns the completed schema.
func (b *SchemaBuilder) Build() types.ToolInputSchema {
	return b.schema
}

// SimpleHandler wraps a simple function as a tool handler.
func SimpleHandler(fn func(input map[string]any) (string, error)) types.ToolHandler {
	return func(ctx context.Context, input map[string]any) (*types.ToolResult, error) {
		result, err := fn(input)
		if err != nil {
			return ErrorResult(err.Error()), nil
		}
		return TextResult(result), nil
	}
}
