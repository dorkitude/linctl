package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type gqlMCPTestRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

func TestDiscoverMCPTools(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlMCPTestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(req.Query, "LinctlMCPIntrospection") {
			t.Fatalf("expected introspection query, got: %s", req.Query)
		}

		body := `{"data":{"__schema":{"queryType":{"name":"Query"},"mutationType":{"name":"Mutation"},"types":[{"kind":"OBJECT","name":"Query","description":"","fields":[{"name":"viewer","description":"Current viewer","args":[],"type":{"kind":"OBJECT","name":"User","ofType":null}},{"name":"issue","description":"Get issue","args":[{"name":"id","description":"Issue ID","type":{"kind":"NON_NULL","name":null,"ofType":{"kind":"SCALAR","name":"String","ofType":null}},"defaultValue":null}],"type":{"kind":"OBJECT","name":"Issue","ofType":null}}],"inputFields":null},{"kind":"OBJECT","name":"Mutation","description":"","fields":[{"name":"issueCreate","description":"Create issue","args":[{"name":"input","description":"Create payload","type":{"kind":"NON_NULL","name":null,"ofType":{"kind":"INPUT_OBJECT","name":"IssueCreateInput","ofType":null}},"defaultValue":null}],"type":{"kind":"OBJECT","name":"IssuePayload","ofType":null}}],"inputFields":null},{"kind":"OBJECT","name":"User","description":"","fields":[{"name":"id","description":"","args":[],"type":{"kind":"NON_NULL","name":null,"ofType":{"kind":"SCALAR","name":"String","ofType":null}}},{"name":"name","description":"","args":[],"type":{"kind":"SCALAR","name":"String","ofType":null}},{"name":"organization","description":"","args":[],"type":{"kind":"OBJECT","name":"Organization","ofType":null}}],"inputFields":null},{"kind":"OBJECT","name":"Issue","description":"","fields":[{"name":"id","description":"","args":[],"type":{"kind":"SCALAR","name":"String","ofType":null}},{"name":"title","description":"","args":[],"type":{"kind":"SCALAR","name":"String","ofType":null}}],"inputFields":null},{"kind":"OBJECT","name":"IssuePayload","description":"","fields":[{"name":"success","description":"","args":[],"type":{"kind":"SCALAR","name":"Boolean","ofType":null}},{"name":"issue","description":"","args":[],"type":{"kind":"OBJECT","name":"Issue","ofType":null}}],"inputFields":null},{"kind":"OBJECT","name":"Organization","description":"","fields":[{"name":"id","description":"","args":[],"type":{"kind":"SCALAR","name":"String","ofType":null}}],"inputFields":null},{"kind":"INPUT_OBJECT","name":"IssueCreateInput","description":"","fields":null,"inputFields":[{"name":"title","description":"","type":{"kind":"NON_NULL","name":null,"ofType":{"kind":"SCALAR","name":"String","ofType":null}},"defaultValue":null}]},{"kind":"SCALAR","name":"String","description":"","fields":null,"inputFields":null},{"kind":"SCALAR","name":"Boolean","description":"","fields":null,"inputFields":null}]}}}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	tools, err := c.DiscoverMCPTools(context.Background())
	if err != nil {
		t.Fatalf("DiscoverMCPTools returned error: %v", err)
	}
	if len(tools) != 3 {
		t.Fatalf("expected 3 tools, got %d", len(tools))
	}

	var sawQueryViewer bool
	var sawMutationCreate bool
	for _, tool := range tools {
		switch tool.Name {
		case "query.viewer":
			sawQueryViewer = true
			if tool.ReturnNamedKind != "OBJECT" {
				t.Fatalf("expected query.viewer object return, got %s", tool.ReturnNamedKind)
			}
			if !strings.Contains(tool.DefaultSelection, "__typename") {
				t.Fatalf("expected default selection for query.viewer, got %q", tool.DefaultSelection)
			}
		case "mutation.issueCreate":
			sawMutationCreate = true
			if len(tool.Args) != 1 || tool.Args[0].Name != "input" || !tool.Args[0].Required {
				t.Fatalf("unexpected mutation args: %#v", tool.Args)
			}
		}
	}
	if !sawQueryViewer {
		t.Fatal("missing query.viewer tool")
	}
	if !sawMutationCreate {
		t.Fatal("missing mutation.issueCreate tool")
	}
}

func TestBuildMCPCall(t *testing.T) {
	tool := MCPTool{
		Name:             "mutation.issueCreate",
		Kind:             "mutation",
		SourceField:      "issueCreate",
		ReturnNamedKind:  "OBJECT",
		DefaultSelection: "{ __typename success }",
		Args: []MCPArg{
			{Name: "input", Type: "IssueCreateInput!", Required: true},
			{Name: "clientMutationId", Type: "String", Required: false},
		},
	}

	query, vars, err := BuildMCPCall(tool, map[string]interface{}{
		"input": map[string]interface{}{"title": "Test"},
	}, "")
	if err != nil {
		t.Fatalf("BuildMCPCall returned error: %v", err)
	}
	if !strings.Contains(query, "mutation LinctlMCPCall") || !strings.Contains(query, "issueCreate(input: $input)") {
		t.Fatalf("unexpected query: %s", query)
	}
	if !strings.Contains(query, "{ __typename success }") {
		t.Fatalf("expected default selection in query, got: %s", query)
	}
	if _, ok := vars["input"]; !ok {
		t.Fatalf("expected input variable, got %#v", vars)
	}

	_, _, err = BuildMCPCall(tool, map[string]interface{}{}, "")
	if err == nil || !strings.Contains(err.Error(), "missing required") {
		t.Fatalf("expected missing required argument error, got: %v", err)
	}

	_, _, err = BuildMCPCall(tool, map[string]interface{}{"input": map[string]interface{}{}, "extra": true}, "")
	if err == nil || !strings.Contains(err.Error(), "unknown argument") {
		t.Fatalf("expected unknown argument error, got: %v", err)
	}
}

func TestBuildMCPCallRejectsSelectionForScalar(t *testing.T) {
	tool := MCPTool{
		Name:            "query.viewerCount",
		Kind:            "query",
		SourceField:     "viewerCount",
		ReturnNamedKind: "SCALAR",
	}

	_, _, err := BuildMCPCall(tool, nil, "{ value }")
	if err == nil || !strings.Contains(err.Error(), "does not accept a selection set") {
		t.Fatalf("expected scalar selection error, got: %v", err)
	}
}
