package cmd

import "testing"

func TestIssueCreateCmd_HelpIncludesProjectFlag(t *testing.T) {
    usage := issueCreateCmd.UsageString()
    if !containsAll(usage, []string{"--project", "Project ID to assign issue to"}) {
        t.Fatalf("create usage missing project flag/help text. got:\n%s", usage)
    }
}

func TestIssueUpdateCmd_HelpIncludesProjectFlag(t *testing.T) {
    usage := issueUpdateCmd.UsageString()
    if !containsAll(usage, []string{"--project", "Project ID to assign issue to", "unassigned"}) {
        t.Fatalf("update usage missing project flag/help text. got:\n%s", usage)
    }
}

// containsAll is a tiny helper for substring checks in tests.
func containsAll(hay string, needles []string) bool {
    for _, n := range needles {
        if !contains(hay, n) {
            return false
        }
    }
    return true
}

func contains(s, sub string) bool { return len(s) >= len(sub) && (s == sub || (len(sub) > 0 && (indexOf(s, sub) >= 0))) }

// indexOf is deliberately simple to avoid importing strings in many files.
func indexOf(s, sub string) int {
    // naive search (small strings)
    n, m := len(s), len(sub)
    if m == 0 { return 0 }
    for i := 0; i+m <= n; i++ {
        if s[i:i+m] == sub { return i }
    }
    return -1
}

