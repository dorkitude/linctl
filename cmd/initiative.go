package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dorkitude/linctl/pkg/api"
	"github.com/dorkitude/linctl/pkg/auth"
	"github.com/dorkitude/linctl/pkg/output"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// initiativeCmd represents the initiative command
var initiativeCmd = &cobra.Command{
	Use:   "initiative",
	Short: "Manage Linear initiatives",
	Long: `Manage Linear initiatives including listing, viewing, creating, and linking to projects.

Examples:
  linctl initiative list                        # List active initiatives
  linctl initiative list --include-completed    # List all initiatives
  linctl initiative get INITIATIVE-ID           # Get initiative details
  linctl initiative create --name "Q3 Goals"    # Create a new initiative
  linctl initiative link INIT-ID --project PROJECT-ID  # Link to project`,
}

var initiativeListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List initiatives",
	Long:    `List all initiatives in your Linear workspace.`,
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)

		status, _ := cmd.Flags().GetString("status")
		limit, _ := cmd.Flags().GetInt("limit")
		includeCompleted, _ := cmd.Flags().GetBool("include-completed")

		filter := make(map[string]interface{})
		if status != "" {
			// Validate and normalize status
			validStatuses := []string{"Planned", "Active", "Completed"}
			normalized := normalizeInitiativeStatus(status)
			if normalized == "" {
				output.Error(fmt.Sprintf("Invalid status '%s'. Valid statuses: %s", status, strings.Join(validStatuses, ", ")), plaintext, jsonOut)
				os.Exit(1)
			}
			filter["status"] = map[string]interface{}{"eq": normalized}
		} else if !includeCompleted {
			filter["status"] = map[string]interface{}{
				"neq": "Completed",
			}
		}

		initiatives, err := client.GetInitiatives(context.Background(), filter, limit, "")
		if err != nil {
			output.Error(fmt.Sprintf("Failed to list initiatives: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(initiatives.Nodes)
			return
		} else if plaintext {
			fmt.Println("# Initiatives")
			for _, init := range initiatives.Nodes {
				fmt.Printf("## %s\n", init.Name)
				fmt.Printf("- **ID**: %s\n", init.ID)
				fmt.Printf("- **Status**: %s\n", init.Status)
				if init.Owner != nil {
					fmt.Printf("- **Owner**: %s\n", init.Owner.Name)
				} else {
					fmt.Printf("- **Owner**: Unassigned\n")
				}
				if init.TargetDate != nil {
					fmt.Printf("- **Target Date**: %s\n", *init.TargetDate)
				}
				if init.Projects != nil && len(init.Projects.Nodes) > 0 {
					projects := ""
					for i, p := range init.Projects.Nodes {
						if i > 0 {
							projects += ", "
						}
						projects += p.Name
					}
					fmt.Printf("- **Projects**: %s\n", projects)
				}
				fmt.Printf("- **Created**: %s\n", init.CreatedAt.Format("2006-01-02"))
				fmt.Printf("- **Updated**: %s\n", init.UpdatedAt.Format("2006-01-02"))
				fmt.Printf("- **URL**: %s\n", init.URL)
				if init.Description != "" {
					fmt.Printf("- **Description**: %s\n", init.Description)
				}
				fmt.Println()
			}
			fmt.Printf("\nTotal: %d initiatives\n", len(initiatives.Nodes))
			return
		}

		// Table output
		headers := []string{"Name", "Status", "Owner", "Projects", "Target Date", "Created", "URL"}
		rows := [][]string{}

		for _, init := range initiatives.Nodes {
			owner := color.New(color.FgYellow).Sprint("Unassigned")
			if init.Owner != nil {
				owner = init.Owner.Name
			}

			projectCount := ""
			if init.Projects != nil {
				projectCount = fmt.Sprintf("%d", len(init.Projects.Nodes))
			}

			targetDate := ""
			if init.TargetDate != nil {
				targetDate = *init.TargetDate
			}

			statusColor := initiativeStatusColor(init.Status)

			rows = append(rows, []string{
				truncateString(init.Name, 30),
				statusColor.Sprint(init.Status),
				owner,
				projectCount,
				targetDate,
				init.CreatedAt.Format("2006-01-02"),
				init.URL,
			})
		}

		output.Table(output.TableData{
			Headers: headers,
			Rows:    rows,
		}, plaintext, jsonOut)

		fmt.Printf("\n%s %d initiatives\n",
			color.New(color.FgGreen).Sprint("✓"),
			len(initiatives.Nodes))
	},
}

