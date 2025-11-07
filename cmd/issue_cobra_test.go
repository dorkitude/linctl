package cmd

import "testing"

// These tests exercise Cobra flag parsing on the real command objects
// without invoking the Run functions (no network/API side-effects).

func TestIssueCreateCmd_ProjectFlag_Parsing(t *testing.T) {
    // Ensure the flag exists
    f := issueCreateCmd.Flags().Lookup("project")
    if f == nil {
        t.Fatalf("expected --project flag on issueCreateCmd")
    }

    // Parse and read back
    uuid := "123e4567-e89b-12d3-a456-426614174000"
    if err := issueCreateCmd.Flags().Set("project", uuid); err != nil {
        t.Fatalf("failed to set project flag: %v", err)
    }
    got, err := issueCreateCmd.Flags().GetString("project")
    if err != nil {
        t.Fatalf("failed to get project flag: %v", err)
    }
    if got != uuid {
        t.Errorf("project flag parsed as %q, want %q", got, uuid)
    }
}

func TestIssueUpdateCmd_ProjectFlag_Unassigned(t *testing.T) {
    // Ensure the flag exists
    f := issueUpdateCmd.Flags().Lookup("project")
    if f == nil {
        t.Fatalf("expected --project flag on issueUpdateCmd")
    }

    if err := issueUpdateCmd.Flags().Set("project", "unassigned"); err != nil {
        t.Fatalf("failed to set project flag: %v", err)
    }
    got, err := issueUpdateCmd.Flags().GetString("project")
    if err != nil {
        t.Fatalf("failed to get project flag: %v", err)
    }
    if got != "unassigned" {
        t.Errorf("project flag parsed as %q, want %q", got, "unassigned")
    }

    // Check helper integration contract
    if val, ok, err := buildProjectInput(got); err != nil || !ok || val != nil {
        t.Errorf("buildProjectInput('unassigned') => (%v,%v,%v), want (nil,true,nil)", val, ok, err)
    }
}

