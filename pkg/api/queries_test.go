package api

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
)

// helper to create a mock GraphQL server with programmable handler
func newMockGraphQLServer(t *testing.T, handler func(query string, w http.ResponseWriter)) *httptest.Server {
    t.Helper()
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var body struct {
            Query string `json:"query"`
        }
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
            t.Fatalf("failed to decode request: %v", err)
            return
        }
        handler(body.Query, w)
    }))
}

func TestGetTeamByKey(t *testing.T) {
    srv := newMockGraphQLServer(t, func(query string, w http.ResponseWriter) {
        if strings.Contains(query, "teams(") {
            io := map[string]any{
                "data": map[string]any{
                    "teams": map[string]any{
                        "nodes": []any{
                            map[string]any{"id": "team-1", "key": "ENG", "name": "Engineering", "issueCount": 42},
                        },
                    },
                },
            }
            _ = json.NewEncoder(w).Encode(io)
            return
        }
        // default empty
        _ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"teams": map[string]any{"nodes": []any{}}}})
    })
    defer srv.Close()

    c := NewClientWithURL(srv.URL, "Bearer test")
    got, err := c.GetTeam(context.Background(), "ENG")
    if err != nil {
        t.Fatalf("GetTeam returned error: %v", err)
    }
    if got == nil || got.Key != "ENG" || got.Name != "Engineering" {
        t.Fatalf("unexpected team: %+v", got)
    }
}

func TestGetTeamFallbackByID(t *testing.T) {
    call := 0
    srv := newMockGraphQLServer(t, func(query string, w http.ResponseWriter) {
        call++
        if strings.Contains(query, "teams(") {
            // first call returns no teams by key
            _ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"teams": map[string]any{"nodes": []any{}}}})
            return
        }
        if strings.Contains(query, "team(") {
            // second call returns direct lookup by id
            _ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"team": map[string]any{"id": "team-1", "key": "ENG", "name": "Engineering", "issueCount": 10}}})
            return
        }
        t.Fatalf("unexpected query: %s", query)
    })
    defer srv.Close()

    c := NewClientWithURL(srv.URL, "Bearer test")
    got, err := c.GetTeam(context.Background(), "team-1")
    if err != nil {
        t.Fatalf("GetTeam returned error: %v", err)
    }
    if got == nil || got.ID != "team-1" || got.Key != "ENG" {
        t.Fatalf("unexpected team: %+v", got)
    }
    if call < 2 {
        t.Fatalf("expected at least 2 calls, got %d", call)
    }
}

func TestCreateArchiveAndGetProject(t *testing.T) {
    srv := newMockGraphQLServer(t, func(query string, w http.ResponseWriter) {
        switch {
        case strings.Contains(query, "mutation CreateProject"):
            _ = json.NewEncoder(w).Encode(map[string]any{
                "data": map[string]any{
                    "projectCreate": map[string]any{
                        "success": true,
                        "project": map[string]any{"id": "p1", "name": "Alpha"},
                    },
                },
            })
        case strings.Contains(query, "mutation ArchiveProject"):
            _ = json.NewEncoder(w).Encode(map[string]any{
                "data": map[string]any{
                    "projectArchive": map[string]any{
                        "success": true,
                        "project": map[string]any{"id": "p1", "name": "Alpha"},
                    },
                },
            })
        case strings.Contains(query, "query Project("):
            _ = json.NewEncoder(w).Encode(map[string]any{
                "data": map[string]any{
                    "project": map[string]any{"id": "p1", "name": "Alpha"},
                },
            })
        default:
            t.Fatalf("unexpected query: %s", query)
        }
    })
    defer srv.Close()

    c := NewClientWithURL(srv.URL, "Bearer test")
    proj, err := c.CreateProject(context.Background(), map[string]any{"name": "Alpha"})
    if err != nil {
        t.Fatalf("CreateProject error: %v", err)
    }
    if proj == nil || proj.ID != "p1" || proj.Name != "Alpha" {
        t.Fatalf("unexpected project: %+v", proj)
    }

    ok, err := c.ArchiveProject(context.Background(), "p1")
    if err != nil || !ok {
        t.Fatalf("ArchiveProject error/ok: %v %v", err, ok)
    }

    got, err := c.GetProject(context.Background(), "p1")
    if err != nil {
        t.Fatalf("GetProject error: %v", err)
    }
    if got == nil || got.Name != "Alpha" {
        t.Fatalf("unexpected GetProject: %+v", got)
    }
}