var initiativeGetCmd = &cobra.Command{
	Use:     "get INITIATIVE-ID",
	Aliases: []string{"show"},
	Short:   "Get initiative details",
	Long:    `Get detailed information about a specific initiative.`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		initiativeID := args[0]

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)

		initiative, err := client.GetInitiative(context.Background(), initiativeID)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to get initiative: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(initiative)
			return
		} else if plaintext {
			fmt.Printf("# %s\n\n", initiative.Name)

			if initiative.Description != "" {
				fmt.Printf("## Description\n%s\n\n", initiative.Description)
			}

			if initiative.Content != "" {
				fmt.Printf("## Content\n%s\n\n", initiative.Content)
			}

			fmt.Printf("## Core Details\n")
			fmt.Printf("- **ID**: %s\n", initiative.ID)
			fmt.Printf("- **Slug ID**: %s\n", initiative.SlugId)
			fmt.Printf("- **Status**: %s\n", initiative.Status)
			if initiative.Icon != nil && *initiative.Icon != "" {
				fmt.Printf("- **Icon**: %s\n", *initiative.Icon)
			}
			if initiative.Color != nil && *initiative.Color != "" {
				fmt.Printf("- **Color**: %s\n", *initiative.Color)
			}

			fmt.Printf("\n## Timeline\n")
			if initiative.TargetDate != nil {
				fmt.Printf("- **Target Date**: %s\n", *initiative.TargetDate)
			}
			fmt.Printf("- **Created**: %s\n", initiative.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("- **Updated**: %s\n", initiative.UpdatedAt.Format("2006-01-02 15:04:05"))
			if initiative.StartedAt != nil {
				fmt.Printf("- **Started**: %s\n", initiative.StartedAt.Format("2006-01-02 15:04:05"))
			}
			if initiative.CompletedAt != nil {
				fmt.Printf("- **Completed**: %s\n", initiative.CompletedAt.Format("2006-01-02 15:04:05"))
			}
			if initiative.ArchivedAt != nil {
				fmt.Printf("- **Archived**: %s\n", initiative.ArchivedAt.Format("2006-01-02 15:04:05"))
			}

			fmt.Printf("\n## People\n")
			if initiative.Owner != nil {
				fmt.Printf("- **Owner**: %s (%s)\n", initiative.Owner.Name, initiative.Owner.Email)
			} else {
				fmt.Printf("- **Owner**: Unassigned\n")
			}
			if initiative.Creator != nil {
				fmt.Printf("- **Creator**: %s (%s)\n", initiative.Creator.Name, initiative.Creator.Email)
			}

			// Hierarchy
			if initiative.ParentInitiative != nil {
				fmt.Printf("\n## Parent Initiative\n")
				fmt.Printf("- **%s** (%s)\n", initiative.ParentInitiative.Name, initiative.ParentInitiative.Status)
			}

			if initiative.SubInitiatives != nil && len(initiative.SubInitiatives.Nodes) > 0 {
				fmt.Printf("\n## Sub-Initiatives\n")
				for _, sub := range initiative.SubInitiatives.Nodes {
					owner := "Unassigned"
					if sub.Owner != nil {
						owner = sub.Owner.Name
					}
					fmt.Printf("- **%s** [%s] (%s)\n", sub.Name, sub.Status, owner)
				}
			}

			// Projects
			if initiative.Projects != nil && len(initiative.Projects.Nodes) > 0 {
				fmt.Printf("\n## Projects (%d)\n", len(initiative.Projects.Nodes))
				for _, project := range initiative.Projects.Nodes {
					lead := "Unassigned"
					if project.Lead != nil {
						lead = project.Lead.Name
					}
					fmt.Printf("- **%s** [%s] (Lead: %s, Progress: %.0f%%)\n",
						project.Name, project.State, lead, project.Progress*100)
				}
			}

			// Initiative Updates
			if initiative.InitiativeUpdates != nil && len(initiative.InitiativeUpdates.Nodes) > 0 {
				fmt.Printf("\n## Recent Updates\n")
				for _, update := range initiative.InitiativeUpdates.Nodes {
					fmt.Printf("\n### %s by %s\n", update.CreatedAt.Format("2006-01-02 15:04"), update.User.Name)
					fmt.Printf("- **Health**: %s\n", update.Health)
					fmt.Printf("\n%s\n", update.Body)
				}
			}

			fmt.Printf("\n## URL\n")
			fmt.Printf("- %s\n", initiative.URL)
		} else {
			// Formatted output
			fmt.Println()
			fmt.Printf("%s %s\n", color.New(color.FgCyan, color.Bold).Sprint("🎯 Initiative:"), initiative.Name)
			fmt.Println(strings.Repeat("─", 50))

			fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("ID:"), initiative.ID)

			if initiative.Description != "" {
				fmt.Printf("\n%s\n%s\n", color.New(color.Bold).Sprint("Description:"), initiative.Description)
			}

			statusColor := initiativeStatusColor(initiative.Status)
			fmt.Printf("\n%s %s\n", color.New(color.Bold).Sprint("Status:"), statusColor.Sprint(initiative.Status))

			if initiative.TargetDate != nil {
				fmt.Printf("%s %s\n", color.New(color.Bold).Sprint("Target Date:"), *initiative.TargetDate)
			}

			if initiative.Owner != nil {
				fmt.Printf("\n%s %s (%s)\n",
					color.New(color.Bold).Sprint("Owner:"),
					initiative.Owner.Name,
					color.New(color.FgCyan).Sprint(initiative.Owner.Email))
			}

			// Hierarchy
			if initiative.ParentInitiative != nil {
				fmt.Printf("\n%s %s [%s]\n",
					color.New(color.Bold).Sprint("Parent:"),
					initiative.ParentInitiative.Name,
					initiative.ParentInitiative.Status)
			}

			if initiative.SubInitiatives != nil && len(initiative.SubInitiatives.Nodes) > 0 {
				fmt.Printf("\n%s\n", color.New(color.Bold).Sprint("Sub-Initiatives:"))
				for _, sub := range initiative.SubInitiatives.Nodes {
					subStatus := initiativeStatusColor(sub.Status)
					owner := "Unassigned"
					if sub.Owner != nil {
						owner = sub.Owner.Name
					}
					fmt.Printf("  • %s %s (%s)\n",
						subStatus.Sprint(sub.Status),
						sub.Name,
						color.New(color.FgWhite, color.Faint).Sprint(owner))
				}
			}

			// Projects
			if initiative.Projects != nil && len(initiative.Projects.Nodes) > 0 {
				fmt.Printf("\n%s\n", color.New(color.Bold).Sprint("Projects:"))
				for _, project := range initiative.Projects.Nodes {
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
					fmt.Printf("  • %s %s\n",
						stateColor.Sprintf("[%s]", project.State),
						project.Name)
				}
			}

			// Timeline
			fmt.Printf("\n%s\n", color.New(color.Bold).Sprint("Timeline:"))
			fmt.Printf("  Created: %s\n", initiative.CreatedAt.Format("2006-01-02"))
			fmt.Printf("  Updated: %s\n", initiative.UpdatedAt.Format("2006-01-02"))
			if initiative.StartedAt != nil {
				fmt.Printf("  Started: %s\n", initiative.StartedAt.Format("2006-01-02"))
			}
			if initiative.CompletedAt != nil {
				fmt.Printf("  Completed: %s\n", initiative.CompletedAt.Format("2006-01-02"))
			}

			if initiative.URL != "" {
				fmt.Printf("\n%s %s\n",
					color.New(color.Bold).Sprint("URL:"),
					color.New(color.FgBlue, color.Underline).Sprint(initiative.URL))
			}

			fmt.Println()
		}
	},
}

var initiativeCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"new"},
	Short:   "Create a new initiative",
	Long: `Create a new initiative in Linear.

Examples:
  linctl initiative create --name "Q3 Goals"
  linctl initiative create --name "Platform Migration" --description "Migrate to new platform" --owner me
  linctl initiative create --name "Revenue Growth" --status Active --target-date 2024-12-31`,
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)

		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			output.Error("Initiative name is required (--name)", plaintext, jsonOut)
			os.Exit(1)
		}

		input := map[string]interface{}{
			"name": name,
		}

		if cmd.Flags().Changed("description") {
			description, _ := cmd.Flags().GetString("description")
			input["description"] = description
		}

		if cmd.Flags().Changed("status") {
			status, _ := cmd.Flags().GetString("status")
			normalized := normalizeInitiativeStatus(status)
			if normalized == "" {
				output.Error(fmt.Sprintf("Invalid status '%s'. Valid statuses: Planned, Active, Completed", status), plaintext, jsonOut)
				os.Exit(1)
			}
			input["status"] = normalized
		}

		if cmd.Flags().Changed("owner") {
			ownerValue, _ := cmd.Flags().GetString("owner")
			ownerID, err := resolveUserID(client, ownerValue)
			if err != nil {
				output.Error(fmt.Sprintf("Failed to resolve owner: %v", err), plaintext, jsonOut)
				os.Exit(1)
			}
			input["ownerId"] = ownerID
		}

		if cmd.Flags().Changed("target-date") {
			targetDate, _ := cmd.Flags().GetString("target-date")
			input["targetDate"] = targetDate
		}

		if cmd.Flags().Changed("color") {
			colorValue, _ := cmd.Flags().GetString("color")
			input["color"] = colorValue
		}

		initiative, err := client.CreateInitiative(context.Background(), input)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to create initiative: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(initiative)
		} else if plaintext {
			fmt.Printf("Created initiative: %s\n", initiative.Name)
			fmt.Printf("ID: %s\n", initiative.ID)
			fmt.Printf("URL: %s\n", initiative.URL)
		} else {
			fmt.Printf("%s Created initiative %s\n",
				color.New(color.FgGreen).Sprint("✓"),
				color.New(color.FgCyan, color.Bold).Sprint(initiative.Name))
			fmt.Printf("  ID: %s\n", initiative.ID)
			fmt.Printf("  URL: %s\n", color.New(color.FgBlue, color.Underline).Sprint(initiative.URL))
		}
	},
}

