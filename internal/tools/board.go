package tools

import (
	"context"
	"fmt"

	"github.com/artschekoff/jira-mcp/internal/acli"
	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
)

const (
	eventBoardSearch      = "board.search"
	eventBoardListSprints = "board.list_sprints"
)

// BoardHandlers groups board-related tool handlers.
type BoardHandlers struct {
	baseHandler
}

// NewBoardHandlers creates a BoardHandlers.
func NewBoardHandlers(runner *acli.Runner, logger *zap.Logger) *BoardHandlers {
	return &BoardHandlers{baseHandler: newBaseHandler(runner, logger)}
}

// HandleBoardSearch searches for Jira boards.
func (h *BoardHandlers) HandleBoardSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	flags := []string{"--json"}

	if name := req.GetString("name", ""); name != "" {
		flags = append(flags, "--name", name)
	}
	if project := req.GetString("project", ""); project != "" {
		flags = append(flags, "--project", project)
	}
	if boardType := req.GetString("type", ""); boardType != "" {
		flags = append(flags, "--type", boardType)
	}
	if limit := req.GetInt("limit", 0); limit > 0 {
		flags = append(flags, "--limit", fmt.Sprintf("%d", limit))
	}

	return h.run(ctx, eventBoardSearch, []string{"jira", "board", "search"}, flags)
}

// HandleBoardListSprints lists all sprints for a given board.
func (h *BoardHandlers) HandleBoardListSprints(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if r := requireParams(req, "id"); r != nil {
		return r, nil
	}

	flags := []string{"--id", req.GetString("id", ""), "--json"}

	if state := req.GetString("state", ""); state != "" {
		flags = append(flags, "--state", state)
	}
	if limit := req.GetInt("limit", 0); limit > 0 {
		flags = append(flags, "--limit", fmt.Sprintf("%d", limit))
	}

	return h.run(ctx, eventBoardListSprints, []string{"jira", "board", "list-sprints"}, flags)
}
