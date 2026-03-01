# Linear API Master Reference

Comprehensive, practical reference for Linear's GraphQL API and how `linctl` maps onto it.

Last verified against official docs: March 1, 2026.

## Canonical Sources
- Developers hub: https://linear.app/developers
- GraphQL endpoint: `https://api.linear.app/graphql`
- Schema explorer: https://studio.apollographql.com/public/Linear-API/variant/current/explorer
- Changelog/deprecations: https://linear.app/changelog and https://linear.app/developers/deprecations

## 1) Authentication

### Personal API key (common for CLI)
- Header format: `Authorization: <API_KEY>`
- Good fit for local automation and operators.

### OAuth 2.0
- Header format: `Authorization: Bearer <ACCESS_TOKEN>`
- Supported grants include user authorization flow and app-actor flow (`actor=app`).
- Modern app capabilities include app actor tokens and app scopes.

### Current scope model (OAuth)
Common scopes (workspace-specific):
- `read`
- `write`
- `issues:create`
- `comments:create`
- `timeSchedule:write`
- `admin`
- `app:assignable`
- `app:mentionable`

Notes:
- Older scope lists like `issues:write` / `teams:read` are stale.
- For older apps, follow OAuth migration guidance (refresh-token / actor updates) in the official docs.

## 2) GraphQL Conventions

### Endpoint
```text
https://api.linear.app/graphql
```

### Introspection
```graphql
query IntrospectionQuery {
  __schema {
    queryType { name }
    mutationType { name }
    types { name }
  }
}
```

### IDs, keys, identifiers
- Most node lookups accept a Linear ID (`id`).
- Some API paths use domain keys/identifiers (for example team key like `ENG`, issue identifier like `ENG-123`) after resolving via query logic.
- In `linctl`, user-facing commands often accept keys/identifiers and resolve IDs internally.

### Pagination pattern
- Connections use cursor pagination.
- Standard args: `first`, `after`, `last`, `before`.
- Standard response: `nodes` + `pageInfo { hasNextPage endCursor ... }`.

### Filtering + sorting
- Filtering is typed per resource (`IssueFilter`, `ProjectFilter`, etc.).
- Sorting is typed per resource (`IssueOrderBy`, `ProjectOrderBy`, `TeamOrderBy`, `UserOrderBy`, `CommentOrderBy`).
- Do not assume global `PaginationOrderBy` for every query.

## 3) Rate Limiting (Current Model)

Linear enforces both request-count and complexity budgets.

### Request budgets
- API key: 5,000 requests/hour
- OAuth key: 5,000 requests/hour
- Unauthenticated: 60 requests/hour
- Endpoint burst protection: 30 requests/minute per endpoint

### Complexity budgets
- Max single request complexity: 10,000
- API key hourly complexity budget: 3,000,000
- OAuth hourly complexity budget: 2,000,000
- Unauthenticated hourly complexity budget: 10,000

Always trust live response headers for enforcement details in your workspace.

### Key response headers
Request-based:
- `X-RateLimit-Requests-Limit`
- `X-RateLimit-Requests-Remaining`
- `X-RateLimit-Requests-Reset`

Complexity-based:
- `X-RateLimit-Complexity-Limit`
- `X-RateLimit-Complexity-Remaining`
- `X-RateLimit-Complexity-Reset`

Endpoint-specific (when applicable):
- `X-RateLimit-Endpoint-Requests-Limit`
- `X-RateLimit-Endpoint-Requests-Remaining`
- `X-RateLimit-Endpoint-Requests-Reset`

## 4) Core Domain Coverage

Linear's GraphQL schema is broad and evolving. This section lists the practical domains you should expect to query/mutate.

### Issues
- List/search/get issues
- Create/update issues
- Assign/delegate, set labels, parent/sub-issue links
- State transitions (workflow state)
- Attachments, comments, SLA/triage related metadata

Representative examples:
```graphql
query Issues($filter: IssueFilter, $orderBy: IssueOrderBy, $first: Int, $after: String) {
  issues(filter: $filter, orderBy: $orderBy, first: $first, after: $after) {
    nodes {
      id
      identifier
      title
      priority
      state { id name type color }
      team { id key name }
      assignee { id name email }
      createdAt
      updatedAt
    }
    pageInfo { hasNextPage endCursor }
  }
}
```

```graphql
mutation IssueUpdate($id: String!, $input: IssueUpdateInput!) {
  issueUpdate(id: $id, input: $input) {
    success
    issue { id identifier title }
  }
}
```

### Projects + milestones
- List/get/create/update/archive/delete projects
- Manage lead, state, target dates, teams
- Query milestones and link issues to milestones

```graphql
query Projects($filter: ProjectFilter, $orderBy: ProjectOrderBy, $first: Int, $after: String) {
  projects(filter: $filter, orderBy: $orderBy, first: $first, after: $after) {
    nodes {
      id
      name
      state
      progress
      startDate
      targetDate
      lead { id name email }
      teams { nodes { id key name } }
    }
    pageInfo { hasNextPage endCursor }
  }
}
```

### Initiatives + roadmap (modern planning layer)
- Roadmap entities exist beyond projects (initiatives and updates).
- Use schema explorer for exact query/mutation names and current fields in your workspace tier.

