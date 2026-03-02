package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func resetIssueCreateStateFlags(t *testing.T) {
	t.Helper()
	for _, name := range []string{"title", "team", "state"} {
		flag := issueCreateCmd.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("missing flag %q", name)
		}
		if err := flag.Value.Set(flag.DefValue); err != nil {
			t.Fatalf("reset flag %q: %v", name, err)
		}
		flag.Changed = false
	}
}

func TestIssueCreateCmdStateResolvesToStateID(t *testing.T) {
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	t.Setenv("LINCTL_API_KEY", "test-key")
	viper.Set("plaintext", true)
	viper.Set("json", false)
	resetIssueCreateStateFlags(t)

	var sawStateID string
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var gqlReq gqlCommandTestRequest
		if err := json.NewDecoder(req.Body).Decode(&gqlReq); err != nil {
			t.Fatalf("decode GraphQL request: %v", err)
		}

		switch {
		case strings.Contains(gqlReq.Query, "query Team("):
			body := `{"data":{"team":{"id":"team-1","key":"ENG","name":"Engineering","description":"","private":false,"issueCount":0}}}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		case strings.Contains(gqlReq.Query, "query TeamStates("):
			body := `{"data":{"team":{"states":{"nodes":[{"id":"state-2","name":"In Progress","type":"started","color":"#00ff00","description":"","position":2}]}}}}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		case strings.Contains(gqlReq.Query, "mutation CreateIssue("):
			input, ok := gqlReq.Variables["input"].(map[string]interface{})
			if !ok {
				t.Fatalf("expected input map, got %#v", gqlReq.Variables["input"])
			}
			if v, ok := input["stateId"].(string); ok {
				sawStateID = v
			}
			body := `{"data":{"issueCreate":{"issue":{"id":"i1","identifier":"ENG-1","title":"Fix bug","description":"","priority":3,"estimate":0,"createdAt":"2026-03-02T00:00:00Z","updatedAt":"2026-03-02T00:00:00Z","dueDate":"","state":{"id":"state-2","name":"In Progress","type":"started","color":"#00ff00"},"assignee":null,"team":{"id":"team-1","key":"ENG","name":"Engineering"},"labels":{"nodes":[]},"project":null,"projectMilestone":null,"parent":null}}}}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		default:
			t.Fatalf("unexpected GraphQL operation: %s", gqlReq.Query)
			return nil, nil
		}
	})

	_ = issueCreateCmd.Flags().Set("title", "Fix bug")
	_ = issueCreateCmd.Flags().Set("team", "ENG")
	_ = issueCreateCmd.Flags().Set("state", "In Progress")
	issueCreateCmd.Run(issueCreateCmd, nil)

	if sawStateID != "state-2" {
		t.Fatalf("expected stateId state-2, got %q", sawStateID)
	}
}
