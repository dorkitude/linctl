package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dorkitude/linctl/pkg/api"
	"github.com/dorkitude/linctl/pkg/auth"
	"github.com/dorkitude/linctl/pkg/output"
	"github.com/dorkitude/linctl/pkg/utils"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Create API client
		client := api.NewClient(authHeader)

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
			headers := []string{"Name", "State", "Lead", "Teams", "Created", "Updated", "URL"}
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

				rows = append(rows, []string{
					truncateString(project.Name, 25),
					stateColor.Sprint(project.State),
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
		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Create API client
		client := api.NewClient(authHeader)

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
			fmt.Printf("- **Progress**: %.0f%%\n", project.Progress*100)
			fmt.Printf("- **Health**: %s\n", project.Health)
			fmt.Printf("- **Scope**: %d\n", project.Scope)
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

			progressColor := color.New(color.FgRed)
			if project.Progress >= 0.75 {
				progressColor = color.New(color.FgGreen)
			} else if project.Progress >= 0.5 {
				progressColor = color.New(color.FgYellow)
			}
			fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("Progress:"), progressColor.Sprintf("%.0f%%", project.Progress*100))

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
	Use:     "create",
	Aliases: []string{"new"},
	Short:   "Create a new project",
	Long: `Create a new project in Linear.

Examples:
  linctl project create --name "Q1 Release" --team ENG
  linctl project create --name "Auth Overhaul" --team ENG --description "Rewrite authentication system"
  linctl project create --name "Mobile App" --team ENG --lead me --start-date 2024-01-01 --target-date 2024-06-30
  linctl project create --name "Bug Bash" --team ENG,QA --state started`,
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)

		// Get required flags
		name, _ := cmd.Flags().GetString("name")
		teamKeys, _ := cmd.Flags().GetStringSlice("team")

		if name == "" {
			output.Error("Project name is required (--name)", plaintext, jsonOut)
			os.Exit(1)
		}

		if len(teamKeys) == 0 {
			output.Error("At least one team is required (--team)", plaintext, jsonOut)
			os.Exit(1)
		}

		// Build input
		input := map[string]interface{}{
			"name": name,
		}

		// Resolve team IDs
		var teamIDs []string
		for _, teamKey := range teamKeys {
			team, err := client.GetTeam(context.Background(), teamKey)
			if err != nil {
				output.Error(fmt.Sprintf("Failed to find team '%s': %v", teamKey, err), plaintext, jsonOut)
				os.Exit(1)
			}
			teamIDs = append(teamIDs, team.ID)
		}
		input["teamIds"] = teamIDs

		// Optional fields
		if cmd.Flags().Changed("description") {
			description, _ := cmd.Flags().GetString("description")
			input["description"] = description
		}

		if cmd.Flags().Changed("state") {
			state, _ := cmd.Flags().GetString("state")
			validStates := []string{"planned", "started", "paused", "completed", "canceled"}
			isValid := false
			for _, vs := range validStates {
				if strings.EqualFold(state, vs) {
					input["state"] = strings.ToLower(state)
					isValid = true
					break
				}
			}
			if !isValid {
				output.Error(fmt.Sprintf("Invalid state '%s'. Valid states: %s", state, strings.Join(validStates, ", ")), plaintext, jsonOut)
				os.Exit(1)
			}
		}

		if cmd.Flags().Changed("lead") {
			leadValue, _ := cmd.Flags().GetString("lead")
			if leadValue == "me" {
				viewer, err := client.GetViewer(context.Background())
				if err != nil {
					output.Error(fmt.Sprintf("Failed to get current user: %v", err), plaintext, jsonOut)
					os.Exit(1)
				}
				input["leadId"] = viewer.ID
			} else {
				users, err := client.GetUsers(context.Background(), 100, "", "")
				if err != nil {
					output.Error(fmt.Sprintf("Failed to get users: %v", err), plaintext, jsonOut)
					os.Exit(1)
				}
				var foundUser *api.User
				for _, user := range users.Nodes {
					if user.Email == leadValue || user.Name == leadValue {
						foundUser = &user
						break
					}
				}
				if foundUser == nil {
					output.Error(fmt.Sprintf("User not found: %s", leadValue), plaintext, jsonOut)
					os.Exit(1)
				}
				input["leadId"] = foundUser.ID
			}
		}

		if cmd.Flags().Changed("start-date") {
			startDate, _ := cmd.Flags().GetString("start-date")
			input["startDate"] = startDate
		}

		if cmd.Flags().Changed("target-date") {
			targetDate, _ := cmd.Flags().GetString("target-date")
			input["targetDate"] = targetDate
		}

		if cmd.Flags().Changed("color") {
			colorValue, _ := cmd.Flags().GetString("color")
			input["color"] = colorValue
		}

		// Create project
		project, err := client.CreateProject(context.Background(), input)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to create project: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Output
		if jsonOut {
			output.JSON(project)
		} else if plaintext {
			fmt.Printf("Created project: %s\n", project.Name)
			fmt.Printf("ID: %s\n", project.ID)
			fmt.Printf("URL: %s\n", constructProjectURL(project.ID, project.URL))
		} else {
			fmt.Printf("%s Created project %s\n",
				color.New(color.FgGreen).Sprint("✓"),
				color.New(color.FgCyan, color.Bold).Sprint(project.Name))
			fmt.Printf("  ID: %s\n", project.ID)
			fmt.Printf("  URL: %s\n", color.New(color.FgBlue, color.Underline).Sprint(constructProjectURL(project.ID, project.URL)))
		}
	},
}

