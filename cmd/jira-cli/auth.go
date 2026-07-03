package main

import "github.com/spf13/cobra"

func newAuthCmd() *cobra.Command {
	auth := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
	}
	auth.AddCommand(
		newAuthStatusCmd(),
		newAuthLoginCmd(),
		newAuthLogoutCmd(),
	)
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

func newAuthLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Log in to Jira via acli (interactive)",
		Long: `Wrapper around "acli jira auth login". Prompts for site, email, and API token.

The child process is attached to your terminal — output is not JSON.

Example:
  jira-cli auth login`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCmdInteractive([]string{"jira", "auth", "login"}, nil)
		},
	}
}

func newAuthLogoutCmd() *cobra.Command {
	// ponytail: same path as login, one helper covers both
	return &cobra.Command{
		Use:   "logout",
		Short: "Log out of Jira via acli",
		Long: `Wrapper around "acli jira auth logout". Clears the local acli credentials.

Example:
  jira-cli auth logout`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCmdInteractive([]string{"jira", "auth", "logout"}, nil)
		},
	}
}
