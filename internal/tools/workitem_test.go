package tools

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/artschekoff/jira-mcp/internal/acli"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// newTestRequest builds a CallToolRequest with the given arguments map.
func newTestRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "jira_comment_list",
			Arguments: args,
		},
	}
}

// resultText extracts the text string from a CallToolResult.
func resultText(t *testing.T, r *mcp.CallToolResult) string {
	t.Helper()
	require.NotNil(t, r)
	require.NotEmpty(t, r.Content)
	tc, ok := r.Content[0].(mcp.TextContent)
	require.True(t, ok, "expected TextContent, got %T", r.Content[0])
	return tc.Text
}

func TestHandleCommentList_Flags(t *testing.T) {
	logger := zap.NewNop()
	runner := acli.NewRunner(logger, acli.WithBinPath("echo"), acli.WithTimeout(5*time.Second))
	h := NewWorkitemHandlers(runner, logger)

	tests := []struct {
		name     string
		args     map[string]any
		wantAll  []string
		wantNone []string
	}{
		{
			name: "only required key",
			args: map[string]any{"key": "PROJ-123"},
			wantAll: []string{
				"--key", "PROJ-123", "--json",
			},
			wantNone: []string{"--limit", "--order", "--paginate"},
		},
		{
			name: "with limit",
			args: map[string]any{"key": "PROJ-123", "limit": float64(25)},
			wantAll: []string{
				"--key", "PROJ-123", "--json", "--limit", "25",
			},
		},
		{
			name: "with order",
			args: map[string]any{"key": "PROJ-123", "order": "desc"},
			wantAll: []string{
				"--key", "PROJ-123", "--json", "--order", "desc",
			},
		},
		{
			name: "with paginate true",
			args: map[string]any{"key": "PROJ-123", "paginate": true},
			wantAll: []string{
				"--key", "PROJ-123", "--json", "--paginate",
			},
		},
		{
			name:     "paginate false omitted",
			args:     map[string]any{"key": "PROJ-123", "paginate": false},
			wantNone: []string{"--paginate"},
		},
		{
			name: "all flags combined",
			args: map[string]any{
				"key":      "PROJ-123",
				"limit":    float64(10),
				"order":    "asc",
				"paginate": true,
			},
			wantAll: []string{
				"--key", "PROJ-123", "--json",
				"--limit", "10",
				"--order", "asc",
				"--paginate",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newTestRequest(tt.args)
			res, err := h.HandleCommentList(context.Background(), req)
			require.NoError(t, err)
			require.False(t, res.IsError, "unexpected tool error: %s", resultText(t, res))

			out := resultText(t, res)
			for _, want := range tt.wantAll {
				assert.Contains(t, out, want)
			}
			for _, nope := range tt.wantNone {
				assert.NotContains(t, out, nope)
			}
		})
	}
}

func TestHandleCommentList_MissingKey(t *testing.T) {
	logger := zap.NewNop()
	runner := acli.NewRunner(logger, acli.WithBinPath("echo"), acli.WithTimeout(5*time.Second))
	h := NewWorkitemHandlers(runner, logger)

	req := newTestRequest(map[string]any{})
	res, err := h.HandleCommentList(context.Background(), req)
	require.NoError(t, err)
	require.True(t, res.IsError)
	assert.Contains(t, resultText(t, res), `"key" is required`)
}

