package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func resetProjectUpdateFlags(t *testing.T) {
	t.Helper()
	for _, name := range []string{"name", "description", "state", "lead", "start-date", "target-date", "color", "content"} {
		flag := projectUpdateCmd.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("missing flag %q", name)
		}
		if err := flag.Value.Set(flag.DefValue); err != nil {
			t.Fatalf("reset flag %q: %v", name, err)
		}
		flag.Changed = false
	}
}

func TestProjectUpdateCmdContentFlag(t *testing.T) {
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	t.Setenv("LINCTL_API_KEY", "test-key")
	viper.Set("plaintext", true)
	viper.Set("json", false)
	resetProjectUpdateFlags(t)

	var sawContent string
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var gqlReq gqlCommandTestRequest
		if err := json.NewDecoder(req.Body).Decode(&gqlReq); err != nil {
			t.Fatalf("decode GraphQL request: %v", err)
		}

		if strings.Contains(gqlReq.Query, "mutation UpdateProject(") {
			input, ok := gqlReq.Variables["input"].(map[string]interface{})
			if !ok {
				t.Fatalf("expected input map, got %#v", gqlReq.Variables["input"])
			}
			if v, ok := input["content"].(string); ok {
				sawContent = v
			}
			body := `{"data":{"projectUpdate":{"success":true,"project":{"id":"proj-1","name":"Test Project","description":"","content":"# Hello world","state":"started","progress":0.5,"startDate":null,"targetDate":null,"url":"https://linear.app/test/project/proj-1","icon":null,"color":"#000","createdAt":"2026-03-01T00:00:00Z","updatedAt":"2026-03-02T00:00:00Z","lead":null,"teams":{"nodes":[]}}}}}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		}

		t.Fatalf("unexpected GraphQL operation: %s", gqlReq.Query)
		return nil, nil
	})

	_ = projectUpdateCmd.Flags().Set("content", "# Hello world")
	projectUpdateCmd.Run(projectUpdateCmd, []string{"proj-1"})

	if sawContent != "# Hello world" {
		t.Fatalf("expected content %q, got %q", "# Hello world", sawContent)
	}
}
