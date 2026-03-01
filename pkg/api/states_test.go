package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetTeamStates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlTestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(req.Query, "query TeamStates(") {
			t.Fatalf("expected TeamStates query, got: %s", req.Query)
		}
		if req.Variables["key"] != "ENG" {
			t.Fatalf("expected key ENG, got %v", req.Variables["key"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"team":{"states":{"nodes":[{"id":"s1","name":"Backlog","type":"backlog","color":"#bec2c8","description":"","position":0},{"id":"s2","name":"In Progress","type":"started","color":"#f2c94c","description":"Work in progress","position":1}]}}}}`))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	states, err := c.GetTeamStates(context.Background(), "ENG")
	if err != nil {
		t.Fatalf("GetTeamStates returned error: %v", err)
	}
	if len(states) != 2 {
		t.Fatalf("expected 2 states, got %d", len(states))
	}
	if states[0].Name != "Backlog" {
		t.Fatalf("expected first state Backlog, got %s", states[0].Name)
	}
	if states[1].Type != "started" {
		t.Fatalf("expected second state type started, got %s", states[1].Type)
	}
}

func TestUpdateWorkflowState(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlTestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(req.Query, "mutation WorkflowStateUpdate(") {
			t.Fatalf("expected WorkflowStateUpdate mutation, got: %s", req.Query)
		}
		if req.Variables["id"] != "s1" {
			t.Fatalf("expected id s1, got %v", req.Variables["id"])
		}

		input, ok := req.Variables["input"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected input map, got %#v", req.Variables["input"])
		}
		if input["name"] != "Ready" {
			t.Fatalf("expected name Ready, got %v", input["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"workflowStateUpdate":{"success":true,"workflowState":{"id":"s1","name":"Ready","type":"backlog","color":"#bec2c8","description":"","position":0}}}}`))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	state, err := c.UpdateWorkflowState(context.Background(), "s1", map[string]interface{}{
		"name": "Ready",
	})
	if err != nil {
		t.Fatalf("UpdateWorkflowState returned error: %v", err)
	}
	if state.Name != "Ready" {
		t.Fatalf("expected Ready, got %s", state.Name)
	}
}

func TestUpdateWorkflowStateFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"workflowStateUpdate":{"success":false,"workflowState":null}}}`))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	_, err := c.UpdateWorkflowState(context.Background(), "s1", map[string]interface{}{
		"name": "Bad",
	})
	if err == nil {
		t.Fatal("expected error for unsuccessful update")
	}
	if !strings.Contains(err.Error(), "failed to update workflow state") {
		t.Fatalf("unexpected error: %v", err)
	}
}
