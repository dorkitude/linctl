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

var agentCmd = &cobra.Command{
	Use:   "agent [issue-id]",
	Short: "View agent session for an issue",
	Long: `View agent delegation/session state for an issue.

Examples:
  linctl agent ENG-80
  linctl agent ENG-80 --json`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)
		issue, err := client.GetIssueAgentSession(context.Background(), args[0])
		if err != nil {
			output.Error(fmt.Sprintf("Failed to fetch issue session: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		session := latestAgentSession(issue)
		if session == nil && issue.Delegate == nil {
			output.Info(fmt.Sprintf("No agent delegation/session found for %s", issue.Identifier), plaintext, jsonOut)
			return
		}

		if jsonOut {
			result := map[string]interface{}{
				"issue": issue.Identifier,
				"title": issue.Title,
			}
			if issue.Delegate != nil {
				result["delegate"] = issue.Delegate
			}
			if session != nil {
				result["agentSession"] = session
			}
			output.JSON(result)
			return
		}

		if plaintext {
			fmt.Printf("# Agent Session for %s\n\n", issue.Identifier)
			fmt.Printf("- **Title**: %s\n", issue.Title)
			if issue.Delegate != nil {
				fmt.Printf("- **Delegate**: %s\n", userDisplay(issue.Delegate))
			}
			if session == nil {
				fmt.Println("- **Status**: delegated (session not started)")
				return
			}
			fmt.Printf("- **Status**: %s\n", session.Status)
			if session.AppUser != nil {
				fmt.Printf("- **Agent**: %s\n", userDisplay(session.AppUser))
			}
			fmt.Printf("- **Started**: %s\n", session.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("- **Updated**: %s\n", session.UpdatedAt.Format("2006-01-02 15:04:05"))
			if session.Activities != nil && len(session.Activities.Nodes) > 0 {
				fmt.Println("\n## Activity Stream")
				for _, activity := range session.Activities.Nodes {
					activityType, body := activitySummary(activity)
					fmt.Printf("- [%s] %s", activity.CreatedAt.Format("15:04:05"), activityType)
					if body != "" {
						fmt.Printf(": %s", body)
					}
					fmt.Println()
				}
			}
			return
		}

		fmt.Printf("%s %s\n",
			color.New(color.FgCyan, color.Bold).Sprint(issue.Identifier),
			color.New(color.FgWhite, color.Bold).Sprint(issue.Title))
		if issue.Delegate != nil {
			fmt.Printf("%s %s\n",
				color.New(color.FgYellow).Sprint("Delegate:"),
				color.New(color.FgCyan).Sprint(userDisplay(issue.Delegate)))
		}

		if session == nil {
			fmt.Printf("%s\n", color.New(color.FgWhite, color.Faint).Sprint("Delegated but session not started yet"))
			return
		}

		statusColor := color.New(color.FgWhite)
		switch session.Status {
		case "active":
			statusColor = color.New(color.FgGreen)
		case "complete":
			statusColor = color.New(color.FgBlue)
		case "awaitingInput":
			statusColor = color.New(color.FgYellow)
		case "error":
			statusColor = color.New(color.FgRed)
		case "pending":
			statusColor = color.New(color.FgMagenta)
		}
		fmt.Printf("%s %s\n", color.New(color.FgYellow).Sprint("Status:"), statusColor.Sprint(session.Status))
		if session.AppUser != nil {
			fmt.Printf("%s %s\n", color.New(color.FgYellow).Sprint("Agent:"), color.New(color.FgCyan).Sprint(userDisplay(session.AppUser)))
		}
		fmt.Printf("%s %s\n", color.New(color.FgYellow).Sprint("Started:"), session.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("%s %s\n", color.New(color.FgYellow).Sprint("Updated:"), session.UpdatedAt.Format("2006-01-02 15:04:05"))

		if session.Activities != nil && len(session.Activities.Nodes) > 0 {
			fmt.Printf("\n%s\n", color.New(color.FgYellow, color.Bold).Sprint("Activity Stream:"))
			for _, activity := range session.Activities.Nodes {
				activityType, body := activitySummary(activity)
				fmt.Printf("  %s [%s]\n",
					color.New(color.FgWhite, color.Faint).Sprint(activity.CreatedAt.Format("15:04:05")),
					activityType)
				if body != "" {
					for _, line := range strings.Split(body, "\n") {
						fmt.Printf("    %s\n", line)
					}
				}
			}
		}
	},
}

var agentMentionCmd = &cobra.Command{
	Use:   "mention [issue-id] [message...]",
	Short: "Send a message to the issue's delegated agent",
	Long: `Create a comment that @mentions the delegated/active agent with a message.

Examples:
  linctl agent mention ENG-80 "Please pick this up"
  linctl agent mention ENG-80 --agent my-agent "Can you rerun tests?"`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
			os.Exit(1)
		}

		issueID := args[0]
		message := strings.TrimSpace(strings.Join(args[1:], " "))
		if message == "" {
			output.Error("Message cannot be empty", plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)
		issue, err := client.GetIssueAgentSession(context.Background(), issueID)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to fetch issue session: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		agentHandle, _ := cmd.Flags().GetString("agent")
		if strings.TrimSpace(agentHandle) == "" {
			agentHandle = pickAgentHandle(issue)
		}
		if strings.TrimSpace(agentHandle) == "" {
			output.Error("No delegated/active agent found. Pass --agent to specify one explicitly.", plaintext, jsonOut)
			os.Exit(1)
		}

		commentID, err := client.MentionAgent(context.Background(), issue.ID, agentHandle, message)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to mention agent: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(map[string]interface{}{
				"success":   true,
				"commentId": commentID,
				"issue":     issue.Identifier,
				"agent":     agentHandle,
				"message":   message,
			})
			return
		}

		if plaintext {
			fmt.Printf("Mentioned @%s on %s (comment: %s)\n", agentHandle, issue.Identifier, commentID)
			return
		}

		fmt.Printf("%s Mentioned @%s on %s\n",
			color.New(color.FgGreen).Sprint("✓"),
			color.New(color.FgCyan).Sprint(agentHandle),
			color.New(color.FgCyan, color.Bold).Sprint(issue.Identifier))
	},
}

func latestAgentSession(issue *api.Issue) *api.AgentSession {
	if issue == nil || issue.Comments == nil {
		return nil
	}

	var selected *api.AgentSession
	for _, comment := range issue.Comments.Nodes {
		if comment.AgentSession == nil {
			continue
		}
		if selected == nil || comment.AgentSession.UpdatedAt.After(selected.UpdatedAt) {
			selected = comment.AgentSession
		}
	}
	return selected
}

func pickAgentHandle(issue *api.Issue) string {
	if issue == nil {
		return ""
	}
	if issue.Delegate != nil {
		if issue.Delegate.DisplayName != "" {
			return issue.Delegate.DisplayName
		}
		if issue.Delegate.Name != "" {
			return issue.Delegate.Name
		}
	}

	session := latestAgentSession(issue)
	if session != nil && session.AppUser != nil {
		if session.AppUser.DisplayName != "" {
			return session.AppUser.DisplayName
		}
		return session.AppUser.Name
	}

	return ""
}

func activitySummary(activity api.AgentActivity) (string, string) {
	activityType := "unknown"
	body := ""

	if value, ok := activity.Content["type"].(string); ok && value != "" {
		activityType = value
	}
	if value, ok := activity.Content["body"].(string); ok {
		body = value
	} else if action, ok := activity.Content["action"].(string); ok {
		parameter, _ := activity.Content["parameter"].(string)
		body = strings.TrimSpace(action + " " + parameter)
	}

	return activityType, body
}

func userDisplay(user *api.User) string {
	if user == nil {
		return ""
	}
	if strings.TrimSpace(user.DisplayName) != "" {
		if strings.TrimSpace(user.Name) != "" && !strings.EqualFold(user.DisplayName, user.Name) {
			return fmt.Sprintf("%s (%s)", user.Name, user.DisplayName)
		}
		return user.DisplayName
	}
	if strings.TrimSpace(user.Name) != "" {
		return user.Name
	}
	return user.Email
}

func init() {
	rootCmd.AddCommand(agentCmd)
	agentCmd.AddCommand(agentMentionCmd)
	agentMentionCmd.Flags().String("agent", "", "Agent handle to mention (defaults to delegated/active agent)")
}
