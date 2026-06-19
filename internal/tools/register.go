// Package tools registers all Jira MCP tools on the MCP server.
package tools

import (
	"github.com/artschekoff/jira-mcp/internal/acli"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// RegisterAll registers every Jira MCP tool on the provided server.
func RegisterAll(s *server.MCPServer, runner *acli.Runner, logger *zap.Logger) {
	wi := NewWorkitemHandlers(runner, logger)
	sp := NewSprintHandlers(runner, logger)
	pr := NewProjectHandlers(runner, logger)
	bo := NewBoardHandlers(runner, logger)
	au := NewAuthHandlers(runner, logger)

	// --- Auth ---

	s.AddTool(
		mcp.NewTool("jira_auth_status",
			mcp.WithDescription("Check the current Atlassian CLI authentication status. Use this first to verify acli is installed and authenticated before calling other tools."),
		),
		au.HandleAuthStatus,
	)

	// --- Work Items ---

	s.AddTool(
		mcp.NewTool("jira_search",
			mcp.WithDescription("Search for Jira work items using a JQL (Jira Query Language) query. Returns a JSON list of matching issues. Examples: 'project = TEAM AND status = \"In Progress\"', 'assignee = currentUser() ORDER BY updated DESC'."),
			mcp.WithString("jql",
				mcp.Required(),
				mcp.Description("JQL query string. E.g. 'project = TEAM AND type = Bug'."),
			),
			mcp.WithString("fields",
				mcp.Description("Comma-separated list of fields to include. Default: issuetype,key,assignee,priority,status,summary."),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of results to return."),
			),
		),
		wi.HandleSearch,
	)

	s.AddTool(
		mcp.NewTool("jira_view",
			mcp.WithDescription("Retrieve full details of a Jira work item by its key (e.g. PROJ-123). Returns JSON with all requested fields including summary, status, assignee, description, and comments."),
			mcp.WithString("key",
				mcp.Required(),
				mcp.Description("Work item key, e.g. 'PROJ-123'."),
			),
			mcp.WithString("fields",
				mcp.Description("Comma-separated list of fields to return. Use '*all' for all fields. Default: key,issuetype,summary,status,assignee,description."),
			),
		),
		wi.HandleView,
	)

	s.AddTool(
		mcp.NewTool("jira_create",
			mcp.WithDescription("Create a new Jira work item. Returns the created item's key and details as JSON."),
			mcp.WithString("summary",
				mcp.Required(),
				mcp.Description("Summary (title) for the new work item."),
			),
			mcp.WithString("project",
				mcp.Required(),
				mcp.Description("Project key, e.g. 'TEAM'."),
			),
			mcp.WithString("type",
				mcp.Required(),
				mcp.Description("Work item type: Task, Bug, Story, Epic, etc."),
			),
			mcp.WithString("description",
				mcp.Description("Optional description in plain text or Atlassian Document Format (ADF)."),
			),
			mcp.WithString("assignee",
				mcp.Description("Assignee email or account ID. Use '@me' to self-assign."),
			),
			mcp.WithString("labels",
				mcp.Description("Comma-separated list of labels, e.g. 'backend,urgent'."),
			),
			mcp.WithString("parent",
				mcp.Description("Parent work item key for sub-tasks, e.g. 'PROJ-10'."),
			),
			mcp.WithString("custom_fields",
				mcp.Description("JSON object of custom field key→value pairs to set at creation time, e.g. '{\"components\":[{\"name\":\"Backend\"}],\"priority\":{\"name\":\"High\"},\"customfield_10016\":5}'. Triggers the --from-json code path so any Jira field is reachable."),
			),
		),
		wi.HandleCreate,
	)

	s.AddTool(
		mcp.NewTool("jira_edit",
			mcp.WithDescription("Edit one or more Jira work items. Provide the work item key(s) and only the fields you want to change."),
			mcp.WithString("key",
				mcp.Required(),
				mcp.Description("Comma-separated list of work item keys, e.g. 'PROJ-1,PROJ-2'."),
			),
			mcp.WithString("summary",
				mcp.Description("New summary text."),
			),
			mcp.WithString("description",
				mcp.Description("New description in plain text or ADF."),
			),
			mcp.WithString("assignee",
				mcp.Description("New assignee email or account ID. Use '@me' to self-assign."),
			),
			mcp.WithString("labels",
				mcp.Description("Replace labels with this comma-separated list."),
			),
			mcp.WithString("type",
				mcp.Description("New work item type."),
			),
		),
		wi.HandleEdit,
	)

	s.AddTool(
		mcp.NewTool("jira_transition",
			mcp.WithDescription("Transition a Jira work item to a new status (e.g. move to 'In Progress', 'Done', 'To Do')."),
			mcp.WithString("key",
				mcp.Required(),
				mcp.Description("Comma-separated list of work item keys to transition, e.g. 'PROJ-1'."),
			),
			mcp.WithString("status",
				mcp.Required(),
				mcp.Description("Target status name, e.g. 'In Progress', 'Done', 'To Do'."),
			),
		),
		wi.HandleTransition,
	)

	s.AddTool(
		mcp.NewTool("jira_assign",
			mcp.WithDescription("Assign a Jira work item to a user."),
			mcp.WithString("key",
				mcp.Required(),
				mcp.Description("Comma-separated list of work item keys, e.g. 'PROJ-1'."),
			),
			mcp.WithString("assignee",
				mcp.Required(),
				mcp.Description("Assignee email or account ID. Use '@me' to self-assign, 'default' for the project default."),
			),
		),
		wi.HandleAssign,
	)

	s.AddTool(
		mcp.NewTool("jira_comment_add",
			mcp.WithDescription("Add a comment to a Jira work item."),
			mcp.WithString("key",
				mcp.Required(),
				mcp.Description("Work item key, e.g. 'PROJ-123'."),
			),
			mcp.WithString("body",
				mcp.Required(),
				mcp.Description("Comment text in plain text."),
			),
		),
		wi.HandleCommentAdd,
	)

	s.AddTool(
		mcp.NewTool("jira_comment_list",
			mcp.WithDescription("List comments on a Jira work item. Supports pagination, ordering, and date filtering."),
			mcp.WithString("key",
				mcp.Required(),
				mcp.Description("Work item key, e.g. 'PROJ-123'."),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of comments to return."),
			),
			mcp.WithString("order",
				mcp.Description("Sort order for comments: 'asc' (oldest first) or 'desc' (newest first)."),
			),
			mcp.WithBoolean("paginate",
				mcp.Description("Set to true to enable paginated output from acli."),
			),
			mcp.WithString("start_date",
				mcp.Description("ISO 8601 date (YYYY-MM-DD). Only return comments created on or after this date. Applied as a post-processing filter."),
			),
		),
		wi.HandleCommentList,
	)

	// --- Sprints ---

	s.AddTool(
		mcp.NewTool("jira_sprint_view",
			mcp.WithDescription("View details of a Jira sprint by its numeric ID."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("Sprint ID, e.g. '42'."),
			),
		),
		sp.HandleSprintView,
	)

	s.AddTool(
		mcp.NewTool("jira_sprint_list_workitems",
			mcp.WithDescription("List all work items in a given sprint. Both sprint ID and board ID are required."),
			mcp.WithNumber("sprint",
				mcp.Required(),
				mcp.Description("Numeric sprint ID."),
			),
			mcp.WithNumber("board",
				mcp.Required(),
				mcp.Description("Numeric board ID that owns the sprint."),
			),
			mcp.WithString("jql",
				mcp.Description("Additional JQL filter to narrow down work items within the sprint."),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of work items to return. Default: 50."),
			),
			mcp.WithString("fields",
				mcp.Description("Comma-separated list of fields. Default: key,issuetype,summary,assignee,priority,status."),
			),
		),
		sp.HandleSprintListWorkitems,
	)

	s.AddTool(
		mcp.NewTool("jira_sprint_create",
			mcp.WithDescription("Create a new sprint on a Jira board."),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Sprint name, e.g. 'Sprint 5'."),
			),
			mcp.WithNumber("board",
				mcp.Required(),
				mcp.Description("Numeric board ID to create the sprint on."),
			),
			mcp.WithString("start",
				mcp.Description("Sprint start date in ISO 8601 format, e.g. '2025-01-01'."),
			),
			mcp.WithString("end",
				mcp.Description("Sprint end date in ISO 8601 format, e.g. '2025-01-14'."),
			),
			mcp.WithString("goal",
				mcp.Description("Sprint goal text."),
			),
		),
		sp.HandleSprintCreate,
	)

	// --- Projects ---

	s.AddTool(
		mcp.NewTool("jira_project_list",
			mcp.WithDescription("List Jira projects visible to the authenticated user."),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of projects to return. Default: 30."),
			),
			mcp.WithBoolean("recent",
				mcp.Description("Set to true to return only recently viewed projects (up to 20)."),
			),
		),
		pr.HandleProjectList,
	)

	s.AddTool(
		mcp.NewTool("jira_project_view",
			mcp.WithDescription("View details of a Jira project by its key."),
			mcp.WithString("key",
				mcp.Required(),
				mcp.Description("Project key, e.g. 'TEAM'."),
			),
		),
		pr.HandleProjectView,
	)

	// --- Boards ---

	s.AddTool(
		mcp.NewTool("jira_board_search",
			mcp.WithDescription("Search for Jira boards. Filter by name, project key, or board type (scrum, kanban, simple)."),
			mcp.WithString("name",
				mcp.Description("Filter by board name (partial match)."),
			),
			mcp.WithString("project",
				mcp.Description("Filter by project key, e.g. 'TEAM'."),
			),
			mcp.WithString("type",
				mcp.Description("Filter by board type: scrum, kanban, or simple."),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of boards to return. Default: 50."),
			),
		),
		bo.HandleBoardSearch,
	)

	s.AddTool(
		mcp.NewTool("jira_board_list_sprints",
			mcp.WithDescription("List all sprints for a given Jira board. Filter by state (future, active, closed)."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("Board ID, e.g. '123'."),
			),
			mcp.WithString("state",
				mcp.Description("Filter by sprint state. Valid values: future, active, closed. Comma-separated for multiple, e.g. 'active,closed'."),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of sprints to return. Default: 50."),
			),
		),
		bo.HandleBoardListSprints,
	)
}
