package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/dorkitude/linctl/pkg/api"
	"github.com/dorkitude/linctl/pkg/auth"
	"github.com/dorkitude/linctl/pkg/mcpcache"
	"github.com/dorkitude/linctl/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const mcpCacheMaxAge = 12 * time.Hour

var (
	mcpCmd = &cobra.Command{
		Use:   "mcp",
		Short: "Dynamic GraphQL-backed Linear tooling",
		Long: `Dynamic Linear capability surface generated from GraphQL introspection.

Use this for newly added Linear features before first-class linctl commands exist.

Examples:
  linctl mcp sync
  linctl mcp tools
  linctl mcp call query.viewer
  linctl mcp call mutation.issueCreate --json '{"input":{"title":"Test issue","teamId":"..."}}'`,
	}
	mcpSyncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Refresh MCP tool registry cache",
		Run: func(cmd *cobra.Command, args []string) {
			_ = args
			quiet, _ := cmd.Flags().GetBool("quiet")
			if err := runMCPSync(cmd, quiet); err != nil {
				plaintext := viper.GetBool("plaintext")
				jsonOut := viper.GetBool("json")
				output.Error(err.Error(), plaintext, jsonOut)
				os.Exit(1)
			}
		},
	}
	mcpToolsCmd = &cobra.Command{
		Use:   "tools",
		Short: "List cached MCP tools",
		Run: func(cmd *cobra.Command, args []string) {
			_ = args
			plaintext := viper.GetBool("plaintext")
			jsonOut := viper.GetBool("json")

			cache, err := mcpcache.Load()
			if err != nil {
				output.Error("MCP cache not found. Run 'linctl mcp sync' first.", plaintext, jsonOut)
				os.Exit(1)
			}

			rows := make([][]string, 0, len(cache.Tools))
			for _, tool := range cache.Tools {
				rows = append(rows, []string{
					tool.Name,
					tool.Kind,
					summarizeArgs(tool.Args),
					tool.ReturnType,
					summarizeDescription(tool.Description),
				})
			}
			sort.Slice(rows, func(i, j int) bool { return rows[i][0] < rows[j][0] })

			if jsonOut {
				output.JSON(map[string]interface{}{
					"fetched_at": cache.FetchedAt,
					"ttl_hours":  int(mcpCacheMaxAge.Hours()),
					"tools":      cache.Tools,
				})
				return
			}

			output.Table(output.TableData{
				Headers: []string{"Name", "Kind", "Args", "Returns", "Description"},
				Rows:    rows,
			}, plaintext, jsonOut)

			stale := mcpcache.IsStale(cache, mcpCacheMaxAge)
			age := time.Since(cache.FetchedAt).Round(time.Minute)
			if age < 0 {
				age = 0
			}
			if stale {
				output.Info(fmt.Sprintf("MCP cache is stale (%s old). Refresh with 'linctl mcp sync'.", age), plaintext, jsonOut)
			} else {
				output.Info(fmt.Sprintf("MCP cache age: %s (TTL %dh)", age, int(mcpCacheMaxAge.Hours())), plaintext, jsonOut)
			}
		},
	}
	mcpCallCmd = &cobra.Command{
		Use:   "call <tool-name>",
		Short: "Call a cached MCP tool by name",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			plaintext := viper.GetBool("plaintext")
			jsonOut := viper.GetBool("json")

			cache, err := mcpcache.Load()
			if err != nil {
				output.Error("MCP cache not found. Run 'linctl mcp sync' first.", plaintext, jsonOut)
				os.Exit(1)
			}

			tool, err := resolveMCPTool(cache.Tools, args[0])
			if err != nil {
				output.Error(err.Error(), plaintext, jsonOut)
				os.Exit(1)
			}

			argJSON, _ := cmd.Flags().GetString("json")
			arguments, err := loadMCPArguments(argJSON)
			if err != nil {
				output.Error(err.Error(), plaintext, jsonOut)
				os.Exit(1)
			}

			selection, _ := cmd.Flags().GetString("selection")
			query, variables, err := api.BuildMCPCall(tool, arguments, selection)
			if err != nil {
				output.Error(fmt.Sprintf("failed to build call: %v", err), plaintext, jsonOut)
				os.Exit(1)
			}

			authHeader, err := auth.GetAuthHeader()
			if err != nil {
				output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
				os.Exit(1)
			}
			client := api.NewClient(authHeader)

			resp, err := client.ExecuteRaw(context.Background(), query, variables)
			if err != nil {
				output.Error(fmt.Sprintf("MCP call failed: %v", err), plaintext, jsonOut)
				os.Exit(1)
			}

			output.JSON(resp)
			if len(resp.Errors) > 0 {
				os.Exit(1)
			}
		},
	}
)

