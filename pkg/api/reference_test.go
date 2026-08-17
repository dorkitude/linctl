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
	cases := []struct {
		name  string
		raw   string
		want  LinearRef
		wantK bool
	}{
		{
			name:  "issue url with slug",
			raw:   "https://linear.app/glif/issue/API-379/block-signups-from-example",
			want:  LinearRef{Kind: LinearRefIssue, Workspace: "glif", ID: "API-379"},
			wantK: true,
		},
		{
			name:  "issue url without slug",
			raw:   "https://linear.app/glif/issue/API-379",
			want:  LinearRef{Kind: LinearRefIssue, Workspace: "glif", ID: "API-379"},
			wantK: true,
		},
		{
			name:  "issue url with comment fragment",
			raw:   "https://linear.app/glif/issue/GTM-580/make-it-generic#comment-b68a4bf5",
			want:  LinearRef{Kind: LinearRefIssue, Workspace: "glif", ID: "GTM-580", CommentID: "b68a4bf5"},
			wantK: true,
		},
		{
			name:  "issue url with commentId query",
			raw:   "https://linear.app/glif/issue/GTM-580?commentId=b68a4bf5-8a34-473e-af4c-b8892a78a9af",
			want:  LinearRef{Kind: LinearRefIssue, Workspace: "glif", ID: "GTM-580", CommentID: "b68a4bf5-8a34-473e-af4c-b8892a78a9af"},
			wantK: true,
		},
		{
			name:  "issue url with agent session fragment",
			raw:   "https://linear.app/glif/issue/API-379/slug#agent-session-56881231",
			want:  LinearRef{Kind: LinearRefIssue, Workspace: "glif", ID: "API-379"},
			wantK: true,
		},
		{
			name:  "scheme-less url",
			raw:   "linear.app/glif/issue/API-379/slug",
			want:  LinearRef{Kind: LinearRefIssue, Workspace: "glif", ID: "API-379"},
			wantK: true,
		},
		{
			name:  "desktop deep link",
			raw:   "linear://linear.app/glif/issue/API-379",
			want:  LinearRef{Kind: LinearRefIssue, Workspace: "glif", ID: "API-379"},
			wantK: true,
		},
		{
			name:  "project url",
			raw:   "https://linear.app/glif/project/benchmarkmaxx-d05c5c7e8a5c/overview",
			want:  LinearRef{Kind: LinearRefProject, Workspace: "glif", ID: "benchmarkmaxx-d05c5c7e8a5c"},
			wantK: true,
		},
		{
			name:  "team url",
			raw:   "https://linear.app/glif/team/API/active",
			want:  LinearRef{Kind: LinearRefTeam, Workspace: "glif", ID: "API"},
			wantK: true,
		},
		{
			name:  "document url",
			raw:   "https://linear.app/glif/document/glif-tgim-sync-507735cb56c1",
			want:  LinearRef{Kind: LinearRefDocument, Workspace: "glif", ID: "glif-tgim-sync-507735cb56c1"},
			wantK: true,
		},
		{
			name:  "review url",
			raw:   "https://linear.app/glif/review/replace-polymorphic-apitokenid-ca4153a35dc0",
			want:  LinearRef{Kind: LinearRefReview, Workspace: "glif", ID: "replace-polymorphic-apitokenid-ca4153a35dc0"},
			wantK: true,
		},
		{name: "bare identifier", raw: "API-379"},
		{name: "uuid", raw: "b68a4bf5-8a34-473e-af4c-b8892a78a9af"},
		{name: "github url", raw: "https://github.com/glifxyz/glif-graph/pull/6153"},
		{name: "empty", raw: ""},
		{name: "linear url with no entity segment", raw: "https://linear.app/glif/settings"},
		{name: "issue segment with nothing after it", raw: "https://linear.app/glif/issue"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ParseLinearURL(tc.raw)
			if ok != tc.wantK {
				t.Fatalf("ParseLinearURL(%q) ok = %v, want %v", tc.raw, ok, tc.wantK)
			}
			if got != tc.want {
				t.Fatalf("ParseLinearURL(%q) = %#v, want %#v", tc.raw, got, tc.want)
			}
		})
	}
}

