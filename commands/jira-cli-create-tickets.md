---
description: 'Create Jira work items via jira-cli — fast path with template cloning, required-field enforcement, minimal discovery'
targets: ["*"]
---

# Create Jira Tickets

All calls use `jira-cli` (JSON on stdout). Run any subcommand with `--help` before guessing flags.

**Goals:** few calls, no test tickets, every creation-required field set before create.

## 1. Auth

`jira-cli auth status` → not authenticated → stop: *Run `jira-cli auth login`.*

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
| `jira-cli search --jql "key = $PARENT"` | Validate parent exists |
| `jira-cli search --jql "project = $P AND parent = $PARENT AND issuetype = $TYPE ORDER BY created DESC" --limit 1` | Template issue |

**Template fields:**

```bash
jira-cli view $TEMPLATE --fields "*all"
```

Copy every **non-null** custom field from the template into the new payload. Override only what the user specified.

**Skip:** sprint lookup unless user mentioned sprint. `Epic Link` JQL (next-gen uses `parent`).

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

**Parent epic (next-gen):** inside `--custom-fields`:

```json
"parent": { "key": "JH-2890" }
```

Not `parentIssueId`. Do not rely on a separate `--parent` flag for epics — pass parent inside `--custom-fields`.

## 5. `--custom-fields` value shapes

Single-select option: `{"value": "Option Name"}`
Named entity (priority): `{"name": "High"}`
Multi-select: `[{"value": "A"}, {"value": "B"}]`
Number: bare number
Parent epic: `{"key": "JH-123"}` under `parent`

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

```bash
jira-cli create \
  --project JH \
  --type Issue \
  --summary "$SUMMARY" \
  --description "$DESCRIPTION" \
  --assignee "$EMAIL_OR_@me" \
  --custom-fields '{"customfield_10160":{"value":"juna-api"},"customfield_10226":{"value":"Bug"},"customfield_10905":{"value":"Analytics - Coding Pattern"},"parent":{"key":"JH-2890"}}'
```

`--custom-fields` must include every required custom field + `parent` for epics. `@me` is a valid `--assignee` value.

**Success:** `✅ Created $KEY — $SUMMARY` + `https://junahealth.atlassian.net/browse/$KEY`

**Verify parent:** `jira-cli view $KEY --fields parent` → if `null`, parent was not in `--custom-fields`; fix in Jira UI or recreate (do not leave duplicates — mark dupes in summary if recreate unavoidable).

## 8. Errors → fix, don't spiral

| Error | Action |
|---|---|
| Product Pod / Scope Repository / Task Type required | Add missing keys to `--custom-fields`; use §4 |
| `parent` null after create | `parent` must be in `--custom-fields`, not a separate flag |
| `Please select valid parent issue` | Use `{"key":"JH-…"}` in `--custom-fields` |
| `Epic Link` JQL invalid | Use `parent = $KEY` |
| Assignee by name only | Ask for email |
| Custom field value rejected | Show error + allowed values from template; ask user |
| Unknown flag | Run `jira-cli <subcommand> --help` |

## 9. Another ticket?

`yes` + same project → skip §3 if type/parent unchanged; reuse template/cheat sheet.