var initiativeUpdateCmd = &cobra.Command{
	Use:   "update [initiative-id]",
	Short: "Update an initiative",
	Long: `Update an existing initiative's properties.

Examples:
  linctl initiative update abc123 --name "New Name"
  linctl initiative update abc123 --status Active
  linctl initiative update abc123 --owner me --target-date 2024-12-31`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		initiativeID := args[0]

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)

		input := map[string]interface{}{}

		if cmd.Flags().Changed("name") {
			name, _ := cmd.Flags().GetString("name")
			input["name"] = name
		}

		if cmd.Flags().Changed("description") {
			description, _ := cmd.Flags().GetString("description")
			input["description"] = description
		}

		if cmd.Flags().Changed("status") {
			status, _ := cmd.Flags().GetString("status")
			normalized := normalizeInitiativeStatus(status)
			if normalized == "" {
				output.Error(fmt.Sprintf("Invalid status '%s'. Valid statuses: Planned, Active, Completed", status), plaintext, jsonOut)
				os.Exit(1)
			}
			input["status"] = normalized
		}

		if cmd.Flags().Changed("owner") {
			ownerValue, _ := cmd.Flags().GetString("owner")
			if ownerValue == "none" {
				input["ownerId"] = nil
			} else {
				ownerID, err := resolveUserID(client, ownerValue)
				if err != nil {
					output.Error(fmt.Sprintf("Failed to resolve owner: %v", err), plaintext, jsonOut)
					os.Exit(1)
				}
				input["ownerId"] = ownerID
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

		if len(input) == 0 {
			output.Error("No update flags provided. Use --name, --description, --status, --owner, --target-date, or --color.", plaintext, jsonOut)
			os.Exit(1)
		}

		initiative, err := client.UpdateInitiative(context.Background(), initiativeID, input)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to update initiative: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(initiative)
		} else if plaintext {
			fmt.Printf("Updated initiative: %s\n", initiative.Name)
			fmt.Printf("ID: %s\n", initiative.ID)
		} else {
			fmt.Printf("%s Updated initiative %s\n",
				color.New(color.FgGreen).Sprint("✓"),
				color.New(color.FgCyan, color.Bold).Sprint(initiative.Name))
		}
	},
}

