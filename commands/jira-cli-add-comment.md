---
name: jira-cli-add-comment
description: 'Add a comment to a Jira issue via jira-cli'
targets: ["*"]
---

# Add Jira Comment

Uses `jira-cli comment add`. JSON on stdout, errors on stderr.

## 1. Auth

```bash
jira-cli auth status
```

Not authenticated → stop: *Run `jira-cli auth login`.*

## 2. Resolve inputs

Need two things: `$KEY` and `$BODY`.

- **Key** — user provided a URL (`https?://.*/browse/([A-Z]+-\d+)`), a bare key (`^[A-Z]+-\d+$`), or digits only (`^\d+$` → prepend `JH-`). Anything else → ask for the key.
- **Body** — take the user's comment text verbatim. Missing/empty → ask once: *"Comment text?"* Do not paraphrase or invent content.

## 3. Confirm (one screen)

```
Adding comment to $KEY:
  $BODY
```

**Post this comment? (yes / edit / cancel)**

Skip confirmation only if the user's original message clearly said "post" / "add" with the full body inline.

## 4. Post

```bash
jira-cli comment add $KEY --body "$BODY"
```

Multi-line body → pass via `$(cat <<'EOF' ... EOF)` heredoc to preserve newlines and quotes.

**Success:** `✅ Comment added to $KEY` + `https://junahealth.atlassian.net/browse/$KEY`

## 5. Errors

| Error | Action |
|---|---|
| Auth failed | Run `jira-cli auth login` |
| Issue not found | Verify `$KEY` — check for typo, wrong project prefix |
| Empty body rejected | Ask user for comment text |
| Unknown flag | `jira-cli comment add --help` |
