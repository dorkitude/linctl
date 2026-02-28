package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type gqlTestRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

func TestGetComment(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlTestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(req.Query, "query Comment(") {
			t.Fatalf("expected Comment query, got: %s", req.Query)
		}
		if req.Variables["id"] != "comment-123" {
			t.Fatalf("expected id variable comment-123, got %v", req.Variables["id"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"comment":{"id":"comment-123","body":"hello","createdAt":"2026-02-28T00:00:00Z","updatedAt":"2026-02-28T00:01:00Z","editedAt":"2026-02-28T00:02:00Z","user":{"id":"u1","name":"Alice","email":"alice@example.com"},"parent":{"id":"comment-1"}}}}`))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	comment, err := c.GetComment(context.Background(), "comment-123")
	if err != nil {
		t.Fatalf("GetComment returned error: %v", err)
	}

	if comment.ID != "comment-123" {
		t.Fatalf("expected comment ID comment-123, got %s", comment.ID)
	}
	if comment.Body != "hello" {
		t.Fatalf("expected body hello, got %s", comment.Body)
	}
	if comment.User == nil || comment.User.Email != "alice@example.com" {
		t.Fatalf("expected user email alice@example.com, got %#v", comment.User)
	}
	if comment.Parent == nil || comment.Parent.ID != "comment-1" {
		t.Fatalf("expected parent comment-1, got %#v", comment.Parent)
	}
}

func TestUpdateComment(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlTestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(req.Query, "mutation UpdateComment(") {
			t.Fatalf("expected UpdateComment mutation, got: %s", req.Query)
		}
		if req.Variables["id"] != "comment-456" {
			t.Fatalf("expected id variable comment-456, got %v", req.Variables["id"])
		}

		input, ok := req.Variables["input"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected input map, got %#v", req.Variables["input"])
		}
		if input["body"] != "updated body" {
			t.Fatalf("expected body updated body, got %v", input["body"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"commentUpdate":{"success":true,"comment":{"id":"comment-456","body":"updated body","createdAt":"2026-02-28T00:00:00Z","updatedAt":"2026-02-28T00:03:00Z","editedAt":"2026-02-28T00:03:00Z","user":{"id":"u2","name":"Bob","email":"bob@example.com"},"parent":null}}}}`))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	comment, err := c.UpdateComment(context.Background(), "comment-456", "updated body")
	if err != nil {
		t.Fatalf("UpdateComment returned error: %v", err)
	}

	if comment.ID != "comment-456" {
		t.Fatalf("expected comment ID comment-456, got %s", comment.ID)
	}
	if comment.Body != "updated body" {
		t.Fatalf("expected body updated body, got %s", comment.Body)
	}
}

func TestUpdateCommentFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"commentUpdate":{"success":false,"comment":{"id":"comment-456","body":"updated body","createdAt":"2026-02-28T00:00:00Z","updatedAt":"2026-02-28T00:03:00Z","editedAt":"2026-02-28T00:03:00Z","user":{"id":"u2","name":"Bob","email":"bob@example.com"},"parent":null}}}}`))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	_, err := c.UpdateComment(context.Background(), "comment-456", "updated body")
	if err == nil {
		t.Fatal("expected UpdateComment error when success=false")
	}
	if !strings.Contains(err.Error(), "failed to update comment") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteComment(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlTestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(req.Query, "mutation DeleteComment(") {
			t.Fatalf("expected DeleteComment mutation, got: %s", req.Query)
		}
		if req.Variables["id"] != "comment-789" {
			t.Fatalf("expected id variable comment-789, got %v", req.Variables["id"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"commentDelete":{"success":true}}}`))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	if err := c.DeleteComment(context.Background(), "comment-789"); err != nil {
		t.Fatalf("DeleteComment returned error: %v", err)
	}
}

func TestDeleteCommentFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"commentDelete":{"success":false}}}`))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	err := c.DeleteComment(context.Background(), "comment-789")
	if err == nil {
		t.Fatal("expected DeleteComment error when success=false")
	}
	if !strings.Contains(err.Error(), "failed to delete comment") {
		t.Fatalf("unexpected error: %v", err)
	}
}
