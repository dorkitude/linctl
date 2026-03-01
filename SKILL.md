---
name: linctl
description: Use linctl to read and update Linear issues, projects, teams, users, labels, comments, and agent sessions from the terminal. Prefer JSON reads, validate auth first, and use command-specific --help for exact flags.
---

# linctl Agent Guide

Use this skill when the user wants to inspect or modify Linear data through `linctl`.

## Quick Rules

- Always verify auth before substantive work: `linctl auth status` or `linctl whoami`.
- For read operations, prefer `--json` and parse results with `jq` when needed.
- Before writing, inspect current state first (`get` / `list --json`).
- Use command-specific help for exact flags and validation rules: `linctl <command> <subcommand> --help`.
- Be explicit with filters; defaults can hide expected results.

## High-Impact Gotchas

- `issue list`, `issue search`, and `project list` default to `--newer-than 6_months_ago`.
- `issue list` and `issue search` exclude completed/canceled by default.
- `issue search` may also need `--include-archived` for archived matches.
- `issue list --cycle current` can validly return no rows if no active cycle exists.
- Parent/sub-issue links are set via `issue update --parent` (not `issue create`).
- If results look incomplete, retry with:
  - `--newer-than all_time`
  - `--include-completed`
  - `--include-archived` (search)

## Auth + Credential Behavior

- Interactive login: `linctl auth` (or `linctl auth login`).
- Credentials are stored in `~/.linctl-auth.json`.
- `LINCTL_API_KEY` overrides stored credentials.
- Precedence: `LINCTL_API_KEY` > `~/.linctl-auth.json`.

## Core Workflow

1. Confirm identity and access.
2. Discover relevant entities (team key, issue ID, project ID, user email) via `list/search` with `--json`.
3. Read exact target with `get`.
4. Apply changes with `create/update/assign/comment/label/agent` commands.
5. Re-read target and summarize concrete outcome to the user.

## Command Map

- Issues: `linctl issue ...`
- Projects: `linctl project ...`
- Teams: `linctl team ...` (includes `statuses` and `status-update`)
- Users: `linctl user ...` and `linctl whoami`
- Comments: `linctl comment ...`
- Labels: `linctl label ...`
- Agent Sessions: `linctl agent ...`
- Auth: `linctl auth ...`

## Suggested Read Patterns

```bash
# Issues assigned to me, no date-limit surprise
linctl issue list --assignee me --newer-than all_time --include-completed --json

# Issues in current cycle (if cycle exists)
linctl issue list --cycle current --limit 20 --json

# Find issue by text, including archived/completed
linctl issue search "<query>" --newer-than all_time --include-completed --include-archived --json

# Project inventory without default date filter
linctl project list --newer-than all_time --json

# Team labels for resolution/validation
linctl label list --team <TEAM_KEY> --json

# Team workflow statuses
linctl team statuses <TEAM_KEY> --json

# Agent session status for an issue
linctl agent <ISSUE_ID> --json
```

## Suggested Write Patterns

```bash
# Update an issue after inspecting it
linctl issue get LIN-123 --json
linctl issue update LIN-123 --state "In Progress" --assignee me
linctl issue get LIN-123 --json

# Delegate issue to an agent/user
linctl issue update LIN-123 --delegate "<displayName|email|name>"

# Set labels by name/ID, or clear
linctl issue update LIN-123 --labels bug,urgent
linctl issue update LIN-123 --clear-labels

# Attach PR or URL
linctl issue attach LIN-123 --pr https://github.com/org/repo/pull/123
linctl issue attach LIN-123 --url https://example.com/spec --title "Spec"

# Add execution note
linctl comment create LIN-123 --body "Implemented and verified locally."

# Comment lifecycle by comment ID
linctl comment get <COMMENT_ID>
linctl comment update <COMMENT_ID> --body "Updated note"
linctl comment delete <COMMENT_ID>

# Update a workflow status
linctl team status-update <STATE_ID> --name "Ready" --color "#00ff00"

# Mention delegated/active agent with message
linctl agent mention LIN-123 "Please pick this up."
```

## Troubleshooting

- `Not authenticated`: run `linctl auth` then `linctl auth status`.
- Empty or partial lists: remove default filters (`--newer-than all_time`, `--include-completed`).
- `unknown command` / `unknown flag`: binary is old; verify `linctl --version` and reinstall/upgrade.
- Validation errors on flags: run the exact subcommand help and retry:
  - `linctl issue update --help`
  - `linctl project create --help`
  - `linctl comment update --help`
  - `linctl label create --help`
  - `linctl agent mention --help`

## Minimal Discovery Commands

```bash
linctl --help
linctl issue --help
linctl project --help
linctl team --help
linctl user --help
linctl comment --help
linctl label --help
linctl agent --help
linctl auth --help
```
