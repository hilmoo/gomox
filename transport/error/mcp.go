package errort

import (
	"context"
	"fmt"
	"strings"

	mlog "github.com/hilmoo/gomox/transport/middleware/log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ory/herodot"
)

// McpError logs err's herodot attributes onto ctx and formats it as an MCP tool error
// result with a human-readable summary of the error's ID, status, reason, and details.
func McpError(ctx context.Context, err *herodot.DefaultError) (*mcp.CallToolResult, any, error) {
	mlog.AddHerodotErrorAttributes(ctx, err)

	var sb strings.Builder
	fmt.Fprintf(&sb, "Error ID: %s\n", err.IDField)
	fmt.Fprintf(&sb, "Status: %s (Code: %d)\n", err.StatusField, err.CodeField)
	fmt.Fprintf(&sb, "Reason: %s\n", err.ReasonField)

	if len(err.DetailsField) > 0 {
		sb.WriteString("Details:\n")
		for key, value := range err.DetailsField {
			fmt.Fprintf(&sb, "- %s: %v\n", key, value)
		}
	}

	finalText := sb.String()

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: finalText,
			},
		},
		IsError: true,
	}, nil, nil
}
