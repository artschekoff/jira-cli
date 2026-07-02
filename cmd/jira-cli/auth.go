package main

import "github.com/spf13/cobra"

func newAuthCmd() *cobra.Command {
	auth := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
	}
	auth.AddCommand(newAuthStatusCmd())
	return auth
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check acli authentication status",
		Long: `Check whether acli is installed and authenticated with Jira.

Output: plain text status message from acli (not JSON).

Example:
  jira-cli auth status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCmd([]string{"jira", "auth", "status"}, nil)
		},
	}
}
