package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	plaintext bool
	jsonOut   bool
)

// version is set at build time via -ldflags
// default value is for local dev builds
var version = "dev"

// generateHeader creates a nice header box with proper Unicode box drawing
func generateHeader() string {
	lines := []string{
		"🚀 linctl",
		"Linear CLI - Built with ❤️",
	}

	// Find the longest line
	maxLen := 0
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}

	// Add padding
	width := maxLen + 8

	// Build the box
	var result strings.Builder

	// Top border
	result.WriteString("┌")
	result.WriteString(strings.Repeat("─", width))
	result.WriteString("┐\n")

	// Content lines
	for _, line := range lines {
		padding := (width - len(line)) / 2
		result.WriteString("│")
		result.WriteString(strings.Repeat(" ", padding))
		result.WriteString(line)
		result.WriteString(strings.Repeat(" ", width-padding-len(line)))
		result.WriteString("│\n")
	}

	// Bottom border
	result.WriteString("└")
	result.WriteString(strings.Repeat("─", width))
	result.WriteString("┘")

	return result.String()
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "linctl",
	Short:   "A comprehensive Linear CLI tool",
	Long:    color.New(color.FgCyan).Sprintf("%s\nA comprehensive CLI tool for Linear's API featuring:\n• Issue management (create, list, update, archive)\n• Project tracking and collaboration  \n• Team and user management\n• Comments and attachments\n• Webhook configuration\n• Table/plaintext/JSON output formats\n", generateHeader()),
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// GetRootCmd returns the root command for testing
func GetRootCmd() *cobra.Command {
	return rootCmd
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.linctl.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&plaintext, "plaintext", "p", false, "plaintext output (non-interactive)")
	rootCmd.PersistentFlags().BoolVarP(&jsonOut, "json", "j", false, "JSON output")

	// Bind flags to viper
	_ = viper.BindPFlag("plaintext", rootCmd.PersistentFlags().Lookup("plaintext"))
	_ = viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".linctl" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".linctl")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if !plaintext && !jsonOut {
			fmt.Fprintln(os.Stderr, color.New(color.FgGreen).Sprintf("✅ Using config file: %s", viper.ConfigFileUsed()))
		}
	}
}
