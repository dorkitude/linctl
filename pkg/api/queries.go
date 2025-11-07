package api

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
)

// User represents a Linear user
type User struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	AvatarURL   string     `json:"avatarUrl"`
	DisplayName string     `json:"displayName"`
	IsMe        bool       `json:"isMe"`
	Active      bool       `json:"active"`
	Admin       bool       `json:"admin"`
	CreatedAt   *time.Time `json:"createdAt"`
}

// Team represents a Linear team
type Team struct {
	ID                 string  `json:"id"`
	Key                string  `json:"key"`
	Name               string  `json:"name"`
	Description        string  `json:"description"`
	Icon               *string `json:"icon"`
	Color              string  `json:"color"`
	Private            bool    `json:"private"`
	IssueCount         int     `json:"issueCount"`
	CyclesEnabled      bool    `json:"cyclesEnabled"`
	CycleStartDay      int     `json:"cycleStartDay"`
	CycleDuration      int     `json:"cycleDuration"`
	UpcomingCycleCount int     `json:"upcomingCycleCount"`
}

// Issue represents a Linear issue
type Issue struct {
	ID                  string       `json:"id"`
	Identifier          string       `json:"identifier"`
	Title               string       `json:"title"`
	Description         string       `json:"description"`
	Priority            int          `json:"priority"`
	Estimate            *float64     `json:"estimate"`
	CreatedAt           time.Time    `json:"createdAt"`
	UpdatedAt           time.Time    `json:"updatedAt"`
	DueDate             *string      `json:"dueDate"`
	State               *State       `json:"state"`
	Assignee            *User        `json:"assignee"`
	Team                *Team        `json:"team"`
	Labels              *Labels      `json:"labels"`
	Children            *Issues      `json:"children"`
	Parent              *Issue       `json:"parent"`
	URL                 string       `json:"url"`
	BranchName          string       `json:"branchName"`
	Cycle               *Cycle       `json:"cycle"`
	Project             *Project     `json:"project"`
	Attachments         *Attachments `json:"attachments"`
	Comments            *Comments    `json:"comments"`
	SnoozedUntilAt      *time.Time   `json:"snoozedUntilAt"`
	CompletedAt         *time.Time   `json:"completedAt"`
	CanceledAt          *time.Time   `json:"canceledAt"`
	ArchivedAt          *time.Time   `json:"archivedAt"`
	TriagedAt           *time.Time   `json:"triagedAt"`
	CustomerTicketCount int          `json:"customerTicketCount"`
	PreviousIdentifiers []string     `json:"previousIdentifiers"`
	// Additional fields
	Number                int              `json:"number"`
	BoardOrder            float64          `json:"boardOrder"`
	SubIssueSortOrder     float64          `json:"subIssueSortOrder"`
	PriorityLabel         string           `json:"priorityLabel"`
	IntegrationSourceType *string          `json:"integrationSourceType"`
	Creator               *User            `json:"creator"`
	Subscribers           *Users           `json:"subscribers"`
	Relations             *IssueRelations  `json:"relations"`
	History               *IssueHistory    `json:"history"`
	Reactions             []Reaction       `json:"reactions"`
	SlackIssueComments    []SlackComment   `json:"slackIssueComments"`
	ExternalUserCreator   *ExternalUser    `json:"externalUserCreator"`
	CustomerTickets       []CustomerTicket `json:"customerTickets"`
}

// State represents an issue state
type State struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Color       string  `json:"color"`
	Description *string `json:"description"`
	Position    float64 `json:"position"`
}

