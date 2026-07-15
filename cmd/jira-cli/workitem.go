package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	var jql, fields string
	var limit int

	cmd := &cobra.Command{
		Use:   "search --jql <query>",
		Short: "Search work items with JQL",
		Long: `Search Jira work items using JQL (Jira Query Language).

Output: JSON array of work item objects. Default fields per item:
  issuetype, key, assignee, priority, status, summary

Flags:
  --jql     JQL query string (required)
  --fields  comma-separated list of fields to return
  --limit   maximum number of results (default: server default)

Examples:
  jira-cli search --jql 'project = TEAM AND status = "In Progress"'
  jira-cli search --jql 'assignee = currentUser() ORDER BY updated DESC' --limit 20
  jira-cli search --jql 'project = TEAM AND type = Bug' --fields summary,status,assignee`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if jql == "" {
				return fmt.Errorf("flag --jql is required")
			}
			flags := []string{"--jql", jql, "--json"}
			if fields != "" {
				flags = append(flags, "--fields", fields)
			}
			if limit > 0 {
				flags = append(flags, "--limit", fmt.Sprintf("%d", limit))
			}
			return runCmd([]string{"jira", "workitem", "search"}, flags)
		},
	}
	cmd.Flags().StringVar(&jql, "jql", "", "JQL query string (required)")
	cmd.Flags().StringVar(&fields, "fields", "", "comma-separated list of fields to return")
	cmd.Flags().IntVar(&limit, "limit", 0, "maximum number of results")
	return cmd
}

func newViewCmd() *cobra.Command {
	var fields string

	cmd := &cobra.Command{
		Use:   "view <KEY>",
		Short: "View full details of a work item",
		Long: `View the full details of a Jira work item by its key.

Output: JSON object with work item fields. Default fields:
  key, issuetype, summary, status, assignee, description

Flags:
  --fields  comma-separated list of fields to return (use "*all" for everything)

Examples:
  jira-cli view PROJ-123
  jira-cli view PROJ-123 --fields summary,status,assignee,description,comments`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// acli takes the key as a positional arg for workitem view, but it
			// must sit outside the allowlisted subcommand prefix (which is
			// matched by exact join), so pass it in the args, not subCmdPath.
			flags := []string{args[0], "--json"}
			if fields != "" {
				flags = append(flags, "--fields", fields)
			}
			return runCmd([]string{"jira", "workitem", "view"}, flags)
		},
	}
	cmd.Flags().StringVar(&fields, "fields", "", `comma-separated fields to return; use "*all" for all`)
	return cmd
}

