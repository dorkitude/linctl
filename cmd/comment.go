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
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// commentCmd represents the comment command
var commentCmd = &cobra.Command{
	Use:   "comment",
	Short: "Manage issue comments",
	Long: `Manage comments on Linear issues with full CRUD support.

Examples:
  linctl comment list LIN-123        # List comments for an issue
  linctl comment create LIN-123 --body "This is fixed"  # Add a comment
  linctl comment get COMMENT-ID
  linctl comment update COMMENT-ID --body "Updated content"
  linctl comment delete COMMENT-ID`,
}

var commentListCmd = &cobra.Command{
	Use:     "list ISSUE-ID",
	Aliases: []string{"ls"},
	Short:   "List comments for an issue",
	Long:    `List all comments for a specific issue.`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		issueID := args[0]

		// Get auth header
		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Create API client
		client := api.NewClient(authHeader)

		// Get limit
		limit, _ := cmd.Flags().GetInt("limit")

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

		// Get comments
		comments, err := client.GetIssueComments(context.Background(), issueID, limit, "", orderBy)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to list comments: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Handle output
		if jsonOut {
			output.JSON(comments.Nodes)
		} else if plaintext {
			for i, comment := range comments.Nodes {
				if i > 0 {
					fmt.Println("---")
				}
				fmt.Printf("ID: %s\n", comment.ID)
				fmt.Printf("Author: %s\n", commentAuthorName(&comment))
				fmt.Printf("Date: %s\n", comment.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("Comment:\n%s\n", comment.Body)
			}
		} else {
			// Rich display
			if len(comments.Nodes) == 0 {
				fmt.Printf("\n%s No comments on issue %s\n",
					color.New(color.FgYellow).Sprint("ℹ️"),
					color.New(color.FgCyan).Sprint(issueID))
				return
			}

			fmt.Printf("\n%s Comments on %s (%d)\n\n",
				color.New(color.FgCyan, color.Bold).Sprint("💬"),
				color.New(color.FgCyan).Sprint(issueID),
				len(comments.Nodes))

			for i, comment := range comments.Nodes {
				if i > 0 {
					fmt.Println(strings.Repeat("─", 50))
				}

				// Header with author and time
				timeAgo := formatTimeAgo(comment.CreatedAt)
				fmt.Printf("%s %s %s %s %s\n",
					color.New(color.FgCyan, color.Bold).Sprint(commentAuthorName(&comment)),
					color.New(color.FgWhite, color.Faint).Sprint("•"),
					color.New(color.FgWhite, color.Faint).Sprint(timeAgo),
					color.New(color.FgWhite, color.Faint).Sprint("•"),
					color.New(color.FgWhite, color.Faint).Sprintf("id=%s", comment.ID))

				// Comment body
				fmt.Printf("\n%s\n\n", comment.Body)
			}
		}
	},
}

var commentCreateCmd = &cobra.Command{
	Use:     "create ISSUE-ID",
	Aliases: []string{"add", "new"},
	Short:   "Create a comment on an issue",
	Long:    `Add a new comment to a specific issue.`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		issueID := args[0]

		// Get auth header
		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Create API client
		client := api.NewClient(authHeader)

		// Get comment body
		body, _ := cmd.Flags().GetString("body")
		if body == "" {
			output.Error("Comment body is required (--body)", plaintext, jsonOut)
			os.Exit(1)
		}

		// Create comment
		comment, err := client.CreateComment(context.Background(), issueID, body)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to create comment: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Handle output
		if jsonOut {
			output.JSON(comment)
		} else if plaintext {
			fmt.Printf("Created comment on %s\n", issueID)
			fmt.Printf("Author: %s\n", commentAuthorName(comment))
			fmt.Printf("Date: %s\n", comment.CreatedAt.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("%s Added comment to %s\n",
				color.New(color.FgGreen).Sprint("✓"),
				color.New(color.FgCyan, color.Bold).Sprint(issueID))
			fmt.Printf("\n%s\n", comment.Body)
		}
	},
}

