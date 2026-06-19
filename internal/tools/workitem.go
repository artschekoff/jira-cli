package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/artschekoff/jira-mcp/internal/acli"
	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
)

const (
	eventSearch      = "workitem.search"
	eventView        = "workitem.view"
	eventCreate      = "workitem.create"
	eventEdit        = "workitem.edit"
	eventTransition  = "workitem.transition"
	eventAssign      = "workitem.assign"
	eventCommentAdd  = "workitem.comment.add"
	eventCommentList = "workitem.comment.list"
)

// WorkitemHandlers groups work-item related tool handlers.
type WorkitemHandlers struct {
	baseHandler
}

// NewWorkitemHandlers creates a WorkitemHandlers.
func NewWorkitemHandlers(runner *acli.Runner, logger *zap.Logger) *WorkitemHandlers {
	return &WorkitemHandlers{baseHandler: newBaseHandler(runner, logger)}
}

// HandleSearch searches for Jira work items using a JQL query.
func (h *WorkitemHandlers) HandleSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if r := requireParams(req, "jql"); r != nil {
		return r, nil
	}

	flags := []string{"--jql", req.GetString("jql", ""), "--json"}

	if fields := req.GetString("fields", ""); fields != "" {
		flags = append(flags, "--fields", fields)
	}
	if limit := req.GetInt("limit", 0); limit > 0 {
		flags = append(flags, "--limit", fmt.Sprintf("%d", limit))
	}

	return h.run(ctx, eventSearch, []string{"jira", "workitem", "search"}, flags)
}

// HandleView retrieves detailed information about a Jira work item.
func (h *WorkitemHandlers) HandleView(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if r := requireParams(req, "key"); r != nil {
		return r, nil
	}

	flags := []string{"--json"}
	if fields := req.GetString("fields", ""); fields != "" {
		flags = append(flags, "--fields", fields)
	}

	return h.run(ctx, eventView, []string{"jira", "workitem", "view", req.GetString("key", "")}, flags)
}

// HandleCreate creates a new Jira work item.
func (h *WorkitemHandlers) HandleCreate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if r := requireParams(req, "summary", "project", "type"); r != nil {
		return r, nil
	}

	if cf := req.GetString("custom_fields", ""); cf != "" {
		return h.handleCreateJSON(ctx, req, cf)
	}

	flags := []string{
		"--summary", req.GetString("summary", ""),
		"--project", req.GetString("project", ""),
		"--type", req.GetString("type", ""),
		"--json",
	}

	if desc := req.GetString("description", ""); desc != "" {
		flags = append(flags, "--description", desc)
	}
	if assignee := req.GetString("assignee", ""); assignee != "" {
		flags = append(flags, "--assignee", assignee)
	}
	if labels := req.GetString("labels", ""); labels != "" {
		flags = append(flags, "--label", labels)
	}
	if parent := req.GetString("parent", ""); parent != "" {
		flags = append(flags, "--parent", parent)
	}

	return h.run(ctx, eventCreate, []string{"jira", "workitem", "create"}, flags)
}

