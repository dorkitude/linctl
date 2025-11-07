package cmd

import "testing"

func TestConstructProjectURL(t *testing.T) {
    // Happy path: workspace URL should be preserved, slug replaced by ID
    id := "1234"
    original := "https://linear.app/acme/project/some-project-slug"
    want := "https://linear.app/acme/project/1234"
    got := constructProjectURL(id, original)
    if got != want {
        t.Fatalf("constructProjectURL mismatch: got %q want %q", got, want)
    }

    // Empty original → empty result
    if s := constructProjectURL("abc", ""); s != "" {
        t.Fatalf("expected empty for empty original URL, got %q", s)
    }
}

