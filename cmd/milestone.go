package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dorkitude/linctl/pkg/api"
	"github.com/dorkitude/linctl/pkg/auth"
	"github.com/dorkitude/linctl/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// milestoneAPI defines the interface for milestone operations
type milestoneAPI interface {
	ListProjectMilestones(ctx context.Context, projectID string, includeArchived bool) (*api.ProjectMilestones, error)
	GetProjectMilestone(ctx context.Context, milestoneID string) (*api.ProjectMilestone, error)
	CreateProjectMilestone(ctx context.Context, input map[string]interface{}) (*api.ProjectMilestone, error)
	UpdateProjectMilestone(ctx context.Context, milestoneID string, input map[string]interface{}) (*api.ProjectMilestone, error)
	DeleteProjectMilestone(ctx context.Context, milestoneID string) error
}

// Injection points for testing
var newMilestoneAPIClient = func(authHeader string) milestoneAPI { return api.NewClient(authHeader) }
var getMilestoneAuthHeader = auth.GetAuthHeader

var milestoneCmd = &cobra.Command{
	Use:   "milestone",
	Short: "Manage project milestones",
	Long:  `Create, list, update, and delete project milestones.`,
}

var milestoneListCmd = &cobra.Command{
	Use:   "list <project-id>",
	Short: "List milestones for a project",
	Long:  `List all milestones for a specific project.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		authHeader, err := getMilestoneAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		client := newMilestoneAPIClient(authHeader)
		runMilestoneList(cmd, client, args[0], plaintext, jsonOut)
	},
}

var milestoneGetCmd = &cobra.Command{
	Use:   "get <milestone-id>",
	Short: "Get a specific milestone",
	Long:  `Get details of a specific project milestone.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		authHeader, err := getMilestoneAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		client := newMilestoneAPIClient(authHeader)
		runMilestoneGet(cmd, client, args[0], plaintext, jsonOut)
	},
}

var milestoneCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new milestone",
	Long:  `Create a new milestone in a project.`,
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		authHeader, err := getMilestoneAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		client := newMilestoneAPIClient(authHeader)
		runMilestoneCreate(cmd, client, plaintext, jsonOut)
	},
}

var milestoneUpdateCmd = &cobra.Command{
	Use:   "update <milestone-id>",
	Short: "Update a milestone",
	Long:  `Update an existing project milestone.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		authHeader, err := getMilestoneAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		client := newMilestoneAPIClient(authHeader)
		runMilestoneUpdate(cmd, client, args[0], plaintext, jsonOut)
	},
}

var milestoneDeleteCmd = &cobra.Command{
	Use:   "delete <milestone-id>",
	Short: "Delete a milestone",
	Long:  `Delete (archive) a project milestone.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		authHeader, err := getMilestoneAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		client := newMilestoneAPIClient(authHeader)
		runMilestoneDelete(cmd, client, args[0], plaintext, jsonOut)
	},
}

func init() {
	rootCmd.AddCommand(milestoneCmd)
	milestoneCmd.AddCommand(milestoneListCmd)
	milestoneCmd.AddCommand(milestoneGetCmd)
	milestoneCmd.AddCommand(milestoneCreateCmd)
	milestoneCmd.AddCommand(milestoneUpdateCmd)
	milestoneCmd.AddCommand(milestoneDeleteCmd)

	// List flags
	milestoneListCmd.Flags().Bool("include-archived", false, "Include archived milestones")

	// Create flags
	milestoneCreateCmd.Flags().String("project", "", "Project ID (required)")
	milestoneCreateCmd.Flags().String("name", "", "Milestone name (required)")
	milestoneCreateCmd.Flags().String("description", "", "Milestone description")
	milestoneCreateCmd.Flags().String("target-date", "", "Target date (YYYY-MM-DD)")
	milestoneCreateCmd.MarkFlagRequired("project")
	milestoneCreateCmd.MarkFlagRequired("name")

	// Update flags
	milestoneUpdateCmd.Flags().String("name", "", "Milestone name")
	milestoneUpdateCmd.Flags().String("description", "", "Milestone description")
	milestoneUpdateCmd.Flags().String("target-date", "", "Target date (YYYY-MM-DD)")
}

func runMilestoneList(cmd *cobra.Command, client milestoneAPI, projectID string, plaintext, jsonOut bool) {
	includeArchived, _ := cmd.Flags().GetBool("include-archived")

	milestones, err := client.ListProjectMilestones(context.Background(), projectID, includeArchived)
	if err != nil {
		output.Error(fmt.Sprintf("Failed to list milestones: %v", err), plaintext, jsonOut)
		os.Exit(1)
	}

	if len(milestones.Nodes) == 0 {
		if jsonOut {
			output.JSON([]interface{}{})
		} else {
			output.Info("No milestones found", plaintext, jsonOut)
		}
		return
	}

	if jsonOut {
		output.JSON(milestones.Nodes)
		return
	}

	// Table output
	headers := []string{"ID", "Name", "Target Date", "Status", "Progress", "Created"}
	rows := [][]string{}

	for _, milestone := range milestones.Nodes {
		targetDate := "Not set"
		if milestone.TargetDate != nil {
			targetDate = *milestone.TargetDate
		}

		progress := fmt.Sprintf("%.0f%%", milestone.Progress*100)
		created := milestone.CreatedAt.Format("2006-01-02")

		rows = append(rows, []string{
			milestone.ID,
			milestone.Name,
			targetDate,
			milestone.Status,
			progress,
			created,
		})
	}

	output.Table(output.TableData{Headers: headers, Rows: rows}, plaintext, jsonOut)
}

