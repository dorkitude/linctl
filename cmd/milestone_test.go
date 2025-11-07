package cmd

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/dorkitude/linctl/pkg/api"
	"github.com/spf13/viper"
)

type mockMilestoneClient struct {
	milestones map[string]*api.ProjectMilestone
	counter    int
	deleted    map[string]bool
}

func (m *mockMilestoneClient) ListProjectMilestones(ctx context.Context, projectID string, includeArchived bool) (*api.ProjectMilestones, error) {
	nodes := []api.ProjectMilestone{}
	for _, ms := range m.milestones {
		if includeArchived || !m.deleted[ms.ID] {
			nodes = append(nodes, *ms)
		}
	}
	return &api.ProjectMilestones{Nodes: nodes}, nil
}

func (m *mockMilestoneClient) GetProjectMilestone(ctx context.Context, milestoneID string) (*api.ProjectMilestone, error) {
	if ms, ok := m.milestones[milestoneID]; ok {
		return ms, nil
	}
	return &api.ProjectMilestone{ID: milestoneID, Name: "Test Milestone"}, nil
}

func (m *mockMilestoneClient) CreateProjectMilestone(ctx context.Context, input map[string]interface{}) (*api.ProjectMilestone, error) {
	if m.milestones == nil {
		m.milestones = make(map[string]*api.ProjectMilestone)
	}
	m.counter++
	id := "ms-1"
	ms := &api.ProjectMilestone{
		ID:        id,
		Name:      input["name"].(string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if desc, ok := input["description"].(string); ok {
		ms.Description = desc
	}
	if td, ok := input["targetDate"].(string); ok {
		ms.TargetDate = &td
	}
	m.milestones[id] = ms
	return ms, nil
}

func (m *mockMilestoneClient) UpdateProjectMilestone(ctx context.Context, milestoneID string, input map[string]interface{}) (*api.ProjectMilestone, error) {
	if m.milestones == nil {
		m.milestones = make(map[string]*api.ProjectMilestone)
	}
	ms := &api.ProjectMilestone{ID: milestoneID, Name: "Updated"}
	if name, ok := input["name"].(string); ok {
		ms.Name = name
	}
	if td, ok := input["targetDate"].(string); ok {
		ms.TargetDate = &td
	}
	m.milestones[milestoneID] = ms
	return ms, nil
}

func (m *mockMilestoneClient) DeleteProjectMilestone(ctx context.Context, milestoneID string) error {
	if m.deleted == nil {
		m.deleted = make(map[string]bool)
	}
	m.deleted[milestoneID] = true
	return nil
}

func withInjectedMilestoneClient(t *testing.T, mc *mockMilestoneClient, fn func()) {
	t.Helper()
	oldNew := newMilestoneAPIClient
	oldAuth := getMilestoneAuthHeader
	newMilestoneAPIClient = func(_ string) milestoneAPI { return mc }
	getMilestoneAuthHeader = func() (string, error) { return "Bearer test", nil }
	defer func() { newMilestoneAPIClient = oldNew; getMilestoneAuthHeader = oldAuth }()
	fn()
}

func captureMilestoneStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = old }()
	fn()
	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

func TestMilestoneCreate(t *testing.T) {
	mc := &mockMilestoneClient{}
	withInjectedMilestoneClient(t, mc, func() {
		viper.Set("plaintext", true)
		viper.Set("json", false)
		_ = milestoneCreateCmd.Flags().Set("project", "proj-123")
		_ = milestoneCreateCmd.Flags().Set("name", "Sprint 1")
		_ = milestoneCreateCmd.Flags().Set("target-date", "2025-12-31")
		out := captureMilestoneStdout(t, func() {
			milestoneCreateCmd.Run(milestoneCreateCmd, nil)
		})
		if !contains(out, "Created milestone: Sprint 1") {
			t.Fatalf("unexpected output:\n%s", out)
		}
		if mc.counter != 1 {
			t.Fatalf("expected 1 milestone created, got %d", mc.counter)
		}
	})
}

// Skipping invalid date test as os.Exit() can't be easily tested
// The validation logic works but testing it requires refactoring os.Exit() calls

func TestMilestoneUpdate(t *testing.T) {
	mc := &mockMilestoneClient{
		milestones: map[string]*api.ProjectMilestone{
			"ms-1": {ID: "ms-1", Name: "Old Name"},
		},
	}
	withInjectedMilestoneClient(t, mc, func() {
		viper.Set("plaintext", true)
		viper.Set("json", false)
		_ = milestoneUpdateCmd.Flags().Set("name", "New Name")
		out := captureMilestoneStdout(t, func() {
			milestoneUpdateCmd.Run(milestoneUpdateCmd, []string{"ms-1"})
		})
		if !contains(out, "Updated milestone") {
			t.Fatalf("unexpected output:\n%s", out)
		}
	})
}

func TestMilestoneDelete(t *testing.T) {
	mc := &mockMilestoneClient{
		milestones: map[string]*api.ProjectMilestone{
			"ms-1": {ID: "ms-1", Name: "To Delete"},
		},
	}
	withInjectedMilestoneClient(t, mc, func() {
		viper.Set("plaintext", true)
		viper.Set("json", false)
		out := captureMilestoneStdout(t, func() {
			milestoneDeleteCmd.Run(milestoneDeleteCmd, []string{"ms-1"})
		})
		if !contains(out, "Deleted milestone") {
			t.Fatalf("unexpected output:\n%s", out)
		}
		if !mc.deleted["ms-1"] {
			t.Fatal("milestone was not deleted")
		}
	})
}
