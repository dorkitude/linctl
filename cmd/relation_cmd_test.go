package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/dorkitude/linctl/pkg/api"
	"github.com/spf13/viper"
)

func resetRelationAddFlags(t *testing.T) {
	t.Helper()
	for _, name := range []string{"blocks", "blocked-by", "related", "duplicate", "similar"} {
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
			body := `{"data":{"issueRelationCreate":{"success":true,"issueRelation":{"id":"rel-1","type":"blocks","issue":{"id":"uuid-100","identifier":"LIN-100","title":"Parent issue"},"relatedIssue":{"id":"uuid-200","identifier":"LIN-200","title":"Child issue"}}}}}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		}

		t.Fatalf("unexpected query: %s", gqlReq.Query)
		return nil, nil
	})

	// "LIN-100 --blocks LIN-200" means LIN-100 blocks LIN-200.
	// Linear's type describes how the source relates to the target, so the
	// blocker is the source: issueId=LIN-100 (uuid-100),
	// relatedIssueId=LIN-200 (uuid-200), type=blocks.
	issueRelationAddCmd.Flags().Set("blocks", "LIN-200")
	issueRelationAddCmd.Run(issueRelationAddCmd, []string{"LIN-100"})

	// Verify the blocker is sent as the source issue, not the target
	if capturedInput["issueId"] != "uuid-100" {
		t.Fatalf("expected issueId=uuid-100 (the blocker), got %v", capturedInput["issueId"])
	}
	if capturedInput["relatedIssueId"] != "uuid-200" {
		t.Fatalf("expected relatedIssueId=uuid-200 (the blocked issue), got %v", capturedInput["relatedIssueId"])
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
			body := `{"data":{"issueRelationCreate":{"success":true,"issueRelation":{"id":"rel-2","type":"blocks","issue":{"id":"uuid-200","identifier":"LIN-200","title":"Blocker issue"},"relatedIssue":{"id":"uuid-100","identifier":"LIN-100","title":"Blocked issue"}}}}}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		}

		t.Fatalf("unexpected query: %s", gqlReq.Query)
		return nil, nil
	})

	// "LIN-100 --blocked-by LIN-200" means LIN-200 blocks LIN-100, so LIN-200 is
	// the source: issueId=LIN-200 (uuid-200), relatedIssueId=LIN-100 (uuid-100),
	// type=blocks.
	issueRelationAddCmd.Flags().Set("blocked-by", "LIN-200")
	issueRelationAddCmd.Run(issueRelationAddCmd, []string{"LIN-100"})

	if capturedInput["issueId"] != "uuid-200" {
		t.Fatalf("expected issueId=uuid-200 (the blocker), got %v", capturedInput["issueId"])
	}
	if capturedInput["relatedIssueId"] != "uuid-100" {
		t.Fatalf("expected relatedIssueId=uuid-100 (the blocked issue), got %v", capturedInput["relatedIssueId"])
	}
	if capturedInput["type"] != "blocks" {
		t.Fatalf("expected type=blocks, got %v", capturedInput["type"])
	}
}

// runRelationList runs 'issue relation list' in plaintext mode against a mocked
// IssueRelations response and returns everything written to stdout.
func runRelationList(t *testing.T, issueID, body string) string {
	t.Helper()

	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	t.Setenv("LINCTL_API_KEY", "test-key")
	viper.Set("plaintext", true)
	viper.Set("json", false)

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var gqlReq gqlCommandTestRequest
		if err := json.NewDecoder(req.Body).Decode(&gqlReq); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(gqlReq.Query, "IssueRelations") {
			t.Fatalf("expected IssueRelations query, got: %s", gqlReq.Query)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	issueRelationListCmd.Run(issueRelationListCmd, []string{issueID})

	w.Close()
	var buf strings.Builder
	io.Copy(&buf, r)
	os.Stdout = oldStdout
	return buf.String()
}

// relationLine returns the plaintext output line describing the given relation ID.
func relationLine(t *testing.T, out, relationID string) string {
	t.Helper()
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, relationID) {
			return line
		}
	}
	t.Fatalf("no output line for relation %q in:\n%s", relationID, out)
	return ""
}

