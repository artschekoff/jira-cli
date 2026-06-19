---
name: jira-create-tickets
author: artschekoff
description: "Create Jira work items (tasks, stories, epics, bugs) via Jira MCP — discovers available issue types and custom fields from the project, matches them to the user's request, and asks when ambiguous"
targets: ["*"]
---

# Create Jira Tickets

All MCP calls use `server: "user-jira"`.

## Step 1: Verify auth

`jira_auth_status` (no args). Not authenticated → "Jira CLI not authenticated. Run `acli jira auth` first." → stop.

## Step 2: Resolve project

If the user provided a project key (e.g. `JH`) → use it directly → Step 3.

Otherwise `jira_project_list` with `recent: true`. Present as numbered list (`$KEY — $NAME`). Ask user to pick or type a key.

## Step 3: Discover project fields

Run these **in parallel**:

1. **Issue types** — `jira_project_view` with `key: $PROJECT`. Extract available issue types (e.g. `Task`, `Story`, `Epic`, `Bug`, `Subtask`). Use only what the project exposes.

2. **Custom field discovery** — `jira_search` with:
   - `jql: "project = $PROJECT ORDER BY updated DESC"`
   - `fields: "*all"`
   - `limit: 1`

   Parse the response to build a **field catalogue**: for every field key that is not a core field (`summary`, `description`, `status`, `assignee`, `issuetype`, `project`, `labels`, `parent`, `creator`, `reporter`, `created`, `updated`, `comment`, `attachment`, `watches`, `votes`), record:
   - `key` — the field key (e.g. `customfield_10016`, `components`, `priority`, `fixVersions`, `versions`)
   - `name` — the human label (e.g. `Story Points`, `Components`, `Priority`, `Sprint`)
   - `currentValue` — the value from that sample issue (helps infer the value format)

3. **Parent validation** — if the user's request mentions a parent/epic (e.g. "under JH-42"), `jira_view` with `key: $PARENT_KEY`, `fields: "key,summary,issuetype"` to confirm it exists.

**Known field key mappings** (common across Jira Cloud):

| User term | Field key | Value format |
|---|---|---|
| priority | `priority` | `{"name": "High"}` / `{"name": "Medium"}` / `{"name": "Low"}` |
| component / components | `components` | `[{"name": "Backend"}]` |
| fix version / version | `fixVersions` | `[{"name": "v1.2"}]` |
| story points / points / SP | `customfield_10016` | number, e.g. `5` |
| sprint | `customfield_10020` | `[{"id": N}]` — resolve sprint name → ID first |
| epic link | `customfield_10014` | `"JH-42"` (string key) |

If a field key is not in the table above, use the key from the discovered catalogue.

## Step 4: Match request to fields

Parse the user's request and map every mentioned concept to a field:

| Field | Delivery param | Notes |
|---|---|---|
| `type` | `type` (top-level) | Fuzzy-match to available issue types. Default: `Task`. |
| `summary` | `summary` (top-level) | Required. Extract or ask. |
| `description` | `description` (top-level) | Optional. |
| `parent` | `parent` (top-level) | Epic/parent key. Required for `Subtask`. |
| `assignee` | `assignee` (top-level) | Email or `@me`. |
| `labels` | `labels` (top-level) | Comma-separated string. |
| Everything else | `custom_fields` JSON | One JSON object, all extra keys merged in. |

**Ambiguity rules — ask before assuming:**

- Type not recognizable → ask: "What type? _(list available types from Step 3)_"
- Summary missing → ask for it directly.
- Parent mentioned by description (not key) → search: `jira_search` with `jql: "text ~ \"$PHRASE\" AND issuetype = Epic ORDER BY updated DESC"`, `limit: 10` → show results → ask user to pick.
- Assignee mentioned by name (not email) → ask to confirm email or use `@me`.
- **User mentions a concept that maps to a field with multiple valid values** (e.g. "high priority", "sprint 5", "backend component") but the exact value doesn't match anything in the catalogue → show available values and ask which one.
- **Unknown concept** (user mentions something not in the catalogue) → ask: "I don't see a field for '$X' in this project. Did you mean one of these? _(list closest catalogue entries)_ Or is it a different field?"

One question at a time. Do not ask about fields the user clearly did not mention.

## Step 5: Resolve sprint (when mentioned)

If the user mentioned a sprint, resolve its numeric ID before building the payload:

1. `jira_board_search` with `project: $PROJECT` → get the board ID.
2. `jira_board_list_sprints` with `id: $BOARD_ID`, `state: "active,future"` → find the sprint by name (fuzzy OK).
3. Extract the sprint's numeric `id`. If ambiguous → show matches and ask.

Pass as `"customfield_10020": [{"id": $SPRINT_ID}]` in `custom_fields`.

## Step 6: Present creation plan

Show before creating anything:

```
Creating in $PROJECT:

  Type:          $TYPE
  Summary:       $SUMMARY
  Parent:        $PARENT_KEY — $PARENT_SUMMARY     ← omit if not set
  Assignee:      $ASSIGNEE                          ← omit if not set
  Labels:        $LABELS                            ← omit if not set
  Priority:      $PRIORITY                          ← omit if not set
  Component(s):  $COMPONENTS                        ← omit if not set
  Sprint:        $SPRINT_NAME                       ← omit if not set
  Story points:  $SP                                ← omit if not set
  Other fields:  $KEY = $VALUE, ...                 ← omit if none
  Description:   (yes / no)
```

Ask: **"Create this ticket? (yes / edit / cancel)"**

- `edit` → ask which field to change → update and loop back here.
- `cancel` → stop.
- `yes` → Step 7.

## Step 7: Create

Build the call:

- If **no custom fields** were collected → use top-level params only:
  ```
  jira_create(project, type, summary, description?, parent?, assignee?, labels?)
  ```

- If **any custom fields** were collected → pass `custom_fields` as a JSON object in addition to top-level params:
  ```
  jira_create(
    project, type, summary, description?, parent?, assignee?, labels?,
    custom_fields: "{\"components\":[{\"name\":\"Backend\"}], \"priority\":{\"name\":\"High\"}, ...}"
  )
  ```
  `assignee` is always a top-level param (even in custom-field mode) — `@me` only works as a CLI flag, not inside JSON.

On success → print: `✅ Created $KEY — $SUMMARY` with link `https://junahealth.atlassian.net/browse/$KEY`.

On error → print the error verbatim, then ask: "Fix and retry, or cancel?"

## Step 8: More tickets?

Ask: **"Create another ticket? (yes / no)"**

- `yes` → back to Step 3 (project and field catalogue are remembered; skip re-discovery if same project).
- `no` → done.

## Error handling

| Error | Response |
|---|---|
| Auth failed | "Run `acli jira auth` to authenticate." |
| Project not found | Re-ask for project key. |
| Parent not found | "Parent $KEY not found. Enter a different key or omit." |
| Issue type not valid | List types from Step 3, ask to pick. |
| Custom field value rejected | Show the error, ask for a corrected value. |
| Field not available on this issue type | Note it, offer to omit and retry. |
| MCP server not running | "Jira MCP server is not running. Check Settings → MCP." |