var commentGetCmd = &cobra.Command{
	Use:     "get COMMENT-ID",
	Aliases: []string{"show"},
	Short:   "Get a comment by ID",
	Long:    `Get details for a specific comment.`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		commentID := args[0]

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)
		comment, err := client.GetComment(context.Background(), commentID)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to get comment: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(comment)
		} else if plaintext {
			fmt.Printf("ID: %s\n", comment.ID)
			fmt.Printf("Author: %s\n", commentAuthorName(comment))
			fmt.Printf("Created: %s\n", comment.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("Updated: %s\n", comment.UpdatedAt.Format("2006-01-02 15:04:05"))
			if comment.EditedAt != nil {
				fmt.Printf("Edited: %s\n", comment.EditedAt.Format("2006-01-02 15:04:05"))
			}
			fmt.Printf("Comment:\n%s\n", comment.Body)
		} else {
			fmt.Printf("%s %s\n",
				color.New(color.FgCyan, color.Bold).Sprint("Comment"),
				color.New(color.FgWhite, color.Faint).Sprintf("(%s)", comment.ID))
			fmt.Printf("%s %s\n",
				color.New(color.FgCyan).Sprint(commentAuthorName(comment)),
				color.New(color.FgWhite, color.Faint).Sprint(formatTimeAgo(comment.CreatedAt)))
			fmt.Printf("\n%s\n", comment.Body)
		}
	},
}

var commentUpdateCmd = &cobra.Command{
	Use:     "update COMMENT-ID",
	Aliases: []string{"edit"},
	Short:   "Update a comment by ID",
	Long:    `Update comment body for an existing comment.`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		commentID := args[0]

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		body, _ := cmd.Flags().GetString("body")
		if body == "" {
			output.Error("Comment body is required (--body)", plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)
		comment, err := client.UpdateComment(context.Background(), commentID, body)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to update comment: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(comment)
		} else if plaintext {
			fmt.Printf("Updated comment %s\n", comment.ID)
		} else {
			fmt.Printf("%s Updated comment %s\n",
				color.New(color.FgGreen).Sprint("✓"),
				color.New(color.FgCyan, color.Bold).Sprint(comment.ID))
		}
	},
}

var commentDeleteCmd = &cobra.Command{
	Use:     "delete COMMENT-ID",
	Aliases: []string{"remove", "rm"},
	Short:   "Delete a comment by ID",
	Long:    `Delete an existing comment.`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")
		commentID := args[0]

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)
		err = client.DeleteComment(context.Background(), commentID)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to delete comment: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(map[string]interface{}{
				"deleted":   true,
				"commentId": commentID,
			})
		} else if plaintext {
			fmt.Printf("Deleted comment %s\n", commentID)
		} else {
			fmt.Printf("%s Deleted comment %s\n",
				color.New(color.FgGreen).Sprint("✓"),
				color.New(color.FgCyan, color.Bold).Sprint(commentID))
		}
	},
}

// formatTimeAgo formats a time as a human-readable "time ago" string
func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if duration < 30*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else if duration < 365*24*time.Hour {
		months := int(duration.Hours() / (24 * 30))
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	} else {
		years := int(duration.Hours() / (24 * 365))
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

func init() {
	rootCmd.AddCommand(commentCmd)
	commentCmd.AddCommand(commentListCmd)
	commentCmd.AddCommand(commentCreateCmd)
	commentCmd.AddCommand(commentGetCmd)
	commentCmd.AddCommand(commentUpdateCmd)
	commentCmd.AddCommand(commentDeleteCmd)

	// List command flags
	commentListCmd.Flags().IntP("limit", "l", 50, "Maximum number of comments to return")
	commentListCmd.Flags().StringP("sort", "o", "linear", "Sort order: linear (default), created, updated")

	// Create command flags
	commentCreateCmd.Flags().StringP("body", "b", "", "Comment body (required)")
	_ = commentCreateCmd.MarkFlagRequired("body")

	// Update command flags
	commentUpdateCmd.Flags().StringP("body", "b", "", "Updated comment body (required)")
	_ = commentUpdateCmd.MarkFlagRequired("body")
}
