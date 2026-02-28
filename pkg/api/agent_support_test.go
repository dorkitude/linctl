package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFindUserByIdentifier_Paginates(t *testing.T) {
	call := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		var req gqlTestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(req.Query, "query Users(") {
			t.Fatalf("expected Users query, got: %s", req.Query)
		}

		w.Header().Set("Content-Type", "application/json")
		if call == 1 {
			_, _ = w.Write([]byte(`{"data":{"users":{"nodes":[{"id":"u1","name":"Alice","displayName":"alice","email":"alice@example.com"}],"pageInfo":{"hasNextPage":true,"endCursor":"cursor-1"}}}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":{"users":{"nodes":[{"id":"u2","name":"Agent Runner","displayName":"agent-runner","email":"agent@example.com"}],"pageInfo":{"hasNextPage":false,"endCursor":""}}}}`))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	user, err := c.FindUserByIdentifier(context.Background(), "agent-runner")
	if err != nil {
		t.Fatalf("FindUserByIdentifier returned error: %v", err)
	}
	if user.ID != "u2" {
		t.Fatalf("expected u2, got %s", user.ID)
	}
	if call < 2 {
		t.Fatalf("expected paginated calls, got %d", call)
	}
}

func TestGetIssueAgentSession(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlTestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(req.Query, "query IssueAgentSession(") {
			t.Fatalf("expected IssueAgentSession query, got: %s", req.Query)
		}
		if req.Variables["id"] != "ENG-80" {
			t.Fatalf("expected id ENG-80, got %v", req.Variables["id"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issue":{"id":"i1","identifier":"ENG-80","title":"Agent task","delegate":{"id":"u2","name":"Agent Runner","displayName":"agent-runner","email":"agent@example.com"},"comments":{"nodes":[{"id":"c1","body":"test","createdAt":"2026-02-28T00:00:00Z","user":{"id":"u1","name":"Alice","displayName":"alice","email":"alice@example.com"},"agentSession":{"id":"s1","status":"active","createdAt":"2026-02-28T00:00:00Z","updatedAt":"2026-02-28T00:01:00Z","appUser":{"id":"u2","name":"Agent Runner","displayName":"agent-runner","email":"agent@example.com"},"activities":{"nodes":[{"id":"a1","createdAt":"2026-02-28T00:00:30Z","ephemeral":false,"content":{"type":"thought","body":"Thinking"}}],"pageInfo":{"hasNextPage":false,"endCursor":""}}}}]}}}}`))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	issue, err := c.GetIssueAgentSession(context.Background(), "ENG-80")
	if err != nil {
		t.Fatalf("GetIssueAgentSession returned error: %v", err)
	}
	if issue.Identifier != "ENG-80" {
		t.Fatalf("expected ENG-80, got %s", issue.Identifier)
	}
	if issue.Delegate == nil || issue.Delegate.DisplayName != "agent-runner" {
		t.Fatalf("expected delegate agent-runner, got %#v", issue.Delegate)
	}
	if issue.Comments == nil || len(issue.Comments.Nodes) != 1 {
		t.Fatalf("expected 1 comment, got %#v", issue.Comments)
	}
	if issue.Comments.Nodes[0].AgentSession == nil {
		t.Fatalf("expected comment agent session")
	}
}

func TestMentionAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlTestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(req.Query, "mutation CreateComment(") {
			t.Fatalf("expected CreateComment mutation, got: %s", req.Query)
		}
		input, ok := req.Variables["input"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected input map, got %#v", req.Variables["input"])
		}
		body, _ := input["body"].(string)
		if !strings.HasPrefix(body, "@agent-runner") {
			t.Fatalf("expected mention prefix, got %q", body)
		}
		if !strings.Contains(body, "Please review") {
			t.Fatalf("expected message content, got %q", body)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"commentCreate":{"comment":{"id":"c-new","body":"@agent-runner\n\nPlease review","createdAt":"2026-02-28T00:00:00Z","updatedAt":"2026-02-28T00:00:00Z","user":{"id":"u1","name":"Alice","email":"alice@example.com"}}}}}`))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	commentID, err := c.MentionAgent(context.Background(), "issue-id", "agent-runner", "Please review")
	if err != nil {
		t.Fatalf("MentionAgent returned error: %v", err)
	}
	if commentID != "c-new" {
		t.Fatalf("expected c-new, got %s", commentID)
	}
}