var mcpInputStdin = os.Stdin

func init() {
	rootCmd.AddCommand(mcpCmd)

	mcpCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		_ = args
		maybeStartAutoMCPSync(cmd)
	}

	mcpCmd.AddCommand(mcpSyncCmd)
	mcpCmd.AddCommand(mcpToolsCmd)
	mcpCmd.AddCommand(mcpCallCmd)

	mcpSyncCmd.Flags().Bool("quiet", false, "suppress success output")
	mcpCallCmd.Flags().String("json", "", "JSON object of tool arguments; if omitted, piped stdin is used or an empty object is sent")
	mcpCallCmd.Flags().String("selection", "", "Override selection set (for object/interface/union returns)")
}

func runMCPSync(cmd *cobra.Command, quiet bool) error {
	authHeader, err := auth.GetAuthHeader()
	if err != nil {
		return fmt.Errorf("not authenticated; run 'linctl auth' first")
	}
	client := api.NewClient(authHeader)

	tools, err := client.DiscoverMCPTools(context.Background())
	if err != nil {
		return fmt.Errorf("discover MCP tools: %w", err)
	}

	cache := mcpcache.Cache{
		FetchedAt: time.Now().UTC(),
		Tools:     tools,
	}
	if err := mcpcache.Save(cache); err != nil {
		return fmt.Errorf("save MCP cache: %w", err)
	}

	if quiet {
		return nil
	}

	cachePath, err := mcpcache.GetPath()
	if err != nil {
		return fmt.Errorf("resolve cache path: %w", err)
	}

	plaintext := viper.GetBool("plaintext")
	jsonOut := viper.GetBool("json")
	if jsonOut {
		output.JSON(map[string]interface{}{
			"status":       "ok",
			"synced_tools": len(tools),
			"fetched_at":   cache.FetchedAt,
			"cache":        cachePath,
		})
		return nil
	}

	output.Success(fmt.Sprintf("Synced %d MCP tool(s)", len(tools)), plaintext, jsonOut)
	output.Info(fmt.Sprintf("Cache: %s", cachePath), plaintext, jsonOut)
	return nil
}

func loadMCPArguments(flagValue string) (map[string]interface{}, error) {
	trimmed := strings.TrimSpace(flagValue)
	if trimmed != "" {
		return decodeMCPArguments(trimmed)
	}

	stdinPiped, err := isMCPStdinPiped()
	if err != nil {
		return nil, err
	}
	if !stdinPiped {
		return map[string]interface{}{}, nil
	}

	data, err := io.ReadAll(mcpInputStdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read stdin: %w", err)
	}
	if strings.TrimSpace(string(data)) == "" {
		return map[string]interface{}{}, nil
	}
	return decodeMCPArguments(string(data))
}

func decodeMCPArguments(raw string) (map[string]interface{}, error) {
	var arguments map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &arguments); err != nil {
		return nil, fmt.Errorf("tool arguments must be a JSON object: %w", err)
	}
	if arguments == nil {
		return nil, fmt.Errorf("tool arguments must be a JSON object")
	}
	return arguments, nil
}

func isMCPStdinPiped() (bool, error) {
	info, err := mcpInputStdin.Stat()
	if err != nil {
		return false, err
	}
	return (info.Mode() & os.ModeCharDevice) == 0, nil
}

func resolveMCPTool(tools []api.MCPTool, requested string) (api.MCPTool, error) {
	want := strings.TrimSpace(requested)
	if want == "" {
		return api.MCPTool{}, fmt.Errorf("tool name is required")
	}

	for _, tool := range tools {
		if tool.Name == want {
			return tool, nil
		}
	}

	var candidates []api.MCPTool
	for _, tool := range tools {
		if tool.SourceField == want {
			candidates = append(candidates, tool)
		}
	}
	if len(candidates) == 1 {
		return candidates[0], nil
	}
	if len(candidates) > 1 {
		fullNames := make([]string, 0, len(candidates))
		for _, c := range candidates {
			fullNames = append(fullNames, c.Name)
		}
		sort.Strings(fullNames)
		return api.MCPTool{}, fmt.Errorf("tool %q is ambiguous; use one of: %s", want, strings.Join(fullNames, ", "))
	}

	return api.MCPTool{}, fmt.Errorf("tool %q not found in cache; run 'linctl mcp tools' to inspect available tools", want)
}

func summarizeArgs(args []api.MCPArg) string {
	if len(args) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		marker := ""
		if arg.Required {
			marker = "*"
		}
		parts = append(parts, arg.Name+marker)
	}
	return strings.Join(parts, ",")
}

func summarizeDescription(s string) string {
	text := strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
	if text == "" {
		return "-"
	}
	const maxLen = 88
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}