### Customers + requests (modern product feedback layer)
- Customer and customer-request objects are first-class API surfaces.
- Webhooks include customer/customerRequest events.

### Teams + workflow states
- List/get teams
- List team members
- List/update team workflow statuses

```graphql
query TeamStates($key: String!) {
  team(id: $key) {
    states {
      nodes {
        id
        name
        type
        color
        description
        position
      }
    }
  }
}
```

```graphql
mutation WorkflowStateUpdate($id: String!, $input: WorkflowStateUpdateInput!) {
  workflowStateUpdate(id: $id, input: $input) {
    success
    workflowState { id name type color description position }
  }
}
```

### Users
- Viewer/current-user query
- Workspace user listing and lookup
- Admin/active/guest metadata

### Labels
- Team label listing
- Create/update/delete labels
- Parent/group label relationships

### Comments
- List by issue
- Create/get/update/delete comment

### Attachments
- Create URL/PR attachments
- Rich metadata supported via attachment input payloads

### Agent sessions + mentions
- Agent session/delegation state is queryable from issue context.
- Mentioning an agent from issue context is supported (with app scopes for agent actors where applicable).

### Webhooks
- Manage webhook endpoints via API.
- Event coverage includes classic issue/project/comment events plus modern domains (initiative, customer, customerRequest, issue SLA, and more).

## 5) Webhooks: Practical Notes

### Representative query
```graphql
query Webhooks($first: Int, $after: String) {
  webhooks(first: $first, after: $after) {
    nodes {
      id
      url
      label
      enabled
      resourceTypes
    }
    pageInfo { hasNextPage endCursor }
  }
}
```

### Operational guidance
- Verify signatures and reject unsigned payloads.
- Design handlers to be idempotent.
- Expect retries and out-of-order arrivals.
- Persist webhook event IDs to suppress duplicates.

## 6) Filtering + Search Patterns

### Typical issue filter shape
```graphql
{
  team: { id: { eq: "TEAM_ID" } }
  assignee: { id: { eq: "USER_ID" } }
  state: { name: { eq: "In Progress" } }
  priority: { eq: 1 }
  createdAt: { gte: "2026-01-01T00:00:00Z" }
}
```

### Common comparators
- `eq`, `neq`
- `in`, `nin`
- `contains`, `containsIgnoreCase`
- `startsWith`, `endsWith`
- date/time comparators such as `gte`, `lte`

Always validate exact comparator support per field/type in the schema explorer.

## 7) linctl -> API Mapping

`linctl` covers a focused subset of the full GraphQL surface.

### Auth
```bash
linctl auth
linctl auth login
linctl auth status
linctl auth logout
linctl whoami
```

### Raw GraphQL
```bash
linctl graphql [query]
linctl graphql --query 'query { viewer { id } }'
linctl graphql --file query.graphql --variables-file vars.json
```

### Issues
```bash
linctl issue list|ls
linctl issue search|find
linctl issue get|show
linctl issue create|new
linctl issue update
linctl issue assign
linctl issue attach
```

### Projects
```bash
linctl project list|ls
linctl project get|show
linctl project create|new
linctl project update
linctl project delete|rm|remove
```

### Teams
```bash
linctl team list|ls
linctl team get|show
linctl team members
linctl team statuses|states
linctl team status-update|state-update
```

### Users
```bash
linctl user list|ls
linctl user get|show
linctl user me
```

### Labels
```bash
linctl label list|ls
linctl label get
linctl label create
linctl label update
linctl label delete|rm|remove
```

### Comments
```bash
linctl comment list|ls
linctl comment create|add|new
linctl comment get|show
linctl comment update|edit
linctl comment delete|rm|remove
```

### Agents
```bash
linctl agent <issue-id>
linctl agent mention <issue-id> [message...]
```

### Global flags
- `--json, -j`
- `--plaintext, -p`
- `--config`
- `--help, -h`
- `--version, -v`

## 8) Gaps Between Full API and linctl

Not all modern Linear API domains are exposed in `linctl` commands yet (for example deeper roadmap/customer/admin/app management surfaces). For unsupported flows:
1. Query schema explorer for exact operation and input names.
2. Execute the operation via `linctl graphql` so auth/config/output stay in one tool.
3. Add CLI coverage only when workflow value is clear.

## 9) Reliability Checklist

Before shipping CLI/API docs changes:
1. Validate claims against official docs and schema explorer.
2. Run `go test ./...`.
3. For command changes, update `README.md`, `SKILL.md`, and this file in the same PR.
4. Avoid hardcoding stale limits/scopes unless source-linked.

## Sources
- https://linear.app/developers/graphql
- https://linear.app/developers/oauth-2-0-authentication
- https://linear.app/developers/oauth-actor-authorization
- https://linear.app/developers/rate-limiting
- https://linear.app/developers/filtering
- https://linear.app/developers/pagination
- https://linear.app/developers/webhooks
- https://linear.app/developers/attachments
- https://linear.app/developers/agents
- https://linear.app/developers/managing-customers
- https://linear.app/developers/deprecations
- https://linear.app/changelog
