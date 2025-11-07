package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dorkitude/linctl/pkg/api"
	"github.com/dorkitude/linctl/pkg/auth"
	"github.com/dorkitude/linctl/pkg/output"
	"github.com/dorkitude/linctl/pkg/utils"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// projectAPI captures the subset of API client used by project commands.
// This enables dependency injection in tests without changing public API types.
type projectAPI interface {
	GetTeam(ctx context.Context, key string) (*api.Team, error)
	GetProjects(ctx context.Context, filter map[string]interface{}, first int, after string, orderBy string) (*api.Projects, error)
	CreateProject(ctx context.Context, input map[string]interface{}) (*api.Project, error)
	UpdateProject(ctx context.Context, id string, input map[string]interface{}) (*api.Project, error)
	ArchiveProject(ctx context.Context, id string) (bool, error)
	GetProject(ctx context.Context, id string) (*api.Project, error)
	CreateProjectUpdate(ctx context.Context, input map[string]interface{}) (*api.ProjectUpdate, error)
	ListProjectUpdates(ctx context.Context, projectID string) (*api.ProjectUpdates, error)
	GetProjectUpdate(ctx context.Context, updateID string) (*api.ProjectUpdate, error)
}

// Injection points for testing
var newAPIClient = func(authHeader string) projectAPI { return api.NewClient(authHeader) }
var getAuthHeader = auth.GetAuthHeader

// constructProjectURL constructs an ID-based project URL
func constructProjectURL(projectID string, originalURL string) string {
	// Extract workspace from the original URL
	// Format: https://linear.app/{workspace}/project/{slug}
	if originalURL == "" {
		return ""
	}

	parts := strings.Split(originalURL, "/")
	if len(parts) >= 5 {
		workspace := parts[3]
		return fmt.Sprintf("https://linear.app/%s/project/%s", workspace, projectID)
	}

	// Fallback to original URL if we can't parse it
	return originalURL
}

// validateHexColor validates a hex color code format
func validateHexColor(color string) error {
	if color == "" {
		return nil
	}
	if !strings.HasPrefix(color, "#") || (len(color) != 4 && len(color) != 7) {
		return fmt.Errorf("invalid hex color format (expected #RGB or #RRGGBB)")
	}
	// Check if all characters after # are valid hex
	for _, c := range color[1:] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return fmt.Errorf("invalid hex color format")
		}
	}
	return nil
}

// lookupUserIDsByEmails looks up user IDs from comma-separated emails
func lookupUserIDsByEmails(ctx context.Context, client projectAPI, emails string) ([]string, error) {
	if emails == "" {
		return nil, nil
	}

	emailList := strings.Split(emails, ",")
	userIDs := make([]string, 0, len(emailList))

	// Type assert to get the full API client
	fullClient, ok := client.(*api.Client)
	if !ok {
		return nil, fmt.Errorf("client type assertion failed")
	}

	for _, email := range emailList {
		email = strings.TrimSpace(email)
		if email == "" {
			continue
		}

		user, err := fullClient.GetUser(ctx, email)
		if err != nil {
			return nil, fmt.Errorf("user not found with email '%s': %v", email, err)
		}
		userIDs = append(userIDs, user.ID)
	}

	return userIDs, nil
}

// lookupLabelIDsByNames looks up project label IDs from comma-separated names
func lookupLabelIDsByNames(ctx context.Context, client projectAPI, names string) ([]string, error) {
	if names == "" {
		return nil, nil
	}

	nameList := strings.Split(names, ",")
	labelIDs := make([]string, 0, len(nameList))

	// Type assert to get the full API client
	fullClient, ok := client.(*api.Client)
	if !ok {
		return nil, fmt.Errorf("client type assertion failed")
	}

	// Get all project labels
	labels, err := fullClient.GetProjectLabels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get project labels: %v", err)
	}

	// Build a map of label names to IDs for quick lookup
	labelMap := make(map[string]string)
	for _, label := range labels.Nodes {
		labelMap[strings.ToLower(label.Name)] = label.ID
	}

	// Look up each requested label
	for _, name := range nameList {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		labelID, found := labelMap[strings.ToLower(name)]
		if !found {
			return nil, fmt.Errorf("project label not found: '%s'", name)
		}
		labelIDs = append(labelIDs, labelID)
	}

	return labelIDs, nil
}

// projectCmd represents the project command
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage Linear projects",
	Long: `Manage Linear projects including listing, viewing, and creating projects.

Examples:
  linctl project list                      # List active projects
  linctl project list --include-completed  # List all projects including completed
  linctl project list --newer-than 1_month_ago  # List projects from last month
  linctl project get PROJECT-ID            # Get project details
  linctl project create                    # Create a new project`,
}

var projectListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List projects",
	Long:    `List all projects in your Linear workspace.`,
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		// Get auth header
		authHeader, err := getAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Create API client
		client := newAPIClient(authHeader)

		// Get filters
		teamKey, _ := cmd.Flags().GetString("team")
		state, _ := cmd.Flags().GetString("state")
		limit, _ := cmd.Flags().GetInt("limit")
		includeCompleted, _ := cmd.Flags().GetBool("include-completed")

		// Build filter
		filter := make(map[string]interface{})
		if teamKey != "" {
			// Get team ID from key
			team, err := client.GetTeam(context.Background(), teamKey)
			if err != nil {
				output.Error(fmt.Sprintf("Failed to find team '%s': %v", teamKey, err), plaintext, jsonOut)
				os.Exit(1)
			}
			filter["team"] = map[string]interface{}{"id": team.ID}
		}
		if state != "" {
			filter["state"] = map[string]interface{}{"eq": state}
		} else if !includeCompleted {
			// Only filter out completed projects if no specific state is requested
			filter["state"] = map[string]interface{}{
				"nin": []string{"completed", "canceled"},
			}
		}

		// Handle newer-than filter
		newerThan, _ := cmd.Flags().GetString("newer-than")
		createdAt, err := utils.ParseTimeExpression(newerThan)
		if err != nil {
			output.Error(fmt.Sprintf("Invalid newer-than value: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}
		if createdAt != "" {
			filter["createdAt"] = map[string]interface{}{"gte": createdAt}
		}

		// Get sort option
		sortBy, _ := cmd.Flags().GetString("sort")
		orderBy := ""
		if sortBy != "" {
			switch sortBy {
			case "created", "createdAt":
				orderBy = "createdAt"
			case "updated", "updatedAt":
				orderBy = "updatedAt"
			case "linear":
				// Use empty string for Linear's default sort
				orderBy = ""
			default:
				output.Error(fmt.Sprintf("Invalid sort option: %s. Valid options are: linear, created, updated", sortBy), plaintext, jsonOut)
				os.Exit(1)
			}
		}

		// Get projects
		projects, err := client.GetProjects(context.Background(), filter, limit, "", orderBy)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to list projects: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Handle output
		if jsonOut {
			output.JSON(projects.Nodes)
			return
		} else if plaintext {
			fmt.Println("# Projects")
			for _, project := range projects.Nodes {
				fmt.Printf("## %s\n", project.Name)
				fmt.Printf("- **ID**: %s\n", project.ID)
				fmt.Printf("- **State**: %s\n", project.State)
				if project.Priority > 0 {
					fmt.Printf("- **Priority**: %d\n", project.Priority)
				}
				fmt.Printf("- **Progress**: %.0f%%\n", project.Progress*100)
				if project.Lead != nil {
					fmt.Printf("- **Lead**: %s\n", project.Lead.Name)
				} else {
					fmt.Printf("- **Lead**: Unassigned\n")
				}
				if project.Teams != nil && len(project.Teams.Nodes) > 0 {
					teams := ""
					for i, team := range project.Teams.Nodes {
						if i > 0 {
							teams += ", "
						}
						teams += team.Key
					}
					fmt.Printf("- **Teams**: %s\n", teams)
				}
				if project.StartDate != nil {
					fmt.Printf("- **Start Date**: %s\n", *project.StartDate)
				}
				if project.TargetDate != nil {
					fmt.Printf("- **Target Date**: %s\n", *project.TargetDate)
				}
				fmt.Printf("- **Created**: %s\n", project.CreatedAt.Format("2006-01-02"))
				fmt.Printf("- **Updated**: %s\n", project.UpdatedAt.Format("2006-01-02"))
				fmt.Printf("- **URL**: %s\n", constructProjectURL(project.ID, project.URL))
				if project.Description != "" {
					fmt.Printf("- **Description**: %s\n", project.Description)
				}
				fmt.Println()
			}
			fmt.Printf("\nTotal: %d projects\n", len(projects.Nodes))
			return
		} else {
			// Table output
			headers := []string{"Name", "State", "Priority", "Lead", "Teams", "Created", "Updated", "URL"}
			rows := [][]string{}

			for _, project := range projects.Nodes {
				lead := color.New(color.FgYellow).Sprint("Unassigned")
				if project.Lead != nil {
					lead = project.Lead.Name
				}

				teams := ""
				if project.Teams != nil && len(project.Teams.Nodes) > 0 {
					for i, team := range project.Teams.Nodes {
						if i > 0 {
							teams += ", "
						}
						teams += team.Key
					}
				}

				stateColor := color.New(color.FgGreen)
				switch project.State {
				case "planned":
					stateColor = color.New(color.FgCyan)
				case "started":
					stateColor = color.New(color.FgBlue)
				case "paused":
					stateColor = color.New(color.FgYellow)
				case "completed":
					stateColor = color.New(color.FgGreen)
				case "canceled":
					stateColor = color.New(color.FgRed)
				}

				// Format priority
				priorityStr := fmt.Sprintf("%d", project.Priority)
				if project.Priority == 0 {
					priorityStr = "-"
				}

				rows = append(rows, []string{
					truncateString(project.Name, 25),
					stateColor.Sprint(project.State),
					priorityStr,
					lead,
					teams,
					project.CreatedAt.Format("2006-01-02"),
					project.UpdatedAt.Format("2006-01-02"),
					constructProjectURL(project.ID, project.URL),
				})
			}

			output.Table(output.TableData{
				Headers: headers,
				Rows:    rows,
			}, plaintext, jsonOut)

			if !plaintext && !jsonOut {
				fmt.Printf("\n%s %d projects\n",
					color.New(color.FgGreen).Sprint("✓"),
					len(projects.Nodes))
			}
		}
	},
}

var projectGetCmd = &cobra.Command{
	Use:     "get PROJECT-ID",
	Aliases: []string{"show"},
	Short:   "Get project details",
	Long:    `Get detailed information about a specific project.`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		projectID := args[0]

		// Get auth header
		authHeader, err := getAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Create API client
		client := newAPIClient(authHeader)

		// Get project details
		project, err := client.GetProject(context.Background(), projectID)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to get project: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Handle output
		if jsonOut {
			output.JSON(project)
		} else if plaintext {
			fmt.Printf("# %s\n\n", project.Name)

			if project.Description != "" {
				fmt.Printf("## Description\n%s\n\n", project.Description)
			}

			if project.Content != "" {
				fmt.Printf("## Content\n%s\n\n", project.Content)
			}

			fmt.Printf("## Core Details\n")
			fmt.Printf("- **ID**: %s\n", project.ID)
			fmt.Printf("- **Slug ID**: %s\n", project.SlugId)
			fmt.Printf("- **State**: %s\n", project.State)
			if project.Priority > 0 {
				fmt.Printf("- **Priority**: %d\n", project.Priority)
			}
			fmt.Printf("- **Progress**: %.0f%%\n", project.Progress*100)
			fmt.Printf("- **Health**: %s\n", project.Health)
			fmt.Printf("- **Scope**: %d\n", project.Scope)
			if project.Initiatives != nil && len(project.Initiatives.Nodes) > 0 {
				initiatives := ""
				for i, initiative := range project.Initiatives.Nodes {
					if i > 0 {
						initiatives += ", "
					}
					initiatives += initiative.Name
				}
				fmt.Printf("- **Initiatives**: %s\n", initiatives)
			}
			if project.Labels != nil && len(project.Labels.Nodes) > 0 {
				labels := ""
				for i, label := range project.Labels.Nodes {
					if i > 0 {
						labels += ", "
					}
					labels += label.Name
				}
				fmt.Printf("- **Labels**: %s\n", labels)
			}
			if project.Icon != nil && *project.Icon != "" {
				fmt.Printf("- **Icon**: %s\n", *project.Icon)
			}
			fmt.Printf("- **Color**: %s\n", project.Color)

			fmt.Printf("\n## Timeline\n")
			if project.StartDate != nil {
				fmt.Printf("- **Start Date**: %s\n", *project.StartDate)
			}
			if project.TargetDate != nil {
				fmt.Printf("- **Target Date**: %s\n", *project.TargetDate)
			}
			fmt.Printf("- **Created**: %s\n", project.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("- **Updated**: %s\n", project.UpdatedAt.Format("2006-01-02 15:04:05"))
			if project.CompletedAt != nil {
				fmt.Printf("- **Completed**: %s\n", project.CompletedAt.Format("2006-01-02 15:04:05"))
			}
			if project.CanceledAt != nil {
				fmt.Printf("- **Canceled**: %s\n", project.CanceledAt.Format("2006-01-02 15:04:05"))
			}
			if project.ArchivedAt != nil {
				fmt.Printf("- **Archived**: %s\n", project.ArchivedAt.Format("2006-01-02 15:04:05"))
			}

			fmt.Printf("\n## People\n")
			if project.Lead != nil {
				fmt.Printf("- **Lead**: %s (%s)\n", project.Lead.Name, project.Lead.Email)
				if project.Lead.DisplayName != "" && project.Lead.DisplayName != project.Lead.Name {
					fmt.Printf("  - Display Name: %s\n", project.Lead.DisplayName)
				}
			} else {
				fmt.Printf("- **Lead**: Unassigned\n")
			}
			if project.Creator != nil {
				fmt.Printf("- **Creator**: %s (%s)\n", project.Creator.Name, project.Creator.Email)
			}

			fmt.Printf("\n## Slack Integration\n")
			fmt.Printf("- **Slack New Issue**: %v\n", project.SlackNewIssue)
			fmt.Printf("- **Slack Issue Comments**: %v\n", project.SlackIssueComments)
			fmt.Printf("- **Slack Issue Statuses**: %v\n", project.SlackIssueStatuses)

			if project.ConvertedFromIssue != nil {
				fmt.Printf("\n## Origin\n")
				fmt.Printf("- **Converted from Issue**: %s - %s\n", project.ConvertedFromIssue.Identifier, project.ConvertedFromIssue.Title)
			}

			if project.LastAppliedTemplate != nil {
				fmt.Printf("\n## Template\n")
				fmt.Printf("- **Last Applied**: %s\n", project.LastAppliedTemplate.Name)
				if project.LastAppliedTemplate.Description != "" {
					fmt.Printf("  - Description: %s\n", project.LastAppliedTemplate.Description)
				}
			}

			// Teams
			if project.Teams != nil && len(project.Teams.Nodes) > 0 {
				fmt.Printf("\n## Teams\n")
				for _, team := range project.Teams.Nodes {
					fmt.Printf("- **%s** (%s)\n", team.Name, team.Key)
					if team.Description != "" {
						fmt.Printf("  - Description: %s\n", team.Description)
					}
					fmt.Printf("  - Cycles Enabled: %v\n", team.CyclesEnabled)
				}
			}

			fmt.Printf("\n## URL\n")
			fmt.Printf("- %s\n", constructProjectURL(project.ID, project.URL))

			// Show members if available
			if project.Members != nil && len(project.Members.Nodes) > 0 {
				fmt.Printf("\n## Members\n")
				for _, member := range project.Members.Nodes {
					fmt.Printf("- %s (%s)", member.Name, member.Email)
					if member.DisplayName != "" && member.DisplayName != member.Name {
						fmt.Printf(" - %s", member.DisplayName)
					}
					if member.Admin {
						fmt.Printf(" [Admin]")
					}
					if !member.Active {
						fmt.Printf(" [Inactive]")
					}
					fmt.Println()
				}
			}

			// Project Updates
			if project.ProjectUpdates != nil && len(project.ProjectUpdates.Nodes) > 0 {
				fmt.Printf("\n## Recent Project Updates\n")
				for _, update := range project.ProjectUpdates.Nodes {
					fmt.Printf("\n### %s by %s\n", update.CreatedAt.Format("2006-01-02 15:04"), update.User.Name)
					if update.EditedAt != nil {
						fmt.Printf("*(edited %s)*\n", update.EditedAt.Format("2006-01-02 15:04"))
					}
					fmt.Printf("- **Health**: %s\n", update.Health)
					fmt.Printf("\n%s\n", update.Body)
				}
			}

			// Documents
			if project.Documents != nil && len(project.Documents.Nodes) > 0 {
				fmt.Printf("\n## Documents\n")
				for _, doc := range project.Documents.Nodes {
					fmt.Printf("\n### %s\n", doc.Title)
					if doc.Icon != nil && *doc.Icon != "" {
						fmt.Printf("- **Icon**: %s\n", *doc.Icon)
					}
					fmt.Printf("- **Color**: %s\n", doc.Color)
					fmt.Printf("- **Created**: %s by %s\n", doc.CreatedAt.Format("2006-01-02"), doc.Creator.Name)
					if doc.UpdatedBy != nil {
						fmt.Printf("- **Updated**: %s by %s\n", doc.UpdatedAt.Format("2006-01-02"), doc.UpdatedBy.Name)
					}
					fmt.Printf("\n%s\n", doc.Content)
				}
			}

			// Show recent issues
			if project.Issues != nil && len(project.Issues.Nodes) > 0 {
				fmt.Printf("\n## Issues (%d total)\n", len(project.Issues.Nodes))
				for _, issue := range project.Issues.Nodes {
					stateStr := ""
					if issue.State != nil {
						switch issue.State.Type {
						case "completed":
							stateStr = "[x]"
						case "started":
							stateStr = "[~]"
						case "canceled":
							stateStr = "[-]"
						default:
							stateStr = "[ ]"
						}
					} else {
						stateStr = "[ ]"
					}

					assignee := "Unassigned"
					if issue.Assignee != nil {
						assignee = issue.Assignee.Name
					}

					fmt.Printf("\n### %s %s (#%d)\n", stateStr, issue.Identifier, issue.Number)
					fmt.Printf("**%s**\n", issue.Title)
					fmt.Printf("- Assignee: %s\n", assignee)
					fmt.Printf("- Priority: %s\n", priorityToString(issue.Priority))
					if issue.Estimate != nil {
						fmt.Printf("- Estimate: %.1f\n", *issue.Estimate)
					}
					if issue.State != nil {
						fmt.Printf("- State: %s\n", issue.State.Name)
					}
					if issue.Labels != nil && len(issue.Labels.Nodes) > 0 {
						labels := []string{}
						for _, label := range issue.Labels.Nodes {
							labels = append(labels, label.Name)
						}
						fmt.Printf("- Labels: %s\n", strings.Join(labels, ", "))
					}
					fmt.Printf("- Updated: %s\n", issue.UpdatedAt.Format("2006-01-02 15:04"))
					if issue.Description != "" {
						// Show first 3 lines of description
						lines := strings.Split(issue.Description, "\n")
						preview := ""
						for i, line := range lines {
							if i >= 3 {
								preview += "\n  ..."
								break
							}
							if i > 0 {
								preview += "\n  "
							}
							preview += line
						}
						fmt.Printf("- Description: %s\n", preview)
					}
				}
			}
		} else {
			// Formatted output
			fmt.Println()
			fmt.Printf("%s %s\n", color.New(color.FgCyan, color.Bold).Sprint("📁 Project:"), project.Name)
			fmt.Println(strings.Repeat("─", 50))

			fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("ID:"), project.ID)

			if project.Description != "" {
				fmt.Printf("\n%s\n%s\n", color.New(color.Bold).Sprint("Description:"), project.Description)
			}

			stateColor := color.New(color.FgGreen)
			switch project.State {
			case "planned":
				stateColor = color.New(color.FgCyan)
			case "started":
				stateColor = color.New(color.FgBlue)
			case "paused":
				stateColor = color.New(color.FgYellow)
			case "completed":
				stateColor = color.New(color.FgGreen)
			case "canceled":
				stateColor = color.New(color.FgRed)
			}
			fmt.Printf("\n%s %s\n", color.New(color.Bold).Sprint("State:"), stateColor.Sprint(project.State))

			if project.Priority > 0 {
				fmt.Printf("%s %d\n", color.New(color.Bold).Sprint("Priority:"), project.Priority)
			}

			progressColor := color.New(color.FgRed)
			if project.Progress >= 0.75 {
				progressColor = color.New(color.FgGreen)
			} else if project.Progress >= 0.5 {
				progressColor = color.New(color.FgYellow)
			}
			fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("Progress:"), progressColor.Sprintf("%.0f%%", project.Progress*100))

			if project.Initiatives != nil && len(project.Initiatives.Nodes) > 0 {
				initiatives := ""
				for i, initiative := range project.Initiatives.Nodes {
					if i > 0 {
						initiatives += ", "
					}
					initiatives += initiative.Name
				}
				fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("Initiatives:"), initiatives)
			}

			if project.Labels != nil && len(project.Labels.Nodes) > 0 {
				labels := ""
				for i, label := range project.Labels.Nodes {
					if i > 0 {
						labels += ", "
					}
					labels += label.Name
				}
				fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("Labels:"), labels)
			}

			if project.StartDate != nil || project.TargetDate != nil {
				fmt.Println()
				if project.StartDate != nil {
					fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("Start Date:"), *project.StartDate)
				}
				if project.TargetDate != nil {
					fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("Target Date:"), *project.TargetDate)
				}
			}

			if project.Lead != nil {
				fmt.Printf("\n%s %s (%s)\n",
					color.New(color.Bold).Sprint("Lead:"),
					project.Lead.Name,
					color.New(color.FgCyan).Sprint(project.Lead.Email))
			}

			if project.Teams != nil && len(project.Teams.Nodes) > 0 {
				fmt.Printf("\n%s\n", color.New(color.Bold).Sprint("Teams:"))
				for _, team := range project.Teams.Nodes {
					fmt.Printf("  • %s - %s\n",
						color.New(color.FgCyan).Sprint(team.Key),
						team.Name)
				}
			}

			// Show members if available
			if project.Members != nil && len(project.Members.Nodes) > 0 {
				fmt.Printf("\n%s\n", color.New(color.Bold).Sprint("Members:"))
				for _, member := range project.Members.Nodes {
					fmt.Printf("  • %s (%s)\n",
						member.Name,
						color.New(color.FgCyan).Sprint(member.Email))
				}
			}

			// Show sample issues if available
			if project.Issues != nil && len(project.Issues.Nodes) > 0 {
				fmt.Printf("\n%s\n", color.New(color.Bold).Sprint("Recent Issues:"))
				for i, issue := range project.Issues.Nodes {
					if i >= 5 {
						break // Show only first 5
					}
					stateIcon := "○"
					if issue.State != nil {
						switch issue.State.Type {
						case "completed":
							stateIcon = color.New(color.FgGreen).Sprint("✓")
						case "started":
							stateIcon = color.New(color.FgBlue).Sprint("◐")
						case "canceled":
							stateIcon = color.New(color.FgRed).Sprint("✗")
						}
					}
					assignee := "Unassigned"
					if issue.Assignee != nil {
						assignee = issue.Assignee.Name
					}
					fmt.Printf("  %s %s %s (%s)\n",
						stateIcon,
						color.New(color.FgCyan).Sprint(issue.Identifier),
						issue.Title,
						color.New(color.FgWhite, color.Faint).Sprint(assignee))
				}
			}

			// Show timestamps
			fmt.Printf("\n%s\n", color.New(color.Bold).Sprint("Timeline:"))
			fmt.Printf("  Created: %s\n", project.CreatedAt.Format("2006-01-02"))
			fmt.Printf("  Updated: %s\n", project.UpdatedAt.Format("2006-01-02"))
			if project.CompletedAt != nil {
				fmt.Printf("  Completed: %s\n", project.CompletedAt.Format("2006-01-02"))
			}
			if project.CanceledAt != nil {
				fmt.Printf("  Canceled: %s\n", project.CanceledAt.Format("2006-01-02"))
			}

			// Show URL
			if project.URL != "" {
				fmt.Printf("\n%s %s\n",
					color.New(color.Bold).Sprint("URL:"),
					color.New(color.FgBlue, color.Underline).Sprint(constructProjectURL(project.ID, project.URL)))
			}

			fmt.Println()
		}
	},
}

var projectCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project",
	Long: `Create a new project in Linear workspace with required and optional configuration.

The --name and --team flags are required. All other fields are optional.

Examples:
  # Create project with minimal required fields
  linctl project create --name "Q1 Backend" --team ENG

  # Create project with all optional fields
  linctl project create --name "Test Project" --team ENG --state started --priority 1 --description "Test project for validation"

  # Create project with target date
  linctl project create --name "Launch" --team PROD --state planned --target-date 2024-12-31`,
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		// Get required flags
		name, _ := cmd.Flags().GetString("name")
		teamKey, _ := cmd.Flags().GetString("team")

		// Validate required fields
		if name == "" || teamKey == "" {
			output.Error("Both --name and --team are required", plaintext, jsonOut)
			os.Exit(1)
		}

		// Get auth header
		authHeader, err := getAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Create API client
		client := newAPIClient(authHeader)

		// Resolve team key to team UUID
		team, err := client.GetTeam(context.Background(), teamKey)
		if err != nil {
			output.Error(fmt.Sprintf("Team '%s' not found. Use 'linctl team list' to see available teams.", teamKey), plaintext, jsonOut)
			os.Exit(1)
		}

		// Get optional fields
		description, _ := cmd.Flags().GetString("description")
		state, _ := cmd.Flags().GetString("state")
		targetDate, _ := cmd.Flags().GetString("target-date")

		// Validate state if provided
		if state != "" {
			allowedStates := []string{"planned", "started", "paused", "completed", "canceled"}
			valid := false
			for _, s := range allowedStates {
				if state == s {
					valid = true
					break
				}
			}
			if !valid {
				output.Error(fmt.Sprintf("Invalid state. Must be one of: %s", strings.Join(allowedStates, ", ")), plaintext, jsonOut)
				os.Exit(1)
			}
		}

		// Validate priority if provided
		var priority int
		if cmd.Flags().Changed("priority") {
			priority, _ = cmd.Flags().GetInt("priority")
			if priority < 0 || priority > 4 {
				output.Error("Priority must be between 0 (None) and 4 (Low)", plaintext, jsonOut)
				os.Exit(1)
			}
		}

		// Validate target-date format if provided (YYYY-MM-DD)
		if targetDate != "" {
			if _, err := time.Parse("2006-01-02", targetDate); err != nil {
				output.Error("Invalid --target-date format. Expected YYYY-MM-DD", plaintext, jsonOut)
				os.Exit(1)
			}
		}

		// Get and validate new optional fields
		startDate, _ := cmd.Flags().GetString("start-date")
		if startDate != "" {
			if _, err := time.Parse("2006-01-02", startDate); err != nil {
				output.Error("Invalid --start-date format. Expected YYYY-MM-DD", plaintext, jsonOut)
				os.Exit(1)
			}
		}

		leadEmail, _ := cmd.Flags().GetString("lead")
		members, _ := cmd.Flags().GetString("members")
		labelNames, _ := cmd.Flags().GetString("label")
		icon, _ := cmd.Flags().GetString("icon")
		projectColor, _ := cmd.Flags().GetString("color")
		links, _ := cmd.Flags().GetStringArray("link")

		// Validate color format
		if err := validateHexColor(projectColor); err != nil {
			output.Error(err.Error(), plaintext, jsonOut)
			os.Exit(1)
		}

		// Look up lead user ID
		var leadID string
		if leadEmail != "" {
			user, err := client.(*api.Client).GetUser(context.Background(), leadEmail)
			if err != nil {
				output.Error(fmt.Sprintf("Lead user not found with email '%s': %v", leadEmail, err), plaintext, jsonOut)
				os.Exit(1)
			}
			leadID = user.ID
		}

		// Look up member user IDs
		memberIDs, err := lookupUserIDsByEmails(context.Background(), client, members)
		if err != nil {
			output.Error(err.Error(), plaintext, jsonOut)
			os.Exit(1)
		}

		// Look up label IDs
		labelIDs, err := lookupLabelIDsByNames(context.Background(), client, labelNames)
		if err != nil {
			output.Error(err.Error(), plaintext, jsonOut)
			os.Exit(1)
		}

		// Build input map
		input := map[string]interface{}{
			"name":    name,
			"teamIds": []string{team.ID},
		}

		if description != "" {
			input["description"] = description
		}
		if state != "" {
			input["state"] = state
		}
		if cmd.Flags().Changed("priority") {
			input["priority"] = priority
		}
		if startDate != "" {
			input["startDate"] = startDate
		}
		if targetDate != "" {
			input["targetDate"] = targetDate
		}
		if leadID != "" {
			input["leadId"] = leadID
		}
		if len(memberIDs) > 0 {
			input["memberIds"] = memberIDs
		}
		if len(labelIDs) > 0 {
			input["labelIds"] = labelIDs
		}
		if icon != "" {
			input["icon"] = icon
		}
		if projectColor != "" {
			input["color"] = projectColor
		}
		if len(links) > 0 {
			// For now, we'll just add the first link as a simple string
			// TODO: Investigate if Linear supports multiple links or structured link objects
			input["links"] = links
		}

		// Create project
		project, err := client.CreateProject(context.Background(), input)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to create project: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Handle output
		if jsonOut {
			output.JSON(project)
		} else if plaintext {
			fmt.Printf("# Project Created\n\n")
			fmt.Printf("- **Name**: %s\n", project.Name)
			fmt.Printf("- **ID**: %s\n", project.ID)
			fmt.Printf("- **State**: %s\n", project.State)
			if project.Description != "" {
				fmt.Printf("- **Description**: %s\n", project.Description)
			}
			if project.Teams != nil && len(project.Teams.Nodes) > 0 {
				teams := ""
				for i, t := range project.Teams.Nodes {
					if i > 0 {
						teams += ", "
					}
					teams += t.Key
				}
				fmt.Printf("- **Teams**: %s\n", teams)
			}
			fmt.Printf("- **URL**: %s\n", constructProjectURL(project.ID, project.URL))
		} else {
			fmt.Println()
			fmt.Printf("%s Project created successfully\n", color.New(color.FgGreen).Sprint("✓"))
			fmt.Println()
			fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("Name:"), project.Name)
			fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("ID:"), project.ID)
			fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("State:"), project.State)
			if project.Teams != nil && len(project.Teams.Nodes) > 0 {
				fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("Team:"), project.Teams.Nodes[0].Key)
			}
			fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("URL:"), color.New(color.FgBlue, color.Underline).Sprint(constructProjectURL(project.ID, project.URL)))
			fmt.Println()
		}
	},
}