var initiativeDeleteCmd = &cobra.Command{
	Use:     "delete [initiative-id]",
	Aliases: []string{"rm", "remove"},
	Short:   "Permanently delete an initiative",
	Long: `Permanently delete an initiative. This cannot be undone.

To soft-delete, use 'linctl initiative archive' instead.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		initiativeID := args[0]

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)

		// Get initiative details for confirmation
		initiative, err := client.GetInitiative(context.Background(), initiativeID)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to get initiative: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		force, _ := cmd.Flags().GetBool("force")
		if !force && !plaintext && !jsonOut {
			fmt.Printf("%s Are you sure you want to permanently delete initiative '%s'? This cannot be undone.\n",
				color.New(color.FgRed).Sprint("⚠"),
				color.New(color.Bold).Sprint(initiative.Name))
			fmt.Print("Type 'yes' to confirm: ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "yes" {
				fmt.Println("Aborted.")
				return
			}
		}

		err = client.DeleteInitiative(context.Background(), initiativeID)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to delete initiative: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(map[string]interface{}{
				"deleted": true,
				"id":      initiativeID,
				"name":    initiative.Name,
			})
		} else if plaintext {
			fmt.Printf("Permanently deleted initiative: %s\n", initiative.Name)
		} else {
			fmt.Printf("%s Permanently deleted initiative %s\n",
				color.New(color.FgRed).Sprint("✗"),
				color.New(color.FgCyan, color.Bold).Sprint(initiative.Name))
		}
	},
}

var initiativeArchiveCmd = &cobra.Command{
	Use:   "archive [initiative-id]",
	Short: "Archive an initiative",
	Long:  `Archive an initiative. Archived initiatives can be restored with 'unarchive'.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		initiativeID := args[0]

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)

		archivedInitiative, err := client.ArchiveInitiative(context.Background(), initiativeID)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to archive initiative: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(archivedInitiative)
		} else if plaintext {
			fmt.Printf("Archived initiative: %s\n", archivedInitiative.Name)
		} else {
			fmt.Printf("%s Archived initiative %s\n",
				color.New(color.FgYellow).Sprint("📦"),
				color.New(color.FgCyan, color.Bold).Sprint(archivedInitiative.Name))
		}
	},
}

