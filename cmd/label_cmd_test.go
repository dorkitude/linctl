package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

type gqlCommandTestRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func resetLabelCreateFlags(t *testing.T) {
	t.Helper()
	for _, name := range []string{"team", "name", "color", "description", "parent", "is-group"} {
		flag := labelCreateCmd.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("missing flag %q", name)
		}
		if err := flag.Value.Set(flag.DefValue); err != nil {
			t.Fatalf("reset flag %q: %v", name, err)
		}
		flag.Changed = false
	}
}

func TestLabelCreateCmdIsGroupSetsInput(t *testing.T) {
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	t.Setenv("LINCTL_API_KEY", "test-key")
	viper.Set("plaintext", true)
	viper.Set("json", false)
	resetLabelCreateFlags(t)

	requestCount := 0
	var sawIsGroup bool

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requestCount++
		if req.Header.Get("Authorization") != "test-key" {
			t.Fatalf("expected auth header test-key, got %q", req.Header.Get("Authorization"))
		}

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
		case strings.Contains(gqlReq.Query, "mutation CreateLabel("):
			input, ok := gqlReq.Variables["input"].(map[string]interface{})
			if !ok {
				t.Fatalf("expected input map, got %#v", gqlReq.Variables["input"])
			}
			val, exists := input["isGroup"]
			sawIsGroup = exists && val == true
			body := `{"data":{"issueLabelCreate":{"success":true,"issueLabel":{"id":"label-1","name":"backend","color":"#0055ff","description":"Backend group","parent":null,"team":{"id":"team-1","key":"ENG","name":"Engineering"}}}}}`
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

	_ = labelCreateCmd.Flags().Set("team", "ENG")
	_ = labelCreateCmd.Flags().Set("name", "backend")
	_ = labelCreateCmd.Flags().Set("is-group", "true")
	labelCreateCmd.Run(labelCreateCmd, nil)

	if requestCount != 2 {
		t.Fatalf("expected 2 GraphQL calls, got %d", requestCount)
	}
	if !sawIsGroup {
		t.Fatalf("expected create payload to include isGroup=true")
	}
}

func TestLabelCreateCmdDefaultOmitsIsGroup(t *testing.T) {
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	t.Setenv("LINCTL_API_KEY", "test-key")
	viper.Set("plaintext", true)
	viper.Set("json", false)
	resetLabelCreateFlags(t)

	var sawIsGroup bool

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
		case strings.Contains(gqlReq.Query, "mutation CreateLabel("):
			input, ok := gqlReq.Variables["input"].(map[string]interface{})
			if !ok {
				t.Fatalf("expected input map, got %#v", gqlReq.Variables["input"])
			}
			_, sawIsGroup = input["isGroup"]
			body := `{"data":{"issueLabelCreate":{"success":true,"issueLabel":{"id":"label-2","name":"bug","color":"#ff0000","description":"Bug label","parent":null,"team":{"id":"team-1","key":"ENG","name":"Engineering"}}}}}`
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

	_ = labelCreateCmd.Flags().Set("team", "ENG")
	_ = labelCreateCmd.Flags().Set("name", "bug")
	labelCreateCmd.Run(labelCreateCmd, nil)

	if sawIsGroup {
		t.Fatalf("expected create payload to omit isGroup by default")
	}
}
