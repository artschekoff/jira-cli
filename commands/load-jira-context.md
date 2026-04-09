---
name: load-jira-context
author: artschekoff
description: "Load Jira issue context (description, comments, subtasks) via Jira MCP server"
targets: ["*"]
---

# Load Jira Context

All MCP calls use `server: "user-jira"`. Only read tool descriptors from `mcps/` if a call fails with an argument error.

## Step 1: Verify auth

`jira_auth_status` (no args). Not authenticated → "Jira CLI not authenticated. Run `acli jira auth` first." → stop.

## Step 2: Ask for the issue

Ask the user (plain text, NOT AskQuestion):

> **Jira issue?** Enter an issue key (e.g. `JH-568`), a full URL, or a search phrase (e.g. `auth login bug`).

### Parse input

- **Full URL** — `https?://.*/browse/([A-Z]+-\d+)` → extract key → Step 3
- **Exact key** — `^[A-Z]+-\d+$` → Step 3
- **Digits only** — `^\d+$` → prepend `JH-` → Step 3
- **Anything else** → search (Step 2b)

### 2b: Time range

AskQuestion — **Time range:**

| Option | JQL clause |
|---|---|
| All time | _(omit)_ |
| Last month | `AND updated >= "YYYY-MM-DD"` (today − 30 days) |
| Last week | `AND updated >= "YYYY-MM-DD"` (today − 7 days) |
| Last day | `AND updated >= "YYYY-MM-DD"` (today − 1 day) |

### 2c–2d: Search & choose

`jira_search` with `jql: "text ~ \"$PHRASE\" $DATE_CLAUSE ORDER BY updated DESC"`, `fields: "key,summary,status,assignee"`, `limit: 15`.

Zero results → report, loop to Step 2. Otherwise AskQuestion: one option per issue (`$KEY — $SUMMARY ($STATUS)`), plus `Refine search...` → loop to 2c.

## Step 3: Fetch issue detail

For a given `$KEY`, fetch two things sequentially:

1. `jira_view` — `key: $KEY`, `fields: "key,issuetype,summary,status,assignee,description"`. Fail → report error, stop.
2. `jira_comment_list` — `key: $KEY`, `order: "asc"`. Extract comments (author + body).

## Step 4: Fetch subtasks

`jira_search` with `jql: "parent = $ISSUE_KEY"`, `fields: "key,summary,status,assignee"`.

For each subtask, run Step 3 sequentially (avoid rate-limiting). Skip if none.

## Step 5: Present context

```markdown
# Jira Context: $ISSUE_KEY — $SUMMARY

## Parent Issue
- **Key:** $ISSUE_KEY | **Status:** $STATUS | **Assignee:** $ASSIGNEE

### Description
$DESCRIPTION

### Comments ($COUNT)
> **$AUTHOR** ($DATE):
> $BODY

---

## Subtasks ($COUNT)

### $SUBTASK_KEY — $SUMMARY
- **Status:** $STATUS | **Assignee:** $ASSIGNEE
#### Description
$DESCRIPTION
#### Comments ($COUNT)
> **$AUTHOR** ($DATE):
> $BODY
---
(repeat per subtask)
```

> Jira context loaded: **$ISSUE_KEY** + **N** subtask(s). Ask me anything about this task.

## Error Handling

- **Auth failed** → "Run `acli jira auth` to authenticate."
- **Issue not found** → report to user.
- **MCP server not running** → "Jira MCP server is not running. Check Cursor Settings → MCP."