// Project represents a Linear project
type Project struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	State       string     `json:"state"`
	Priority    int        `json:"priority"`
	Progress    float64    `json:"progress"`
	StartDate   *string    `json:"startDate"`
	TargetDate  *string    `json:"targetDate"`
	Lead        *User         `json:"lead"`
	Teams       *Teams        `json:"teams"`
	Initiatives *Initiatives  `json:"initiatives"`
	Labels      *Labels       `json:"labels"`
	URL         string     `json:"url"`
	Icon        *string    `json:"icon"`
	Color       string     `json:"color"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	CompletedAt *time.Time `json:"completedAt"`
	CanceledAt  *time.Time `json:"canceledAt"`
	ArchivedAt  *time.Time `json:"archivedAt"`
	Creator     *User      `json:"creator"`
	Members     *Users     `json:"members"`
	Issues      *Issues    `json:"issues"`
	// Additional fields
	SlugId              string          `json:"slugId"`
	Content             string          `json:"content"`
	ConvertedFromIssue  *Issue          `json:"convertedFromIssue"`
	LastAppliedTemplate *Template       `json:"lastAppliedTemplate"`
	ProjectUpdates      *ProjectUpdates `json:"projectUpdates"`
	Documents           *Documents      `json:"documents"`
	Health              string          `json:"health"`
	Scope               int             `json:"scope"`
	SlackNewIssue       bool            `json:"slackNewIssue"`
	SlackIssueComments  bool            `json:"slackIssueComments"`
    SlackIssueStatuses  bool            `json:"slackIssueStatuses"`
}

// ProjectMilestone represents a milestone within a project
type ProjectMilestone struct {
    ID          string     `json:"id"`
    Name        string     `json:"name"`
    Description string     `json:"description"`
    TargetDate  *string    `json:"targetDate"`
    Status      string     `json:"status"`
    Progress    float64    `json:"progress"`
    SortOrder   float64    `json:"sortOrder"`
    Project     *Project   `json:"project"`
    CreatedAt   time.Time  `json:"createdAt"`
    UpdatedAt   time.Time  `json:"updatedAt"`
    ArchivedAt  *time.Time `json:"archivedAt"`
}

// ProjectMilestones represents a collection of project milestones
type ProjectMilestones struct {
    Nodes []ProjectMilestone `json:"nodes"`
}

// Paginated collections
type Issues struct {
	Nodes    []Issue  `json:"nodes"`
	PageInfo PageInfo `json:"pageInfo"`
}

type Teams struct {
	Nodes    []Team   `json:"nodes"`
	PageInfo PageInfo `json:"pageInfo"`
}

type Projects struct {
	Nodes    []Project `json:"nodes"`
	PageInfo PageInfo  `json:"pageInfo"`
}

type Users struct {
	Nodes    []User   `json:"nodes"`
	PageInfo PageInfo `json:"pageInfo"`
}

type Labels struct {
	Nodes []Label `json:"nodes"`
}

type Label struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Color       string  `json:"color"`
	Description *string `json:"description"`
	Parent      *Label  `json:"parent"`
}

// Cycle represents a Linear cycle (sprint)
type Cycle struct {
	ID           string     `json:"id"`
	Number       int        `json:"number"`
	Name         string     `json:"name"`
	Description  *string    `json:"description"`
	StartsAt     string     `json:"startsAt"`
	EndsAt       string     `json:"endsAt"`
	Progress     float64    `json:"progress"`
	CompletedAt  *time.Time `json:"completedAt"`
	ScopeHistory []float64  `json:"scopeHistory"`
}

// Attachment represents a file attachment or link
type Attachment struct {
	ID        string                 `json:"id"`
	Title     string                 `json:"title"`
	Subtitle  *string                `json:"subtitle"`
	URL       string                 `json:"url"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"createdAt"`
	Creator   *User                  `json:"creator"`

	// Use a map to capture any extra fields Linear might return
	Extra map[string]interface{} `json:"-"`
}

// UnmarshalJSON implements custom unmarshaling to handle unexpected fields from Linear API
func (a *Attachment) UnmarshalJSON(data []byte) error {
	// Create an alias to avoid infinite recursion
	type Alias Attachment
	aux := &struct {
		*Alias
		// Capture extra fields that might come from Linear
		Source     interface{} `json:"source,omitempty"`
		SourceType interface{} `json:"sourceType,omitempty"`
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Store unexpected fields in Extra map if needed
	if aux.Source != nil || aux.SourceType != nil {
		a.Extra = make(map[string]interface{})
		if aux.Source != nil {
			a.Extra["source"] = aux.Source
		}
		if aux.SourceType != nil {
			a.Extra["sourceType"] = aux.SourceType
		}
	}

	return nil
}

// Attachments represents a paginated list of attachments
type Attachments struct {
	Nodes []Attachment `json:"nodes"`
}

// Initiative represents a Linear initiative
type Initiative struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Initiatives struct {
	Nodes []Initiative `json:"nodes"`
}

type PageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

// Additional types for expanded fields
type IssueRelations struct {
	Nodes []IssueRelation `json:"nodes"`
}

type IssueRelation struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Issue        *Issue `json:"issue"`
	RelatedIssue *Issue `json:"relatedIssue"`
}

type IssueHistory struct {
	Nodes []IssueHistoryEntry `json:"nodes"`
}

type IssueHistoryEntry struct {
	ID              string    `json:"id"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	Changes         string    `json:"changes"`
	Actor           *User     `json:"actor"`
	FromAssignee    *User     `json:"fromAssignee"`
	ToAssignee      *User     `json:"toAssignee"`
	FromState       *State    `json:"fromState"`
	ToState         *State    `json:"toState"`
	FromPriority    *int      `json:"fromPriority"`
	ToPriority      *int      `json:"toPriority"`
	FromTitle       *string   `json:"fromTitle"`
	ToTitle         *string   `json:"toTitle"`
	FromCycle       *Cycle    `json:"fromCycle"`
	ToCycle         *Cycle    `json:"toCycle"`
	FromProject     *Project  `json:"fromProject"`
	ToProject       *Project  `json:"toProject"`
	AddedLabelIds   []string  `json:"addedLabelIds"`
	RemovedLabelIds []string  `json:"removedLabelIds"`
}

