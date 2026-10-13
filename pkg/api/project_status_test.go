package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetProjectStatuses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlTestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(req.Query, "query ProjectStatuses") {
			t.Fatalf("expected ProjectStatuses query, got: %s", req.Query)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"projectStatuses":{"nodes":[{"id":"status-shaping","name":"Shaping","type":"backlog","color":"#95a2b3","position":1}]}}}`))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	statuses, err := c.GetProjectStatuses(context.Background())
	if err != nil {
		t.Fatalf("GetProjectStatuses returned error: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("expected one status, got %d", len(statuses))
	}
	if got := statuses[0]; got.ID != "status-shaping" || got.Name != "Shaping" || got.Type != "backlog" {
		t.Fatalf("unexpected status: %+v", got)
	}
}

func TestGetProjectsSelectsSlugIDAndStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlTestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		for _, field := range []string{"slugId", "status {"} {
			if !strings.Contains(req.Query, field) {
				t.Fatalf("expected Projects query to select %q, got: %s", field, req.Query)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"projects":{"nodes":[{"id":"project-1","slugId":"abc123","name":"Roof","status":{"id":"status-shaping","name":"Shaping","type":"backlog"}}],"pageInfo":{"hasNextPage":false}}}}`))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	projects, err := c.GetProjects(context.Background(), nil, 10, "", "")
	if err != nil {
		t.Fatalf("GetProjects returned error: %v", err)
	}
	if len(projects.Nodes) != 1 {
		t.Fatalf("expected one project, got %d", len(projects.Nodes))
	}
	if got := projects.Nodes[0]; got.SlugId != "abc123" || got.Status == nil || got.Status.Name != "Shaping" {
		t.Fatalf("unexpected project: %+v", got)
	}
}