var projectUpdateCmd = &cobra.Command{
	Use:   "update [project-id]",
	Short: "Update a project",
	Long: `Update an existing project's properties.

Examples:
  linctl project update abc123 --name "New Name"
  linctl project update abc123 --description "Updated description"
  linctl project update abc123 --state started
  linctl project update abc123 --lead john@company.com
  linctl project update abc123 --target-date 2024-12-31
  linctl project update abc123 --state completed`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		projectID := args[0]

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)

		// Build update input
		input := make(map[string]interface{})

		if cmd.Flags().Changed("name") {
			name, _ := cmd.Flags().GetString("name")
			input["name"] = name
		}

		if cmd.Flags().Changed("description") {
			description, _ := cmd.Flags().GetString("description")
			input["description"] = description
		}

		if cmd.Flags().Changed("state") {
			state, _ := cmd.Flags().GetString("state")
			validStates := []string{"planned", "started", "paused", "completed", "canceled"}
			isValid := false
			for _, vs := range validStates {
				if strings.EqualFold(state, vs) {
					input["state"] = strings.ToLower(state)
					isValid = true
					break
				}
			}
			if !isValid {
				output.Error(fmt.Sprintf("Invalid state '%s'. Valid states: %s", state, strings.Join(validStates, ", ")), plaintext, jsonOut)
				os.Exit(1)
			}
		}

		if cmd.Flags().Changed("lead") {
			leadValue, _ := cmd.Flags().GetString("lead")
			switch strings.ToLower(leadValue) {
			case "none", "unassigned", "":
				input["leadId"] = nil
			case "me":
				viewer, err := client.GetViewer(context.Background())
				if err != nil {
					output.Error(fmt.Sprintf("Failed to get current user: %v", err), plaintext, jsonOut)
					os.Exit(1)
				}
				input["leadId"] = viewer.ID
			default:
				users, err := client.GetUsers(context.Background(), 100, "", "")
				if err != nil {
					output.Error(fmt.Sprintf("Failed to get users: %v", err), plaintext, jsonOut)
					os.Exit(1)
				}
				var foundUser *api.User
				for _, user := range users.Nodes {
					if user.Email == leadValue || user.Name == leadValue {
						foundUser = &user
						break
					}
				}
				if foundUser == nil {
					output.Error(fmt.Sprintf("User not found: %s", leadValue), plaintext, jsonOut)
					os.Exit(1)
				}
				input["leadId"] = foundUser.ID
			}
		}

		if cmd.Flags().Changed("start-date") {
			startDate, _ := cmd.Flags().GetString("start-date")
			if startDate == "" {
				input["startDate"] = nil
			} else {
				input["startDate"] = startDate
			}
		}

		if cmd.Flags().Changed("target-date") {
			targetDate, _ := cmd.Flags().GetString("target-date")
			if targetDate == "" {
				input["targetDate"] = nil
			} else {
				input["targetDate"] = targetDate
			}
		}

		if cmd.Flags().Changed("color") {
			colorValue, _ := cmd.Flags().GetString("color")
			input["color"] = colorValue
		}

		// Check if any updates were specified
		if len(input) == 0 {
			output.Error("No updates specified. Use flags to specify what to update.", plaintext, jsonOut)
			os.Exit(1)
		}

		// Update project
		project, err := client.UpdateProject(context.Background(), projectID, input)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to update project: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Output
		if jsonOut {
			output.JSON(project)
		} else if plaintext {
			fmt.Printf("Updated project: %s\n", project.Name)
		} else {
			fmt.Printf("%s Updated project %s\n",
				color.New(color.FgGreen).Sprint("✓"),
				color.New(color.FgCyan, color.Bold).Sprint(project.Name))
		}
	},
}

