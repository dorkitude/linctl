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

type mockProjectClient struct{
    created *api.Project
    archived bool
}

func (m *mockProjectClient) GetTeam(ctx context.Context, key string) (*api.Team, error) {
    return &api.Team{ID: "team-1", Key: key, Name: "Team-"+key}, nil
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

func withInjectedProjectClient(t *testing.T, mc *mockProjectClient, fn func()) {
    t.Helper()
    oldNew := newAPIClient
    oldAuth := getAuthHeader
    newAPIClient = func(_ string) projectAPI { return mc }
    getAuthHeader = func() (string, error) { return "Bearer test", nil }
    defer func(){ newAPIClient = oldNew; getAuthHeader = oldAuth }()
    fn()
}

func captureStdout(t *testing.T, fn func()) string {
    t.Helper()
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w
    defer func(){ os.Stdout = old }()
    fn()
    _ = w.Close()
    var buf bytes.Buffer
    _, _ = buf.ReadFrom(r)
    return buf.String()
}

func TestProjectCreate_Plaintext_Output(t *testing.T) {
    mc := &mockProjectClient{}
    withInjectedProjectClient(t, mc, func(){
        viper.Set("plaintext", true)
        viper.Set("json", false)
        // Set flags directly on the command and call Run
        _ = projectCreateCmd.Flags().Set("name", "Alpha")
        _ = projectCreateCmd.Flags().Set("team", "ENG")
        _ = projectCreateCmd.Flags().Set("state", "planned")
        _ = projectCreateCmd.Flags().Set("target-date", "2024-12-31")
        out := captureStdout(t, func(){ projectCreateCmd.Run(projectCreateCmd, nil) })
        if !contains(out, "# Project Created") || !contains(out, "**Name**: Alpha") {
            t.Fatalf("unexpected output:\n%s", out)
        }
    })
}

func TestProjectArchive_Plaintext_IncludesName(t *testing.T) {
    mc := &mockProjectClient{}
    withInjectedProjectClient(t, mc, func(){
        viper.Set("plaintext", true)
        viper.Set("json", false)
        out := captureStdout(t, func(){ projectArchiveCmd.Run(projectArchiveCmd, []string{"p1"}) })
        if !contains(out, "# Project Archived") || !contains(out, "**Name**: Alpha") {
            t.Fatalf("unexpected output:\n%s", out)
        }
    })
}
