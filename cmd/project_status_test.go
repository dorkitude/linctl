package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dorkitude/linctl/pkg/api"
)

func TestResolveProjectStatusByName(t *testing.T) {
	client := projectStatusTestClient(t, `[{"id":"status-shaping","name":"Shaping","type":"backlog","color":"#95a2b3","position":1}]`)

	status, err := resolveProjectStatus(context.Background(), client, "shaping")
	if err != nil {
		t.Fatalf("resolveProjectStatus returned error: %v", err)
	}
	if status.ID != "status-shaping" {
		t.Fatalf("expected Shaping status ID, got %q", status.ID)
	}
}

func TestResolveProjectStatusRejectsAmbiguousType(t *testing.T) {
	client := projectStatusTestClient(t, `[{"id":"status-progress","name":"In Progress","type":"started","color":"#95a2b3","position":1},{"id":"status-review","name":"In Review","type":"started","color":"#95a2b3","position":2}]`)

	_, err := resolveProjectStatus(context.Background(), client, "started")
	if err == nil {
		t.Fatal("expected ambiguous status type to fail")
	}
	if !strings.Contains(err.Error(), "use a named status") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func projectStatusTestClient(t *testing.T, statuses string) *api.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlCommandTestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(req.Query, "query ProjectStatuses") {
			t.Fatalf("expected ProjectStatuses query, got: %s", req.Query)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"projectStatuses":{"nodes":` + statuses + `}}}`))
	}))
	t.Cleanup(srv.Close)
	return api.NewClientWithURL(srv.URL, "Bearer test")
}
