package cmd

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/dorkitude/linctl/pkg/api"
	"github.com/spf13/viper"
)

func resetIssueCreateStateFlags(t *testing.T) {
	t.Helper()
	for _, name := range []string{"title", "team", "state"} {
		flag := issueCreateCmd.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("missing flag %q", name)
		}
		if err := flag.Value.Set(flag.DefValue); err != nil {
			t.Fatalf("reset flag %q: %v", name, err)
		}
		flag.Changed = false
	}
}

func TestIssueCreateCmdStateResolvesToStateID(t *testing.T) {
	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()
	t.Setenv("LINCTL_API_KEY", "test-key")
	viper.Set("plaintext", true)
	viper.Set("json", false)
	resetIssueCreateStateFlags(t)

	var sawStateID string
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var gqlReq gqlCommandTestRequest
		if err := json.NewDecoder(req.Body).Decode(&gqlReq); err != nil {
			t.Fatalf("decode GraphQL request: %v", err)
		}

		switch {
		case strings.Contains(gqlReq.Query, "query Team("):
			body := `{"data":{"team":{"id":"team-1","key":"ENG","name":"Engineering","description":"","private":false,"issueCount":0}}}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		case strings.Contains(gqlReq.Query, "query TeamStates("):
			body := `{"data":{"team":{"states":{"nodes":[{"id":"state-2","name":"In Progress","type":"started","color":"#00ff00","description":"","position":2}]}}}}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		case strings.Contains(gqlReq.Query, "mutation CreateIssue("):
			input, ok := gqlReq.Variables["input"].(map[string]interface{})
			if !ok {
				t.Fatalf("expected input map, got %#v", gqlReq.Variables["input"])
			}
			if v, ok := input["stateId"].(string); ok {
				sawStateID = v
			}
			body := `{"data":{"issueCreate":{"issue":{"id":"i1","identifier":"ENG-1","title":"Fix bug","description":"","priority":3,"estimate":0,"createdAt":"2026-03-02T00:00:00Z","updatedAt":"2026-03-02T00:00:00Z","dueDate":"","state":{"id":"state-2","name":"In Progress","type":"started","color":"#00ff00"},"assignee":null,"team":{"id":"team-1","key":"ENG","name":"Engineering"},"labels":{"nodes":[]},"project":null,"projectMilestone":null,"parent":null}}}}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		default:
			t.Fatalf("unexpected GraphQL operation: %s", gqlReq.Query)
			return nil, nil
		}
	})

	_ = issueCreateCmd.Flags().Set("title", "Fix bug")
	_ = issueCreateCmd.Flags().Set("team", "ENG")
	_ = issueCreateCmd.Flags().Set("state", "In Progress")
	issueCreateCmd.Run(issueCreateCmd, nil)

	if sawStateID != "state-2" {
		t.Fatalf("expected stateId state-2, got %q", sawStateID)
	}
}

func TestExtractUploadsLinearURLs(t *testing.T) {
	text := `
Main [doc](https://uploads.linear.app/abc-123/spec.md) and plain https://uploads.linear.app/def-456/log.txt.
Ignore https://example.com/file.txt
`

	got := extractUploadsLinearURLs(text)
	want := []string{
		"https://uploads.linear.app/abc-123/spec.md",
		"https://uploads.linear.app/def-456/log.txt",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestCollectIssueAttachmentEntries(t *testing.T) {
	issue := &api.Issue{
		Description: "See https://uploads.linear.app/path/from-description.md",
		Attachments: &api.Attachments{
			Nodes: []api.Attachment{
				{ID: "att-1", Title: "Canonical", URL: "https://uploads.linear.app/path/from-attachment.md"},
			},
		},
		Comments: &api.Comments{
			Nodes: []api.Comment{
				{Body: "Another link https://uploads.linear.app/path/from-comment.md"},
				{Body: "Duplicate https://uploads.linear.app/path/from-attachment.md should not repeat"},
			},
		},
	}

	entries := collectIssueAttachmentEntries(issue)
	if len(entries) != 3 {
		t.Fatalf("expected 3 unique entries, got %d: %#v", len(entries), entries)
	}

	if entries[0].Source != "attachment" || entries[0].ID != "att-1" {
		t.Fatalf("expected first entry to be canonical attachment, got %#v", entries[0])
	}
}

func TestSelectAttachmentEntriesForDownload(t *testing.T) {
	entries := []issueAttachmentEntry{
		{ID: "a-1", Title: "spec.md", URL: "https://uploads.linear.app/x/spec.md", Source: "attachment"},
		{ID: "a-2", Title: "notes.md", URL: "https://uploads.linear.app/y/notes.md", Source: "attachment"},
	}

	gotAll, skippedAll, err := selectAttachmentEntriesForDownload(entries, true, "", "")
	if err != nil || len(gotAll) != 2 || len(skippedAll) != 0 {
		t.Fatalf("expected all entries selected without skips, got selected=%d skipped=%d err=%v", len(gotAll), len(skippedAll), err)
	}

	gotID, skippedID, err := selectAttachmentEntriesForDownload(entries, false, "a-2", "")
	if err != nil || len(gotID) != 1 || gotID[0].ID != "a-2" {
		t.Fatalf("expected id selection a-2, got %#v err=%v", gotID, err)
	}
	if len(skippedID) != 0 {
		t.Fatalf("expected no skipped entries for id selection, got %#v", skippedID)
	}

	gotName, skippedName, err := selectAttachmentEntriesForDownload(entries, false, "", "spec.md")
	if err != nil || len(gotName) != 1 || gotName[0].ID != "a-1" {
		t.Fatalf("expected name selection spec.md, got %#v err=%v", gotName, err)
	}
	if len(skippedName) != 0 {
		t.Fatalf("expected no skipped entries for name selection, got %#v", skippedName)
	}
}

func TestSelectAttachmentEntriesForDownloadAllSkipsNonDownloadableLinks(t *testing.T) {
	entries := []issueAttachmentEntry{
		{ID: "a-1", Title: "pr", URL: "https://github.com/org/repo/pull/123", Source: "attachment"},
		{ID: "a-2", Title: "file.md", URL: "https://uploads.linear.app/abc/file.md", Source: "markdown"},
	}

	selected, skipped, err := selectAttachmentEntriesForDownload(entries, true, "", "")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(selected) != 1 || selected[0].ID != "a-2" {
		t.Fatalf("expected only uploads entry selected, got %#v", selected)
	}
	if len(skipped) != 1 {
		t.Fatalf("expected one skipped entry, got %#v", skipped)
	}
	if skipped[0].Status != "skipped" || skipped[0].Reason != "non-downloadable-link" {
		t.Fatalf("unexpected skipped metadata: %#v", skipped[0])
	}
}

func TestHasAttachmentDownloadFailuresIgnoresSkipped(t *testing.T) {
	results := []issueAttachmentDownloadResult{
		{Status: "skipped", Success: false},
		{Status: "downloaded", Success: true},
	}
	if hasAttachmentDownloadFailures(results) {
		t.Fatalf("skipped entries should not count as failures")
	}

	results = append(results, issueAttachmentDownloadResult{Status: "failed", Success: false})
	if !hasAttachmentDownloadFailures(results) {
		t.Fatalf("failed entries must count as failures")
	}
}

func TestDownloadAttachmentEntryDoesNotSendAuthToExternalURL(t *testing.T) {
	var sawAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Disposition", `attachment; filename="external.txt"`)
		_, _ = w.Write([]byte("file contents"))
	}))
	defer server.Close()

	outputDir := t.TempDir()
	filePath, err := downloadAttachmentEntry(context.Background(), "linear-secret", issueAttachmentEntry{
		Title:  "external.txt",
		URL:    server.URL + "/external.txt",
		Source: "attachment",
	}, outputDir, "")
	if err != nil {
		t.Fatalf("downloadAttachmentEntry returned error: %v", err)
	}
	if sawAuth != "" {
		t.Fatalf("expected no Authorization header for external URL, got %q", sawAuth)
	}
	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("expected downloaded file at %s: %v", filePath, err)
	}
}

func TestShouldSendAttachmentAuthHeaderOnlyForLinearUploads(t *testing.T) {
	if !shouldSendAttachmentAuthHeader("https://uploads.linear.app/path/file.md") {
		t.Fatalf("expected auth header for uploads.linear.app")
	}
	if shouldSendAttachmentAuthHeader("https://example.com/path/file.md") {
		t.Fatalf("expected no auth header for external host")
	}
}

func TestIssueAttachmentFlagsRegistered(t *testing.T) {
	if issueGetCmd.Flags().Lookup("download-attachments") == nil {
		t.Fatalf("issue get is missing --download-attachments")
	}
	if issueGetCmd.Flags().Lookup("output-dir") == nil {
		t.Fatalf("issue get is missing --output-dir")
	}

	for _, name := range []string{"all", "id", "name", "output", "output-dir"} {
		if issueAttachmentDownloadCmd.Flags().Lookup(name) == nil {
			t.Fatalf("issue attachment download is missing --%s", name)
		}
	}
}
