---
name: linctl
description: Use linctl to read and update Linear issues, projects, teams, users, and comments from the terminal. Prefer JSON reads, validate auth first, and use command-specific --help for exact flags.
---

# linctl Agent Guide

Use this skill when the user wants you to inspect or modify Linear data through `linctl`.

## Quick Rules

- Always verify auth before substantive work: `linctl auth status` or `linctl whoami`.
- For read operations, prefer `--json` and parse results with `jq` when needed.
- Before writing, inspect current state first (`get`/`list --json`).
- Use command-specific help for exact flags and validation rules: `linctl <command> <subcommand> --help`.
- Be explicit with filters; defaults can hide expected results.

## High-Impact Gotchas

- `issue list`, `issue search`, and `project list` default to `--newer-than 6_months_ago`.
- `issue list` and `issue search` exclude completed/canceled by default.
- `issue search` may also need `--include-archived` for archived matches.
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
4. Apply changes with `create/update/assign/comment` commands.
5. Re-read target and summarize concrete outcome to the user.

## Command Map

- Issues: `linctl issue ...`
- Projects: `linctl project ...`
- Teams: `linctl team ...`
- Users: `linctl user ...` and `linctl whoami`
- Comments: `linctl comment ...`
- Auth: `linctl auth ...`

## Suggested Read Patterns

```bash
# Issues assigned to me, no date-limit surprise
linctl issue list --assignee me --newer-than all_time --include-completed --json

# Find issue by text, including archived/completed
linctl issue search "<query>" --newer-than all_time --include-completed --include-archived --json

# Project inventory without default date filter
linctl project list --newer-than all_time --json
```

## Suggested Write Patterns

```bash
# Update an issue after inspecting it
linctl issue get LIN-123 --json
linctl issue update LIN-123 --state "In Progress" --assignee me
linctl issue get LIN-123 --json

# Add execution note
linctl comment create LIN-123 --body "Implemented and verified locally."
```

## Troubleshooting

- `Not authenticated`: run `linctl auth` then `linctl auth status`.
- Empty or partial lists: remove default filters (`--newer-than all_time`, `--include-completed`).
- Validation errors on flags: run the exact subcommand help and retry:
  - `linctl issue update --help`
  - `linctl project create --help`
  - `linctl comment create --help`

## Minimal Discovery Commands

```bash
linctl --help
linctl issue --help
linctl project --help
linctl team --help
linctl user --help
linctl comment --help
linctl auth --help
```