var projectArchiveCmd = &cobra.Command{
	Use:   "archive PROJECT-UUID",
	Short: "Archive a project",
	Long: `Archive a project by its UUID. Archived projects are hidden from most views but can still be accessed.

Examples:
  linctl project archive abc-123-def-456`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		projectID := args[0]

		// Validate argument provided
		if projectID == "" {
			output.Error("Project UUID is required", plaintext, jsonOut)
			os.Exit(1)
		}

		// Get auth header
		authHeader, err := getAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Create API client
		client := newAPIClient(authHeader)

		// Archive project
		success, err := client.ArchiveProject(context.Background(), projectID)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to archive project: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Try to fetch project details to include the name in output (best effort)
		var projectName string
		if success {
			if proj, gerr := client.GetProject(context.Background(), projectID); gerr == nil && proj != nil {
				projectName = proj.Name
			}
		}

		// Handle output
		if jsonOut {
			payload := map[string]interface{}{
				"success":   success,
				"projectId": projectID,
			}
			if projectName != "" {
				payload["projectName"] = projectName
			}
			output.JSON(payload)
		} else if plaintext {
			fmt.Printf("# Project Archived\n\n")
			if projectName != "" {
				fmt.Printf("- **Name**: %s\n", projectName)
			}
			fmt.Printf("- **Project ID**: %s\n", projectID)
			fmt.Printf("- **Status**: Archived\n")
		} else {
			fmt.Println()
			fmt.Printf("%s Project archived successfully\n", color.New(color.FgGreen).Sprint("✓"))
			fmt.Println()
			if projectName != "" {
				fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("Name:"), projectName)
			}
			fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("Project ID:"), projectID)
			fmt.Println()
		}
	},
}