func TestNormalizeIssueRef(t *testing.T) {
	cases := []struct {
		name    string
		ref     string
		want    string
		wantErr string
	}{
		{name: "identifier passes through", ref: "API-379", want: "API-379"},
		{name: "uuid passes through", ref: "b68a4bf5-8a34-473e-af4c-b8892a78a9af", want: "b68a4bf5-8a34-473e-af4c-b8892a78a9af"},
		{name: "surrounding whitespace trimmed", ref: "  API-379\n", want: "API-379"},
		{name: "issue url", ref: "https://linear.app/glif/issue/API-379/some-slug", want: "API-379"},
		{name: "empty", ref: "  ", wantErr: "issue reference cannot be empty"},
		{name: "project url", ref: "https://linear.app/glif/project/benchmarkmaxx-d05c5c7e8a5c", wantErr: "is a Linear project URL"},
		{name: "review url", ref: "https://linear.app/glif/review/replace-polymorphic-ca4153a35dc0", wantErr: "is a Linear review URL"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeIssueRef(tc.ref)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("NormalizeIssueRef(%q) error = %v, want it to contain %q", tc.ref, err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeIssueRef(%q) returned error: %v", tc.ref, err)
			}
			if got != tc.want {
				t.Fatalf("NormalizeIssueRef(%q) = %q, want %q", tc.ref, got, tc.want)
			}
		})
	}
}

func TestNormalizeProjectAndTeamRefs(t *testing.T) {
	projectID, err := NormalizeProjectRef("https://linear.app/glif/project/benchmarkmaxx-d05c5c7e8a5c/issues")
	if err != nil {
		t.Fatalf("NormalizeProjectRef returned error: %v", err)
	}
	if projectID != "benchmarkmaxx-d05c5c7e8a5c" {
		t.Fatalf("expected slug id, got %q", projectID)
	}

	teamKey, err := NormalizeTeamRef("https://linear.app/glif/team/API/all")
	if err != nil {
		t.Fatalf("NormalizeTeamRef returned error: %v", err)
	}
	if teamKey != "API" {
		t.Fatalf("expected team key API, got %q", teamKey)
	}

	if _, err := NormalizeTeamRef("https://linear.app/glif/issue/API-379"); err == nil {
		t.Fatal("expected an error for an issue URL passed as a team ref")
	}
}

func TestGetIssueAcceptsIssueURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlTestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Variables["id"] != "API-379" {
			t.Fatalf("expected id API-379, got %v", req.Variables["id"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issue":{"id":"i1","identifier":"API-379","title":"Block signups"}}}`))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	issue, err := c.GetIssue(context.Background(), "https://linear.app/glif/issue/API-379/block-signups")
	if err != nil {
		t.Fatalf("GetIssue returned error: %v", err)
	}
	if issue.Identifier != "API-379" {
		t.Fatalf("expected API-379, got %s", issue.Identifier)
	}
}

func TestResolveIssueRefFromGitHubPullRequestURL(t *testing.T) {
	var sawURL interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlTestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(req.Query, "query AttachmentsForURL(") {
			t.Fatalf("expected AttachmentsForURL query, got: %s", req.Query)
		}
		sawURL = req.Variables["url"]
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"attachmentsForURL":{"nodes":[{"issue":null},{"issue":{"identifier":"API-285"}}]}}}`))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	ref, err := c.ResolveIssueRef(context.Background(), "https://github.com/glifxyz/glif-graph/pull/6153?w=1")
	if err != nil {
		t.Fatalf("ResolveIssueRef returned error: %v", err)
	}
	if ref != "API-285" {
		t.Fatalf("expected API-285, got %q", ref)
	}
	if sawURL != "https://github.com/glifxyz/glif-graph/pull/6153" {
		t.Fatalf("expected the canonical PR URL, got %v", sawURL)
	}
}

func TestResolveIssueRefUnlinkedPullRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"attachmentsForURL":{"nodes":[]}}}`))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	_, err := c.ResolveIssueRef(context.Background(), "https://github.com/glifxyz/glif-graph/pull/6153")
	if err == nil || !strings.Contains(err.Error(), "no Linear issue is linked") {
		t.Fatalf("expected an unlinked-PR error, got %v", err)
	}
}

func TestResolveCommentRefFromURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlTestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(req.Query, "query IssueComments(") {
			t.Fatalf("expected IssueComments query, got: %s", req.Query)
		}
		if req.Variables["id"] != "GTM-580" {
			t.Fatalf("expected id GTM-580, got %v", req.Variables["id"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issue":{"comments":{"nodes":[{"id":"11111111-1111-1111-1111-111111111111"},{"id":"b68a4bf5-8a34-473e-af4c-b8892a78a9af"}],"pageInfo":{"hasNextPage":false,"endCursor":""}}}}}`))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	id, err := c.ResolveCommentRef(context.Background(), "https://linear.app/glif/issue/GTM-580/make-it-generic#comment-b68a4bf5")
	if err != nil {
		t.Fatalf("ResolveCommentRef returned error: %v", err)
	}
	if id != "b68a4bf5-8a34-473e-af4c-b8892a78a9af" {
		t.Fatalf("expected the full comment UUID, got %q", id)
	}
}

func TestResolveCommentRefWithoutCommentFragment(t *testing.T) {
	c := NewClientWithURL("http://127.0.0.1:0", "Bearer test")
	_, err := c.ResolveCommentRef(context.Background(), "https://linear.app/glif/issue/GTM-580/make-it-generic")
	if err == nil || !strings.Contains(err.Error(), "does not point at a comment") {
		t.Fatalf("expected a missing-comment error, got %v", err)
	}
}

func TestParseLinearURLMoreVariants(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want LinearRef
	}{
		{
			name: "initiative url",
			raw:  "https://linear.app/glif/initiative/reduce-costs-83e5d6e7a371",
			want: LinearRef{Kind: LinearRefInitiative, Workspace: "glif", ID: "reduce-costs-83e5d6e7a371"},
		},
		{
			name: "trailing slash",
			raw:  "https://linear.app/glif/issue/API-379/",
			want: LinearRef{Kind: LinearRefIssue, Workspace: "glif", ID: "API-379"},
		},
		{
			name: "query string",
			raw:  "https://linear.app/glif/issue/API-379/slug?tab=activity",
			want: LinearRef{Kind: LinearRefIssue, Workspace: "glif", ID: "API-379"},
		},
		{
			name: "www host",
			raw:  "https://www.linear.app/glif/issue/API-379",
			want: LinearRef{Kind: LinearRefIssue, Workspace: "glif", ID: "API-379"},
		},
		{
			name: "lowercase identifier",
			raw:  "https://linear.app/glif/issue/api-379/slug",
			want: LinearRef{Kind: LinearRefIssue, Workspace: "glif", ID: "api-379"},
		},
		{
			name: "team cycle view",
			raw:  "https://linear.app/glif/team/API/cycle/12",
			want: LinearRef{Kind: LinearRefTeam, Workspace: "glif", ID: "API"},
		},
		{
			name: "project sub-tab",
			raw:  "https://linear.app/glif/project/benchmarkmaxx-d05c5c7e8a5c/documents",
			want: LinearRef{Kind: LinearRefProject, Workspace: "glif", ID: "benchmarkmaxx-d05c5c7e8a5c"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ParseLinearURL(tc.raw)
			if !ok {
				t.Fatalf("ParseLinearURL(%q) did not parse", tc.raw)
			}
			if got != tc.want {
				t.Fatalf("ParseLinearURL(%q) = %#v, want %#v", tc.raw, got, tc.want)
			}
		})
	}
}

func TestNormalizeIssueRefRejectsNonIssuePaths(t *testing.T) {
	cases := map[string]string{
		"https://linear.app/glif/issue/new":                               "does not contain an issue identifier",
		"https://linear.app/glif/initiative/reduce-costs-83e5d6e7a371":    "is a Linear initiative URL",
		"https://linear.app/glif/document/glif-tgim-sync-507735cb56c1":    "is a Linear document URL",
		"https://linear.app/glif/review/replace-polymorphic-ca4153a35dc0": "linctl issue search \"replace polymorphic\"",
	}

	for ref, want := range cases {
		if _, err := NormalizeIssueRef(ref); err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("NormalizeIssueRef(%q) error = %v, want it to contain %q", ref, err, want)
		}
	}
}
