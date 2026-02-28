# 🚀 linctl - Linear CLI Tool


A comprehensive command-line interface for Linear's API, built with Go and Cobra.

## ✨ Features

- 🔐 **Authentication**: Personal API Key support
- 📋 **Issue Management**: Create, list, view, update, assign, and manage issues with full details
  - Sub-issue hierarchy with parent/child relationships
  - Git branch integration showing linked branches
  - Cycle (sprint) and project associations
  - Project + project milestone assignment on create/update
  - Attachments and recent comments preview
  - Due dates, snoozed status, and completion tracking
  - Full-text search via `linctl issue search`
- 👥 **Team Management**: View teams, get team details, and list team members
- 🚀 **Project Management**: List, view, create, update, archive, and permanently delete projects
- 👤 **User Management**: List all users, view user details, and current user info
- 💬 **Comments**: List and create comments on issues with time-aware formatting
- 📎 **Attachments**: View file uploads and attachments on issues
- 🎨 **Multiple Output Formats**: Table, plaintext, and JSON output
- ⚡ **Performance**: Fast and lightweight CLI tool
- 🔄 **Flexible Sorting**: Sort lists by Linear's default order, creation date, or update date
- 📅 **Time-based Filtering**: Filter lists by creation date with intuitive time expressions
- 📚 **Built-in Documentation**: Access full documentation with `linctl docs`
- 🧪 **Smoke Testing**: Automated smoke tests for all read-only commands

## 🛠️ Installation

### Homebrew (macOS/Linux)
```bash
brew tap dorkitude/linctl
brew install linctl
linctl docs      # Render the README.md
```

### From Source
```bash
git clone https://github.com/dorkitude/linctl.git
cd linctl
make deps        # Install dependencies
make build       # Build the binary
make install     # Install to /usr/local/bin (requires sudo)
linctl docs      # Render the README.md
```

### For Development
```bash
git clone https://github.com/dorkitude/linctl.git
cd linctl
make deps        # Install dependencies
go run main.go   # Run directly without building
make dev         # Or build and run in development mode
make test        # Run smoke tests
make lint        # Run linter
make fmt         # Format code
linctl docs      # Render the README.md
```

## Important: Default Filters

**By default, `issue list`, `issue search`, and `project list` commands only show items created in the last 6 months!**
 
