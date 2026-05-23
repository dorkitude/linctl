package cmd

import (
	"fmt"
	"os"

	"github.com/dorkitude/linctl/pkg/auth"
	"github.com/dorkitude/linctl/pkg/output"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Linear",
	Long: `Authenticate with Linear using Personal API Key.

By default, linctl stores credentials in ~/.linctl-auth.json. Users of the
external pass password manager can opt in by setting LINCTL_PASS_NAME to the
desired pass entry name before running auth commands.

Examples:
  linctl auth              # Interactive authentication
  linctl auth login        # Same as above
  linctl auth status       # Check authentication status
  linctl auth logout       # Clear stored credentials
  LINCTL_PASS_NAME=linear-api-key linctl auth`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior is to run login
		loginCmd.Run(cmd, args)
	},
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Linear",
	Long: `Authenticate with Linear using Personal API Key.

By default, linctl stores credentials in ~/.linctl-auth.json. Users of the
external pass password manager can opt in by setting LINCTL_PASS_NAME to the
desired pass entry name.`,
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		if !plaintext && !jsonOut {
			fmt.Println(color.New(color.FgCyan, color.Bold).Sprint("🔐 Linear Authentication"))
			fmt.Println()
		}

		err := auth.Login(plaintext, jsonOut)
		if err != nil {
			output.Error(fmt.Sprintf("Authentication failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if !plaintext && !jsonOut {
			fmt.Println(color.New(color.FgGreen).Sprint("✅ Successfully authenticated with Linear!"))
		} else if jsonOut {
			output.JSON(map[string]interface{}{
				"status":  "success",
				"message": "Successfully authenticated with Linear",
			})
		} else {
			fmt.Println("Successfully authenticated with Linear")
		}
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Long:  `Check if you are currently authenticated with Linear.`,
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		user, err := auth.GetCurrentUser()
		if err != nil {
			if !plaintext && !jsonOut {
				fmt.Println(color.New(color.FgRed).Sprint("❌ Not authenticated"))
			} else if jsonOut {
				output.JSON(map[string]interface{}{
					"authenticated": false,
					"error":         err.Error(),
				})
			} else {
				fmt.Println("Not authenticated")
			}
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(map[string]interface{}{
				"authenticated": true,
				"user":          user,
			})
		} else if plaintext {
			fmt.Printf("Authenticated as: %s (%s)\n", user.Name, user.Email)
		} else {
			fmt.Println(color.New(color.FgGreen).Sprint("✅ Authenticated"))
			fmt.Printf("User: %s\n", color.New(color.FgCyan).Sprint(user.Name))
			fmt.Printf("Email: %s\n", color.New(color.FgCyan).Sprint(user.Email))
		}
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from Linear",
	Long: `Clear stored Linear credentials.

By default, this removes ~/.linctl-auth.json. When LINCTL_PASS_NAME is set,
linctl also removes that pass entry.`,
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		err := auth.Logout()
		if err != nil {
			output.Error(fmt.Sprintf("Logout failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		if jsonOut {
			output.JSON(map[string]interface{}{
				"status":  "success",
				"message": "Successfully logged out",
			})
		} else if plaintext {
			fmt.Println("Successfully logged out")
		} else {
			fmt.Println(color.New(color.FgGreen).Sprint("✅ Successfully logged out"))
		}
	},
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current user",
	Long:  `Display information about the currently authenticated user.`,
	Run: func(cmd *cobra.Command, args []string) {
		statusCmd.Run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(statusCmd)
	authCmd.AddCommand(logoutCmd)

	// Add whoami as a top-level command too
	rootCmd.AddCommand(whoamiCmd)
}