var projectUpdateCmd = &cobra.Command{
	Use:   "update PROJECT-UUID",
	Short: "Update project fields",
	Long: `Update one or more project fields. At least one field must be provided.

The project UUID is required as the first argument. All field flags are optional, but at least one must be specified.

Examples:
  # Update single field
  linctl project update abc-123 --name "New Project Name"
  linctl project update abc-123 --state started
  linctl project update abc-123 --priority 1

  # Update multiple fields
  linctl project update abc-123 --state started --priority 2
  linctl project update abc-123 --description "Full description" --summary "Short summary"

  # Update with labels
  linctl project update abc-123 --label "urgent,backend"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		projectID := args[0]

		// Validate project UUID provided
		if projectID == "" {
			output.Error("Project UUID is required", plaintext, jsonOut)
			os.Exit(1)
		}

		// Get auth header
		authHeader, err := getAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Create API client
		client := newAPIClient(authHeader)

		// Build input map with only changed fields
		input := make(map[string]interface{})

		if cmd.Flags().Changed("name") {
			name, _ := cmd.Flags().GetString("name")
			input["name"] = name
		}
		if cmd.Flags().Changed("description") {
			description, _ := cmd.Flags().GetString("description")
			input["description"] = description
		}
		if cmd.Flags().Changed("summary") {
			summary, _ := cmd.Flags().GetString("summary")
			input["shortSummary"] = summary
		}
		if cmd.Flags().Changed("state") {
			state, _ := cmd.Flags().GetString("state")
			input["state"] = state
		}
		if cmd.Flags().Changed("priority") {
			priority, _ := cmd.Flags().GetInt("priority")
			input["priority"] = priority
		}
		if cmd.Flags().Changed("start-date") {
			startDate, _ := cmd.Flags().GetString("start-date")
			if startDate != "" {
				if _, err := time.Parse("2006-01-02", startDate); err != nil {
					output.Error("Invalid --start-date format. Expected YYYY-MM-DD", plaintext, jsonOut)
					os.Exit(1)
				}
			}
			input["startDate"] = startDate
		}
		if cmd.Flags().Changed("lead") {
			leadEmail, _ := cmd.Flags().GetString("lead")
			if leadEmail != "" {
				user, err := client.(*api.Client).GetUser(context.Background(), leadEmail)
				if err != nil {
					output.Error(fmt.Sprintf("Lead user not found with email '%s': %v", leadEmail, err), plaintext, jsonOut)
					os.Exit(1)
				}
				input["leadId"] = user.ID
			}
		}
		if cmd.Flags().Changed("members") {
			members, _ := cmd.Flags().GetString("members")
			memberIDs, err := lookupUserIDsByEmails(context.Background(), client, members)
			if err != nil {
				output.Error(err.Error(), plaintext, jsonOut)
				os.Exit(1)
			}
			if len(memberIDs) > 0 {
				input["memberIds"] = memberIDs
			}
		}
		if cmd.Flags().Changed("label") {
			labelNames, _ := cmd.Flags().GetString("label")
			labelIDs, err := lookupLabelIDsByNames(context.Background(), client, labelNames)
			if err != nil {
				output.Error(err.Error(), plaintext, jsonOut)
				os.Exit(1)
			}
			if len(labelIDs) > 0 {
				input["labelIds"] = labelIDs
			}
		}
		if cmd.Flags().Changed("icon") {
			icon, _ := cmd.Flags().GetString("icon")
			input["icon"] = icon
		}
		if cmd.Flags().Changed("color") {
			projectColor, _ := cmd.Flags().GetString("color")
			if err := validateHexColor(projectColor); err != nil {
				output.Error(err.Error(), plaintext, jsonOut)
				os.Exit(1)
			}
			input["color"] = projectColor
		}
		if cmd.Flags().Changed("link") {
			links, _ := cmd.Flags().GetStringArray("link")
			if len(links) > 0 {
				input["links"] = links
			}
		}

		// Validate at least one field provided
		if len(input) == 0 {
			output.Error("At least one field to update is required", plaintext, jsonOut)
			os.Exit(1)
		}

		// Validate state if provided
		if state, ok := input["state"].(string); ok {
			allowedStates := []string{"planned", "started", "paused", "completed", "canceled"}
			valid := false
			for _, s := range allowedStates {
				if state == s {
					valid = true
					break
				}
			}
			if !valid {
				output.Error(fmt.Sprintf("Invalid state. Must be one of: %s", strings.Join(allowedStates, ", ")), plaintext, jsonOut)
				os.Exit(1)
			}
		}

		// Validate priority if provided
		if priority, ok := input["priority"].(int); ok {
			if priority < 0 || priority > 4 {
				output.Error("Priority must be between 0 and 4", plaintext, jsonOut)
				os.Exit(1)
			}
		}

		// Update project
		project, err := client.UpdateProject(context.Background(), projectID, input)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to update project: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Handle output
		if jsonOut {
			output.JSON(project)
		} else if plaintext {
			fmt.Printf("# Project Updated\n\n")
			fmt.Printf("- **Name**: %s\n", project.Name)
			fmt.Printf("- **ID**: %s\n", project.ID)
			if project.State != "" {
				fmt.Printf("- **State**: %s\n", project.State)
			}
			if project.Priority > 0 {
				fmt.Printf("- **Priority**: %d\n", project.Priority)
			}
			if project.Description != "" {
				fmt.Printf("- **Description**: %s\n", project.Description)
			}
			fmt.Printf("- **URL**: %s\n", constructProjectURL(project.ID, project.URL))
		} else {
			fmt.Println()
			fmt.Printf("%s Project updated successfully\n", color.New(color.FgGreen).Sprint("✓"))
			fmt.Println()
			fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("Name:"), project.Name)
			fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("ID:"), project.ID)
			if project.State != "" {
				fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("State:"), project.State)
			}
			if project.Priority > 0 {
				fmt.Printf("%s %d\n", color.New(color.Bold).Sprint("Priority:"), project.Priority)
			}
			fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("URL:"), color.New(color.FgBlue, color.Underline).Sprint(constructProjectURL(project.ID, project.URL)))
			fmt.Println()
		}
	},
}

// Project update-post commands

var projectUpdatePostCmd = &cobra.Command{
	Use:   "update-post",
	Short: "Manage project update posts",
	Long:  `Create, list, and view project update posts.`,
}

var projectUpdatePostCreateCmd = &cobra.Command{
	Use:   "create PROJECT-UUID",
	Short: "Create a project update post",
	Long: `Create a new update post for a project.

The project UUID is required as the first argument.

Examples:
  linctl project update-post create PROJECT-UUID --body "Monthly update..."
  linctl project update-post create PROJECT-UUID --body "Q1 progress" --health "onTrack"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		projectID := args[0]
		body, _ := cmd.Flags().GetString("body")
		health, _ := cmd.Flags().GetString("health")

		// Validate body is provided
		if body == "" {
			output.Error("--body is required", plaintext, jsonOut)
			os.Exit(1)
		}

		// Validate health if provided
		if health != "" {
			allowedHealth := []string{"onTrack", "atRisk", "offTrack"}
			valid := false
			for _, h := range allowedHealth {
				if health == h {
					valid = true
					break
				}
			}
			if !valid {
				output.Error(fmt.Sprintf("Invalid health. Must be one of: %s", strings.Join(allowedHealth, ", ")), plaintext, jsonOut)
				os.Exit(1)
			}
		}

		// Get auth header
		authHeader, err := getAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Create API client
		client := newAPIClient(authHeader)

		// Build input
		input := map[string]interface{}{
			"projectId": projectID,
			"body":      body,
		}
		if health != "" {
			input["health"] = health
		}

		// Create project update
		update, err := client.CreateProjectUpdate(context.Background(), input)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to create project update: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(update)
			return
		}

		if plaintext {
			fmt.Println("✓ Project update created successfully")
			fmt.Printf("ID: %s\n", update.ID)
			fmt.Printf("Created: %s\n", update.CreatedAt.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Println()
			fmt.Printf("%s Project update created successfully\n", color.New(color.FgGreen).Sprint("✓"))
			fmt.Println()
			fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("ID:"), update.ID)
			if update.User != nil {
				fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("Author:"), update.User.Name)
			}
			if update.Health != "" {
				fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("Health:"), update.Health)
			}
			fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("Created:"), update.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Println()
		}
	},
}

