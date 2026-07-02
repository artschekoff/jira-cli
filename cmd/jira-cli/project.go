package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newProjectCmd() *cobra.Command {
	project := &cobra.Command{
		Use:   "project",
		Short: "Manage Jira projects",
	}
	project.AddCommand(newProjectListCmd(), newProjectViewCmd())
	return project
}

func newProjectListCmd() *cobra.Command {
	var limit int
	var recent bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List visible Jira projects",
		Long: `List Jira projects visible to the authenticated user.

Output: JSON array of project objects. Each item contains:
  - key   string  project key (e.g. TEAM)
  - name  string  project name
  - id    string  project ID

Flags:
  --limit   maximum number of projects to return (default: ~30)
  --recent  return only recently accessed projects (up to 20)

Examples:
  jira-cli project list
  jira-cli project list --limit 50
  jira-cli project list --recent`,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := []string{"--json"}
			if limit > 0 {
				flags = append(flags, "--limit", fmt.Sprintf("%d", limit))
			}
			if recent {
				flags = append(flags, "--recent")
			}
			return runCmd([]string{"jira", "project", "list"}, flags)
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 0, "maximum number of projects")
	cmd.Flags().BoolVar(&recent, "recent", false, "return only recently accessed projects")
	return cmd
}

func newProjectViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view <KEY>",
		Short: "View project details",
		Long: `View details of a Jira project by its key.

Output: JSON object with project fields:
  - key          string  project key
  - name         string  project name
  - description  string  project description
  - lead         object  project lead {displayName, emailAddress}

Example:
  jira-cli project view TEAM`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCmd([]string{"jira", "project", "view"}, []string{"--key", args[0], "--json"})
		},
	}
}
