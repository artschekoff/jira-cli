---
description: 'Create Jira work items via Jira MCP — fast path with template cloning, required-field enforcement, minimal discovery'
targets: ["*"]
---

# Create Jira Tickets

MCP: `server: "user-jira"`. Shell fallback: `acli jira workitem …` when noted below.

**Goals:** few tool calls, no test tickets, every creation-required field set before create.

## 1. Auth

`jira_auth_status` → not authenticated → stop: *Run `acli jira auth login`.*

## 2. Parse request → fill gaps

Extract from user message: project, type, summary, description, parent (epic/story URL or key), assignee, labels, priority, sprint, scope hints (`api` → juna-api, `frontend` → juna-frontend, etc.).

**Required before create:** `summary`, `type`, `project`, and every field the project marks required at creation (see §4). `description` required for bugs unless user already gave repro/fix text.

**If anything required is missing or ambiguous** → ask **once**, listing all gaps:

```
Need before I create:
- [field]: (options if known)
```

Do not guess parent, assignee email, scope repo, task type, or product pod. Do not create probe/test issues.

## 3. Fast discovery (max 2 calls, parallel)

Skip full field catalogues. Clone a sibling issue instead.

| Call | Purpose |
|---|---|
| `jira_search` `key = $PARENT` or `parent = $PARENT` | Validate parent exists |
| `jira_search` `project = $P AND parent = $PARENT AND issuetype = $TYPE ORDER BY created DESC` `limit: 1` | Template issue |

**Template fields** — shell only (MCP `jira_view` often blocks `*all`):

```bash
acli jira workitem view $TEMPLATE --json --fields "*all"
```

Copy every **non-null** custom field from the template into the new payload. Override only what the user specified.

**Skip:** `jira_view` MCP for discovery, `Epic Link` JQL (next-gen uses `parent`), sprint lookup unless user mentioned sprint.

Same project + same type as a prior ticket in session → reuse cached template / §4 cheat sheet.

## 4. JH project cheat sheet

| User says | Use |
|---|---|
| Bug | Issue type `Issue` (not `Bug`) |
| API scope | `customfield_10160` → `{"value":"juna-api"}` |
| Frontend scope | `customfield_10160` → `{"value":"juna-frontend"}` |
| RDBMS scope | `customfield_10160` → `{"value":"rdbms"}` |
| Bug task | `customfield_10226` → `{"value":"Bug"}` |
| Procedure / coding analytics epic | `customfield_10905` → `{"value":"Analytics - Coding Pattern"}` |
| Platform pod | `customfield_10905` → `{"value":"Platform"}` |

**JH Issue creation-required trio:** Scope Repository (`customfield_10160`), Task Type (`customfield_10226`), Product Pod (`customfield_10905`). All three must be non-null.

**Parent epic (next-gen):** inside `custom_fields` / `additionalAttributes`:

```json
"parent": { "key": "JH-2890" }
```

Not `parentIssueId`. Not `--parent` combined with `--from-json` (parent is ignored → ticket orphaned).

## 5. `additionalAttributes` formats

Single-select option: `{"value": "Option Name"}`  
Named entity (priority): `{"name": "High"}`  
Multi-select: `[{"value": "A"}, {"value": "B"}]`  
Number: bare number  
Parent epic: `{"key": "JH-123"}` inside `additionalAttributes`

## 6. Confirm (one screen)

```
Creating in $PROJECT:
  Type:     $TYPE
  Summary:  $SUMMARY
  Parent:   $PARENT_KEY — $PARENT_SUMMARY
  Assignee: $EMAIL
  Labels:   $LABELS
  Required: Scope Repository=$X | Task Type=$Y | Product Pod=$Z
  (+ any other non-null fields from template)
  Description: yes
```

**Create this ticket? (yes / edit / cancel)**

## 7. Create (one call)

**JH `Issue` or any required custom fields** → always pass `custom_fields` on `jira_create`:

```
jira_create(
  project, type, summary,
  description?, assignee?, labels?, parent?,   // parent param kept for subtasks; epics → also in custom_fields
  custom_fields: "{\"customfield_10160\":{\"value\":\"juna-api\"},\"customfield_10226\":{\"value\":\"Bug\"},\"customfield_10905\":{\"value\":\"Analytics - Coding Pattern\"},\"parent\":{\"key\":\"JH-2890\"}}"
)
```

`description` and `assignee` stay top-level MCP params (plain text + `@me`).

**Shell equivalent** (if MCP fails):

```bash
acli jira workitem create --from-json /tmp/issue.json --assignee "$EMAIL" --description-file /tmp/desc.txt --json
```

JSON must include `projectKey`, `type`, `summary`, `labels`, `additionalAttributes` (all required custom fields + `parent`). Do not mix partial CLI `--project` with `--from-json`.

**Success:** `✅ Created $KEY — $SUMMARY` + `https://junahealth.atlassian.net/browse/$KEY`

**Verify parent:** `acli jira workitem view $KEY --json --fields parent` → if `null`, parent was not in `additionalAttributes`; fix in Jira UI or recreate (do not leave duplicates — mark dupes in summary if recreate unavoidable).

## 8. Errors → fix, don't spiral

| Error | Action |
|---|---|
| Product Pod / Scope Repository / Task Type required | Add missing keys to `custom_fields`; use §4 |
| `parent` null after create | `parent` must be in `additionalAttributes`, not `parentIssueId` / `--parent` |
| `Please select valid parent issue` | Use `{"key":"JH-…"}` in `additionalAttributes` |
| `no value given for required property projectKey` | Full JSON body with `projectKey` when using `--from-json` |
| `jira_view` / subcommand not allowed | Shell `acli jira workitem view` |
| `Epic Link` JQL invalid | Use `parent = $KEY` |
| `unknown flag: --label` on edit | Use `--labels` (comma-separated) |
| Assignee by name only | Ask for email |
| Custom field value rejected | Show error + allowed values from template; ask user |

## 9. Another ticket?

`yes` + same project → skip §3 if type/parent unchanged; reuse template/cheat sheet.