var projectUpdatePostListCmd = &cobra.Command{
	Use:   "list PROJECT-UUID",
	Short: "List project update posts",
	Long: `List all update posts for a project.

Examples:
  linctl project update-post list PROJECT-UUID
  linctl project update-post list PROJECT-UUID --json`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		projectID := args[0]

		// Get auth header
		authHeader, err := getAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Create API client
		client := newAPIClient(authHeader)

		// List project updates
		updates, err := client.ListProjectUpdates(context.Background(), projectID)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to list project updates: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if len(updates.Nodes) == 0 {
			if jsonOut {
				output.JSON([]interface{}{})
			} else {
				output.Info("No project updates found", plaintext, jsonOut)
			}
			return
		}

		if jsonOut {
			output.JSON(updates.Nodes)
			return
		}

		// Table output
		headers := []string{"ID", "Author", "Health", "Created", "Updated"}
		rows := [][]string{}

		for _, update := range updates.Nodes {
			author := ""
			if update.User != nil {
				author = update.User.Name
			}

			health := update.Health
			if health == "" {
				health = "N/A"
			}

			created := update.CreatedAt.Format("2006-01-02")
			updated := update.UpdatedAt.Format("2006-01-02")

			rows = append(rows, []string{
				update.ID,
				author,
				health,
				created,
				updated,
			})
		}

		output.Table(output.TableData{Headers: headers, Rows: rows}, plaintext, jsonOut)
	},
}