// handleCreateJSON uses acli's --from-json path so custom fields can be set at
// creation time. The acli JSON format (from --generate-json) is a flat object
// with top-level keys (summary, projectKey, type, parentIssueId, labels, …)
// and an "additionalAttributes" map for every non-standard field.
func (h *WorkitemHandlers) handleCreateJSON(ctx context.Context, req mcp.CallToolRequest, customFieldsJSON string) (*mcp.CallToolResult, error) {
	var additionalAttrs map[string]any
	if err := json.Unmarshal([]byte(customFieldsJSON), &additionalAttrs); err != nil {
		return mcp.NewToolResultError("custom_fields must be a valid JSON object: " + err.Error()), nil
	}

	// acli flat format — NOT the Jira REST API {"fields":{}} shape.
	payload := map[string]any{
		"summary":              req.GetString("summary", ""),
		"projectKey":           req.GetString("project", ""),
		"type":                 req.GetString("type", ""),
		"additionalAttributes": additionalAttrs,
	}
	if parent := req.GetString("parent", ""); parent != "" {
		payload["parentIssueId"] = parent
	}
	if labels := req.GetString("labels", ""); labels != "" {
		parts := strings.Split(labels, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		payload["labels"] = parts
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return mcp.NewToolResultError("failed to serialize work item: " + err.Error()), nil
	}

	tmp, err := os.CreateTemp("", "jira-create-*.json")
	if err != nil {
		return mcp.NewToolResultError("failed to create temp file: " + err.Error()), nil
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return mcp.NewToolResultError("failed to write temp file: " + err.Error()), nil
	}
	tmp.Close()

	// description and assignee are passed as CLI flags alongside --from-json:
	// description because acli accepts plain text there (JSON requires ADF);
	// assignee because "@me" is a CLI-only shorthand, not valid inside JSON.
	flags := []string{"--from-json", tmp.Name(), "--json"}
	if desc := req.GetString("description", ""); desc != "" {
		flags = append(flags, "--description", desc)
	}
	if assignee := req.GetString("assignee", ""); assignee != "" {
		flags = append(flags, "--assignee", assignee)
	}
	return h.run(ctx, eventCreate, []string{"jira", "workitem", "create"}, flags)
}

// HandleEdit edits one or more Jira work items.
func (h *WorkitemHandlers) HandleEdit(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if r := requireParams(req, "key"); r != nil {
		return r, nil
	}

	flags := []string{"--key", req.GetString("key", ""), "--yes", "--json"}

	if summary := req.GetString("summary", ""); summary != "" {
		flags = append(flags, "--summary", summary)
	}
	if desc := req.GetString("description", ""); desc != "" {
		flags = append(flags, "--description", desc)
	}
	if assignee := req.GetString("assignee", ""); assignee != "" {
		flags = append(flags, "--assignee", assignee)
	}
	if labels := req.GetString("labels", ""); labels != "" {
		flags = append(flags, "--labels", labels)
	}
	if issueType := req.GetString("type", ""); issueType != "" {
		flags = append(flags, "--type", issueType)
	}

	return h.run(ctx, eventEdit, []string{"jira", "workitem", "edit"}, flags)
}

// HandleTransition transitions a Jira work item to a new status.
func (h *WorkitemHandlers) HandleTransition(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if r := requireParams(req, "key", "status"); r != nil {
		return r, nil
	}

	flags := []string{
		"--key", req.GetString("key", ""),
		"--status", req.GetString("status", ""),
		"--yes",
		"--json",
	}

	return h.run(ctx, eventTransition, []string{"jira", "workitem", "transition"}, flags)
}

// HandleAssign assigns a Jira work item to a user.
func (h *WorkitemHandlers) HandleAssign(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if r := requireParams(req, "key", "assignee"); r != nil {
		return r, nil
	}

	flags := []string{
		"--key", req.GetString("key", ""),
		"--assignee", req.GetString("assignee", ""),
		"--yes",
		"--json",
	}

	return h.run(ctx, eventAssign, []string{"jira", "workitem", "assign"}, flags)
}

// HandleCommentAdd adds a comment to a Jira work item.
func (h *WorkitemHandlers) HandleCommentAdd(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if r := requireParams(req, "key", "body"); r != nil {
		return r, nil
	}

	flags := []string{
		"--key", req.GetString("key", ""),
		"--body", req.GetString("body", ""),
		"--json",
	}

	return h.run(ctx, eventCommentAdd, []string{"jira", "workitem", "comment", "create"}, flags)
}

// HandleCommentList lists comments on a Jira work item.
func (h *WorkitemHandlers) HandleCommentList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if r := requireParams(req, "key"); r != nil {
		return r, nil
	}

	flags := []string{"--key", req.GetString("key", ""), "--json"}

	if limit := req.GetInt("limit", 0); limit > 0 {
		flags = append(flags, "--limit", fmt.Sprintf("%d", limit))
	}
	if order := req.GetString("order", ""); order != "" {
		flags = append(flags, "--order", order)
	}
	if req.GetBool("paginate", false) {
		flags = append(flags, "--paginate")
	}

	startDateStr := req.GetString("start_date", "")
	var startDate time.Time
	if startDateStr != "" {
		var err error
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return mcp.NewToolResultError(
				fmt.Sprintf("invalid start_date %q: expected YYYY-MM-DD format", startDateStr),
			), nil
		}
	}

	h.logger.Info(eventCommentList, zap.Strings("cmd", []string{"jira", "workitem", "comment", "list"}))

	out, err := h.runner.Run(ctx, []string{"jira", "workitem", "comment", "list"}, flags)
	if err != nil {
		h.logger.Error(eventCommentList+" failed", zap.Error(err))
		return mcp.NewToolResultError(errMsgAcliUnavailable + ": " + err.Error()), nil
	}
	if out == "" {
		return mcp.NewToolResultText("(no output)"), nil
	}

	if !startDate.IsZero() {
		filtered, fErr := filterCommentsByStartDate(out, startDate)
		if fErr != nil {
			h.logger.Error(eventCommentList+" filter failed", zap.Error(fErr))
			return mcp.NewToolResultError("failed to filter comments by start_date: " + fErr.Error()), nil
		}
		out = filtered
	}

	return mcp.NewToolResultText(out), nil
}

// commentDateFormats lists the date formats we attempt when parsing the
// "created" field in Jira comment JSON. Tried in order; first success wins.
var commentDateFormats = []string{
	"2006-01-02T15:04:05.000-0700",
	"2006-01-02T15:04:05.000Z0700",
	time.RFC3339,
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05.000Z",
	"2006-01-02T15:04:05-0700",
}

// filterCommentsByStartDate filters a JSON array of comments, keeping only
// those whose "created" timestamp is on or after startDate. Comments without
// a parseable "created" field are dropped.
func filterCommentsByStartDate(jsonData string, startDate time.Time) (string, error) {
	var comments []map[string]any
	if err := json.Unmarshal([]byte(jsonData), &comments); err != nil {
		return "", fmt.Errorf("parsing comment JSON: %w", err)
	}

	filtered := make([]map[string]any, 0, len(comments))
	for _, c := range comments {
		createdRaw, ok := c["created"]
		if !ok {
			continue
		}
		createdStr, ok := createdRaw.(string)
		if !ok {
			continue
		}

		created, err := parseCommentDate(createdStr)
		if err != nil {
			continue
		}

		if !created.Before(startDate) {
			filtered = append(filtered, c)
		}
	}

	out, err := json.Marshal(filtered)
	if err != nil {
		return "", fmt.Errorf("marshaling filtered comments: %w", err)
	}
	return string(out), nil
}

func parseCommentDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, layout := range commentDateFormats {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized date format: %q", s)
}