var projectDeleteCmd = &cobra.Command{
	Use:     "delete [project-id]",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete or archive a project",
	Long: `Delete or archive a project.

By default, this command archives the project (soft delete).
Use --permanent to permanently delete (cannot be undone).

Examples:
  linctl project delete abc123              # Archive project
  linctl project delete abc123 --permanent  # Permanent delete (use with caution)
  linctl project delete abc123 --force      # Skip confirmation prompt`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		projectID := args[0]

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)

		permanent, _ := cmd.Flags().GetBool("permanent")
		force, _ := cmd.Flags().GetBool("force")

		// Get project details for confirmation message
		project, err := client.GetProject(context.Background(), projectID)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to find project '%s': %v", projectID, err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Confirmation prompt (unless --force or --json)
		if !force && !jsonOut {
			action := "archive"
			if permanent {
				action = "PERMANENTLY DELETE"
			}
			fmt.Printf("Are you sure you want to %s project '%s'? [y/N]: ", action, project.Name)

			var response string
			fmt.Scanln(&response)
			response = strings.ToLower(strings.TrimSpace(response))

			if response != "y" && response != "yes" {
				fmt.Println("Cancelled.")
				return
			}
		}

		if permanent {
			// Permanent delete
			err = client.DeleteProject(context.Background(), projectID)
			if err != nil {
				output.Error(fmt.Sprintf("Failed to delete project: %v", err), plaintext, jsonOut)
				os.Exit(1)
			}

			if jsonOut {
				output.JSON(map[string]interface{}{
					"success":   true,
					"action":    "deleted",
					"projectId": projectID,
					"name":      project.Name,
				})
			} else if plaintext {
				fmt.Printf("Deleted project: %s\n", project.Name)
			} else {
				fmt.Printf("%s Permanently deleted project %s\n",
					color.New(color.FgRed).Sprint("✗"),
					color.New(color.FgCyan, color.Bold).Sprint(project.Name))
			}
		} else {
			// Archive (soft delete)
			archivedProject, err := client.ArchiveProject(context.Background(), projectID)
			if err != nil {
				output.Error(fmt.Sprintf("Failed to archive project: %v", err), plaintext, jsonOut)
				os.Exit(1)
			}

			if jsonOut {
				output.JSON(archivedProject)
			} else if plaintext {
				fmt.Printf("Archived project: %s\n", project.Name)
			} else {
				fmt.Printf("%s Archived project %s\n",
					color.New(color.FgYellow).Sprint("📦"),
					color.New(color.FgCyan, color.Bold).Sprint(project.Name))
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectGetCmd)
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectUpdateCmd)
	projectCmd.AddCommand(projectDeleteCmd)

	// List command flags
	projectListCmd.Flags().StringP("team", "t", "", "Filter by team key")
	projectListCmd.Flags().StringP("state", "s", "", "Filter by state (planned, started, paused, completed, canceled)")
	projectListCmd.Flags().IntP("limit", "l", 50, "Maximum number of projects to return")
	projectListCmd.Flags().BoolP("include-completed", "c", false, "Include completed and canceled projects")
	projectListCmd.Flags().StringP("sort", "o", "linear", "Sort order: linear (default), created, updated")
	projectListCmd.Flags().StringP("newer-than", "n", "", "Show projects created after this time (default: 6_months_ago, use 'all_time' for no filter)")

	// Create command flags
	projectCreateCmd.Flags().String("name", "", "Project name (required)")
	projectCreateCmd.Flags().StringSliceP("team", "t", []string{}, "Team key(s) (required, comma-separated for multiple)")
	projectCreateCmd.Flags().StringP("description", "d", "", "Project description")
	projectCreateCmd.Flags().StringP("state", "s", "", "Initial state (planned, started, paused)")
	projectCreateCmd.Flags().String("lead", "", "Project lead (email, name, or 'me')")
	projectCreateCmd.Flags().String("start-date", "", "Start date (YYYY-MM-DD)")
	projectCreateCmd.Flags().String("target-date", "", "Target date (YYYY-MM-DD)")
	projectCreateCmd.Flags().String("color", "", "Project color (hex code)")
	_ = projectCreateCmd.MarkFlagRequired("name")
	_ = projectCreateCmd.MarkFlagRequired("team")

	// Update command flags
	projectUpdateCmd.Flags().String("name", "", "New project name")
	projectUpdateCmd.Flags().StringP("description", "d", "", "New description")
	projectUpdateCmd.Flags().StringP("state", "s", "", "State (planned, started, paused, completed, canceled)")
	projectUpdateCmd.Flags().String("lead", "", "Project lead (email, name, 'me', or 'none' to remove)")
	projectUpdateCmd.Flags().String("start-date", "", "Start date (YYYY-MM-DD, or empty to remove)")
	projectUpdateCmd.Flags().String("target-date", "", "Target date (YYYY-MM-DD, or empty to remove)")
	projectUpdateCmd.Flags().String("color", "", "Project color (hex code)")

	// Delete command flags
	projectDeleteCmd.Flags().Bool("permanent", false, "Permanently delete (cannot be undone)")
	projectDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}
