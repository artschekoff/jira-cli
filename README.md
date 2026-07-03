# jira-cli

CLI wrapper around the [Atlassian CLI](https://www.atlassian.com/software/acli) (`acli`).
Exposes common Jira operations as subcommands. All output is JSON to stdout.

## Install

```bash
go install github.com/artschekoff/jira-cli/cmd/jira-cli@latest
# or
git clone https://github.com/artschekoff/jira-cli.git && cd jira-cli && make install
```

Requires Go 1.23+.

## Prerequisites

Install and authenticate the [Atlassian CLI](https://www.atlassian.com/software/acli):

```bash
acli --version         # confirm installed
jira-cli auth login    # interactive: site, email, API token
```

Verify:

```bash
jira-cli auth status
```

## Commands

All commands print JSON to stdout. Errors go to stderr with exit code 1.

### Auth

| Command | Description |
|---------|-------------|
| `jira-cli auth login` | Log in via acli (interactive prompt) |
| `jira-cli auth logout` | Clear acli credentials |
| `jira-cli auth status` | Check acli authentication status |

### Work Items

| Command | Description |
|---------|-------------|
| `jira-cli search --jql <query>` | Search with JQL → JSON array |
| `jira-cli view <KEY>` | View work item → JSON object |
| `jira-cli create --summary <s> --project <p> --type <t>` | Create work item → JSON object |
| `jira-cli edit <KEY>` | Edit fields → JSON object |
| `jira-cli transition <KEY> --status <s>` | Change status → JSON object |
| `jira-cli assign <KEY> --assignee <u>` | Assign user → JSON object |
| `jira-cli comment add <KEY> --body <text>` | Add comment → JSON object |
| `jira-cli comment list <KEY>` | List comments → JSON array |

### Sprint

| Command | Description |
|---------|-------------|
| `jira-cli sprint view <ID>` | Sprint details → JSON object |
| `jira-cli sprint list-workitems --sprint <id> --board <id>` | Work items in sprint → JSON array |
| `jira-cli sprint create --name <n> --board <id>` | Create sprint → JSON object |

### Project

| Command | Description |
|---------|-------------|
| `jira-cli project list` | All projects → JSON array |
| `jira-cli project view <KEY>` | Project details → JSON object |

### Board

| Command | Description |
|---------|-------------|
| `jira-cli board search` | Find boards → JSON array |
| `jira-cli board list-sprints <ID>` | Sprints for board → JSON array |

## Usage examples

```bash
# Find all open bugs in TEAM, newest first
jira-cli search --jql 'project = TEAM AND type = Bug AND status != Done ORDER BY created DESC'

# View a ticket with specific fields
jira-cli view PROJ-123 --fields summary,status,assignee,description

# Create a story and self-assign it
jira-cli create --summary "Implement OAuth" --project TEAM --type Story --assignee "@me"

# Create with custom fields (e.g. story points, components)
jira-cli create --summary "API work" --project TEAM --type Story \
  --custom-fields '{"customfield_10016": 5, "components": [{"name": "Backend"}]}'

# Move ticket to In Progress
jira-cli transition PROJ-123 --status "In Progress"

# List recent comments (post June 1)
jira-cli comment list PROJ-123 --order desc --start-date 2024-06-01

# Find the active sprint on board 7
jira-cli board list-sprints 7 --state active

# List tickets in sprint 42 on board 7
jira-cli sprint list-workitems --sprint 42 --board 7 --fields summary,status,assignee

# Pipe to jq for filtering
jira-cli search --jql 'project = TEAM' | jq '[.[] | {key, summary: .fields.summary}]'
```

## Run `--help` for full flag reference

Every command has detailed help with flag descriptions and I/O format:

```bash
jira-cli --help
jira-cli create --help
jira-cli comment list --help
```
