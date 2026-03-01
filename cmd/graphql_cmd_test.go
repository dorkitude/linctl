package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func resetGraphQLFlags(t *testing.T) {
	t.Helper()
	for _, name := range []string{"query", "file", "variables", "variables-file"} {
		flag := graphqlCmd.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("missing flag %q", name)
		}
		if err := flag.Value.Set(flag.DefValue); err != nil {
			t.Fatalf("reset flag %q: %v", name, err)
		}
		flag.Changed = false
	}
}

func TestResolveGraphQLQueryInputValidation(t *testing.T) {
	resetGraphQLFlags(t)

	if _, err := resolveGraphQLQueryInput(graphqlCmd, nil); err == nil {
		t.Fatal("expected error when no query source is provided")
	}

	_ = graphqlCmd.Flags().Set("query", "query { viewer { id } }")
	if _, err := resolveGraphQLQueryInput(graphqlCmd, []string{"query { viewer { name } }"}); err == nil {
		t.Fatal("expected error when multiple query sources are provided")
	}
}

func TestResolveGraphQLInputsFromFiles(t *testing.T) {
	resetGraphQLFlags(t)

	queryFile, err := os.CreateTemp("", "linctl-query-*.graphql")
	if err != nil {
		t.Fatalf("create temp query file: %v", err)
	}
	defer os.Remove(queryFile.Name())
	_, _ = queryFile.WriteString("query($k:String!){ team(id:$k){ id key } }")
	_ = queryFile.Close()

	varsFile, err := os.CreateTemp("", "linctl-vars-*.json")
	if err != nil {
		t.Fatalf("create temp vars file: %v", err)
	}
	defer os.Remove(varsFile.Name())
	_, _ = varsFile.WriteString(`{"k":"ENG"}`)
	_ = varsFile.Close()

	_ = graphqlCmd.Flags().Set("file", queryFile.Name())
	_ = graphqlCmd.Flags().Set("variables-file", varsFile.Name())

	query, err := resolveGraphQLQueryInput(graphqlCmd, nil)
	if err != nil {
		t.Fatalf("resolveGraphQLQueryInput returned error: %v", err)
	}
	if !strings.Contains(query, "team(id:$k)") {
		t.Fatalf("unexpected query: %q", query)
	}

	vars, err := resolveGraphQLVariablesInput(graphqlCmd)
	if err != nil {
		t.Fatalf("resolveGraphQLVariablesInput returned error: %v", err)
	}
	if vars["k"] != "ENG" {
		t.Fatalf("expected vars[k]=ENG, got %#v", vars["k"])
	}
}

func TestGraphQLCmdExecutesRawQuery(t *testing.T) {
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	t.Setenv("LINCTL_API_KEY", "test-key")
	viper.Set("plaintext", true)
	viper.Set("json", false)
	resetGraphQLFlags(t)

	var sawQuery string
	var sawVars map[string]interface{}
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("Authorization") != "test-key" {
			t.Fatalf("expected auth header test-key, got %q", req.Header.Get("Authorization"))
		}

		var gqlReq gqlCommandTestRequest
		if err := json.NewDecoder(req.Body).Decode(&gqlReq); err != nil {
			t.Fatalf("decode GraphQL request: %v", err)
		}

		sawQuery = gqlReq.Query
		sawVars = gqlReq.Variables

		body := `{"data":{"viewer":{"id":"u1","name":"Test User"}}}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})

	_ = graphqlCmd.Flags().Set("query", "query($k:String!){ team(id:$k){ id key name } }")
	_ = graphqlCmd.Flags().Set("variables", `{"k":"ENG"}`)
	graphqlCmd.Run(graphqlCmd, nil)

	if !strings.Contains(sawQuery, "team(id:$k)") {
		t.Fatalf("unexpected query: %q", sawQuery)
	}
	if sawVars["k"] != "ENG" {
		t.Fatalf("expected variables[k]=ENG, got %#v", sawVars["k"])
	}
}