func TestRelationListShowsRelations(t *testing.T) {
	// Two "blocks" edges seen from LIN-100:
	//   rel-1 (forward):  LIN-100 blocks LIN-200
	//   rel-2 (inverse):  LIN-300 blocks LIN-100
	// So LIN-100 should read as "blocks LIN-200" and "blocked by LIN-300",
	// each naming the counterpart rather than LIN-100 itself.
	body := `{"data":{"issue":{
		"relations":{"nodes":[
			{"id":"rel-1","type":"blocks","issue":{"id":"uuid-100","identifier":"LIN-100","title":"This issue"},"relatedIssue":{"id":"uuid-200","identifier":"LIN-200","title":"Blocked task"}}
		]},
		"inverseRelations":{"nodes":[
			{"id":"rel-2","type":"blocks","issue":{"id":"uuid-300","identifier":"LIN-300","title":"Blocker task"},"relatedIssue":{"id":"uuid-100","identifier":"LIN-100","title":"This issue"}}
		]}
	}}}`

	got := runRelationList(t, "LIN-100", body)

	// Forward "blocks" relation: the queried issue is the blocker.
	forward := relationLine(t, got, "rel-1")
	if strings.Contains(forward, "blocked by") || !strings.Contains(forward, "blocks") {
		t.Fatalf("expected forward blocks relation labeled 'blocks', got:\n%s", forward)
	}
	if !strings.Contains(forward, "LIN-200") {
		t.Fatalf("expected forward relation to name counterpart LIN-200, got:\n%s", forward)
	}

	// Inverse "blocks" relation: the queried issue is the one being blocked.
	inverse := relationLine(t, got, "rel-2")
	if !strings.Contains(inverse, "blocked by") {
		t.Fatalf("expected inverse blocks relation labeled 'blocked by', got:\n%s", inverse)
	}
	if !strings.Contains(inverse, "LIN-300") {
		t.Fatalf("expected inverse relation to name counterpart LIN-300, got:\n%s", inverse)
	}
	// Regression: the counterpart must never be the issue that was queried.
	if strings.Contains(inverse, "LIN-100") {
		t.Fatalf("inverse relation named the queried issue as its own counterpart:\n%s", inverse)
	}
}

// TestRelationListDirectionIsSymmetric reads the same edge from both ends and
// asserts the two views are consistent: opposite labels, each naming the other
// issue. This is the property that fails when either the label mapping or the
// counterpart selection is inverted, so it covers both defects at once.
func TestRelationListDirectionIsSymmetric(t *testing.T) {
	// The single edge under test: LIN-500 blocks LIN-600.
	edge := `{"id":"rel-sym","type":"blocks","issue":{"id":"uuid-500","identifier":"LIN-500","title":"Blocker"},"relatedIssue":{"id":"uuid-600","identifier":"LIN-600","title":"Blocked"}}`

	// From LIN-500 (the blocker) the edge arrives via relations.
	fromBlocker := runRelationList(t, "LIN-500", `{"data":{"issue":{
		"relations":{"nodes":[`+edge+`]},
		"inverseRelations":{"nodes":[]}
	}}}`)

	// From LIN-600 (the blocked) the same edge arrives via inverseRelations.
	fromBlocked := runRelationList(t, "LIN-600", `{"data":{"issue":{
		"relations":{"nodes":[]},
		"inverseRelations":{"nodes":[`+edge+`]}
	}}}`)

	blockerView := relationLine(t, fromBlocker, "rel-sym")
	if strings.Contains(blockerView, "blocked by") || !strings.Contains(blockerView, "blocks") {
		t.Fatalf("blocker should read as 'blocks', got:\n%s", blockerView)
	}
	if !strings.Contains(blockerView, "LIN-600") || strings.Contains(blockerView, "LIN-500") {
		t.Fatalf("blocker view should name LIN-600 and not itself, got:\n%s", blockerView)
	}

	blockedView := relationLine(t, fromBlocked, "rel-sym")
	if !strings.Contains(blockedView, "blocked by") {
		t.Fatalf("blocked issue should read as 'blocked by', got:\n%s", blockedView)
	}
	if !strings.Contains(blockedView, "LIN-500") || strings.Contains(blockedView, "LIN-600") {
		t.Fatalf("blocked view should name LIN-500 and not itself, got:\n%s", blockedView)
	}
}

