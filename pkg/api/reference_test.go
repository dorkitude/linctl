package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseLinearURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want linearRef
		ok   bool
	}{
		{"issue", "https://linear.app/glif/issue/API-379/a-title", linearRef{kind: refIssue, id: "API-379"}, true},
		{"comment fragment", "https://linear.app/glif/issue/GTM-580/title#comment-b68a4bf5", linearRef{kind: refIssue, id: "GTM-580", commentID: "b68a4bf5"}, true},
		{"comment query", "https://linear.app/glif/issue/GTM-580?commentId=b68a4bf5-8a34-473e-af4c-b8892a78a9af", linearRef{kind: refIssue, id: "GTM-580", commentID: "b68a4bf5-8a34-473e-af4c-b8892a78a9af"}, true},
		{"other fragment", "https://linear.app/glif/issue/API-379/title#agent-session-56881231", linearRef{kind: refIssue, id: "API-379"}, true},
		{"without scheme", "linear.app/glif/issue/API-379", linearRef{kind: refIssue, id: "API-379"}, true},
		{"desktop link", "linear://linear.app/glif/issue/API-379", linearRef{kind: refIssue, id: "API-379"}, true},
		{"www host", "https://www.linear.app/glif/team/API/active", linearRef{kind: refTeam, id: "API"}, true},
		{"project", "https://linear.app/glif/project/roadmap-d05c5c7e8a5c/overview", linearRef{kind: refProject, id: "roadmap-d05c5c7e8a5c"}, true},
		{"document", "https://linear.app/glif/document/notes-507735cb56c1", linearRef{kind: refDocument, id: "notes-507735cb56c1"}, true},
		{"initiative", "https://linear.app/glif/initiative/costs-83e5d6e7a371", linearRef{kind: refInitiative, id: "costs-83e5d6e7a371"}, true},
		{"review", "https://linear.app/glif/review/change-ca4153a35dc0", linearRef{kind: refReview, id: "change-ca4153a35dc0"}, true},
		{"bare identifier", "API-379", linearRef{}, false},
		{"wrong host", "https://github.com/acme/api/pull/1", linearRef{}, false},
		{"wrong scheme", "ftp://linear.app/glif/issue/API-379", linearRef{}, false},
		{"malformed URL", "https://linear.app/%zz", linearRef{}, false},
		{"missing workspace", "https://linear.app/issue/API-379", linearRef{}, false},
		{"missing id", "https://linear.app/glif/issue", linearRef{}, false},
		{"settings", "https://linear.app/glif/settings", linearRef{}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := parseLinearURL(test.raw)
			if got != test.want || ok != test.ok {
				t.Fatalf("parseLinearURL(%q) = %#v, %v; want %#v, %v", test.raw, got, ok, test.want, test.ok)
			}
		})
	}
}

func TestNormalizeRefs(t *testing.T) {
	tests := []struct {
		name    string
		fn      func(string) (string, error)
		raw     string
		want    string
		wantErr string
	}{
		{"issue identifier", NormalizeIssueRef, " API-379\n", "API-379", ""},
		{"issue UUID", NormalizeIssueRef, "b68a4bf5-8a34-473e-af4c-b8892a78a9af", "b68a4bf5-8a34-473e-af4c-b8892a78a9af", ""},
		{"issue URL", NormalizeIssueRef, "https://linear.app/glif/issue/API-379/title", "API-379", ""},
		{"project URL", NormalizeProjectRef, "https://linear.app/glif/project/roadmap-d05c5c7e8a5c/issues", "roadmap-d05c5c7e8a5c", ""},
		{"team URL", NormalizeTeamRef, "https://linear.app/glif/team/API/all", "API", ""},
		{"empty issue", NormalizeIssueRef, " ", "", "issue reference cannot be empty"},
		{"wrong kind", NormalizeTeamRef, "https://linear.app/glif/issue/API-379", "", "Linear issue URL, not a team URL"},
		{"new issue page", NormalizeIssueRef, "https://linear.app/glif/issue/new", "", "does not contain an issue identifier"},
		{"review URL", NormalizeIssueRef, "https://linear.app/glif/review/change-ca4153a35dc0", "", "Pass the GitHub pull request URL"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.fn(test.raw)
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("error = %v, want it to contain %q", err, test.wantErr)
				}
				return
			}
			if err != nil || got != test.want {
				t.Fatalf("got %q, %v; want %q, nil", got, err, test.want)
			}
		})
	}
}

