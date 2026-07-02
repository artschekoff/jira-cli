package main

import (
	"context"
	"fmt"
	"os"

	"github.com/artschekoff/jira-cli/internal/acli"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// ponytail: package-level runner; CLI is single-command, no concurrency concern
var runner *acli.Runner

func main() {
	runner = acli.NewRunner(zap.NewNop())

	root := &cobra.Command{
		Use:   "jira-cli",
		Short: "Jira CLI — a thin wrapper around the Atlassian CLI (acli)",
		Long: `jira-cli wraps the Atlassian CLI (acli) and exposes Jira operations as
subcommands. All output is JSON printed to stdout. Errors go to stderr
with a non-zero exit code.

Prerequisites:
  1. Install acli: https://www.atlassian.com/software/acli
  2. Authenticate:  acli jira auth login
  3. Verify:        jira-cli auth status`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(
		newAuthCmd(),
		newSearchCmd(),
		newViewCmd(),
		newCreateCmd(),
		newEditCmd(),
		newTransitionCmd(),
		newAssignCmd(),
		newCommentCmd(),
		newSprintCmd(),
		newProjectCmd(),
		newBoardCmd(),
	)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "jira-cli:", err)
		os.Exit(1)
	}
}

// runCmd executes an acli subcommand and prints its output to stdout.
func runCmd(subCmd []string, flags []string) error {
	out, err := runner.Run(context.Background(), subCmd, flags)
	if err != nil {
		return err
	}
	if out != "" {
		fmt.Println(out)
	}
	return nil
}
