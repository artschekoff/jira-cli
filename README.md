# jira-mcp

MCP server that gives AI assistants access to Jira — search issues, manage sprints, create and edit work items. Wraps the Atlassian CLI (`acli`) and exposes operations as MCP tools over stdio.

## Install

```bash
go install github.com/artschekoff/jira-mcp/cmd/jira-mcp@latest
# or
git clone https://github.com/artschekoff/jira-mcp.git && cd jira-mcp && make install
```

Requires Go 1.23+.

## Prerequisites

Install and authenticate the [Atlassian CLI](https://www.atlassian.com/software/acli):

```bash
acli jira auth login
```

Verify it works:

```bash
acli jira auth status
```

## MCP server setup

Add to your MCP client config.

**Cursor** (`.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "jira": {
      "type": "stdio",
      "command": "jira-mcp"
    }
  }
}
```

**Claude Desktop** (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "jira": {
      "command": "jira-mcp"
    }
  }
}
```

## Tools

| Tool | Parameters | Description |
|------|-----------|-------------|
| `jira_auth_status` | — | Check acli authentication status |
| `jira_search` | `jql`, `fields?`, `limit?` | Search work items with JQL |
| `jira_view` | `key`, `fields?` | View full details of a work item |
| `jira_create` | `summary`, `project`, `type`, `description?`, `assignee?`, `labels?`, `parent?` | Create a new work item |
| `jira_edit` | `key`, `summary?`, `description?`, `assignee?`, `labels?`, `type?` | Edit one or more work items |
| `jira_transition` | `key`, `status` | Transition work items to a new status |
| `jira_assign` | `key`, `assignee` | Assign work items to a user |
| `jira_comment_add` | `key`, `body` | Add a comment to a work item |
| `jira_comment_list` | `key`, `limit?`, `order?`, `paginate?`, `start_date?` | List comments on a work item |
| `jira_sprint_view` | `id` | View sprint details by ID |
| `jira_sprint_list_workitems` | `sprint`, `board`, `jql?`, `limit?`, `fields?` | List work items in a sprint |
| `jira_sprint_create` | `name`, `board`, `start?`, `end?`, `goal?` | Create a new sprint |
| `jira_project_list` | `limit?`, `recent?` | List visible projects |
| `jira_project_view` | `key` | View project details |
| `jira_board_search` | `name?`, `project?`, `type?`, `limit?` | Search for boards |
| `jira_board_list_sprints` | `id`, `state?`, `limit?` | List sprints for a board |

## CLI

```
jira-mcp    # Start MCP server (stdio)
```