var projectUpdatePostGetCmd = &cobra.Command{
	Use:   "get UPDATE-UUID",
	Short: "Get a project update post",
	Long: `Get details of a specific project update post.

Examples:
  linctl project update-post get UPDATE-UUID
  linctl project update-post get UPDATE-UUID --json`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		updateID := args[0]

		// Get auth header
		authHeader, err := getAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Create API client
		client := newAPIClient(authHeader)

		// Get project update
		update, err := client.GetProjectUpdate(context.Background(), updateID)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to get project update: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(update)
			return
		}

		// Format output
		if plaintext {
			fmt.Printf("ID: %s\n", update.ID)
			if update.User != nil {
				fmt.Printf("Author: %s (%s)\n", update.User.Name, update.User.Email)
			}
			if update.Health != "" {
				fmt.Printf("Health: %s\n", update.Health)
			}
			fmt.Printf("Created: %s\n", update.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("Updated: %s\n", update.UpdatedAt.Format("2006-01-02 15:04:05"))
			if update.EditedAt != nil {
				fmt.Printf("Edited: %s\n", update.EditedAt.Format("2006-01-02 15:04:05"))
			}
			fmt.Println()
			fmt.Println("Body:")
			fmt.Println(update.Body)
		} else {
			fmt.Println()
			fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("ID:"), update.ID)
			if update.User != nil {
				fmt.Printf("%s %s (%s)\n", color.New(color.Bold).Sprint("Author:"), update.User.Name, update.User.Email)
			}
			if update.Health != "" {
				fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("Health:"), update.Health)
			}
			fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("Created:"), update.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("Updated:"), update.UpdatedAt.Format("2006-01-02 15:04:05"))
			if update.EditedAt != nil {
				fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("Edited:"), update.EditedAt.Format("2006-01-02 15:04:05"))
			}
			fmt.Println()
			fmt.Println(color.New(color.Bold).Sprint("Body:"))
			fmt.Println(update.Body)
			fmt.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectGetCmd)
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectArchiveCmd)
	projectCmd.AddCommand(projectUpdateCmd)
	projectCmd.AddCommand(projectUpdatePostCmd)

	// Project update-post subcommands
	projectUpdatePostCmd.AddCommand(projectUpdatePostCreateCmd)
	projectUpdatePostCmd.AddCommand(projectUpdatePostListCmd)
	projectUpdatePostCmd.AddCommand(projectUpdatePostGetCmd)

	// List command flags
	projectListCmd.Flags().StringP("team", "t", "", "Filter by team key")
	projectListCmd.Flags().StringP("state", "s", "", "Filter by state (planned, started, paused, completed, canceled)")
	projectListCmd.Flags().IntP("limit", "l", 50, "Maximum number of projects to return")
	projectListCmd.Flags().BoolP("include-completed", "c", false, "Include completed and canceled projects")
	projectListCmd.Flags().StringP("sort", "o", "linear", "Sort order: linear (default), created, updated")
	projectListCmd.Flags().StringP("newer-than", "n", "", "Show projects created after this time (default: 6_months_ago, use 'all_time' for no filter)")

	// Create command flags
	projectCreateCmd.Flags().String("name", "", "Project name (required)")
	projectCreateCmd.Flags().String("team", "", "Team key (required)")
	projectCreateCmd.Flags().String("description", "", "Project description")
	projectCreateCmd.Flags().String("state", "", "Project state (planned|started|paused|completed|canceled)")
	projectCreateCmd.Flags().Int("priority", 0, "Priority (0-4: None, Urgent, High, Normal, Low)")
	projectCreateCmd.Flags().String("start-date", "", "Start date (YYYY-MM-DD)")
	projectCreateCmd.Flags().String("target-date", "", "Target date (YYYY-MM-DD)")
	projectCreateCmd.Flags().String("lead", "", "Project lead (email)")
	projectCreateCmd.Flags().String("members", "", "Project members (comma-separated emails)")
	projectCreateCmd.Flags().String("label", "", "Project labels (comma-separated names)")
	projectCreateCmd.Flags().String("icon", "", "Project icon (emoji)")
	projectCreateCmd.Flags().String("color", "", "Project color (hex code, e.g., #ff6b6b)")
	projectCreateCmd.Flags().StringArray("link", []string{}, "External link URL (can be specified multiple times)")

	// Update command flags
	projectUpdateCmd.Flags().String("name", "", "Project name")
	projectUpdateCmd.Flags().String("description", "", "Project description")
	projectUpdateCmd.Flags().String("summary", "", "Project short summary")
	projectUpdateCmd.Flags().String("state", "", "Project state (planned|started|paused|completed|canceled)")
	projectUpdateCmd.Flags().Int("priority", 0, "Priority (0-4: None, Urgent, High, Normal, Low)")
	projectUpdateCmd.Flags().String("start-date", "", "Start date (YYYY-MM-DD)")
	projectUpdateCmd.Flags().String("lead", "", "Project lead (email)")
	projectUpdateCmd.Flags().String("members", "", "Project members (comma-separated emails)")
	projectUpdateCmd.Flags().String("label", "", "Project labels (comma-separated names)")
	projectUpdateCmd.Flags().String("icon", "", "Project icon (emoji)")
	projectUpdateCmd.Flags().String("color", "", "Project color (hex code, e.g., #ff6b6b)")
	projectUpdateCmd.Flags().StringArray("link", []string{}, "External link URL (can be specified multiple times)")

	// Project update-post create flags
	projectUpdatePostCreateCmd.Flags().String("body", "", "Update post body (required)")
	projectUpdatePostCreateCmd.Flags().String("health", "", "Project health (onTrack|atRisk|offTrack)")
}
