package cmd

import (
    "strings"
    "testing"
)

func TestIssueCreateCmd_ProjectFlag_DefaultsAndHelp(t *testing.T) {
    f := issueCreateCmd.Flags().Lookup("project")
    if f == nil {
        t.Fatalf("expected --project flag on issueCreateCmd")
    }
    if f.DefValue != "" {
        t.Errorf("default value = %q, want empty string", f.DefValue)
    }
    if !strings.Contains(f.Usage, "Project ID to assign issue to") {
        t.Errorf("usage text %q does not contain expected phrase", f.Usage)
    }
}

func TestIssueUpdateCmd_ProjectFlag_DefaultsAndHelp(t *testing.T) {
    f := issueUpdateCmd.Flags().Lookup("project")
    if f == nil {
        t.Fatalf("expected --project flag on issueUpdateCmd")
    }
    if f.DefValue != "" {
        t.Errorf("default value = %q, want empty string", f.DefValue)
    }
    if !strings.Contains(f.Usage, "Project ID to assign issue to") {
        t.Errorf("usage text %q does not contain expected phrase", f.Usage)
    }
    if !strings.Contains(f.Usage, "unassigned") {
        t.Errorf("usage text %q does not mention 'unassigned' handling", f.Usage)
    }
}

