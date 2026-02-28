package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetTeamLabels(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlTestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(req.Query, "query TeamLabels(") {
			t.Fatalf("expected TeamLabels query, got: %s", req.Query)
		}
		if req.Variables["key"] != "ENG" {
			t.Fatalf("expected key ENG, got %v", req.Variables["key"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"team":{"labels":{"nodes":[{"id":"l1","name":"bug","color":"#ff0000","description":"Bug label","parent":null,"team":{"id":"t1","key":"ENG","name":"Engineering"}}],"pageInfo":{"hasNextPage":false,"endCursor":""}}}}}`))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	labels, err := c.GetTeamLabels(context.Background(), "ENG")
	if err != nil {
		t.Fatalf("GetTeamLabels returned error: %v", err)
	}
	if len(labels) != 1 {
		t.Fatalf("expected 1 label, got %d", len(labels))
	}
	if labels[0].Name != "bug" {
		t.Fatalf("expected label bug, got %s", labels[0].Name)
	}
	if labels[0].Team == nil || labels[0].Team.Key != "ENG" {
		t.Fatalf("expected team ENG, got %#v", labels[0].Team)
	}
}

func TestGetLabel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlTestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(req.Query, "query Label(") {
			t.Fatalf("expected Label query, got: %s", req.Query)
		}
		if req.Variables["id"] != "label-1" {
			t.Fatalf("expected id label-1, got %v", req.Variables["id"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issueLabel":{"id":"label-1","name":"bug","color":"#ff0000","description":"Bug label","parent":null,"team":{"id":"t1","key":"ENG","name":"Engineering"}}}}`))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	label, err := c.GetLabel(context.Background(), "label-1")
	if err != nil {
		t.Fatalf("GetLabel returned error: %v", err)
	}
	if label.ID != "label-1" {
		t.Fatalf("expected label-1, got %s", label.ID)
	}
}

func TestCreateLabel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlTestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(req.Query, "mutation CreateLabel(") {
			t.Fatalf("expected CreateLabel mutation, got: %s", req.Query)
		}

		input, ok := req.Variables["input"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected input map, got %#v", req.Variables["input"])
		}
		if input["name"] != "bug" {
			t.Fatalf("expected name bug, got %v", input["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issueLabelCreate":{"success":true,"issueLabel":{"id":"label-1","name":"bug","color":"#ff0000","description":"Bug label","parent":null,"team":{"id":"t1","key":"ENG","name":"Engineering"}}}}}`))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	label, err := c.CreateLabel(context.Background(), map[string]interface{}{
		"name":   "bug",
		"teamId": "t1",
		"color":  "#ff0000",
	})
	if err != nil {
		t.Fatalf("CreateLabel returned error: %v", err)
	}
	if label.Name != "bug" {
		t.Fatalf("expected bug, got %s", label.Name)
	}
}

func TestUpdateLabel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlTestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(req.Query, "mutation UpdateLabel(") {
			t.Fatalf("expected UpdateLabel mutation, got: %s", req.Query)
		}
		if req.Variables["id"] != "label-1" {
			t.Fatalf("expected id label-1, got %v", req.Variables["id"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issueLabelUpdate":{"success":true,"issueLabel":{"id":"label-1","name":"critical bug","color":"#ff0000","description":"Bug label","parent":null,"team":{"id":"t1","key":"ENG","name":"Engineering"}}}}}`))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	label, err := c.UpdateLabel(context.Background(), "label-1", map[string]interface{}{
		"name": "critical bug",
	})
	if err != nil {
		t.Fatalf("UpdateLabel returned error: %v", err)
	}
	if label.Name != "critical bug" {
		t.Fatalf("expected critical bug, got %s", label.Name)
	}
}

func TestDeleteLabel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req gqlTestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !strings.Contains(req.Query, "mutation DeleteLabel(") {
			t.Fatalf("expected DeleteLabel mutation, got: %s", req.Query)
		}
		if req.Variables["id"] != "label-1" {
			t.Fatalf("expected id label-1, got %v", req.Variables["id"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issueLabelDelete":{"success":true}}}`))
	}))
	defer srv.Close()

	c := NewClientWithURL(srv.URL, "Bearer test")
	if err := c.DeleteLabel(context.Background(), "label-1"); err != nil {
		t.Fatalf("DeleteLabel returned error: %v", err)
	}
}
