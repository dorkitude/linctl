package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/dorkitude/linctl/pkg/api"
	"github.com/spf13/viper"
)

type mockProjectClient struct {
	created        *api.Project
	archived       bool
	projectUpdates map[string]*api.ProjectUpdate
	updateCounter  int
}

func (m *mockProjectClient) GetTeam(ctx context.Context, key string) (*api.Team, error) {
	return &api.Team{ID: "team-1", Key: key, Name: "Team-" + key}, nil
}

func (m *mockProjectClient) GetProjects(ctx context.Context, filter map[string]interface{}, first int, after string, orderBy string) (*api.Projects, error) {
	return &api.Projects{}, nil
}

func (m *mockProjectClient) CreateProject(ctx context.Context, input map[string]interface{}) (*api.Project, error) {
	name, _ := input["name"].(string)
	m.created = &api.Project{ID: "p1", Name: name, State: fmt.Sprint(input["state"])}
	return m.created, nil
}

func (m *mockProjectClient) ArchiveProject(ctx context.Context, id string) (bool, error) {
	m.archived = true
	return true, nil
}

func (m *mockProjectClient) UpdateProject(ctx context.Context, id string, input map[string]interface{}) (*api.Project, error) {
	project := &api.Project{ID: id, Name: "Alpha"}
	if name, ok := input["name"].(string); ok {
		project.Name = name
	}
	if state, ok := input["state"].(string); ok {
		project.State = state
	}
	if priority, ok := input["priority"].(int); ok {
		project.Priority = priority
	}
	return project, nil
}

func (m *mockProjectClient) GetProject(ctx context.Context, id string) (*api.Project, error) {
	return &api.Project{ID: id, Name: "Alpha"}, nil
}

func (m *mockProjectClient) CreateProjectUpdate(ctx context.Context, input map[string]interface{}) (*api.ProjectUpdate, error) {
	if m.projectUpdates == nil {
		m.projectUpdates = make(map[string]*api.ProjectUpdate)
	}
	m.updateCounter++
	id := fmt.Sprintf("update-%d", m.updateCounter)
	update := &api.ProjectUpdate{
		ID:   id,
		Body: input["body"].(string),
	}
	if health, ok := input["health"].(string); ok {
		update.Health = health
	}
	m.projectUpdates[id] = update
	return update, nil
}

func (m *mockProjectClient) ListProjectUpdates(ctx context.Context, projectID string) (*api.ProjectUpdates, error) {
	updates := []api.ProjectUpdate{}
	for _, u := range m.projectUpdates {
		updates = append(updates, *u)
	}
	return &api.ProjectUpdates{Nodes: updates}, nil
}

func (m *mockProjectClient) GetProjectUpdate(ctx context.Context, updateID string) (*api.ProjectUpdate, error) {
	if m.projectUpdates == nil {
		m.projectUpdates = make(map[string]*api.ProjectUpdate)
	}
	if update, ok := m.projectUpdates[updateID]; ok {
		return update, nil
	}
	return &api.ProjectUpdate{ID: updateID, Body: "Test update body"}, nil
}

func withInjectedProjectClient(t *testing.T, mc *mockProjectClient, fn func()) {
	t.Helper()
	oldNew := newAPIClient
	oldAuth := getAuthHeader
	newAPIClient = func(_ string) projectAPI { return mc }
	getAuthHeader = func() (string, error) { return "Bearer test", nil }
	defer func() { newAPIClient = oldNew; getAuthHeader = oldAuth }()
	fn()
}

func captureStdout(t *testing.T, fn func()) string {
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

func TestProjectCreate_Plaintext_Output(t *testing.T) {
	mc := &mockProjectClient{}
	withInjectedProjectClient(t, mc, func() {
		viper.Set("plaintext", true)
		viper.Set("json", false)
		// Set flags directly on the command and call Run
		_ = projectCreateCmd.Flags().Set("name", "Alpha")
		_ = projectCreateCmd.Flags().Set("team", "ENG")
		_ = projectCreateCmd.Flags().Set("state", "planned")
		_ = projectCreateCmd.Flags().Set("target-date", "2024-12-31")
		out := captureStdout(t, func() { projectCreateCmd.Run(projectCreateCmd, nil) })
		if !contains(out, "# Project Created") || !contains(out, "**Name**: Alpha") {
			t.Fatalf("unexpected output:\n%s", out)
		}
	})
}

func TestProjectArchive_Plaintext_IncludesName(t *testing.T) {
	mc := &mockProjectClient{}
	withInjectedProjectClient(t, mc, func() {
		viper.Set("plaintext", true)
		viper.Set("json", false)
		out := captureStdout(t, func() { projectArchiveCmd.Run(projectArchiveCmd, []string{"p1"}) })
		if !contains(out, "# Project Archived") || !contains(out, "**Name**: Alpha") {
			t.Fatalf("unexpected output:\n%s", out)
		}
	})
}

func TestProjectUpdatePostCreate(t *testing.T) {
	mc := &mockProjectClient{}
	withInjectedProjectClient(t, mc, func() {
		viper.Set("plaintext", true)
		viper.Set("json", false)
		_ = projectUpdatePostCreateCmd.Flags().Set("body", "Monthly progress update")
		_ = projectUpdatePostCreateCmd.Flags().Set("health", "onTrack")
		out := captureStdout(t, func() {
			projectUpdatePostCreateCmd.Run(projectUpdatePostCreateCmd, []string{"proj-123"})
		})
		if !contains(out, "Project update created successfully") {
			t.Fatalf("unexpected output:\n%s", out)
		}
		if mc.updateCounter != 1 {
			t.Fatalf("expected 1 update created, got %d", mc.updateCounter)
		}
	})
}

// Skipping validation error tests as os.Exit() can't be easily tested
// The validation logic works but testing it requires refactoring os.Exit() calls