func runMilestoneGet(cmd *cobra.Command, client milestoneAPI, milestoneID string, plaintext, jsonOut bool) {
	milestone, err := client.GetProjectMilestone(context.Background(), milestoneID)
	if err != nil {
		output.Error(fmt.Sprintf("Failed to get milestone: %v", err), plaintext, jsonOut)
		os.Exit(1)
	}

	if jsonOut {
		output.JSON(milestone)
		return
	}

	// Format output
	output.Info(fmt.Sprintf("Milestone: %s", milestone.Name), plaintext, jsonOut)
	output.Info(fmt.Sprintf("ID: %s", milestone.ID), plaintext, jsonOut)
	if milestone.Project != nil {
		output.Info(fmt.Sprintf("Project: %s (%s)", milestone.Project.Name, milestone.Project.ID), plaintext, jsonOut)
	}
	output.Info(fmt.Sprintf("Description: %s", milestone.Description), plaintext, jsonOut)
	output.Info(fmt.Sprintf("Status: %s", milestone.Status), plaintext, jsonOut)
	output.Info(fmt.Sprintf("Progress: %.0f%%", milestone.Progress*100), plaintext, jsonOut)

	if milestone.TargetDate != nil {
		output.Info(fmt.Sprintf("Target Date: %s", *milestone.TargetDate), plaintext, jsonOut)
	}

	output.Info(fmt.Sprintf("Created: %s", milestone.CreatedAt.Format("2006-01-02 15:04:05")), plaintext, jsonOut)
	output.Info(fmt.Sprintf("Updated: %s", milestone.UpdatedAt.Format("2006-01-02 15:04:05")), plaintext, jsonOut)

	if milestone.ArchivedAt != nil {
		output.Info(fmt.Sprintf("Archived: %s", milestone.ArchivedAt.Format("2006-01-02 15:04:05")), plaintext, jsonOut)
	}
}

func runMilestoneCreate(cmd *cobra.Command, client milestoneAPI, plaintext, jsonOut bool) {
	projectID, _ := cmd.Flags().GetString("project")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	targetDate, _ := cmd.Flags().GetString("target-date")

	// Validate target date format if provided
	if targetDate != "" {
		if _, err := time.Parse("2006-01-02", targetDate); err != nil {
			output.Error("Invalid --target-date format. Expected YYYY-MM-DD", plaintext, jsonOut)
			os.Exit(1)
		}
	}

	// Build input map
	input := map[string]interface{}{
		"projectId": projectID,
		"name":      name,
	}

	if description != "" {
		input["description"] = description
	}
	if targetDate != "" {
		input["targetDate"] = targetDate
	}

	// Create milestone
	milestone, err := client.CreateProjectMilestone(context.Background(), input)
	if err != nil {
		output.Error(fmt.Sprintf("Failed to create milestone: %v", err), plaintext, jsonOut)
		os.Exit(1)
	}

	if jsonOut {
		output.JSON(milestone)
		return
	}

	output.Success(fmt.Sprintf("Created milestone: %s (ID: %s)", milestone.Name, milestone.ID), plaintext, jsonOut)
	if milestone.Project != nil {
		output.Info(fmt.Sprintf("Project: %s", milestone.Project.Name), plaintext, jsonOut)
	}
	if milestone.TargetDate != nil {
		output.Info(fmt.Sprintf("Target Date: %s", *milestone.TargetDate), plaintext, jsonOut)
	}
}

func runMilestoneUpdate(cmd *cobra.Command, client milestoneAPI, milestoneID string, plaintext, jsonOut bool) {
	input := make(map[string]interface{})

	// Only add changed fields to input
	if cmd.Flags().Changed("name") {
		name, _ := cmd.Flags().GetString("name")
		input["name"] = name
	}
	if cmd.Flags().Changed("description") {
		description, _ := cmd.Flags().GetString("description")
		input["description"] = description
	}
	if cmd.Flags().Changed("target-date") {
		targetDate, _ := cmd.Flags().GetString("target-date")
		if targetDate != "" {
			if _, err := time.Parse("2006-01-02", targetDate); err != nil {
				output.Error("Invalid --target-date format. Expected YYYY-MM-DD", plaintext, jsonOut)
				os.Exit(1)
			}
		}
		input["targetDate"] = targetDate
	}

	// Validate at least one field provided
	if len(input) == 0 {
		output.Error("At least one field to update is required", plaintext, jsonOut)
		os.Exit(1)
	}

	// Update milestone
	milestone, err := client.UpdateProjectMilestone(context.Background(), milestoneID, input)
	if err != nil {
		output.Error(fmt.Sprintf("Failed to update milestone: %v", err), plaintext, jsonOut)
		os.Exit(1)
	}

	if jsonOut {
		output.JSON(milestone)
		return
	}

	output.Success(fmt.Sprintf("Updated milestone: %s", milestone.Name), plaintext, jsonOut)
	output.Info(fmt.Sprintf("ID: %s", milestone.ID), plaintext, jsonOut)
}

func runMilestoneDelete(cmd *cobra.Command, client milestoneAPI, milestoneID string, plaintext, jsonOut bool) {
	err := client.DeleteProjectMilestone(context.Background(), milestoneID)
	if err != nil {
		output.Error(fmt.Sprintf("Failed to delete milestone: %v", err), plaintext, jsonOut)
		os.Exit(1)
	}

	if jsonOut {
		output.JSON(map[string]interface{}{"success": true, "milestoneId": milestoneID})
		return
	}

	output.Success(fmt.Sprintf("Deleted milestone: %s", milestoneID), plaintext, jsonOut)
}
