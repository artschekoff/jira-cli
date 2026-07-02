package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newBoardCmd() *cobra.Command {
	board := &cobra.Command{
		Use:   "board",
		Short: "Manage Jira boards",
	}
	board.AddCommand(newBoardSearchCmd(), newBoardListSprintsCmd())
	return board
}

func newBoardSearchCmd() *cobra.Command {
	var name, project, boardType string
	var limit int

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for Jira boards",
		Long: `Search for Jira boards, optionally filtered by name, project, or type.

Output: JSON array of board objects. Each item contains:
  - id    int     board ID (use this with sprint commands)
  - name  string  board name
  - type  string  "scrum", "kanban", or "simple"

Flags:
  --name     filter by board name (substring match)
  --project  filter by project key (e.g. TEAM)
  --type     filter by board type: "scrum", "kanban", or "simple"
  --limit    maximum number of results (default: 50)

Examples:
  jira-cli board search
  jira-cli board search --project TEAM
  jira-cli board search --type scrum --limit 10`,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := []string{"--json"}
			if name != "" {
				flags = append(flags, "--name", name)
			}
			if project != "" {
				flags = append(flags, "--project", project)
			}
			if boardType != "" {
				flags = append(flags, "--type", boardType)
			}
			if limit > 0 {
				flags = append(flags, "--limit", fmt.Sprintf("%d", limit))
			}
			return runCmd([]string{"jira", "board", "search"}, flags)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "filter by board name")
	cmd.Flags().StringVar(&project, "project", "", "filter by project key")
	cmd.Flags().StringVar(&boardType, "type", "", `board type: "scrum", "kanban", or "simple"`)
	cmd.Flags().IntVar(&limit, "limit", 0, "maximum number of results")
	return cmd
}

func newBoardListSprintsCmd() *cobra.Command {
	var state string
	var limit int

	cmd := &cobra.Command{
		Use:   "list-sprints <ID>",
		Short: "List sprints for a board",
		Long: `List all sprints for a given board ID.

Output: JSON array of sprint objects. Each item contains:
  - id         int     sprint ID (use this with sprint commands)
  - name       string  sprint name
  - state      string  "future", "active", or "closed"
  - startDate  string  ISO 8601 start date
  - endDate    string  ISO 8601 end date
  - goal       string  sprint goal text

Flags:
  --state  filter by state: "future", "active", or "closed"
           comma-separated for multiple, e.g. "active,closed"
  --limit  maximum number of sprints to return (default: 50)

Examples:
  jira-cli board list-sprints 7
  jira-cli board list-sprints 7 --state active
  jira-cli board list-sprints 7 --state active,closed --limit 20`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := []string{"--id", args[0], "--json"}
			if state != "" {
				flags = append(flags, "--state", state)
			}
			if limit > 0 {
				flags = append(flags, "--limit", fmt.Sprintf("%d", limit))
			}
			return runCmd([]string{"jira", "board", "list-sprints"}, flags)
		},
	}
	cmd.Flags().StringVar(&state, "state", "", `sprint state: "future", "active", "closed"`)
	cmd.Flags().IntVar(&limit, "limit", 0, "maximum number of sprints")
	return cmd
}