type Reaction struct {
	ID        string    `json:"id"`
	Emoji     string    `json:"emoji"`
	User      *User     `json:"user"`
	CreatedAt time.Time `json:"createdAt"`
}

type SlackComment struct {
	ID   string `json:"id"`
	Body string `json:"body"`
}

type ExternalUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CustomerTicket struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	CreatedAt  time.Time `json:"createdAt"`
	ExternalId string    `json:"externalId"`
}

type Template struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Milestone struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	TargetDate  *string   `json:"targetDate"`
	Projects    *Projects `json:"projects"`
}

type Roadmaps struct {
	Nodes []Roadmap `json:"nodes"`
}

type Roadmap struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Creator     *User     `json:"creator"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ProjectUpdates struct {
	Nodes []ProjectUpdate `json:"nodes"`
}

type ProjectUpdate struct {
	ID        string     `json:"id"`
	Body      string     `json:"body"`
	User      *User      `json:"user"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	EditedAt  *time.Time `json:"editedAt"`
	Health    string     `json:"health"`
}

type Documents struct {
	Nodes []Document `json:"nodes"`
}

type Document struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Icon      *string   `json:"icon"`
	Color     string    `json:"color"`
	Creator   *User     `json:"creator"`
	UpdatedBy *User     `json:"updatedBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ProjectLinks struct {
	Nodes []ProjectLink `json:"nodes"`
}

type ProjectLink struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Label     string    `json:"label"`
	Creator   *User     `json:"creator"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// GetViewer returns the current authenticated user
func (c *Client) GetViewer(ctx context.Context) (*User, error) {
	query := `
		query Me {
			viewer {
				id
				name
				email
				avatarUrl
				isMe
				active
				admin
			}
		}
	`

	var response struct {
		Viewer User `json:"viewer"`
	}

	err := c.Execute(ctx, query, nil, &response)
	if err != nil {
		return nil, err
	}

	return &response.Viewer, nil
}

// GetIssues returns a list of issues with optional filtering
func (c *Client) GetIssues(ctx context.Context, filter map[string]interface{}, first int, after string, orderBy string) (*Issues, error) {
	query := `
		query Issues($filter: IssueFilter, $first: Int, $after: String, $orderBy: PaginationOrderBy) {
			issues(filter: $filter, first: $first, after: $after, orderBy: $orderBy) {
				nodes {
					id
					identifier
					title
					description
					priority
					estimate
					createdAt
					updatedAt
					dueDate
					url
					state {
						id
						name
						type
						color
					}
					assignee {
						id
						name
						email
					}
					team {
						id
						key
						name
					}
					project {
						id
						name
					}
					labels {
						nodes {
							id
							name
							color
						}
					}
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
	`

	variables := map[string]interface{}{
		"first": first,
	}
	if filter != nil {
		variables["filter"] = filter
	}
	if after != "" {
		variables["after"] = after
	}
	if orderBy != "" {
		variables["orderBy"] = orderBy
	}

	var response struct {
		Issues Issues `json:"issues"`
	}

	err := c.Execute(ctx, query, variables, &response)
	if err != nil {
		return nil, err
	}

	return &response.Issues, nil
}

// IssueSearch returns issues that match a full-text query
func (c *Client) IssueSearch(ctx context.Context, term string, filter map[string]interface{}, first int, after string, orderBy string, includeArchived bool) (*Issues, error) {
	query := `
		query IssueSearch($term: String!, $filter: IssueFilter, $first: Int, $after: String, $orderBy: PaginationOrderBy, $includeArchived: Boolean) {
			searchIssues(term: $term, filter: $filter, first: $first, after: $after, orderBy: $orderBy, includeArchived: $includeArchived) {
				nodes {
					id
					identifier
					title
					description
					priority
					estimate
					createdAt
					updatedAt
					dueDate
					url
					state {
						id
						name
						type
						color
					}
					assignee {
						id
						name
						email
					}
					team {
						id
						key
						name
					}
					project {
						id
						name
					}
					labels {
						nodes {
							id
							name
							color
						}
					}
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
	`

	variables := map[string]interface{}{
		"term":            term,
		"first":           first,
		"includeArchived": includeArchived,
	}
	if filter != nil {
		variables["filter"] = filter
	}
	if after != "" {
		variables["after"] = after
	}
	if orderBy != "" {
		variables["orderBy"] = orderBy
	}

	var response struct {
		SearchIssues struct {
			Nodes    []Issue  `json:"nodes"`
			PageInfo PageInfo `json:"pageInfo"`
		} `json:"searchIssues"`
	}

	err := c.Execute(ctx, query, variables, &response)
	if err != nil {
		return nil, err
	}

	return &Issues{
		Nodes:    response.SearchIssues.Nodes,
		PageInfo: response.SearchIssues.PageInfo,
	}, nil
}

