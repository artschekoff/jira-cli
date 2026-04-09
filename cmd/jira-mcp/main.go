// Command jira-mcp is a Model Context Protocol server that wraps the Atlassian
// CLI (acli) and exposes Jira operations as MCP tools over stdio.
//
// Run as a stdio MCP server:
//
//	./jira-mcp
//
// Register in Cursor's mcp.json:
//
//	{
//	  "mcpServers": {
//	    "jira": {
//	      "type": "stdio",
//	      "command": "/path/to/bin/jira-mcp"
//	    }
//	  }
//	}
//
// Prerequisites: acli must be installed and authenticated via `acli jira auth login`.
package main

import (
	"fmt"
	"os"

	"github.com/artschekoff/jira-mcp/internal/acli"
	"github.com/artschekoff/jira-mcp/internal/tools"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	serverName    = "jira-mcp"
	serverVersion = "1.0.0"
)

func main() {
	root := &cobra.Command{
		Use:           "jira-mcp",
		Short:         "Jira MCP server wrapping the Atlassian CLI (acli)",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runMCP()
		},
	}

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "jira-mcp: %v\n", err)
		os.Exit(1)
	}
}

func runMCP() error {
	logger, err := buildLogger()
	if err != nil {
		return fmt.Errorf("initializing logger: %w", err)
	}
	defer logger.Sync() //nolint:errcheck

	runner := acli.NewRunner(logger)

	s := server.NewMCPServer(
		serverName,
		serverVersion,
		server.WithToolCapabilities(true),
	)

	tools.RegisterAll(s, runner, logger)

	logger.Info("jira-mcp starting", zap.String("version", serverVersion))

	if err := server.ServeStdio(s); err != nil {
		return fmt.Errorf("serving stdio: %w", err)
	}
	return nil
}

// buildLogger creates a zap production logger that writes only to stderr,
// keeping stdout free for MCP stdio JSON-RPC communication.
func buildLogger() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stderr"}
	cfg.ErrorOutputPaths = []string{"stderr"}

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("building zap logger: %w", err)
	}
	return logger, nil
}
