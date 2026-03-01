package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func resetStateUpdateFlags(t *testing.T) {
	t.Helper()
	for _, name := range []string{"name", "color", "description"} {
		flag := teamStateUpdateCmd.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("missing flag %q", name)
		}
		if err := flag.Value.Set(flag.DefValue); err != nil {
			t.Fatalf("reset flag %q: %v", name, err)
		}
		flag.Changed = false
	}
}

func TestTeamStateUpdateCmdSetsInput(t *testing.T) {
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	t.Setenv("LINCTL_API_KEY", "test-key")
	viper.Set("plaintext", true)
	viper.Set("json", false)
	resetStateUpdateFlags(t)

	var sawName string
	var sawColor string

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var gqlReq gqlCommandTestRequest
		if err := json.NewDecoder(req.Body).Decode(&gqlReq); err != nil {
			t.Fatalf("decode GraphQL request: %v", err)
		}

		if !strings.Contains(gqlReq.Query, "mutation WorkflowStateUpdate(") {
			t.Fatalf("expected WorkflowStateUpdate mutation, got: %s", gqlReq.Query)
		}

		input, ok := gqlReq.Variables["input"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected input map, got %#v", gqlReq.Variables["input"])
		}
		if v, ok := input["name"].(string); ok {
			sawName = v
		}
		if v, ok := input["color"].(string); ok {
			sawColor = v
		}

		body := `{"data":{"workflowStateUpdate":{"success":true,"workflowState":{"id":"s1","name":"Ready","type":"backlog","color":"#abc","description":"","position":0}}}}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})

	_ = teamStateUpdateCmd.Flags().Set("name", "Ready")
	_ = teamStateUpdateCmd.Flags().Set("color", "#abc")
	teamStateUpdateCmd.Run(teamStateUpdateCmd, []string{"s1"})

	if sawName != "Ready" {
		t.Fatalf("expected name Ready, got %q", sawName)
	}
	if sawColor != "#abc" {
		t.Fatalf("expected color #abc, got %q", sawColor)
	}
}
