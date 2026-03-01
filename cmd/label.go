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

var labelCmd = &cobra.Command{
	Use:   "label",
	Short: "Manage issue labels",
	Long: `Manage Linear issue labels with full CRUD support.

Examples:
  linctl label list --team ENG
  linctl label get <label-id>
  linctl label create --team ENG --name bug --color "#ff0000"
  linctl label update <label-id> --name "critical bug"
  linctl label delete <label-id>`,
}

var labelListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List labels for a team",
	Long:    `List issue labels for a specific team.`,
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
			os.Exit(1)
		}

		teamKey, _ := cmd.Flags().GetString("team")
		if strings.TrimSpace(teamKey) == "" {
			output.Error("--team is required", plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)
		labels, err := client.GetTeamLabels(context.Background(), teamKey)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to list labels: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(labels)
			return
		}

		if plaintext {
			fmt.Printf("# Labels (%s)\n", teamKey)
			for _, label := range labels {
				fmt.Printf("- %s (%s)\n", label.Name, label.ID)
				fmt.Printf("  Color: %s\n", label.Color)
				if label.Description != nil && *label.Description != "" {
					fmt.Printf("  Description: %s\n", *label.Description)
				}
				if label.Parent != nil {
					fmt.Printf("  Parent: %s (%s)\n", label.Parent.Name, label.Parent.ID)
				}
			}
			fmt.Printf("\nTotal: %d labels\n", len(labels))
			return
		}

		headers := []string{"Name", "ID", "Color", "Parent"}
		rows := make([][]string, 0, len(labels))
		for _, label := range labels {
			parent := "-"
			if label.Parent != nil {
				parent = label.Parent.Name
			}
			rows = append(rows, []string{label.Name, label.ID, label.Color, parent})
		}
		output.Table(output.TableData{Headers: headers, Rows: rows}, false, false)
		fmt.Printf("\n%s %d labels\n", color.New(color.FgGreen).Sprint("✓"), len(labels))
	},
}

