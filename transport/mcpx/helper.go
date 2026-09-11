// Package mcpx provides helpers for building MCP (Model Context Protocol) servers and
// handlers.
package mcpx

import (
	"encoding/json"
	"fmt"
)

// ToMcpText formats data as an indented JSON block under a Markdown heading labeled
// label, suitable for returning as MCP tool output text.
func ToMcpText(label string, data any) string {
	jsonData, _ := json.MarshalIndent(data, "", "  ")

	return fmt.Sprintf("### %s\n%s", label, string(jsonData))
}
