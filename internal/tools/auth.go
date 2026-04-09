package tools

import (
	"context"

	"github.com/artschekoff/jira-mcp/internal/acli"
	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
)

const eventAuthStatus = "auth.status"

// AuthHandlers groups authentication-related tool handlers.
type AuthHandlers struct {
	baseHandler
}

// NewAuthHandlers creates an AuthHandlers.
func NewAuthHandlers(runner *acli.Runner, logger *zap.Logger) *AuthHandlers {
	return &AuthHandlers{baseHandler: newBaseHandler(runner, logger)}
}

// HandleAuthStatus reports the current acli authentication status.
// This is useful for verifying that acli is installed and authenticated before
// using other tools. No tokens or secrets are exposed in the output.
func (h *AuthHandlers) HandleAuthStatus(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return h.run(ctx, eventAuthStatus, []string{"jira", "auth", "status"}, nil)
}