var labelGetCmd = &cobra.Command{
	Use:   "get [label-id]",
	Short: "Get label details",
	Long: `Get detailed information for a label.

Use a label ID directly, or pass a label name with --team.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		labelRef := strings.TrimSpace(args[0])

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)

		label, err := client.GetLabel(context.Background(), labelRef)
		if err != nil {
			teamKey, _ := cmd.Flags().GetString("team")
			if strings.TrimSpace(teamKey) == "" {
				output.Error(fmt.Sprintf("Failed to fetch label by ID: %v. If using a label name, pass --team.", err), plaintext, jsonOut)
				os.Exit(1)
			}

			labels, listErr := client.GetTeamLabels(context.Background(), teamKey)
			if listErr != nil {
				output.Error(fmt.Sprintf("Failed to list team labels: %v", listErr), plaintext, jsonOut)
				os.Exit(1)
			}

			label = findLabelByNameOrID(labels, labelRef)
			if label == nil {
				output.Error(fmt.Sprintf("Label %q not found in team %s", labelRef, teamKey), plaintext, jsonOut)
				os.Exit(1)
			}
		}

		if jsonOut {
			output.JSON(label)
			return
		}

		if plaintext {
			fmt.Printf("ID: %s\n", label.ID)
			fmt.Printf("Name: %s\n", label.Name)
			fmt.Printf("Color: %s\n", label.Color)
			if label.Team != nil {
				fmt.Printf("Team: %s\n", label.Team.Key)
			}
			if label.Description != nil && *label.Description != "" {
				fmt.Printf("Description: %s\n", *label.Description)
			}
			if label.Parent != nil {
				fmt.Printf("Parent: %s (%s)\n", label.Parent.Name, label.Parent.ID)
			}
			return
		}

		fmt.Printf("%s %s\n", color.New(color.FgCyan, color.Bold).Sprint("Label"), color.New(color.FgWhite, color.Bold).Sprint(label.Name))
		fmt.Printf("ID: %s\n", color.New(color.FgCyan).Sprint(label.ID))
		fmt.Printf("Color: %s\n", label.Color)
		if label.Team != nil {
			fmt.Printf("Team: %s\n", label.Team.Key)
		}
		if label.Description != nil && *label.Description != "" {
			fmt.Printf("Description: %s\n", *label.Description)
		}
		if label.Parent != nil {
			fmt.Printf("Parent: %s (%s)\n", label.Parent.Name, label.Parent.ID)
		}
	},
}

var labelCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a label",
	Long:  `Create a new issue label in a team.`,
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
			os.Exit(1)
		}

		name, _ := cmd.Flags().GetString("name")
		teamKey, _ := cmd.Flags().GetString("team")
		colorValue, _ := cmd.Flags().GetString("color")
		description, _ := cmd.Flags().GetString("description")
		parentValue, _ := cmd.Flags().GetString("parent")
		isGroup, _ := cmd.Flags().GetBool("is-group")

		if strings.TrimSpace(name) == "" {
			output.Error("--name is required", plaintext, jsonOut)
			os.Exit(1)
		}
		if strings.TrimSpace(teamKey) == "" {
			output.Error("--team is required", plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)
		team, err := client.GetTeam(context.Background(), teamKey)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to find team '%s': %v", teamKey, err), plaintext, jsonOut)
			os.Exit(1)
		}

		input := map[string]interface{}{
			"name":   name,
			"teamId": team.ID,
		}

		if isGroup {
			input["isGroup"] = true
		}

		if colorValue != "" {
			input["color"] = colorValue
		}

		if cmd.Flags().Changed("description") {
			if strings.TrimSpace(description) == "" {
				input["description"] = nil
			} else {
				input["description"] = description
			}
		}

		if cmd.Flags().Changed("parent") {
			if isUnsetValue(parentValue) {
				input["parentId"] = nil
			} else {
				labels, err := client.GetTeamLabels(context.Background(), team.Key)
				if err != nil {
					output.Error(fmt.Sprintf("Failed to list team labels: %v", err), plaintext, jsonOut)
					os.Exit(1)
				}
				parent := findLabelByNameOrID(labels, parentValue)
				if parent == nil {
					output.Error(fmt.Sprintf("Parent label %q not found in team %s", parentValue, team.Key), plaintext, jsonOut)
					os.Exit(1)
				}
				input["parentId"] = parent.ID
			}
		}

		label, err := client.CreateLabel(context.Background(), input)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to create label: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(label)
		} else if plaintext {
			fmt.Printf("Created label %s (%s)\n", label.Name, label.ID)
		} else {
			fmt.Printf("%s Created label %s (%s)\n",
				color.New(color.FgGreen).Sprint("✓"),
				color.New(color.FgCyan, color.Bold).Sprint(label.Name),
				label.ID)
		}
	},
}

var labelUpdateCmd = &cobra.Command{
	Use:   "update [label-id]",
	Short: "Update a label",
	Long:  `Update an existing issue label.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		labelID := args[0]

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
			os.Exit(1)
		}

		name, _ := cmd.Flags().GetString("name")
		colorValue, _ := cmd.Flags().GetString("color")
		description, _ := cmd.Flags().GetString("description")
		parentValue, _ := cmd.Flags().GetString("parent")
		clearParent, _ := cmd.Flags().GetBool("clear-parent")

		if cmd.Flags().Changed("parent") && clearParent {
			output.Error("Cannot combine --parent with --clear-parent", plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)
		currentLabel, err := client.GetLabel(context.Background(), labelID)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to fetch label: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		input := make(map[string]interface{})
		if cmd.Flags().Changed("name") {
			if strings.TrimSpace(name) == "" {
				output.Error("--name cannot be empty", plaintext, jsonOut)
				os.Exit(1)
			}
			input["name"] = name
		}
		if cmd.Flags().Changed("color") {
			if strings.TrimSpace(colorValue) == "" {
				output.Error("--color cannot be empty", plaintext, jsonOut)
				os.Exit(1)
			}
			input["color"] = colorValue
		}
		if cmd.Flags().Changed("description") {
			if strings.TrimSpace(description) == "" {
				input["description"] = nil
			} else {
				input["description"] = description
			}
		}
		if clearParent {
			input["parentId"] = nil
		}
		if cmd.Flags().Changed("parent") {
			if isUnsetValue(parentValue) {
				input["parentId"] = nil
			} else {
				if currentLabel.Team == nil || currentLabel.Team.Key == "" {
					output.Error("Cannot resolve parent by name: label has no associated team key", plaintext, jsonOut)
					os.Exit(1)
				}
				labels, err := client.GetTeamLabels(context.Background(), currentLabel.Team.Key)
				if err != nil {
					output.Error(fmt.Sprintf("Failed to list team labels: %v", err), plaintext, jsonOut)
					os.Exit(1)
				}
				parent := findLabelByNameOrID(labels, parentValue)
				if parent == nil {
					output.Error(fmt.Sprintf("Parent label %q not found in team %s", parentValue, currentLabel.Team.Key), plaintext, jsonOut)
					os.Exit(1)
				}
				input["parentId"] = parent.ID
			}
		}

		if len(input) == 0 {
			output.Error("No updates specified. Use flags to specify what to update.", plaintext, jsonOut)
			os.Exit(1)
		}

		updated, err := client.UpdateLabel(context.Background(), labelID, input)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to update label: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(updated)
		} else if plaintext {
			fmt.Printf("Updated label %s (%s)\n", updated.Name, updated.ID)
		} else {
			fmt.Printf("%s Updated label %s (%s)\n",
				color.New(color.FgGreen).Sprint("✓"),
				color.New(color.FgCyan, color.Bold).Sprint(updated.Name),
				updated.ID)
		}
	},
}