This improves performance and prevents overwhelming data loads. To see older items:
 - Use `--newer-than 1_year_ago` for items from the last year
 - Use `--newer-than all_time` to see ALL items ever created
 - See the [Time-based Filtering](#-time-based-filtering) section for details

**By default, `issue list` and `issue search` also filter out canceled and completed items. To see all items, use the `--include-completed` flag.**
- Need archived matches? Add `--include-archived` when using `issue search`.


## 🚀 Quick Start

> **IMPORTANT**  Agents like Claude Code, Cursor, and Gemini should use the `--json` flag on all read operations.

### 1. Authentication
```bash
# Interactive authentication
linctl auth

# Check authentication status
linctl auth status

# Show current user
linctl whoami

# View full documentation
linctl docs | less
```

### 2. Issue Management
```bash
# List all issues
linctl issue list

# List issues assigned to you
linctl issue list --assignee me

# List issues in a specific state
linctl issue list --state "In Progress"

# List issues sorted by update date
linctl issue list --sort updated

# Search issues using Linear's full-text index (shares the same filters as list)
linctl issue search "login bug" --team ENG
linctl issue search "customer:" --include-completed --include-archived

# List recent issues (last 2 weeks instead of default 6 months)
linctl issue list --newer-than 2_weeks_ago

# List ALL issues ever created (override 6-month default)
linctl issue list --newer-than all_time

# List today's issues
linctl issue list --newer-than 1_day_ago

# Get issue details (now includes git branch, cycle, project, attachments, and comments)
linctl issue get LIN-123

# Create a new issue
linctl issue create --title "Bug fix" --team ENG
linctl issue create --title "Bug fix" --team ENG --project "Q1 Platform"
linctl issue create --title "Bug fix" --team ENG --project "Q1 Platform" --project-milestone "Phase 1"

# Assign issue to yourself
linctl issue assign LIN-123

# Update issue fields
linctl issue update LIN-123 --title "New title"
linctl issue update LIN-123 --description "Updated description"
linctl issue update LIN-123 --assignee john.doe@company.com
linctl issue update LIN-123 --assignee me  # Assign to yourself
linctl issue update LIN-123 --assignee unassigned  # Remove assignee
linctl issue update LIN-123 --state "In Progress"
linctl issue update LIN-123 --priority 1  # 0=None, 1=Urgent, 2=High, 3=Normal, 4=Low
linctl issue update LIN-123 --due-date "2024-12-31"
linctl issue update LIN-123 --due-date ""  # Remove due date
linctl issue update LIN-123 --project "Q1 Platform"
linctl issue update LIN-123 --project "none"  # Remove project assignment
linctl issue update LIN-123 --project "Q1 Platform" --project-milestone "Phase 1"
linctl issue update LIN-123 --project-milestone "none"  # Remove project milestone
linctl issue update LIN-123 --parent LIN-100
linctl issue update LIN-123 --parent none  # Remove parent

# Set up parent-child issue relationships
linctl issue update LIN-124 --parent LIN-123  # Make LIN-124 a sub-issue of LIN-123
linctl issue update LIN-125 --parent LIN-123  # Make LIN-125 also a sub-issue

# Remove parent-child relationships
linctl issue update LIN-124 --parent none

# Update multiple fields at once
linctl issue update LIN-123 --title "Critical Bug" --assignee me --priority 1 --project "Q1 Platform"
```

### 3. Project Management
```bash
# List all projects (shows IDs)
linctl project list

# Filter projects by state
linctl project list --state started

# List projects created in the last month (instead of default 6 months)
linctl project list --newer-than 1_month_ago

# List ALL projects regardless of age
linctl project list --newer-than all_time

# Get project details (use ID from list command)
linctl project get 65a77a62-ec5e-491e-b1d9-84aebee01b33

# Create a project
linctl project create --name "Q1 Platform" --team ENG --state started

# Update a project
linctl project update PROJECT-ID --lead me --target-date 2026-06-30

# Archive or delete a project
linctl project delete PROJECT-ID
linctl project delete PROJECT-ID --permanent --force
```

### 4. Team Management
```bash
# List all teams
linctl team list

# Get team details
linctl team get ENG

# List team members
linctl team members ENG
```

### 5. User Management
```bash
# List all users
linctl user list

# Show only active users
linctl user list --active

# Get user details by email
linctl user get john@example.com

# Show your own profile
linctl user me
```

### 6. Comments
```bash
# List comments on an issue
linctl comment list LIN-123

# Add a comment to an issue
linctl comment create LIN-123 --body "Fixed the authentication bug"
```

## 📖 Command Reference

### Global Flags
- `--plaintext, -p`: Plain text output (non-interactive)
- `--json, -j`: JSON output for scripting
- `--help, -h`: Show help
- `--version, -v`: Show version

### Authentication Commands
```bash
linctl auth               # Interactive authentication
linctl auth login         # Same as above
linctl auth status        # Check authentication status
linctl auth logout        # Clear stored credentials
linctl whoami            # Show current user
```

### Issue Commands
```bash
# List issues with filters
linctl issue list [flags]
linctl issue ls [flags]     # Short alias

# Flags:
  -a, --assignee string     Filter by assignee (email or 'me')
  -c, --include-completed   Include completed and canceled issues
  -s, --state string       Filter by state name
  -t, --team string        Filter by team key
  -r, --priority int       Filter by priority (0-4, default: -1)
  -l, --limit int          Maximum results (default 50)
  -o, --sort string        Sort order: linear (default), created, updated
  -n, --newer-than string  Show items created after this time (default: 6_months_ago, use 'all_time' for no filter)

# Get issue details (shows parent and sub-issues)
linctl issue get <issue-id>
linctl issue show <issue-id>  # Alias

# Create issue
linctl issue create [flags]
linctl issue new [flags]      # Alias
# Flags:
  --title string           Issue title (required)
  -d, --description string Issue description
  -t, --team string        Team key (required)
  --priority int       Priority 0-4 (default 3)
  -m, --assign-me          Assign to yourself
  --project string         Project name or ID to assign the issue to
  --project-milestone string  Project milestone name or ID (requires --project)

# Assign issue to yourself
linctl issue assign <issue-id>

# Update issue
linctl issue update <issue-id> [flags]
linctl issue edit <issue-id> [flags]    # Alias
# Flags:
  --title string           New title
  -d, --description string New description
  -a, --assignee string    Assignee (email, name, 'me', or 'unassigned')
  -s, --state string       State name (e.g., 'Todo', 'In Progress', 'Done')
  --priority int           Priority (0=None, 1=Urgent, 2=High, 3=Normal, 4=Low)
  --due-date string        Due date (YYYY-MM-DD format, or empty to remove)
  --parent string          Parent issue ID/identifier (or 'none' to remove parent)
  --project string         Project name or ID (or 'none' to remove project assignment)
  --project-milestone string  Project milestone name or ID (or 'none' to remove milestone)

```

### Team Commands
```bash
# List all teams with issue counts
linctl team list
linctl team ls              # Alias
# Flags:
  -l, --limit int          Maximum results (default 50)
  -o, --sort string        Sort order: linear (default), created, updated

# Get team details
linctl team get <team-key>
linctl team show <team-key> # Alias

# Examples:
linctl team get ENG         # Shows Engineering team details
linctl team get DESIGN      # Shows Design team details

# List team members with roles and status
linctl team members <team-key>

# Examples:
linctl team members ENG     # Lists all Engineering team members
```

### Project Commands
```bash
# List projects
linctl project list [flags]
linctl project ls [flags]     # Alias
# Flags:
  -t, --team string        Filter by team key
  -s, --state string       Filter by state (planned, started, paused, completed, canceled)
  -l, --limit int          Maximum results (default 50)
  -o, --sort string        Sort order: linear (default), created, updated
  -n, --newer-than string  Show items created after this time (default: 6_months_ago)
  -c, --include-completed  Include completed and canceled projects

# Get project details
linctl project get <project-id>
linctl project show <project-id>  # Alias

# Create project
linctl project create [flags]
# Key flags:
  --name string          Project name (required)
  -t, --team strings     Team key(s) (required)
  -d, --description      Description
  -s, --state            planned|started|paused
  --lead                 email|name|me
  --start-date           YYYY-MM-DD
  --target-date          YYYY-MM-DD
  --color                Hex color

# Update project
linctl project update <project-id> [flags]
# Key flags:
  --name string
  -d, --description
  -s, --state            planned|started|paused|completed|canceled
  --lead                 email|name|me|none
  --start-date           YYYY-MM-DD or empty to clear
  --target-date          YYYY-MM-DD or empty to clear
  --color                Hex color

# Delete/archive project
linctl project delete <project-id> [flags]
linctl project rm <project-id> [flags]
linctl project remove <project-id> [flags]
# Key flags:
  --permanent            Permanently delete instead of archive
  -f, --force            Skip confirmation prompt
```

### User Commands
```bash
# List all users in workspace
linctl user list [flags]
linctl user ls [flags]      # Alias
# Flags:
  -a, --active             Show only active users
  -l, --limit int          Maximum results (default 50)
  -o, --sort string        Sort order: linear (default), created, updated

# Examples:
linctl user list            # List all users
linctl user list --active   # List only active users

# Get user details by email
linctl user get <email>
linctl user show <email>    # Alias

# Examples:
linctl user get john@example.com
linctl user get jane.doe@company.com

# Show current authenticated user
linctl user me              # Shows your profile with admin status
```

### Comment Commands
```bash
# List all comments for an issue
linctl comment list <issue-id> [flags]
linctl comment ls <issue-id> [flags]    # Alias
# Flags:
  -l, --limit int          Maximum results (default 50)
  -o, --sort string        Sort order: linear (default), created, updated

# Examples:
linctl comment list LIN-123      # Shows all comments with timestamps
linctl comment list LIN-456 -l 10 # Show latest 10 comments

# Add comment to issue
linctl comment create <issue-id> --body "Comment text"
linctl comment add <issue-id> -b "Comment text"    # Alias
linctl comment new <issue-id> -b "Comment text"    # Alias

# Examples:
linctl comment create LIN-123 --body "I've started working on this"
linctl comment add LIN-123 -b "Fixed in commit abc123"
linctl comment create LIN-456 --body "@john please review this PR"
```

## 🎨 Output Formats

### Table Format (Default)
```bash
linctl issue list
```
```
ID       Title                State        Assignee    Team  Priority
LIN-123  Fix authentication   In Progress  john@co.com ENG   High
LIN-124  Update documentation Done         jane@co.com DOC   Normal
```

### Plaintext Format
```bash
linctl issue list --plaintext
```
```
# Issues
## BUG: Fix login button alignment
- **ID**: FAK-123
- **State**: In Progress
- **Assignee**: Jane Doe
- **Team**: WEB
- **Created**: 2025-07-12
- **URL**: https://linear.app/example/issue/FAK-123/bug-fix-login-button-alignment
- **Description**: The login button on the main page is misaligned on mobile devices.

Steps to reproduce:
1. Open the website on a mobile browser.
2. Navigate to the login page.
3. Observe the button alignment.

## FEAT: Add dark mode support
- **ID**: FAK-124
- **State**: Todo
- **Assignee**: John Smith
- **Team**: APP
- **Created**: 2025-07-11
- **URL**: https://linear.app/example/issue/FAK-124/feat-add-dark-mode-support
- **Description**: Implement a dark mode theme for the entire application to improve user experience in low-light environments.
```

### JSON Format
```bash
linctl issue list --json
```
```json
[
  {
    "id": "LIN-123",
    "title": "Fix authentication",
    "state": "In Progress",
    "assignee": "john@co.com",
    "team": "ENG",
    "priority": "High"
  }
]
```

## ⚙️ Configuration

Configuration is stored in `~/.linctl.yaml`:

```yaml
# Default output format
output: table

# Default pagination limit
limit: 50

# API settings
api:
  timeout: 30s
  retries: 3
```

Authentication credentials are stored securely in `~/.linctl-auth.json`.

## 🔒 Authentication

### Personal API Key (Recommended)
1. Go to [Linear Settings > API](https://linear.app/settings/api)
2. Create a new Personal API Key
3. Run `linctl auth` and paste your key

### Temporary Override (Current Session)

You can temporarily override your stored credentials using the `LINCTL_API_KEY` environment variable. This is useful for:
- CI/CD pipelines
- Testing with a different account
- Running one-off commands without changing your stored credentials

```bash
# Override for a single command
LINCTL_API_KEY="lin_api_..." linctl issue list

# Override for the current shell session
export LINCTL_API_KEY="lin_api_..."
linctl issue list
linctl whoami  # Shows the user for the env var token

# Clear the override
unset LINCTL_API_KEY
```

**Precedence:** Environment variable > Config file (`~/.linctl-auth.json`)

## 📅 Time-based Filtering

**⚠️ Default Behavior**: To improve performance and prevent overwhelming data loads, list commands **only show items created in the last 6 months by default**. This is especially important for large workspaces.

### Using the --newer-than Flag

The `--newer-than` (or `-n`) flag is available on `issue list` and `project list` commands:

```bash
# Default behavior (last 6 months)
linctl issue list

# Show items from a specific time period
linctl issue list --newer-than 2_weeks_ago
linctl project list --newer-than 1_month_ago

# Show ALL items regardless of age
linctl issue list --newer-than all_time
```

### Supported Time Formats

1. **Relative time expressions**: `N_units_ago`
   - Units: `minutes`, `hours`, `days`, `weeks`, `months`, `years`
   - Examples: `30_minutes_ago`, `2_hours_ago`, `3_days_ago`, `1_week_ago`, `6_months_ago`

2. **Special values**:
   - `all_time` - Shows all items without any date filter
   - ISO dates - `2025-07-01` or `2025-07-01T15:30:00Z`

3. **Default value**: `6_months_ago` (when flag is not specified)

### Quick Reference

| Time Expression | Description | Example Command |
|----------------|-------------|-----------------|
| *(no flag)* | Last 6 months (default) | `linctl issue list` |
| `1_day_ago` | Last 24 hours | `linctl issue list --newer-than 1_day_ago` |
| `1_week_ago` | Last 7 days | `linctl issue list --newer-than 1_week_ago` |
| `2_weeks_ago` | Last 14 days | `linctl issue list --newer-than 2_weeks_ago` |
| `1_month_ago` | Last month | `linctl issue list --newer-than 1_month_ago` |
| `3_months_ago` | Last quarter | `linctl issue list --newer-than 3_months_ago` |
| `6_months_ago` | Last 6 months | `linctl issue list --newer-than 6_months_ago` |
| `1_year_ago` | Last year | `linctl issue list --newer-than 1_year_ago` |
| `all_time` | No date filter | `linctl issue list --newer-than all_time` |
| `2025-07-01` | Since specific date | `linctl issue list --newer-than 2025-07-01` |

### Common Use Cases

```bash
# Recent activity - issues from last week
linctl issue list --newer-than 1_week_ago

# Sprint planning - issues from current month
linctl issue list --newer-than 1_month_ago --state "Todo"

# Quarterly review - all projects from last 3 months
linctl project list --newer-than 3_months_ago

# Historical analysis - ALL issues ever created
linctl issue list --newer-than all_time --sort created

# Today's issues
linctl issue list --newer-than 1_day_ago

# Combine with other filters
linctl issue list --newer-than 2_weeks_ago --assignee me --sort updated
```

## 🔄 Sorting Options

All list commands support sorting with the `--sort` or `-o` flag:

- **linear** (default): Linear's built-in sorting order (respects manual ordering in the UI)
- **created**: Sort by creation date (newest first)
- **updated**: Sort by last update date (most recently updated first)

### Examples
```bash
# Get recently updated issues
linctl issue list --sort updated

# Get oldest projects first
linctl project list --sort created

# Get recently joined users
linctl user list --sort created --active

# Get latest comments on an issue
linctl comment list LIN-123 --sort created

# Combine sorting with filters
linctl issue list --assignee me --state "In Progress" --sort updated

# Combine time filtering with sorting
linctl issue list --newer-than 1_week_ago --sort updated

# Get all projects sorted by creation date
linctl project list --newer-than all_time --sort created
```

### Performance Tips

- The 6-month default filter significantly improves performance for large workspaces
- Use specific time ranges when possible instead of `all_time`
- Combine time filtering with other filters (assignee, state, team) for faster results

## 🧪 Testing

Current test paths:

```bash
# Go unit tests
go test ./...

# CLI smoke tests
make test
# or:
./smoke_test.sh
```

## 🤖 Scripting & Automation

Use `--plaintext` or `--json` flags for scripting:

```bash
#!/bin/bash

# Get all urgent issues in JSON format
urgent_issues=$(linctl issue list --priority 1 --json)

# Parse with jq
echo "$urgent_issues" | jq '.[] | select(.assignee == "me") | .id'

# Plaintext output for simple parsing
linctl issue list --assignee me --plaintext | cut -f1 | tail -n +2

# Get issue count for different time periods
echo "Last week: $(linctl issue list --newer-than 1_week_ago --json | jq '. | length')"
echo "Last month: $(linctl issue list --newer-than 1_month_ago --json | jq '. | length')"
echo "All time: $(linctl issue list --newer-than all_time --json | jq '. | length')"

# Create and assign issue in one command
linctl issue create --title "Fix bug" --team ENG --assign-me --json

# Get all projects with progress
linctl project list --json | jq '.[] | {name, progress}'

# List all admin users
linctl user list --json | jq '.[] | select(.admin == true) | {name, email}'

# Get team member count
linctl team members ENG --json | jq '. | length'

# Export issue comments
linctl comment list LIN-123 --json > issue-comments.json
```

## 📡 Real-World Examples

### Team Workflows
```bash
# Find which team a user belongs to
for team in $(linctl team list --json | jq -r '.[].key'); do
  echo "Checking team: $team"
  linctl team members $team --json | jq '.[] | select(.email == "john@example.com")'
done

# List all private teams
linctl team list --json | jq '.[] | select(.private == true) | {key, name}'

# Get teams with more than 50 issues
linctl team list --json | jq '.[] | select(.issueCount > 50) | {key, name, issueCount}'
```

### User Management
```bash
# Find inactive users
linctl user list --json | jq '.[] | select(.active == false) | {name, email}'

# Check if you're an admin
linctl user me --json | jq '.admin'

# List users who are admins but not the current user
linctl user list --json | jq '.[] | select(.admin == true and .isMe == false) | .email'
```

### Issue Comments
```bash
# Add a comment mentioning the issue is blocked
linctl comment create LIN-123 --body "Blocked by LIN-456. Waiting for API changes."

# Get all comments by a specific user
linctl comment list LIN-123 --json | jq '.[] | select(.user.email == "john@example.com") | .body'

# Count comments per issue
for issue in LIN-123 LIN-124 LIN-125; do
  count=$(linctl comment list $issue --json | jq '. | length')
  echo "$issue: $count comments"
done
```

### Project Tracking
```bash
# List projects nearing completion (>80% progress)
linctl project list --json | jq '.[] | select(.progress > 0.8) | {name, progress}'

# Get all paused projects
linctl project list --state paused

# Show project timeline
linctl project get PROJECT-ID --json | jq '{name, startDate, targetDate, progress}'
```

### Daily Standup Helper
```bash
#!/bin/bash
# Show my recent activity
echo "=== My Issues ==="
linctl issue list --assignee me --limit 10

echo -e "\n=== Recent Comments ==="
for issue in $(linctl issue list --assignee me --json | jq -r '.[].identifier'); do
  echo "Comments on $issue:"
  linctl comment list $issue --limit 3
done
```

## 🐛 Troubleshooting

### Authentication Issues
```bash
# Check authentication status
linctl auth status

# Re-authenticate
linctl auth logout
linctl auth
```

### API Rate Limits
Linear has the following rate limits:
- Personal API Keys: 5,000 requests/hour

### Common Errors
- `Not authenticated`: Run `linctl auth` first
- `Team not found`: Use team key (e.g., "ENG") not display name
- `Invalid priority`: Use numbers 0-4 (0=None, 1=Urgent, 2=High, 3=Normal, 4=Low)

### Time Filtering Issues
- **Missing old issues?** Remember that list commands default to showing only the last 6 months
  - Solution: Use `--newer-than all_time` to see all issues
- **Invalid time expression?** Check the format: `N_units_ago` (e.g., `3_weeks_ago`)
  - Valid units: `minutes`, `hours`, `days`, `weeks`, `months`, `years`
- **Performance issues?** Avoid using `all_time` on large workspaces
  - Solution: Use specific time ranges like `--newer-than 1_year_ago`

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

See CONTRIBUTING.md for a detailed release checklist and the Homebrew tap auto-bump workflow.

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🔗 Links

- [Linear API Documentation](https://developers.linear.app/)
- [GitHub Repository](https://github.com/dorkitude/linctl)
- [Issue Tracker](https://github.com/dorkitude/linctl/issues)

---

**Built with ❤️ using Go, Cobra, and the Linear API**
