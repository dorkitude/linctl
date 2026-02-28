package cmd

import (
	"testing"
	"time"

	"github.com/dorkitude/linctl/pkg/api"
)

func TestLatestAgentSession(t *testing.T) {
	t1 := time.Date(2026, 2, 28, 10, 0, 0, 0, time.UTC)
	t2 := t1.Add(2 * time.Minute)

	issue := &api.Issue{
		Comments: &api.Comments{
			Nodes: []api.Comment{
				{ID: "c1", AgentSession: &api.AgentSession{ID: "s1", UpdatedAt: t1}},
				{ID: "c2", AgentSession: &api.AgentSession{ID: "s2", UpdatedAt: t2}},
			},
		},
	}

	got := latestAgentSession(issue)
	if got == nil || got.ID != "s2" {
		t.Fatalf("expected latest session s2, got %#v", got)
	}
}

func TestPickAgentHandle(t *testing.T) {
	issue := &api.Issue{
		Delegate: &api.User{DisplayName: "agent-delegate"},
		Comments: &api.Comments{
			Nodes: []api.Comment{
				{AgentSession: &api.AgentSession{AppUser: &api.User{DisplayName: "agent-session"}}},
			},
		},
	}

	got := pickAgentHandle(issue)
	if got != "agent-delegate" {
		t.Fatalf("expected delegate handle, got %q", got)
	}
}

func TestActivitySummary(t *testing.T) {
	activity := api.AgentActivity{
		Content: map[string]interface{}{
			"type":      "action",
			"action":    "Bash",
			"parameter": "go test ./...",
		},
	}

	typ, body := activitySummary(activity)
	if typ != "action" {
		t.Fatalf("expected action type, got %q", typ)
	}
	if body != "Bash go test ./..." {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestAgentMentionArgs(t *testing.T) {
	if err := agentMentionCmd.Args(agentMentionCmd, []string{"ENG-1"}); err == nil {
		t.Fatal("expected error for missing message args")
	}
	if err := agentMentionCmd.Args(agentMentionCmd, []string{"ENG-1", "hello"}); err != nil {
		t.Fatalf("unexpected args error: %v", err)
	}
}