var labelDeleteCmd = &cobra.Command{
	Use:     "delete [label-id]",
	Aliases: []string{"remove", "rm"},
	Short:   "Delete a label",
	Long:    `Delete an issue label by ID.`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		labelID := args[0]

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)
		if err := client.DeleteLabel(context.Background(), labelID); err != nil {
			output.Error(fmt.Sprintf("Failed to delete label: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(map[string]interface{}{
				"deleted": true,
				"id":      labelID,
			})
		} else if plaintext {
			fmt.Printf("Deleted label %s\n", labelID)
		} else {
			fmt.Printf("%s Deleted label %s\n",
				color.New(color.FgGreen).Sprint("✓"),
				color.New(color.FgCyan, color.Bold).Sprint(labelID))
		}
	},
}

func init() {
	rootCmd.AddCommand(labelCmd)
	labelCmd.AddCommand(labelListCmd)
	labelCmd.AddCommand(labelGetCmd)
	labelCmd.AddCommand(labelCreateCmd)
	labelCmd.AddCommand(labelUpdateCmd)
	labelCmd.AddCommand(labelDeleteCmd)

	labelListCmd.Flags().StringP("team", "t", "", "Team key (required)")

	labelGetCmd.Flags().StringP("team", "t", "", "Team key (required when using label name)")

	labelCreateCmd.Flags().StringP("team", "t", "", "Team key (required)")
	labelCreateCmd.Flags().String("name", "", "Label name (required)")
	labelCreateCmd.Flags().String("color", "", "Label color hex (e.g. #ff0000)")
	labelCreateCmd.Flags().String("description", "", "Label description")
	labelCreateCmd.Flags().String("parent", "", "Parent label name or ID")
	labelCreateCmd.Flags().Bool("is-group", false, "Create as a group label (for organizing child labels)")
	_ = labelCreateCmd.MarkFlagRequired("team")
	_ = labelCreateCmd.MarkFlagRequired("name")

	labelUpdateCmd.Flags().String("name", "", "New label name")
	labelUpdateCmd.Flags().String("color", "", "New label color hex")
	labelUpdateCmd.Flags().String("description", "", "New description (empty string clears description)")
	labelUpdateCmd.Flags().String("parent", "", "New parent label name or ID")
	labelUpdateCmd.Flags().Bool("clear-parent", false, "Remove parent label")
}