func newCreateCmd() *cobra.Command {
	var summary, project, issueType, description, assignee, labels, parent, customFields string

	cmd := &cobra.Command{
		Use:   "create --summary <s> --project <p> --type <t>",
		Short: "Create a new work item",
		Long: `Create a new Jira work item.

Output: JSON object of the created work item, including its new key.

Required flags:
  --summary   title of the work item
  --project   project key (e.g. TEAM)
  --type      issue type (e.g. Story, Bug, Task, Epic)

Optional flags:
  --description  plain-text description
  --assignee     username, account ID, or "@me" to self-assign
  --labels       comma-separated labels (e.g. "backend,urgent")
  --parent       parent issue key for sub-tasks (e.g. PROJ-10)
  --custom-fields JSON object of extra fields, e.g.:
                   '{"priority":{"name":"High"},"customfield_10016":5}'
                   Triggers --from-json path; any Jira field is reachable.

Examples:
  jira-cli create --summary "Fix login bug" --project TEAM --type Bug
  jira-cli create --summary "New feature" --project TEAM --type Story \
    --assignee "@me" --labels "backend,api"
  jira-cli create --summary "Epic work" --project TEAM --type Epic \
    --custom-fields '{"priority":{"name":"High"}}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if summary == "" || project == "" || issueType == "" {
				return fmt.Errorf("--summary, --project, and --type are required")
			}
			if customFields != "" {
				return createWithCustomFields(summary, project, issueType, description, assignee, labels, parent, customFields)
			}
			flags := []string{
				"--summary", summary,
				"--project", project,
				"--type", issueType,
				"--json",
			}
			if description != "" {
				flags = append(flags, "--description", description)
			}
			if assignee != "" {
				flags = append(flags, "--assignee", assignee)
			}
			if labels != "" {
				flags = append(flags, "--label", labels)
			}
			if parent != "" {
				flags = append(flags, "--parent", parent)
			}
			return runCmd([]string{"jira", "workitem", "create"}, flags)
		},
	}
	cmd.Flags().StringVar(&summary, "summary", "", "work item title (required)")
	cmd.Flags().StringVar(&project, "project", "", "project key, e.g. TEAM (required)")
	cmd.Flags().StringVar(&issueType, "type", "", "issue type, e.g. Story, Bug, Task (required)")
	cmd.Flags().StringVar(&description, "description", "", "plain-text description")
	cmd.Flags().StringVar(&assignee, "assignee", "", `username, account ID, or "@me"`)
	cmd.Flags().StringVar(&labels, "labels", "", "comma-separated labels")
	cmd.Flags().StringVar(&parent, "parent", "", "parent issue key (for sub-tasks)")
	cmd.Flags().StringVar(&customFields, "custom-fields", "", "JSON object of extra Jira field values")
	return cmd
}

// createWithCustomFields uses the acli --from-json path to set arbitrary fields.
func createWithCustomFields(summary, project, issueType, description, assignee, labels, parent, customFieldsJSON string) error {
	var additionalAttrs map[string]any
	if err := json.Unmarshal([]byte(customFieldsJSON), &additionalAttrs); err != nil {
		return fmt.Errorf("--custom-fields must be a valid JSON object: %w", err)
	}

	// acli next-gen epics: parent must live inside additionalAttributes as {"key":"..."}
	if parent != "" {
		if _, exists := additionalAttrs["parent"]; !exists {
			additionalAttrs["parent"] = map[string]string{"key": parent}
		}
	}

	payload := map[string]any{
		"summary":              summary,
		"projectKey":           project,
		"type":                 issueType,
		"additionalAttributes": additionalAttrs,
	}
	if labels != "" {
		parts := strings.Split(labels, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		payload["labels"] = parts
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("serializing work item: %w", err)
	}

	tmp, err := os.CreateTemp("", "jira-create-*.json")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("writing temp file: %w", err)
	}
	tmp.Close()

	flags := []string{"--from-json", tmp.Name(), "--json"}
	if description != "" {
		flags = append(flags, "--description", description)
	}
	if assignee != "" {
		flags = append(flags, "--assignee", assignee)
	}
	return runCmd([]string{"jira", "workitem", "create"}, flags)
}

func newEditCmd() *cobra.Command {
	var summary, description, assignee, labels, issueType string

	cmd := &cobra.Command{
		Use:   "edit <KEY>",
		Short: "Edit a work item",
		Long: `Edit one or more fields of a Jira work item. Only provided flags are changed.

KEY may be a comma-separated list to bulk-edit multiple items.

Output: JSON object(s) of the updated work item(s).

Examples:
  jira-cli edit PROJ-123 --summary "Updated title"
  jira-cli edit PROJ-123 --assignee jane --labels "backend,urgent"
  jira-cli edit PROJ-1,PROJ-2,PROJ-3 --type Task`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if summary == "" && description == "" && assignee == "" && labels == "" && issueType == "" {
				return fmt.Errorf("at least one of --summary, --description, --assignee, --labels, --type is required")
			}
			flags := []string{"--key", args[0], "--yes", "--json"}
			if summary != "" {
				flags = append(flags, "--summary", summary)
			}
			if description != "" {
				flags = append(flags, "--description", description)
			}
			if assignee != "" {
				flags = append(flags, "--assignee", assignee)
			}
			if labels != "" {
				flags = append(flags, "--labels", labels)
			}
			if issueType != "" {
				flags = append(flags, "--type", issueType)
			}
			return runCmd([]string{"jira", "workitem", "edit"}, flags)
		},
	}
	cmd.Flags().StringVar(&summary, "summary", "", "new summary/title")
	cmd.Flags().StringVar(&description, "description", "", "new plain-text description")
	cmd.Flags().StringVar(&assignee, "assignee", "", `new assignee username, account ID, or "@me"`)
	cmd.Flags().StringVar(&labels, "labels", "", "replace labels with this comma-separated list")
	cmd.Flags().StringVar(&issueType, "type", "", "new issue type")
	return cmd
}

func newTransitionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transition <KEY> --status <status>",
		Short: "Transition a work item to a new status",
		Long: `Move a Jira work item to a new workflow status.

KEY may be a comma-separated list to bulk-transition multiple items.

Output: JSON object(s) of the updated work item(s).

Examples:
  jira-cli transition PROJ-123 --status "In Progress"
  jira-cli transition PROJ-123 --status Done
  jira-cli transition PROJ-1,PROJ-2 --status "To Do"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			status, _ := cmd.Flags().GetString("status")
			if status == "" {
				return fmt.Errorf("--status is required")
			}
			return runCmd([]string{"jira", "workitem", "transition"}, []string{
				"--key", args[0], "--status", status, "--yes", "--json",
			})
		},
	}
	cmd.Flags().String("status", "", "target status name, e.g. \"In Progress\", \"Done\" (required)")
	return cmd
}

func newAssignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assign <KEY> --assignee <user>",
		Short: "Assign a work item to a user",
		Long: `Assign a Jira work item to a user.

KEY may be a comma-separated list to bulk-assign.

Output: JSON object(s) of the updated work item(s).

Examples:
  jira-cli assign PROJ-123 --assignee johndoe
  jira-cli assign PROJ-123 --assignee "@me"
  jira-cli assign PROJ-1,PROJ-2 --assignee jane@company.com`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			assignee, _ := cmd.Flags().GetString("assignee")
			if assignee == "" {
				return fmt.Errorf("--assignee is required")
			}
			return runCmd([]string{"jira", "workitem", "assign"}, []string{
				"--key", args[0], "--assignee", assignee, "--yes", "--json",
			})
		},
	}
	cmd.Flags().String("assignee", "", `username, account ID, "@me", or "default" (required)`)
	return cmd
}

func newCommentCmd() *cobra.Command {
	comment := &cobra.Command{
		Use:   "comment",
		Short: "Manage work item comments",
	}
	comment.AddCommand(newCommentAddCmd(), newCommentListCmd())
	return comment
}

func newCommentAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <KEY> --body <text>",
		Short: "Add a comment to a work item",
		Long: `Add a plain-text comment to a Jira work item.

Output: JSON object of the created comment:
  - id       string  comment ID
  - author   object  {displayName, emailAddress}
  - body     string  comment text
  - created  string  ISO 8601 timestamp

Example:
  jira-cli comment add PROJ-123 --body "Looking into this now"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, _ := cmd.Flags().GetString("body")
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			return runCmd([]string{"jira", "workitem", "comment", "create"}, []string{
				"--key", args[0], "--body", body, "--json",
			})
		},
	}
	cmd.Flags().String("body", "", "comment text (required)")
	return cmd
}

func newCommentListCmd() *cobra.Command {
	var limit int
	var order, startDate string
	var paginate bool

	cmd := &cobra.Command{
		Use:   "list <KEY>",
		Short: "List comments on a work item",
		Long: `List comments on a Jira work item.

Output: JSON array of comment objects. Each comment contains:
  - id        string  comment ID
  - author    object  {displayName, emailAddress}
  - body      string  comment text
  - created   string  ISO 8601 timestamp (e.g. "2024-01-15T10:30:00.000-0700")
  - updated   string  ISO 8601 timestamp

Flags:
  --limit       maximum number of comments to return
  --order       "asc" = oldest first (default), "desc" = newest first
  --paginate    fetch all pages from acli
  --start-date  return only comments on or after YYYY-MM-DD (post-processing filter)

Examples:
  jira-cli comment list PROJ-123
  jira-cli comment list PROJ-123 --order desc --limit 10
  jira-cli comment list PROJ-123 --start-date 2024-06-01`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := []string{"--key", args[0], "--json"}
			if limit > 0 {
				flags = append(flags, "--limit", fmt.Sprintf("%d", limit))
			}
			if order != "" {
				flags = append(flags, "--order", order)
			}
			if paginate {
				flags = append(flags, "--paginate")
			}

			if startDate == "" {
				return runCmd([]string{"jira", "workitem", "comment", "list"}, flags)
			}

			sd, err := time.Parse("2006-01-02", startDate)
			if err != nil {
				return fmt.Errorf("invalid --start-date %q: expected YYYY-MM-DD", startDate)
			}
			out, err := runner.Run(context.Background(), []string{"jira", "workitem", "comment", "list"}, flags)
			if err != nil {
				return err
			}
			filtered, err := filterCommentsByDate(out, sd)
			if err != nil {
				return err
			}
			if filtered != "" {
				fmt.Println(filtered)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 0, "maximum number of comments")
	cmd.Flags().StringVar(&order, "order", "", `sort order: "asc" (default) or "desc"`)
	cmd.Flags().BoolVar(&paginate, "paginate", false, "fetch all pages of comments")
	cmd.Flags().StringVar(&startDate, "start-date", "", "only return comments on/after YYYY-MM-DD")
	return cmd
}

var commentDateFormats = []string{
	"2006-01-02T15:04:05.000-0700",
	"2006-01-02T15:04:05.000Z0700",
	time.RFC3339,
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05.000Z",
	"2006-01-02T15:04:05-0700",
}

func filterCommentsByDate(raw string, since time.Time) (string, error) {
	var comments []map[string]any
	if err := json.Unmarshal([]byte(raw), &comments); err != nil {
		return "", fmt.Errorf("parsing comment JSON: %w", err)
	}
	result := make([]map[string]any, 0, len(comments))
	for _, c := range comments {
		created, ok := c["created"].(string)
		if !ok {
			continue
		}
		var t time.Time
		for _, layout := range commentDateFormats {
			if parsed, err := time.Parse(layout, strings.TrimSpace(created)); err == nil {
				t = parsed
				break
			}
		}
		if !t.IsZero() && !t.Before(since) {
			result = append(result, c)
		}
	}
	out, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshaling filtered comments: %w", err)
	}
	return string(out), nil
}
