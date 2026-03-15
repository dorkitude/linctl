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

// validRelationTypes lists the valid relation types accepted by the Linear API.
var validRelationTypes = []string{"blocks", "blocked-by", "duplicate", "related"}

// issueRelationCmd is the parent command for issue relation management.
var issueRelationCmd = &cobra.Command{
	Use:   "relation",
	Short: "Manage issue relations",
	Long: `Manage relations between Linear issues (blocking, blocked-by, related, duplicate).

Examples:
  linctl issue relation list LIN-123
  linctl issue relation add LIN-123 --blocks LIN-456
  linctl issue relation add LIN-123 --blocked-by LIN-456
  linctl issue relation add LIN-123 --related LIN-456
  linctl issue relation add LIN-123 --duplicate LIN-456
  linctl issue relation remove RELATION-ID`,
}

var issueRelationListCmd = &cobra.Command{
	Use:     "list ISSUE-ID",
	Aliases: []string{"ls"},
	Short:   "List relations for an issue",
	Long: `List all relations (blocking, blocked-by, related, duplicate) for an issue.

Examples:
  linctl issue relation list LIN-123
  linctl issue relation ls LIN-123 -j`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		issueID := args[0]

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)

		relations, err := client.GetIssueRelations(context.Background(), issueID)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to fetch relations: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if len(relations) == 0 {
			output.Info(fmt.Sprintf("No relations found for %s", issueID), plaintext, jsonOut)
			return
		}

		if jsonOut {
			output.JSON(relations)
			return
		}

		if plaintext {
			for _, rel := range relations {
				other := relationOtherIssue(&rel)
				fmt.Printf("%s\t%s\t%s\t%s\n", rel.ID, rel.Type, other.Identifier, other.Title)
			}
			return
		}

		// Rich display
		fmt.Printf("\n%s Relations for %s (%d)\n\n",
			color.New(color.FgCyan, color.Bold).Sprint("🔗"),
			color.New(color.FgCyan).Sprint(issueID),
			len(relations))

		for _, rel := range relations {
			other := relationOtherIssue(&rel)
			typeLabel := relationTypeLabel(rel.Type)
			fmt.Printf("  %s %s %s\n",
				color.New(color.FgYellow).Sprint(typeLabel),
				color.New(color.FgCyan, color.Bold).Sprint(other.Identifier),
				other.Title)
			fmt.Printf("    %s\n\n",
				color.New(color.FgWhite, color.Faint).Sprintf("relation-id: %s", rel.ID))
		}
	},
}