// GetIssue returns a single issue by ID
func (c *Client) GetIssue(ctx context.Context, id string) (*Issue, error) {
	query := `
		query Issue($id: String!) {
			issue(id: $id) {
				id
				identifier
				number
				title
				description
				priority
				priorityLabel
				estimate
				boardOrder
				subIssueSortOrder
				createdAt
				updatedAt
				dueDate
				url
				branchName
				snoozedUntilAt
				completedAt
				canceledAt
				archivedAt
				triagedAt
				customerTicketCount
				previousIdentifiers
				integrationSourceType
				state {
					id
					name
					type
					color
					description
					position
				}
				assignee {
					id
					name
					email
					avatarUrl
					displayName
					active
					admin
					createdAt
				}
				creator {
					id
					name
					email
					avatarUrl
					displayName
					active
				}
				team {
					id
					key
					name
					description
					icon
					color
					cyclesEnabled
					cycleStartDay
					cycleDuration
					upcomingCycleCount
				}
				labels {
					nodes {
						id
						name
						color
						description
						parent {
							id
							name
						}
					}
				}
				parent {
					id
					identifier
					title
					state {
						name
						type
					}
				}
				children {
					nodes {
						id
						identifier
						title
						priority
						createdAt
						state {
							name
							type
							color
						}
						assignee {
							name
							email
						}
					}
				}
				cycle {
					id
					number
					name
					description
					startsAt
					endsAt
					progress
					completedAt
					scopeHistory
				}
				project {
					id
					name
					description
					state
					progress
					startDate
					targetDate
					health
					lead {
						name
						email
					}
				}
				attachments(first: 20) {
					nodes {
						id
						title
						subtitle
						url
						metadata
						createdAt
						creator {
							name
							email
						}
					}
				}
				comments(first: 10) {
					nodes {
						id
						body
						createdAt
						updatedAt
						editedAt
						user {
							name
							email
							avatarUrl
						}
						parent {
							id
						}
						children {
							nodes {
								id
								body
								user {
									name
								}
							}
						}
					}
				}
				subscribers {
					nodes {
						id
						name
						email
						avatarUrl
					}
				}
				relations {
					nodes {
						id
						type
						relatedIssue {
							id
							identifier
							title
							state {
								name
								type
							}
						}
					}
				}
				history(first: 10) {
					nodes {
						id
						createdAt
						updatedAt
						actor {
							name
							email
						}
						fromAssignee {
							name
						}
						toAssignee {
							name
						}
						fromState {
							name
						}
						toState {
							name
						}
						fromPriority
						toPriority
						fromTitle
						toTitle
						fromCycle {
							name
						}
						toCycle {
							name
						}
						fromProject {
							name
						}
						toProject {
							name
						}
						addedLabelIds
						removedLabelIds
					}
				}
				reactions {
					id
					emoji
					user {
						name
						email
					}
					createdAt
				}
				externalUserCreator {
					id
					name
					email
					avatarUrl
				}
			}
		}
	`

	variables := map[string]interface{}{
		"id": id,
	}

	var response struct {
		Issue Issue `json:"issue"`
	}

	err := c.Execute(ctx, query, variables, &response)
	if err != nil {
		return nil, err
	}

	return &response.Issue, nil
}

// GetTeams returns a list of teams
func (c *Client) GetTeams(ctx context.Context, first int, after string, orderBy string) (*Teams, error) {
	query := `
		query Teams($first: Int, $after: String, $orderBy: PaginationOrderBy) {
			teams(first: $first, after: $after, orderBy: $orderBy) {
				nodes {
					id
					key
					name
					description
					private
					issueCount
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
	`

	variables := map[string]interface{}{
		"first": first,
	}
	if after != "" {
		variables["after"] = after
	}
	if orderBy != "" {
		variables["orderBy"] = orderBy
	}

	var response struct {
		Teams Teams `json:"teams"`
	}

	err := c.Execute(ctx, query, variables, &response)
	if err != nil {
		return nil, err
	}

	return &response.Teams, nil
}