func TestGetIssueAcceptsIssueURL(t *testing.T) {
	server := graphqlTestServer(t, func(req gqlTestRequest) string {
		if req.Variables["id"] != "API-379" {
			t.Fatalf("id = %v, want API-379", req.Variables["id"])
		}
		return `{"data":{"issue":{"id":"i1","identifier":"API-379"}}}`
	})
	defer server.Close()

	issue, err := NewClientWithURL(server.URL, "test").GetIssue(context.Background(), "https://linear.app/glif/issue/API-379/title")
	if err != nil || issue.Identifier != "API-379" {
		t.Fatalf("GetIssue() = %#v, %v", issue, err)
	}
}

func TestGetProjectMilestonesAcceptsProjectURL(t *testing.T) {
	server := graphqlTestServer(t, func(req gqlTestRequest) string {
		if req.Variables["id"] != "roadmap-d05c5c7e8a5c" {
			t.Fatalf("id = %v, want project URL identifier", req.Variables["id"])
		}
		return `{"data":{"project":{"projectMilestones":{"nodes":[],"pageInfo":{"hasNextPage":false}}}}}`
	})
	defer server.Close()

	_, err := NewClientWithURL(server.URL, "test").GetProjectMilestones(context.Background(), "https://linear.app/glif/project/roadmap-d05c5c7e8a5c/overview")
	if err != nil {
		t.Fatalf("GetProjectMilestones() error = %v", err)
	}
}

func TestResolveIssueRefFromPullRequest(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    string
		wantErr string
	}{
		{"linked", `{"data":{"attachmentsForURL":{"nodes":[{"issue":null},{"issue":{"identifier":"API-285"}}]}}}`, "API-285", ""},
		{"unlinked", `{"data":{"attachmentsForURL":{"nodes":[]}}}`, "", "no Linear issue is linked"},
		{"ambiguous", `{"data":{"attachmentsForURL":{"nodes":[{"issue":{"identifier":"API-285"}},{"issue":{"identifier":"API-379"}}]}}}`, "", "linked to multiple Linear issues: API-285, API-379"},
		{"truncated", `{"data":{"attachmentsForURL":{"nodes":[{"issue":{"identifier":"API-285"}}],"pageInfo":{"hasNextPage":true}}}}`, "", "too many Linear attachments"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := graphqlTestServer(t, func(req gqlTestRequest) string {
				if req.Variables["url"] != "https://github.com/acme/api/pull/6153" {
					t.Fatalf("url = %v, want canonical URL", req.Variables["url"])
				}
				return test.body
			})
			defer server.Close()

			got, err := NewClientWithURL(server.URL, "test").resolveIssueRef(context.Background(), "https://www.github.com/acme/api/pull/6153/files?w=1")
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("error = %v, want it to contain %q", err, test.wantErr)
				}
				return
			}
			if err != nil || got != test.want {
				t.Fatalf("got %q, %v; want %q, nil", got, err, test.want)
			}
		})
	}
}

func TestResolveCommentRef(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    string
		wantErr string
	}{
		{"unique", `{"data":{"issue":{"comments":{"nodes":[{"id":"b68a4bf5-8a34-473e-af4c-b8892a78a9af"}],"pageInfo":{"hasNextPage":false}}}}}`, "b68a4bf5-8a34-473e-af4c-b8892a78a9af", ""},
		{"ambiguous", `{"data":{"issue":{"comments":{"nodes":[{"id":"b68a4bf5-1111-1111-1111-111111111111"},{"id":"b68a4bf5-2222-2222-2222-222222222222"}],"pageInfo":{"hasNextPage":false}}}}}`, "", "comment prefix \"b68a4bf5\" is ambiguous"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := graphqlTestServer(t, func(req gqlTestRequest) string {
				if req.Variables["id"] != "GTM-580" {
					t.Fatalf("id = %v, want GTM-580", req.Variables["id"])
				}
				return test.body
			})
			defer server.Close()

			got, err := NewClientWithURL(server.URL, "test").resolveCommentRef(context.Background(), "https://linear.app/glif/issue/GTM-580/title#comment-b68a4bf5")
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("error = %v, want it to contain %q", err, test.wantErr)
				}
				return
			}
			if err != nil || got != test.want {
				t.Fatalf("got %q, %v; want %q, nil", got, err, test.want)
			}
		})
	}
}

func TestResolveCommentRefRequiresCommentLink(t *testing.T) {
	client := NewClientWithURL("http://127.0.0.1:0", "test")
	_, err := client.resolveCommentRef(context.Background(), "https://linear.app/glif/issue/GTM-580/title")
	if err == nil || !strings.Contains(err.Error(), "does not point at a comment") {
		t.Fatalf("error = %v, want missing comment error", err)
	}
}

func graphqlTestServer(t *testing.T, respond func(gqlTestRequest) string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request gqlTestRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(respond(request)))
	}))
}