var initiativeUnarchiveCmd = &cobra.Command{
	Use:   "unarchive [initiative-id]",
	Short: "Unarchive an initiative",
	Long:  `Restore a previously archived initiative.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		initiativeID := args[0]

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)

		restoredInitiative, err := client.UnarchiveInitiative(context.Background(), initiativeID)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to unarchive initiative: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(restoredInitiative)
		} else if plaintext {
			fmt.Printf("Unarchived initiative: %s\n", restoredInitiative.Name)
		} else {
			fmt.Printf("%s Unarchived initiative %s\n",
				color.New(color.FgGreen).Sprint("✓"),
				color.New(color.FgCyan, color.Bold).Sprint(restoredInitiative.Name))
		}
	},
}

var initiativeLinkCmd = &cobra.Command{
	Use:   "link [initiative-id]",
	Short: "Link an initiative to a project",
	Long: `Link an initiative to a project.

Examples:
  linctl initiative link INITIATIVE-ID --project PROJECT-ID`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		initiativeID := args[0]

		projectID, _ := cmd.Flags().GetString("project")
		if projectID == "" {
			output.Error("Project ID is required (--project)", plaintext, jsonOut)
			os.Exit(1)
		}

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)

		link, err := client.LinkInitiativeToProject(context.Background(), initiativeID, projectID)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to link initiative to project: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(link)
		} else if plaintext {
			initName := initiativeID
			projName := projectID
			if link.Initiative != nil {
				initName = link.Initiative.Name
			}
			if link.Project != nil {
				projName = link.Project.Name
			}
			fmt.Printf("Linked initiative '%s' to project '%s'\n", initName, projName)
		} else {
			initName := initiativeID
			projName := projectID
			if link.Initiative != nil {
				initName = link.Initiative.Name
			}
			if link.Project != nil {
				projName = link.Project.Name
			}
			fmt.Printf("%s Linked initiative %s to project %s\n",
				color.New(color.FgGreen).Sprint("✓"),
				color.New(color.FgCyan, color.Bold).Sprint(initName),
				color.New(color.FgCyan, color.Bold).Sprint(projName))
		}
	},
}

var initiativeUnlinkCmd = &cobra.Command{
	Use:   "unlink [initiative-id]",
	Short: "Unlink an initiative from a project",
	Long: `Remove the link between an initiative and a project.

Examples:
  linctl initiative unlink INITIATIVE-ID --project PROJECT-ID`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		initiativeID := args[0]

		projectID, _ := cmd.Flags().GetString("project")
		if projectID == "" {
			output.Error("Project ID is required (--project)", plaintext, jsonOut)
			os.Exit(1)
		}

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)

		err = client.UnlinkInitiativeFromProject(context.Background(), initiativeID, projectID)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to unlink initiative from project: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(map[string]interface{}{
				"unlinked":     true,
				"initiativeId": initiativeID,
				"projectId":    projectID,
			})
		} else if plaintext {
			fmt.Printf("Unlinked initiative %s from project %s\n", initiativeID, projectID)
		} else {
			fmt.Printf("%s Unlinked initiative from project\n",
				color.New(color.FgGreen).Sprint("✓"))
		}
	},
}

// normalizeInitiativeStatus normalizes a status string to the GraphQL enum value.
// Returns empty string if the status is invalid.
func normalizeInitiativeStatus(status string) string {
	switch strings.ToLower(status) {
	case "planned":
		return "Planned"
	case "active":
		return "Active"
	case "completed":
		return "Completed"
	default:
		return ""
	}
}

// initiativeStatusColor returns the color for a given initiative status.
func initiativeStatusColor(status string) *color.Color {
	switch status {
	case "Planned":
		return color.New(color.FgCyan)
	case "Active":
		return color.New(color.FgGreen)
	case "Completed":
		return color.New(color.FgBlue)
	default:
		return color.New(color.FgWhite)
	}
}

// resolveUserID resolves a user identifier ("me", email, or name) to a Linear user ID.
func resolveUserID(client *api.Client, identifier string) (string, error) {
	if identifier == "me" {
		viewer, err := client.GetViewer(context.Background())
		if err != nil {
			return "", fmt.Errorf("failed to get current user: %w", err)
		}
		return viewer.ID, nil
	}

	users, err := client.GetUsers(context.Background(), 100, "", "")
	if err != nil {
		return "", fmt.Errorf("failed to get users: %w", err)
	}

	for _, user := range users.Nodes {
		if user.Email == identifier || user.Name == identifier {
			return user.ID, nil
		}
	}

	return "", fmt.Errorf("user not found: %s", identifier)
}

func init() {
	rootCmd.AddCommand(initiativeCmd)
	initiativeCmd.AddCommand(initiativeListCmd)
	initiativeCmd.AddCommand(initiativeGetCmd)
	initiativeCmd.AddCommand(initiativeCreateCmd)
	initiativeCmd.AddCommand(initiativeUpdateCmd)
	initiativeCmd.AddCommand(initiativeDeleteCmd)
	initiativeCmd.AddCommand(initiativeArchiveCmd)
	initiativeCmd.AddCommand(initiativeUnarchiveCmd)
	initiativeCmd.AddCommand(initiativeLinkCmd)
	initiativeCmd.AddCommand(initiativeUnlinkCmd)

	// List command flags
	initiativeListCmd.Flags().StringP("status", "s", "", "Filter by status (Planned, Active, Completed)")
	initiativeListCmd.Flags().IntP("limit", "l", 50, "Maximum number of initiatives to return")
	initiativeListCmd.Flags().BoolP("include-completed", "c", false, "Include completed initiatives")

	// Create command flags
	initiativeCreateCmd.Flags().String("name", "", "Initiative name (required)")
	initiativeCreateCmd.Flags().StringP("description", "d", "", "Initiative description")
	initiativeCreateCmd.Flags().StringP("status", "s", "", "Initial status (Planned, Active, Completed)")
	initiativeCreateCmd.Flags().String("owner", "", "Initiative owner (email, name, or 'me')")
	initiativeCreateCmd.Flags().String("target-date", "", "Target date (YYYY-MM-DD)")
	initiativeCreateCmd.Flags().String("color", "", "Initiative color (hex code)")
	_ = initiativeCreateCmd.MarkFlagRequired("name")

	// Update command flags
	initiativeUpdateCmd.Flags().String("name", "", "New initiative name")
	initiativeUpdateCmd.Flags().StringP("description", "d", "", "New description")
	initiativeUpdateCmd.Flags().StringP("status", "s", "", "Status (Planned, Active, Completed)")
	initiativeUpdateCmd.Flags().String("owner", "", "Initiative owner (email, name, 'me', or 'none' to remove)")
	initiativeUpdateCmd.Flags().String("target-date", "", "Target date (YYYY-MM-DD, or empty to remove)")
	initiativeUpdateCmd.Flags().String("color", "", "Initiative color (hex code)")

	// Delete command flags
	initiativeDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	// Link command flags
	initiativeLinkCmd.Flags().String("project", "", "Project ID to link (required)")
	_ = initiativeLinkCmd.MarkFlagRequired("project")

	// Unlink command flags
	initiativeUnlinkCmd.Flags().String("project", "", "Project ID to unlink (required)")
	_ = initiativeUnlinkCmd.MarkFlagRequired("project")
}
