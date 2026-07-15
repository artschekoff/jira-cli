---
name: jira-cli-load-context
description: 'Load Jira issue context (description, comments, subtasks) via jira-cli'
targets: ["*"]
---

# Load Jira Context

All calls use `jira-cli` (JSON on stdout, errors on stderr). Pipe to `jq` to extract fields.

## Step 1: Verify auth

```bash
jira-cli auth status
```

Not authenticated → "jira-cli not authenticated. Run `jira-cli auth login` first." → stop.

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

```bash
jira-cli search --jql "text ~ \"$PHRASE\" $DATE_CLAUSE ORDER BY updated DESC" \
  --fields "key,summary,status,assignee" --limit 15
```

Zero results → report, loop to Step 2. Otherwise AskQuestion: one option per issue (`$KEY — $SUMMARY ($STATUS)`), plus `Refine search...` → loop to 2c.

## Step 3: Fetch issue detail

For a given `$KEY`, fetch two things sequentially:

1. `jira-cli view $KEY --fields "key,issuetype,summary,status,assignee,description"`. Fail → report error, stop.
   - Need custom fields too? Use `--fields '*all'` to pull every field (incl. `customfield_*`): `jira-cli view $KEY --fields '*all'`.
2. `jira-cli comment list $KEY --order asc`. Extract comments (author + body).

## Step 4: Fetch subtasks

```bash
jira-cli search --jql "parent = $ISSUE_KEY" --fields "key,summary,status,assignee"
```

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

- **Auth failed** → "Run `jira-cli auth login` to authenticate."
- **Issue not found** → report to user.
- **Command not found** → "jira-cli is not installed or not on PATH."
