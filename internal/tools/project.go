package tools

import (
	"context"
	"fmt"

	"github.com/artschekoff/jira-mcp/internal/acli"
	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
)

const (
	eventProjectList = "project.list"
	eventProjectView = "project.view"
)

// ProjectHandlers groups project-related tool handlers.
type ProjectHandlers struct {
	baseHandler
}

// NewProjectHandlers creates a ProjectHandlers.
func NewProjectHandlers(runner *acli.Runner, logger *zap.Logger) *ProjectHandlers {
	return &ProjectHandlers{baseHandler: newBaseHandler(runner, logger)}
}

// HandleProjectList lists Jira projects visible to the authenticated user.
func (h *ProjectHandlers) HandleProjectList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	flags := []string{"--json"}

	if limit := req.GetInt("limit", 0); limit > 0 {
		flags = append(flags, "--limit", fmt.Sprintf("%d", limit))
	}
	if req.GetBool("recent", false) {
		flags = append(flags, "--recent")
	}

	return h.run(ctx, eventProjectList, []string{"jira", "project", "list"}, flags)
}

// HandleProjectView retrieves details of a Jira project by its key.
func (h *ProjectHandlers) HandleProjectView(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if r := requireParams(req, "key"); r != nil {
		return r, nil
	}

	flags := []string{"--key", req.GetString("key", ""), "--json"}

	return h.run(ctx, eventProjectView, []string{"jira", "project", "view"}, flags)
}
