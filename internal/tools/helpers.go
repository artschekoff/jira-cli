package tools

import (
	"context"
	"fmt"

	"github.com/artschekoff/jira-mcp/internal/acli"
	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
)

// Generic user-facing error messages. Internal details are logged, not exposed.
const (
	errMsgAcliUnavailable = "acli command failed; ensure acli is installed and authenticated via `acli jira auth login`"
	errMsgInternalError   = "an unexpected error occurred"
)

// baseHandler holds the dependencies shared by every tool handler group.
type baseHandler struct {
	runner *acli.Runner
	logger *zap.Logger
}

func newBaseHandler(runner *acli.Runner, logger *zap.Logger) baseHandler {
	return baseHandler{runner: runner, logger: logger}
}

// requireParams checks that each named string parameter is present and
// non-empty in the request. Returns the first missing-parameter error result,
// or nil when all are satisfied.
func requireParams(req mcp.CallToolRequest, names ...string) *mcp.CallToolResult {
	for _, name := range names {
		if req.GetString(name, "") == "" {
			return mcp.NewToolResultError(fmt.Sprintf("parameter %q is required", name))
		}
	}
	return nil
}

// run executes the acli subcommand and returns a tool result.
// On success the raw output (typically JSON) is returned as text.
// On error the details are logged and a safe message is returned to the caller.
func (b *baseHandler) run(ctx context.Context, event string, subCmdPath []string, flags []string) (*mcp.CallToolResult, error) {
	b.logger.Info(event, zap.Strings("cmd", subCmdPath))

	out, err := b.runner.Run(ctx, subCmdPath, flags)
	if err != nil {
		b.logger.Error(event+" failed", zap.Error(err))
		return mcp.NewToolResultError(errMsgAcliUnavailable + ": " + err.Error()), nil
	}
	if out == "" {
		return mcp.NewToolResultText("(no output)"), nil
	}
	return mcp.NewToolResultText(out), nil
}
