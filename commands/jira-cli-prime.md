---
description: 'Prime the agent with jira-cli usage — the CLI wrapper around acli that returns JSON on stdout'
targets: ["*"]
---

# jira-cli Prime

Use `jira-cli` (not raw `acli`) for all Jira operations. Every command prints JSON to stdout; errors go to stderr with exit code 1. Pipe to `jq` to extract fields.

## Auth

```bash
jira-cli auth status              # check first — stop and ask user to run login if not authenticated
jira-cli auth login               # interactive: site, email, API token
jira-cli auth logout
```

## Work items

```bash
jira-cli search --jql '<JQL>'                                 # JSON array
jira-cli view <KEY> [--fields summary,status,assignee,...]    # JSON object
jira-cli create --summary <s> --project <p> --type <t> \
  [--assignee <u>] [--custom-fields '<JSON>']                 # returns created issue
jira-cli edit <KEY> [--summary ...] [--custom-fields ...]
jira-cli transition <KEY> --status "<name>"
jira-cli assign <KEY> --assignee <user>
jira-cli comment add <KEY> --body "<text>"
jira-cli comment list <KEY> [--order desc] [--start-date YYYY-MM-DD]
```

## Sprint / Board / Project

```bash
jira-cli sprint view <ID>
jira-cli sprint list-workitems --sprint <id> --board <id> [--fields ...]
jira-cli sprint create --name <n> --board <id>
jira-cli board search
jira-cli board list-sprints <ID> [--state active]
jira-cli project list
jira-cli project view <KEY>
```

## Rules

- Custom fields use raw Jira field IDs (e.g. `customfield_10160`) and Jira value shapes: single-select `{"value":"X"}`, named `{"name":"X"}`, multi-select `[{"value":"A"}]`, parent epic `{"key":"JH-123"}` under `parent`.
- `@me` is a valid `--assignee` value.
- Every command supports `--help` — run it before guessing flags.
- Never invoke `acli` directly when a `jira-cli` subcommand exists.
