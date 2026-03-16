package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func resetRelationAddFlags(t *testing.T) {
	t.Helper()
	for _, name := range []string{"blocks", "blocked-by", "related", "duplicate"} {
		flag := issueRelationAddCmd.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("missing flag %q", name)
		}
		if err := flag.Value.Set(flag.DefValue); err != nil {
			t.Fatalf("reset flag %q: %v", name, err)
		}
		flag.Changed = false
	}
}

// mockIssueResponse returns a mock GraphQL response for an issue query.
func mockIssueResponse(id, identifier, title string) string {
	return `{"data":{"issue":{"id":"` + id + `","identifier":"` + identifier + `","title":"` + title + `","description":"","priority":0,"createdAt":"2024-01-01T00:00:00Z","updatedAt":"2024-01-01T00:00:00Z","url":"","branchName":"","number":0,"boardOrder":0,"subIssueSortOrder":0,"priorityLabel":"","reactions":[],"slackIssueComments":[],"customerTickets":[],"previousIdentifiers":[]}}}`
}

func TestRelationAddBlocksSendsCorrectMutation(t *testing.T) {
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	t.Setenv("LINCTL_API_KEY", "test-key")
	viper.Set("plaintext", true)
	viper.Set("json", false)
	resetRelationAddFlags(t)

	var capturedInput map[string]interface{}

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var gqlReq gqlCommandTestRequest
		if err := json.NewDecoder(req.Body).Decode(&gqlReq); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		switch {
		case strings.Contains(gqlReq.Query, "query Issue("):
			// Determine which issue is being fetched by inspecting the variable
			issueIDVar, _ := gqlReq.Variables["id"].(string)
			if issueIDVar == "LIN-100" {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(mockIssueResponse("uuid-100", "LIN-100", "Parent issue"))),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(mockIssueResponse("uuid-200", "LIN-200", "Child issue"))),
			}, nil

		case strings.Contains(gqlReq.Query, "mutation IssueRelationCreate("):
			input, ok := gqlReq.Variables["input"].(map[string]interface{})
			if !ok {
				t.Fatalf("expected input map, got %#v", gqlReq.Variables["input"])
			}
			capturedInput = input
			body := `{"data":{"issueRelationCreate":{"success":true,"issueRelation":{"id":"rel-1","type":"blocks","issue":{"id":"uuid-200","identifier":"LIN-200","title":"Child issue"},"relatedIssue":{"id":"uuid-100","identifier":"LIN-100","title":"Parent issue"}}}}}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		}

		t.Fatalf("unexpected query: %s", gqlReq.Query)
		return nil, nil
	})

	// "LIN-100 --blocks LIN-200" means LIN-100 blocks LIN-200
	// => LIN-200 is blocked by LIN-100
	// => API: issueId=LIN-200 (uuid-200), relatedIssueId=LIN-100 (uuid-100), type=blocks
	issueRelationAddCmd.Flags().Set("blocks", "LIN-200")
	issueRelationAddCmd.Run(issueRelationAddCmd, []string{"LIN-100"})

	// Verify the API was called with the correct swapped IDs
	if capturedInput["issueId"] != "uuid-200" {
		t.Fatalf("expected issueId=uuid-200 (the blocked issue), got %v", capturedInput["issueId"])
	}
	if capturedInput["relatedIssueId"] != "uuid-100" {
		t.Fatalf("expected relatedIssueId=uuid-100 (the blocker), got %v", capturedInput["relatedIssueId"])
	}
	if capturedInput["type"] != "blocks" {
		t.Fatalf("expected type=blocks, got %v", capturedInput["type"])
	}
}

func TestRelationAddBlockedBySendsCorrectMutation(t *testing.T) {
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	t.Setenv("LINCTL_API_KEY", "test-key")
	viper.Set("plaintext", true)
	viper.Set("json", false)
	resetRelationAddFlags(t)

	var capturedInput map[string]interface{}

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var gqlReq gqlCommandTestRequest
		if err := json.NewDecoder(req.Body).Decode(&gqlReq); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		switch {
		case strings.Contains(gqlReq.Query, "query Issue("):
			issueIDVar, _ := gqlReq.Variables["id"].(string)
			if issueIDVar == "LIN-100" {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(mockIssueResponse("uuid-100", "LIN-100", "Blocked issue"))),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(mockIssueResponse("uuid-200", "LIN-200", "Blocker issue"))),
			}, nil

		case strings.Contains(gqlReq.Query, "mutation IssueRelationCreate("):
			input, ok := gqlReq.Variables["input"].(map[string]interface{})
			if !ok {
				t.Fatalf("expected input map, got %#v", gqlReq.Variables["input"])
			}
			capturedInput = input
			body := `{"data":{"issueRelationCreate":{"success":true,"issueRelation":{"id":"rel-2","type":"blocks","issue":{"id":"uuid-100","identifier":"LIN-100","title":"Blocked issue"},"relatedIssue":{"id":"uuid-200","identifier":"LIN-200","title":"Blocker issue"}}}}}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		}

		t.Fatalf("unexpected query: %s", gqlReq.Query)
		return nil, nil
	})

	// "LIN-100 --blocked-by LIN-200" means LIN-100 is blocked by LIN-200
	// => API: issueId=LIN-100 (uuid-100), relatedIssueId=LIN-200 (uuid-200), type=blocks
	issueRelationAddCmd.Flags().Set("blocked-by", "LIN-200")
	issueRelationAddCmd.Run(issueRelationAddCmd, []string{"LIN-100"})

	if capturedInput["issueId"] != "uuid-100" {
		t.Fatalf("expected issueId=uuid-100, got %v", capturedInput["issueId"])
	}
	if capturedInput["relatedIssueId"] != "uuid-200" {
		t.Fatalf("expected relatedIssueId=uuid-200, got %v", capturedInput["relatedIssueId"])
	}
	if capturedInput["type"] != "blocks" {
		t.Fatalf("expected type=blocks, got %v", capturedInput["type"])
	}
}

func TestRelationTypeLabel(t *testing.T) {
	cases := map[string]string{
		"blocks":    "blocked by",
		"duplicate": "duplicate of",
		"related":   "related to",
		"unknown":   "unknown",
	}

	for input, want := range cases {
		if got := relationTypeLabel(input); got != want {
			t.Fatalf("relationTypeLabel(%q) = %q, want %q", input, got, want)
		}
	}
}
