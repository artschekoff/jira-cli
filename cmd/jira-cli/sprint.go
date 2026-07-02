package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSprintCmd() *cobra.Command {
	sprint := &cobra.Command{
		Use:   "sprint",
		Short: "Manage Jira sprints",
	}
	sprint.AddCommand(newSprintViewCmd(), newSprintListWorkitemsCmd(), newSprintCreateCmd())
	return sprint
}

func newSprintViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view <ID>",
		Short: "View sprint details by ID",
		Long: `View details of a Jira sprint by its numeric ID.

Output: JSON object with sprint fields:
  - id         int     sprint ID
  - name       string  sprint name
  - state      string  "future", "active", or "closed"
  - startDate  string  ISO 8601 start date
  - endDate    string  ISO 8601 end date
  - goal       string  sprint goal text

Example:
  jira-cli sprint view 42`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCmd([]string{"jira", "sprint", "view"}, []string{"--id", args[0], "--json"})
		},
	}
}

func newSprintListWorkitemsCmd() *cobra.Command {
	var sprint, board, limit int
	var jql, fields string

	cmd := &cobra.Command{
		Use:   "list-workitems --sprint <id> --board <id>",
		Short: "List work items in a sprint",
		Long: `List all work items in a given sprint.

Output: JSON array of work item objects. Default fields per item:
  key, issuetype, summary, assignee, priority, status

Required flags:
  --sprint  sprint ID (integer)
  --board   board ID (integer)

Optional flags:
  --jql     additional JQL filter to narrow results within the sprint
  --limit   maximum number of results
  --fields  comma-separated list of fields to return

Examples:
  jira-cli sprint list-workitems --sprint 42 --board 7
  jira-cli sprint list-workitems --sprint 42 --board 7 --fields summary,status,assignee
  jira-cli sprint list-workitems --sprint 42 --board 7 --jql 'assignee = currentUser()'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if sprint == 0 || board == 0 {
				return fmt.Errorf("--sprint and --board are required")
			}
			flags := []string{
				"--sprint", fmt.Sprintf("%d", sprint),
				"--board", fmt.Sprintf("%d", board),
				"--json",
			}
			if jql != "" {
				flags = append(flags, "--jql", jql)
			}
			if limit > 0 {
				flags = append(flags, "--limit", fmt.Sprintf("%d", limit))
			}
			if fields != "" {
				flags = append(flags, "--fields", fields)
			}
			return runCmd([]string{"jira", "sprint", "list-workitems"}, flags)
		},
	}
	cmd.Flags().IntVar(&sprint, "sprint", 0, "sprint ID (required)")
	cmd.Flags().IntVar(&board, "board", 0, "board ID (required)")
	cmd.Flags().StringVar(&jql, "jql", "", "additional JQL filter")
	cmd.Flags().IntVar(&limit, "limit", 0, "maximum number of results")
	cmd.Flags().StringVar(&fields, "fields", "", "comma-separated fields to return")
	return cmd
}

func newSprintCreateCmd() *cobra.Command {
	var board int
	var name, start, end, goal string

	cmd := &cobra.Command{
		Use:   "create --name <name> --board <id>",
		Short: "Create a new sprint",
		Long: `Create a new sprint on a Jira board.

Output: JSON object of the created sprint.

Required flags:
  --name   sprint name
  --board  board ID (integer)

Optional flags:
  --start  start date (YYYY-MM-DD)
  --end    end date (YYYY-MM-DD)
  --goal   sprint goal description

Examples:
  jira-cli sprint create --name "Sprint 10" --board 7
  jira-cli sprint create --name "Sprint 10" --board 7 \
    --start 2024-01-15 --end 2024-01-29 --goal "Ship auth v2"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" || board == 0 {
				return fmt.Errorf("--name and --board are required")
			}
			flags := []string{
				"--name", name,
				"--board", fmt.Sprintf("%d", board),
				"--json",
			}
			if start != "" {
				flags = append(flags, "--start", start)
			}
			if end != "" {
				flags = append(flags, "--end", end)
			}
			if goal != "" {
				flags = append(flags, "--goal", goal)
			}
			return runCmd([]string{"jira", "sprint", "create"}, flags)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "sprint name (required)")
	cmd.Flags().IntVar(&board, "board", 0, "board ID (required)")
	cmd.Flags().StringVar(&start, "start", "", "start date YYYY-MM-DD")
	cmd.Flags().StringVar(&end, "end", "", "end date YYYY-MM-DD")
	cmd.Flags().StringVar(&goal, "goal", "", "sprint goal text")
	return cmd
}