func TestFilterCommentsByStartDate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		startDate string
		wantCount int
		wantErr   bool
		wantIDs   []string
	}{
		{
			name: "filters old comments from array",
			input: `[
				{"id":"1","body":"old","created":"2025-01-01T00:00:00.000+0000"},
				{"id":"2","body":"new","created":"2025-06-15T12:00:00.000+0000"},
				{"id":"3","body":"newest","created":"2025-12-01T08:30:00.000+0000"}
			]`,
			startDate: "2025-06-01",
			wantCount: 2,
			wantIDs:   []string{"2", "3"},
		},
		{
			name: "exact boundary is inclusive",
			input: `[
				{"id":"1","body":"exact","created":"2025-06-01T00:00:00.000+0000"},
				{"id":"2","body":"before","created":"2025-05-31T23:59:59.000+0000"}
			]`,
			startDate: "2025-06-01",
			wantCount: 1,
			wantIDs:   []string{"1"},
		},
		{
			name:      "empty array returns empty array",
			input:     `[]`,
			startDate: "2025-01-01",
			wantCount: 0,
		},
		{
			name: "all comments pass filter",
			input: `[
				{"id":"1","body":"a","created":"2025-07-01T00:00:00.000+0000"},
				{"id":"2","body":"b","created":"2025-08-01T00:00:00.000+0000"}
			]`,
			startDate: "2025-01-01",
			wantCount: 2,
			wantIDs:   []string{"1", "2"},
		},
		{
			name: "all comments filtered out",
			input: `[
				{"id":"1","body":"a","created":"2025-01-01T00:00:00.000+0000"},
				{"id":"2","body":"b","created":"2025-02-01T00:00:00.000+0000"}
			]`,
			startDate: "2025-12-01",
			wantCount: 0,
		},
		{
			name:    "invalid JSON returns error",
			input:   `{broken`,
			wantErr: true,
		},
		{
			name: "missing created field skips comment",
			input: `[
				{"id":"1","body":"no date"},
				{"id":"2","body":"has date","created":"2025-06-15T00:00:00.000+0000"}
			]`,
			startDate: "2025-01-01",
			wantCount: 1,
			wantIDs:   []string{"2"},
		},
		{
			name: "handles ISO 8601 with timezone offset",
			input: `[
				{"id":"1","body":"a","created":"2025-06-15T12:00:00.000+0530"}
			]`,
			startDate: "2025-06-01",
			wantCount: 1,
			wantIDs:   []string{"1"},
		},
		{
			name: "handles RFC3339 format",
			input: `[
				{"id":"1","body":"a","created":"2025-06-15T12:00:00Z"}
			]`,
			startDate: "2025-06-01",
			wantCount: 1,
			wantIDs:   []string{"1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startDate := tt.startDate
			if startDate == "" {
				startDate = "2025-01-01"
			}
			sd, err := time.Parse("2006-01-02", startDate)
			require.NoError(t, err)

			got, err := filterCommentsByStartDate(tt.input, sd)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Parse result to verify count and IDs
			var comments []map[string]any
			if tt.wantCount == 0 {
				assert.Equal(t, "[]", got)
				return
			}
			require.NoError(t, json.Unmarshal([]byte(got), &comments))
			assert.Len(t, comments, tt.wantCount)

			if len(tt.wantIDs) > 0 {
				var gotIDs []string
				for _, c := range comments {
					if id, ok := c["id"].(string); ok {
						gotIDs = append(gotIDs, id)
					}
				}
				assert.Equal(t, tt.wantIDs, gotIDs)
			}
		})
	}
}

func TestHandleCommentList_StartDateFilter(t *testing.T) {
	logger := zap.NewNop()
	runner := acli.NewRunner(logger, acli.WithBinPath("echo"), acli.WithTimeout(5*time.Second))
	h := NewWorkitemHandlers(runner, logger)

	tests := []struct {
		name       string
		startDate  string
		wantErr    bool
		wantIsErr  bool
		errContain string
	}{
		{
			name:       "invalid start_date format returns error",
			startDate:  "not-a-date",
			wantIsErr:  true,
			errContain: "invalid start_date",
		},
		{
			name:      "empty start_date is ignored",
			startDate: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]any{"key": "PROJ-123"}
			if tt.startDate != "" {
				args["start_date"] = tt.startDate
			}
			req := newTestRequest(args)
			res, err := h.HandleCommentList(context.Background(), req)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tt.wantIsErr {
				require.True(t, res.IsError)
				assert.Contains(t, resultText(t, res), tt.errContain)
				return
			}

			require.False(t, res.IsError)
		})
	}
}