func TestRelationOtherIssue(t *testing.T) {
	source := &api.Issue{ID: "uuid-1", Identifier: "LIN-1", Title: "Source"}
	target := &api.Issue{ID: "uuid-2", Identifier: "LIN-2", Title: "Target"}

	cases := []struct {
		name string
		rel  api.IssueRelation
		want string
	}{
		{
			// Queried issue is Issue, so the counterpart is RelatedIssue.
			name: "forward returns related issue",
			rel:  api.IssueRelation{Issue: source, RelatedIssue: target},
			want: "LIN-2",
		},
		{
			// Queried issue is RelatedIssue, so the counterpart is Issue.
			// Both sides are populated — the query selects both on both edges —
			// so this only works if Inverse is honoured.
			name: "inverse returns source issue",
			rel:  api.IssueRelation{Issue: source, RelatedIssue: target, Inverse: true},
			want: "LIN-1",
		},
		{
			name: "forward falls back to issue when related is nil",
			rel:  api.IssueRelation{Issue: source},
			want: "LIN-1",
		},
		{
			name: "inverse falls back to related when issue is nil",
			rel:  api.IssueRelation{RelatedIssue: target, Inverse: true},
			want: "LIN-2",
		},
		{
			name: "both nil yields placeholder",
			rel:  api.IssueRelation{},
			want: "?",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := relationOtherIssue(&tc.rel); got.Identifier != tc.want {
				t.Fatalf("relationOtherIssue() = %q, want %q", got.Identifier, tc.want)
			}
		})
	}
}

func TestRelationRemoveSendsDeleteMutation(t *testing.T) {
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	t.Setenv("LINCTL_API_KEY", "test-key")
	viper.Set("plaintext", true)
	viper.Set("json", false)

	var capturedID string

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var gqlReq gqlCommandTestRequest
		if err := json.NewDecoder(req.Body).Decode(&gqlReq); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if !strings.Contains(gqlReq.Query, "mutation IssueRelationDelete(") {
			t.Fatalf("expected IssueRelationDelete mutation, got: %s", gqlReq.Query)
		}

		capturedID, _ = gqlReq.Variables["id"].(string)

		body := `{"data":{"issueRelationDelete":{"success":true}}}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})

	issueRelationRemoveCmd.Run(issueRelationRemoveCmd, []string{"rel-abc-123"})

	if capturedID != "rel-abc-123" {
		t.Fatalf("expected relation ID rel-abc-123, got %v", capturedID)
	}
}

func TestRelationTypeLabel(t *testing.T) {
	// Forward (non-inverse) labels — the queried issue is the subject acting
	// on the counterpart, so "blocks" and "duplicate of" share one convention.
	forwardCases := map[string]string{
		"blocks":    "blocks",
		"duplicate": "duplicate of",
		"related":   "related to",
		"similar":   "similar to",
		"unknown":   "unknown",
	}
	for input, want := range forwardCases {
		if got := relationTypeLabel(input, false); got != want {
			t.Fatalf("relationTypeLabel(%q, false) = %q, want %q", input, got, want)
		}
	}

	// Inverse labels — direction-sensitive types should flip
	inverseCases := map[string]string{
		"blocks":    "blocked by",
		"duplicate": "has duplicate",
		"related":   "related to",
		"similar":   "similar to",
	}
	for input, want := range inverseCases {
		if got := relationTypeLabel(input, true); got != want {
			t.Fatalf("relationTypeLabel(%q, true) = %q, want %q", input, got, want)
		}
	}
}