var issueRelationAddCmd = &cobra.Command{
	Use:     "add ISSUE-ID",
	Aliases: []string{"create", "new"},
	Short:   "Add a relation to an issue",
	Long: `Create a relation between two issues.

Relation types:
  --blocks ISSUE-ID      This issue blocks the specified issue
  --blocked-by ISSUE-ID  This issue is blocked by the specified issue
  --related ISSUE-ID     Mark issues as related
  --duplicate ISSUE-ID   Mark this issue as a duplicate of the specified issue

Examples:
  linctl issue relation add LIN-123 --blocks LIN-456
  linctl issue relation add LIN-123 --blocked-by LIN-456
  linctl issue relation add LIN-123 --related LIN-456
  linctl issue relation add LIN-123 --duplicate LIN-456`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		issueID := args[0]

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
			os.Exit(1)
		}

		// Determine which relation type was specified
		blocks, _ := cmd.Flags().GetString("blocks")
		blockedBy, _ := cmd.Flags().GetString("blocked-by")
		related, _ := cmd.Flags().GetString("related")
		duplicate, _ := cmd.Flags().GetString("duplicate")

		var relatedIssueID string
		var relationType string

		// Count how many flags were set
		flagCount := 0
		if blocks != "" {
			flagCount++
			relatedIssueID = blocks
			relationType = "blocks"
		}
		if blockedBy != "" {
			flagCount++
			relatedIssueID = blockedBy
			relationType = "blocked-by"
		}
		if related != "" {
			flagCount++
			relatedIssueID = related
			relationType = "related"
		}
		if duplicate != "" {
			flagCount++
			relatedIssueID = duplicate
			relationType = "duplicate"
		}

		if flagCount == 0 {
			output.Error("Must specify one of: --blocks, --blocked-by, --related, --duplicate", plaintext, jsonOut)
			os.Exit(1)
		}
		if flagCount > 1 {
			output.Error("Only one relation type can be specified at a time", plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)

		// Resolve both issue IDs to UUIDs
		issue, err := client.GetIssue(context.Background(), issueID)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to fetch issue %s: %v", issueID, err), plaintext, jsonOut)
			os.Exit(1)
		}
		relatedIssue, err := client.GetIssue(context.Background(), relatedIssueID)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to fetch issue %s: %v", relatedIssueID, err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Map the CLI relation type to the Linear API type.
		// Linear's issueRelationCreate uses:
		//   issueId = the issue that has the relation
		//   relatedIssueId = the other issue
		//   type = "blocks" means issueId is blocked by relatedIssueId
		//
		// CLI semantics:
		//   --blocks TARGET     => "this issue blocks TARGET"
		//                       => TARGET is blocked by THIS
		//                       => API: issueId=TARGET, relatedIssueId=THIS, type=blocks
		//   --blocked-by SOURCE => "this issue is blocked by SOURCE"
		//                       => API: issueId=THIS, relatedIssueId=SOURCE, type=blocks
		//   --related TARGET    => API: issueId=THIS, relatedIssueId=TARGET, type=related
		//   --duplicate TARGET  => API: issueId=THIS, relatedIssueId=TARGET, type=duplicate

		var apiIssueID, apiRelatedIssueID, apiType string

		switch relationType {
		case "blocks":
			// "LIN-123 blocks LIN-456" => LIN-456 is blocked by LIN-123
			apiIssueID = relatedIssue.ID
			apiRelatedIssueID = issue.ID
			apiType = "blocks"
		case "blocked-by":
			// "LIN-123 is blocked by LIN-456" => LIN-123 is blocked by LIN-456
			apiIssueID = issue.ID
			apiRelatedIssueID = relatedIssue.ID
			apiType = "blocks"
		case "related":
			apiIssueID = issue.ID
			apiRelatedIssueID = relatedIssue.ID
			apiType = "related"
		case "duplicate":
			apiIssueID = issue.ID
			apiRelatedIssueID = relatedIssue.ID
			apiType = "duplicate"
		}

		relation, err := client.CreateIssueRelation(context.Background(), apiIssueID, apiRelatedIssueID, apiType)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to create relation: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(relation)
			return
		}

		if plaintext {
			fmt.Printf("Created %s relation between %s and %s\n", relationType, issueID, relatedIssueID)
			fmt.Printf("Relation ID: %s\n", relation.ID)
			return
		}

		// Rich display
		fmt.Printf("%s Created relation: %s %s %s\n",
			color.New(color.FgGreen).Sprint("✓"),
			color.New(color.FgCyan, color.Bold).Sprint(issueID),
			color.New(color.FgYellow).Sprint(relationType),
			color.New(color.FgCyan, color.Bold).Sprint(relatedIssueID))
	},
}

var issueRelationRemoveCmd = &cobra.Command{
	Use:     "remove RELATION-ID",
	Aliases: []string{"delete", "rm"},
	Short:   "Remove a relation by its ID",
	Long: `Remove a relation between two issues.

Use 'linctl issue relation list ISSUE-ID' to find relation IDs.

Examples:
  linctl issue relation remove abc123-def456
  linctl issue relation rm abc123-def456`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		relationID := args[0]

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)

		err = client.DeleteIssueRelation(context.Background(), relationID)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to remove relation: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(map[string]interface{}{
				"deleted":    true,
				"relationId": relationID,
			})
			return
		}

		if plaintext {
			fmt.Printf("Removed relation %s\n", relationID)
			return
		}

		fmt.Printf("%s Removed relation %s\n",
			color.New(color.FgGreen).Sprint("✓"),
			color.New(color.FgCyan, color.Bold).Sprint(relationID))
	},
}

// relationOtherIssue returns the "other" issue in a relation — either Issue or
// RelatedIssue, whichever is populated.
func relationOtherIssue(rel *api.IssueRelation) *api.Issue {
	if rel.RelatedIssue != nil {
		return rel.RelatedIssue
	}
	if rel.Issue != nil {
		return rel.Issue
	}
	return &api.Issue{Identifier: "?", Title: "unknown"}
}

// relationTypeLabel returns a human-readable label for a relation type.
func relationTypeLabel(t string) string {
	switch strings.ToLower(t) {
	case "blocks":
		return "blocked by"
	case "duplicate":
		return "duplicate of"
	case "related":
		return "related to"
	default:
		return t
	}
}

func init() {
	issueCmd.AddCommand(issueRelationCmd)
	issueRelationCmd.AddCommand(issueRelationListCmd)
	issueRelationCmd.AddCommand(issueRelationAddCmd)
	issueRelationCmd.AddCommand(issueRelationRemoveCmd)

	// Add command flags
	issueRelationAddCmd.Flags().String("blocks", "", "Issue that this issue blocks (issue identifier)")
	issueRelationAddCmd.Flags().String("blocked-by", "", "Issue that blocks this issue (issue identifier)")
	issueRelationAddCmd.Flags().String("related", "", "Related issue (issue identifier)")
	issueRelationAddCmd.Flags().String("duplicate", "", "Issue that this is a duplicate of (issue identifier)")
}
