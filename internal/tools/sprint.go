package tools

import (
	"context"
	"fmt"

	"github.com/artschekoff/jira-mcp/internal/acli"
	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
)

const (
	eventSprintView          = "sprint.view"
	eventSprintListWorkitems = "sprint.list_workitems"
	eventSprintCreate        = "sprint.create"
)

// SprintHandlers groups sprint-related tool handlers.
type SprintHandlers struct {
	baseHandler
}

// NewSprintHandlers creates a SprintHandlers.
func NewSprintHandlers(runner *acli.Runner, logger *zap.Logger) *SprintHandlers {
	return &SprintHandlers{baseHandler: newBaseHandler(runner, logger)}
}

// HandleSprintView retrieves details of a Jira sprint by ID.
func (h *SprintHandlers) HandleSprintView(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if r := requireParams(req, "id"); r != nil {
		return r, nil
	}

	flags := []string{"--id", req.GetString("id", ""), "--json"}

	return h.run(ctx, eventSprintView, []string{"jira", "sprint", "view"}, flags)
}

// HandleSprintListWorkitems lists all work items in a given sprint.
func (h *SprintHandlers) HandleSprintListWorkitems(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if r := requireParams(req, "sprint", "board"); r != nil {
		return r, nil
	}

	flags := []string{
		"--sprint", fmt.Sprintf("%d", req.GetInt("sprint", 0)),
		"--board", fmt.Sprintf("%d", req.GetInt("board", 0)),
		"--json",
	}

	if jql := req.GetString("jql", ""); jql != "" {
		flags = append(flags, "--jql", jql)
	}
	if limit := req.GetInt("limit", 0); limit > 0 {
		flags = append(flags, "--limit", fmt.Sprintf("%d", limit))
	}
	if fields := req.GetString("fields", ""); fields != "" {
		flags = append(flags, "--fields", fields)
	}

	return h.run(ctx, eventSprintListWorkitems, []string{"jira", "sprint", "list-workitems"}, flags)
}

// HandleSprintCreate creates a new sprint on a board.
func (h *SprintHandlers) HandleSprintCreate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if r := requireParams(req, "name"); r != nil {
		return r, nil
	}
	if req.GetInt("board", 0) == 0 {
		return mcp.NewToolResultError(`parameter "board" is required`), nil
	}

	flags := []string{
		"--name", req.GetString("name", ""),
		"--board", fmt.Sprintf("%d", req.GetInt("board", 0)),
		"--json",
	}

	if start := req.GetString("start", ""); start != "" {
		flags = append(flags, "--start", start)
	}
	if end := req.GetString("end", ""); end != "" {
		flags = append(flags, "--end", end)
	}
	if goal := req.GetString("goal", ""); goal != "" {
		flags = append(flags, "--goal", goal)
	}

	return h.run(ctx, eventSprintCreate, []string{"jira", "sprint", "create"}, flags)
}
