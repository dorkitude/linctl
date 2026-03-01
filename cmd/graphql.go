package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dorkitude/linctl/pkg/api"
	"github.com/dorkitude/linctl/pkg/auth"
	"github.com/dorkitude/linctl/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var graphqlCmd = &cobra.Command{
	Use:   "graphql [query]",
	Short: "Execute a raw GraphQL operation against Linear",
	Long: `Execute a raw GraphQL query or mutation against Linear using your existing linctl auth.

Use this when the Linear API supports features that aren't exposed by first-class linctl commands yet.

Examples:
  linctl graphql 'query { viewer { id name email } }'
  linctl graphql --file query.graphql
  linctl graphql --query 'query($k:String!){ team(id:$k){ id key name } }' --variables '{"k":"ENG"}'
  linctl graphql --file query.graphql --variables-file vars.json`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
			os.Exit(1)
		}

		query, err := resolveGraphQLQueryInput(cmd, args)
		if err != nil {
			output.Error(err.Error(), plaintext, jsonOut)
			os.Exit(1)
		}

		variables, err := resolveGraphQLVariablesInput(cmd)
		if err != nil {
			output.Error(err.Error(), plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)
		resp, err := client.ExecuteRaw(context.Background(), query, variables)
		if err != nil {
			output.Error(fmt.Sprintf("GraphQL request failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		// Raw GraphQL is always most useful as JSON, regardless of table/plaintext mode.
		output.JSON(resp)
		if len(resp.Errors) > 0 {
			os.Exit(1)
		}
	},
}

func resolveGraphQLQueryInput(cmd *cobra.Command, args []string) (string, error) {
	queryFlag, _ := cmd.Flags().GetString("query")
	fileFlag, _ := cmd.Flags().GetString("file")

	sources := 0
	if strings.TrimSpace(queryFlag) != "" {
		sources++
	}
	if strings.TrimSpace(fileFlag) != "" {
		sources++
	}
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		sources++
	}

	if sources == 0 {
		return "", fmt.Errorf("no query provided; use positional query, --query, or --file")
	}
	if sources > 1 {
		return "", fmt.Errorf("provide only one query source: positional query, --query, or --file")
	}

	if strings.TrimSpace(queryFlag) != "" {
		return strings.TrimSpace(queryFlag), nil
	}
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		return strings.TrimSpace(args[0]), nil
	}

	content, err := os.ReadFile(fileFlag)
	if err != nil {
		return "", fmt.Errorf("failed to read query file %q: %w", fileFlag, err)
	}
	query := strings.TrimSpace(string(content))
	if query == "" {
		return "", fmt.Errorf("query file %q is empty", fileFlag)
	}
	return query, nil
}

func resolveGraphQLVariablesInput(cmd *cobra.Command) (map[string]interface{}, error) {
	varsFlag, _ := cmd.Flags().GetString("variables")
	varsFileFlag, _ := cmd.Flags().GetString("variables-file")

	if strings.TrimSpace(varsFlag) != "" && strings.TrimSpace(varsFileFlag) != "" {
		return nil, fmt.Errorf("provide either --variables or --variables-file, not both")
	}

	raw := "{}"
	if strings.TrimSpace(varsFlag) != "" {
		raw = varsFlag
	} else if strings.TrimSpace(varsFileFlag) != "" {
		content, err := os.ReadFile(varsFileFlag)
		if err != nil {
			return nil, fmt.Errorf("failed to read variables file %q: %w", varsFileFlag, err)
		}
		raw = string(content)
	}

	variables := map[string]interface{}{}
	if err := json.Unmarshal([]byte(raw), &variables); err != nil {
		return nil, fmt.Errorf("variables must be a JSON object: %w", err)
	}
	return variables, nil
}

func init() {
	rootCmd.AddCommand(graphqlCmd)
	graphqlCmd.Flags().StringP("query", "q", "", "GraphQL query/mutation string")
	graphqlCmd.Flags().StringP("file", "f", "", "Path to a .graphql query/mutation file")
	graphqlCmd.Flags().String("variables", "", "Variables as JSON object string")
	graphqlCmd.Flags().String("variables-file", "", "Path to JSON object file for variables")
}

