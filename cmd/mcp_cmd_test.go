package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dorkitude/linctl/pkg/api"
	"github.com/dorkitude/linctl/pkg/mcpcache"
	"github.com/spf13/viper"
)

func resetMCPCacheFlags(t *testing.T) {
	t.Helper()
	mcpInputStdin = os.Stdin
	for _, name := range []string{"json", "selection"} {
		flag := mcpCallCmd.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("missing flag %q", name)
		}
		if err := flag.Value.Set(flag.DefValue); err != nil {
			t.Fatalf("reset flag %q: %v", name, err)
		}
		flag.Changed = false
	}
	quietFlag := mcpSyncCmd.Flags().Lookup("quiet")
	if quietFlag == nil {
		t.Fatal("missing quiet flag")
	}
	if err := quietFlag.Value.Set(quietFlag.DefValue); err != nil {
		t.Fatalf("reset quiet flag: %v", err)
	}
	quietFlag.Changed = false
}

func TestResolveMCPTool(t *testing.T) {
	tools := []api.MCPTool{
		{Name: "query.viewer", SourceField: "viewer", Kind: "query"},
		{Name: "mutation.viewer", SourceField: "viewer", Kind: "mutation"},
		{Name: "mutation.issueCreate", SourceField: "issueCreate", Kind: "mutation"},
	}

	tool, err := resolveMCPTool(tools, "mutation.issueCreate")
	if err != nil {
		t.Fatalf("resolveMCPTool exact match returned error: %v", err)
	}
	if tool.Name != "mutation.issueCreate" {
		t.Fatalf("expected mutation.issueCreate, got %s", tool.Name)
	}

	tool, err = resolveMCPTool(tools, "issueCreate")
	if err != nil {
		t.Fatalf("resolveMCPTool source field match returned error: %v", err)
	}
	if tool.Name != "mutation.issueCreate" {
		t.Fatalf("expected mutation.issueCreate by source field, got %s", tool.Name)
	}

	_, err = resolveMCPTool(tools, "viewer")
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("expected ambiguous error, got %v", err)
	}
}

func TestLoadMCPArguments(t *testing.T) {
	resetMCPCacheFlags(t)

	args, err := loadMCPArguments(`{"input":{"title":"T"}}`)
	if err != nil {
		t.Fatalf("loadMCPArguments with flag returned error: %v", err)
	}
	if _, ok := args["input"]; !ok {
		t.Fatalf("expected input key, got %#v", args)
	}

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	defer reader.Close()
	_, _ = writer.WriteString(`{"team":"ENG"}`)
	_ = writer.Close()
	mcpInputStdin = reader

	args, err = loadMCPArguments("")
	if err != nil {
		t.Fatalf("loadMCPArguments from stdin returned error: %v", err)
	}
	if args["team"] != "ENG" {
		t.Fatalf("expected team ENG, got %#v", args["team"])
	}
}

func TestRunMCPSyncWritesCache(t *testing.T) {
	resetMCPCacheFlags(t)
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	viper.Set("plaintext", true)
	viper.Set("json", false)
	t.Setenv("LINCTL_API_KEY", "test-key")

	home := t.TempDir()
	t.Setenv("HOME", home)

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var gqlReq gqlCommandTestRequest
		if err := json.NewDecoder(req.Body).Decode(&gqlReq); err != nil {
			t.Fatalf("decode GraphQL request: %v", err)
		}
		if !strings.Contains(gqlReq.Query, "LinctlMCPIntrospection") {
			t.Fatalf("unexpected query: %s", gqlReq.Query)
		}
		body := `{"data":{"__schema":{"queryType":{"name":"Query"},"mutationType":{"name":"Mutation"},"types":[{"kind":"OBJECT","name":"Query","fields":[{"name":"viewer","description":"","args":[],"type":{"kind":"OBJECT","name":"User","ofType":null}}],"inputFields":null},{"kind":"OBJECT","name":"Mutation","fields":[],"inputFields":null},{"kind":"OBJECT","name":"User","fields":[{"name":"id","description":"","args":[],"type":{"kind":"SCALAR","name":"String","ofType":null}}],"inputFields":null},{"kind":"SCALAR","name":"String","fields":null,"inputFields":null}]}}}`
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
	})

	if err := runMCPSync(mcpSyncCmd, false); err != nil {
		t.Fatalf("runMCPSync returned error: %v", err)
	}

	cachePath := filepath.Join(home, ".linctl", "mcp-tools-cache.json")
	if _, err := os.Stat(cachePath); err != nil {
		t.Fatalf("expected cache file at %s: %v", cachePath, err)
	}

	cache, err := mcpcache.Load()
	if err != nil {
		t.Fatalf("mcpcache.Load returned error: %v", err)
	}
	if len(cache.Tools) != 1 || cache.Tools[0].Name != "query.viewer" {
		t.Fatalf("unexpected cache tools: %#v", cache.Tools)
	}
}

func TestMCPCmdCallExecutes(t *testing.T) {
	resetMCPCacheFlags(t)
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	viper.Set("plaintext", true)
	viper.Set("json", false)
	t.Setenv("LINCTL_API_KEY", "test-key")

	home := t.TempDir()
	t.Setenv("HOME", home)

	err := mcpcache.Save(mcpcache.Cache{
		FetchedAt: time.Now().UTC(),
		Tools: []api.MCPTool{
			{
				Name:             "query.viewer",
				Kind:             "query",
				SourceField:      "viewer",
				ReturnType:       "User",
				ReturnNamedKind:  "OBJECT",
				DefaultSelection: "{ __typename id }",
			},
		},
	})
	if err != nil {
		t.Fatalf("mcpcache.Save returned error: %v", err)
	}

	var sawQuery string
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var gqlReq gqlCommandTestRequest
		if err := json.NewDecoder(req.Body).Decode(&gqlReq); err != nil {
			t.Fatalf("decode GraphQL request: %v", err)
		}
		sawQuery = gqlReq.Query
		body := `{"data":{"viewer":{"__typename":"User","id":"u1"}}}`
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
	})

	_ = mcpCallCmd.Flags().Set("json", `{}`)
	mcpCallCmd.Run(mcpCallCmd, []string{"query.viewer"})

	if !strings.Contains(sawQuery, "query LinctlMCPCall") || !strings.Contains(sawQuery, "viewer") {
		t.Fatalf("unexpected query: %s", sawQuery)
	}
}