// GetProjects returns a list of projects
func (c *Client) GetProjects(ctx context.Context, filter map[string]interface{}, first int, after string, orderBy string) (*Projects, error) {
	query := `
		query Projects($filter: ProjectFilter, $first: Int, $after: String, $orderBy: PaginationOrderBy) {
			projects(filter: $filter, first: $first, after: $after, orderBy: $orderBy) {
				nodes {
					id
					name
					description
					state
					priority
					progress
					startDate
					targetDate
					url
					createdAt
					updatedAt
					lead {
						id
						name
						email
					}
					teams {
						nodes {
							id
							key
							name
						}
					}
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
	`

	variables := map[string]interface{}{
		"first": first,
	}
	if filter != nil {
		variables["filter"] = filter
	}
	if after != "" {
		variables["after"] = after
	}
	if orderBy != "" {
		variables["orderBy"] = orderBy
	}

	var response struct {
		Projects Projects `json:"projects"`
	}

	err := c.Execute(ctx, query, variables, &response)
	if err != nil {
		return nil, err
	}

	return &response.Projects, nil
}

// GetProject returns a single project by ID
func (c *Client) GetProject(ctx context.Context, id string) (*Project, error) {
	query := `
		query Project($id: String!) {
			project(id: $id) {
				id
				slugId
				name
				description
				content
				state
				priority
				progress
				health
				scope
				startDate
				targetDate
				url
				icon
				color
				createdAt
				updatedAt
				completedAt
				canceledAt
				archivedAt
				slackNewIssue
				slackIssueComments
				slackIssueStatuses
				initiatives {
					nodes {
						id
						name
					}
				}
				labels {
					nodes {
						id
						name
						color
					}
				}
				lead {
					id
					name
					email
					avatarUrl
					displayName
					active
				}
				creator {
					id
					name
					email
					avatarUrl
					active
				}
				convertedFromIssue {
					id
					identifier
					title
				}
				lastAppliedTemplate {
					id
					name
					description
				}
				teams {
					nodes {
						id
						key
						name
						description
						icon
						color
						cyclesEnabled
					}
				}
				members {
					nodes {
						id
						name
						email
						avatarUrl
						displayName
						active
						admin
					}
				}
				issues(first: 50, orderBy: updatedAt) {
					nodes {
						id
						identifier
						number
						title
						description
						priority
						estimate
						createdAt
						updatedAt
						completedAt
						state {
							name
							type
							color
						}
						assignee {
							name
							email
						}
						labels {
							nodes {
								name
								color
							}
						}
					}
				}
				projectUpdates(first: 10) {
					nodes {
						id
						body
						health
						createdAt
						updatedAt
						editedAt
						user {
							name
							email
							avatarUrl
						}
					}
				}
				documents(first: 20) {
					nodes {
						id
						title
						content
						icon
						color
						createdAt
						updatedAt
						creator {
							name
							email
						}
						updatedBy {
							name
							email
						}
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"id": id,
	}

	var response struct {
		Project Project `json:"project"`
	}

	err := c.Execute(ctx, query, variables, &response)
	if err != nil {
		return nil, err
	}

	return &response.Project, nil
}

// UpdateIssue updates an issue's fields
func (c *Client) UpdateIssue(ctx context.Context, id string, input map[string]interface{}) (*Issue, error) {
	query := `
		mutation UpdateIssue($id: String!, $input: IssueUpdateInput!) {
			issueUpdate(id: $id, input: $input) {
				issue {
					id
					identifier
					title
					description
					priority
					estimate
					createdAt
					updatedAt
					dueDate
					state {
						id
						name
						type
						color
					}
					assignee {
						id
						name
						email
					}
					team {
						id
						key
						name
					}
					project {
						id
						name
					}
					labels {
						nodes {
							id
							name
							color
						}
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"id":    id,
		"input": input,
	}

	var response struct {
		IssueUpdate struct {
			Issue Issue `json:"issue"`
		} `json:"issueUpdate"`
	}

	err := c.Execute(ctx, query, variables, &response)
	if err != nil {
		return nil, err
	}

	return &response.IssueUpdate.Issue, nil
}

// CreateIssue creates a new issue
func (c *Client) CreateIssue(ctx context.Context, input map[string]interface{}) (*Issue, error) {
	query := `
		mutation CreateIssue($input: IssueCreateInput!) {
			issueCreate(input: $input) {
				issue {
					id
					identifier
					title
					description
					priority
					estimate
					createdAt
					updatedAt
					dueDate
					state {
						id
						name
						type
						color
					}
					assignee {
						id
						name
						email
					}
					team {
						id
						key
						name
					}
					project {
						id
						name
					}
					labels {
						nodes {
							id
							name
							color
						}
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"input": input,
	}

	var response struct {
		IssueCreate struct {
			Issue Issue `json:"issue"`
		} `json:"issueCreate"`
	}

	err := c.Execute(ctx, query, variables, &response)
	if err != nil {
		return nil, err
	}

	return &response.IssueCreate.Issue, nil
}

// GetTeam returns a single team by key; falls back to id lookup if not found
func (c *Client) GetTeam(ctx context.Context, key string) (*Team, error) {
    // First, attempt lookup by team key via teams connection
    queryByKey := `
        query TeamByKey($key: String!) {
            teams(filter: { key: { eq: $key } }, first: 1) {
                nodes {
                    id
                    key
                    name
                    description
                    private
                    issueCount
                }
            }
        }
    `

    variables := map[string]interface{}{"key": key}

    var respByKey struct {
        Teams struct {
            Nodes []Team `json:"nodes"`
        } `json:"teams"`
    }

    if err := c.Execute(ctx, queryByKey, variables, &respByKey); err == nil {
        if len(respByKey.Teams.Nodes) > 0 {
            t := respByKey.Teams.Nodes[0]
            return &t, nil
        }
    }

    // Fallback: try direct id lookup (in case caller passed an ID)
    queryByID := `
        query TeamByID($id: String!) {
            team(id: $id) {
                id
                key
                name
                description
                private
                issueCount
            }
        }
    `

    var respByID struct {
        Team *Team `json:"team"`
    }
    if err := c.Execute(ctx, queryByID, map[string]interface{}{"id": key}, &respByID); err == nil && respByID.Team != nil {
        return respByID.Team, nil
    }

    return nil, fmt.Errorf("team '%s' not found", key)
}

// Comment represents a Linear comment
type Comment struct {
	ID        string     `json:"id"`
	Body      string     `json:"body"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	EditedAt  *time.Time `json:"editedAt"`
	User      *User      `json:"user"`
	Parent    *Comment   `json:"parent"`
	Children  *Comments  `json:"children"`
}

// Comments represents a paginated list of comments
type Comments struct {
	Nodes    []Comment `json:"nodes"`
	PageInfo PageInfo  `json:"pageInfo"`
}

// WorkflowState represents a Linear workflow state
type WorkflowState struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Color       string  `json:"color"`
	Description string  `json:"description"`
	Position    float64 `json:"position"`
}

// GetTeamStates returns workflow states for a team
func (c *Client) GetTeamStates(ctx context.Context, teamKey string) ([]WorkflowState, error) {
	query := `
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
	`

	variables := map[string]interface{}{
		"key": teamKey,
	}

	var response struct {
		Team struct {
			States struct {
				Nodes []WorkflowState `json:"nodes"`
			} `json:"states"`
		} `json:"team"`
	}

	err := c.Execute(ctx, query, variables, &response)
	if err != nil {
		return nil, err
	}

	return response.Team.States.Nodes, nil
}

// GetTeamMembers returns members of a specific team
func (c *Client) GetTeamMembers(ctx context.Context, teamKey string) (*Users, error) {
	query := `
		query TeamMembers($key: String!) {
			team(id: $key) {
				members {
					nodes {
						id
						name
						email
						avatarUrl
						isMe
						active
						admin
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"key": teamKey,
	}

	var response struct {
		Team struct {
			Members Users `json:"members"`
		} `json:"team"`
	}

	err := c.Execute(ctx, query, variables, &response)
	if err != nil {
		return nil, err
	}

	return &response.Team.Members, nil
}

// GetUsers returns a list of all users
func (c *Client) GetUsers(ctx context.Context, first int, after string, orderBy string) (*Users, error) {
	query := `
		query Users($first: Int, $after: String, $orderBy: PaginationOrderBy) {
			users(first: $first, after: $after, orderBy: $orderBy) {
				nodes {
					id
					name
					email
					avatarUrl
					isMe
					active
					admin
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
	`

	variables := map[string]interface{}{
		"first": first,
	}
	if after != "" {
		variables["after"] = after
	}
	if orderBy != "" {
		variables["orderBy"] = orderBy
	}

	var response struct {
		Users Users `json:"users"`
	}

	err := c.Execute(ctx, query, variables, &response)
	if err != nil {
		return nil, err
	}

	return &response.Users, nil
}

// GetUser returns a specific user by email
func (c *Client) GetUser(ctx context.Context, email string) (*User, error) {
	query := `
		query User($email: String!) {
			user(email: $email) {
				id
				name
				email
				avatarUrl
				isMe
				active
				admin
			}
		}
	`

	variables := map[string]interface{}{
		"email": email,
	}

	var response struct {
		User User `json:"user"`
	}

	err := c.Execute(ctx, query, variables, &response)
	if err != nil {
		return nil, err
	}

	return &response.User, nil
}

// GetIssueComments returns comments for a specific issue
func (c *Client) GetIssueComments(ctx context.Context, issueID string, first int, after string, orderBy string) (*Comments, error) {
	query := `
		query IssueComments($id: String!, $first: Int, $after: String, $orderBy: PaginationOrderBy) {
			issue(id: $id) {
				comments(first: $first, after: $after, orderBy: $orderBy) {
					nodes {
						id
						body
						createdAt
						updatedAt
						user {
							id
							name
							email
						}
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"id":    issueID,
		"first": first,
	}
	if after != "" {
		variables["after"] = after
	}
	if orderBy != "" {
		variables["orderBy"] = orderBy
	}

	var response struct {
		Issue struct {
			Comments Comments `json:"comments"`
		} `json:"issue"`
	}

	err := c.Execute(ctx, query, variables, &response)
	if err != nil {
		return nil, err
	}

	return &response.Issue.Comments, nil
}

// CreateComment creates a new comment on an issue
func (c *Client) CreateComment(ctx context.Context, issueID string, body string) (*Comment, error) {
	query := `
		mutation CreateComment($input: CommentCreateInput!) {
			commentCreate(input: $input) {
				comment {
					id
					body
					createdAt
					updatedAt
					user {
						id
						name
						email
					}
				}
			}
		}
	`

	input := map[string]interface{}{
		"issueId": issueID,
		"body":    body,
	}

	variables := map[string]interface{}{
		"input": input,
	}

	var response struct {
		CommentCreate struct {
			Comment Comment `json:"comment"`
		} `json:"commentCreate"`
	}

	err := c.Execute(ctx, query, variables, &response)
	if err != nil {
		return nil, err
	}

	return &response.CommentCreate.Comment, nil
}

// CreateProject creates a new project
func (c *Client) CreateProject(ctx context.Context, input map[string]interface{}) (*Project, error) {
	query := `
		mutation CreateProject($input: ProjectCreateInput!) {
			projectCreate(input: $input) {
				success
				project {
					id
					name
					description
					state
					progress
					startDate
					targetDate
					url
					icon
					color
					createdAt
					updatedAt
					lead {
						id
						name
						email
					}
					teams {
						nodes {
							id
							key
							name
						}
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"input": input,
	}

	var response struct {
		ProjectCreate struct {
			Success bool    `json:"success"`
			Project Project `json:"project"`
		} `json:"projectCreate"`
	}

	err := c.Execute(ctx, query, variables, &response)
	if err != nil {
		return nil, err
	}

	return &response.ProjectCreate.Project, nil
}

// ArchiveProject archives a project by ID
func (c *Client) ArchiveProject(ctx context.Context, id string) (bool, error) {
	query := `
		mutation ArchiveProject($id: String!) {
			projectArchive(id: $id) {
				success
				project {
					id
					name
					archivedAt
				}
			}
		}
	`

	variables := map[string]interface{}{
		"id": id,
	}

	var response struct {
		ProjectArchive struct {
			Success bool    `json:"success"`
			Project Project `json:"project"`
		} `json:"projectArchive"`
	}

	err := c.Execute(ctx, query, variables, &response)
	if err != nil {
		return false, err
	}

	return response.ProjectArchive.Success, nil
}

// UpdateProject updates a project by ID with partial field updates
func (c *Client) UpdateProject(ctx context.Context, id string, input map[string]interface{}) (*Project, error) {
	query := `
		mutation UpdateProject($id: String!, $input: ProjectUpdateInput!) {
			projectUpdate(id: $id, input: $input) {
				success
				project {
					id
					name
					description
					state
					priority
					progress
					startDate
					targetDate
					url
					icon
					color
					createdAt
					updatedAt
					lead {
						id
						name
						email
					}
					teams {
						nodes {
							id
							key
							name
						}
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"id":    id,
		"input": input,
	}

	var response struct {
		ProjectUpdate struct {
			Success bool    `json:"success"`
			Project Project `json:"project"`
		} `json:"projectUpdate"`
	}

	err := c.Execute(ctx, query, variables, &response)
	if err != nil {
		return nil, err
	}

	return &response.ProjectUpdate.Project, nil
}

// ListProjectMilestones returns milestones for a specific project
func (c *Client) ListProjectMilestones(ctx context.Context, projectID string, includeArchived bool) (*ProjectMilestones, error) {
    query := `
        query ProjectMilestones($projectId: String!, $includeArchived: Boolean) {
            project(id: $projectId) {
                projectMilestones(includeArchived: $includeArchived) {
                    nodes {
                        id
                        name
                        description
                        targetDate
                        status
                        progress
                        sortOrder
                        createdAt
                        updatedAt
                        archivedAt
                    }
                }
            }
        }
    `

    variables := map[string]interface{}{
        "projectId":       projectID,
        "includeArchived": includeArchived,
    }

    var response struct {
        Project struct {
            ProjectMilestones ProjectMilestones `json:"projectMilestones"`
        } `json:"project"`
    }

    err := c.Execute(ctx, query, variables, &response)
    if err != nil {
        return nil, err
    }

    return &response.Project.ProjectMilestones, nil
}

// GetProjectMilestone returns a specific project milestone
func (c *Client) GetProjectMilestone(ctx context.Context, milestoneID string) (*ProjectMilestone, error) {
    query := `
        query ProjectMilestone($id: String!) {
            projectMilestone(id: $id) {
                id
                name
                description
                targetDate
                status
                progress
                sortOrder
                createdAt
                updatedAt
                archivedAt
                project {
                    id
                    name
                }
            }
        }
    `

    variables := map[string]interface{}{
        "id": milestoneID,
    }

    var response struct {
        ProjectMilestone ProjectMilestone `json:"projectMilestone"`
    }

    err := c.Execute(ctx, query, variables, &response)
    if err != nil {
        return nil, err
    }

    return &response.ProjectMilestone, nil
}

// CreateProjectMilestone creates a new project milestone
func (c *Client) CreateProjectMilestone(ctx context.Context, input map[string]interface{}) (*ProjectMilestone, error) {
    query := `
        mutation CreateProjectMilestone($input: ProjectMilestoneCreateInput!) {
            projectMilestoneCreate(input: $input) {
                success
                projectMilestone {
                    id
                    name
                    description
                    targetDate
                    status
                    progress
                    sortOrder
                    createdAt
                    updatedAt
                    project {
                        id
                        name
                    }
                }
            }
        }
    `

    variables := map[string]interface{}{
        "input": input,
    }

    var response struct {
        ProjectMilestoneCreate struct {
            Success          bool             `json:"success"`
            ProjectMilestone ProjectMilestone `json:"projectMilestone"`
        } `json:"projectMilestoneCreate"`
    }

    err := c.Execute(ctx, query, variables, &response)
    if err != nil {
        return nil, err
    }

    return &response.ProjectMilestoneCreate.ProjectMilestone, nil
}

// UpdateProjectMilestone updates an existing project milestone
func (c *Client) UpdateProjectMilestone(ctx context.Context, milestoneID string, input map[string]interface{}) (*ProjectMilestone, error) {
    query := `
        mutation UpdateProjectMilestone($id: String!, $input: ProjectMilestoneUpdateInput!) {
            projectMilestoneUpdate(id: $id, input: $input) {
                success
                projectMilestone {
                    id
                    name
                    description
                    targetDate
                    status
                    progress
                    sortOrder
                    createdAt
                    updatedAt
                    archivedAt
                    project {
                        id
                        name
                    }
                }
            }
        }
    `

    variables := map[string]interface{}{
        "id":    milestoneID,
        "input": input,
    }

    var response struct {
        ProjectMilestoneUpdate struct {
            Success          bool             `json:"success"`
            ProjectMilestone ProjectMilestone `json:"projectMilestone"`
        } `json:"projectMilestoneUpdate"`
    }

    err := c.Execute(ctx, query, variables, &response)
    if err != nil {
        return nil, err
    }

    return &response.ProjectMilestoneUpdate.ProjectMilestone, nil
}

// DeleteProjectMilestone deletes (archives) a project milestone
func (c *Client) DeleteProjectMilestone(ctx context.Context, milestoneID string) error {
    query := `
        mutation DeleteProjectMilestone($id: String!) {
            projectMilestoneDelete(id: $id) {
                success
            }
        }
    `

    variables := map[string]interface{}{
        "id": milestoneID,
    }

    var response struct {
        ProjectMilestoneDelete struct {
            Success bool `json:"success"`
        } `json:"projectMilestoneDelete"`
    }

    err := c.Execute(ctx, query, variables, &response)
    if err != nil {
        return err
    }

    if !response.ProjectMilestoneDelete.Success {
        return fmt.Errorf("failed to delete milestone")
    }

    return nil
}
