package api

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// LinearRefKind identifies the entity a linear.app URL points at.
type LinearRefKind string

const (
	LinearRefIssue      LinearRefKind = "issue"
	LinearRefProject    LinearRefKind = "project"
	LinearRefDocument   LinearRefKind = "document"
	LinearRefTeam       LinearRefKind = "team"
	LinearRefInitiative LinearRefKind = "initiative"
	LinearRefReview     LinearRefKind = "review"
)

// LinearRef is a linear.app URL split into the parts Linear's API understands.
type LinearRef struct {
	Kind      LinearRefKind
	Workspace string
	// ID is the value Linear accepts for the entity: an issue identifier such as
	// ENG-123, a team key such as ENG, or a slug id such as roadmap-d05c5c7e8a5c.
	ID string
	// CommentID is set when the URL points at a comment. Web app links carry only
	// the first 8 characters of the comment UUID.
	CommentID string
}

var (
	uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	// Issue identifiers look like ENG-123. Linear accepts them in any case.
	issueIdentifierPattern = regexp.MustCompile(`^[A-Za-z0-9]+-[0-9]+$`)
	// Projects, documents, initiatives and reviews end their URL slug with a short
	// hex id, which is what Linear calls the slug id.
	slugIDPattern = regexp.MustCompile(`^[0-9a-f]{8,16}$`)
)

// ParseLinearURL parses a linear.app URL. It reports false when raw is not one,
// which is the common case: most references are bare identifiers or UUIDs.
func ParseLinearURL(raw string) (LinearRef, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return LinearRef{}, false
	}

	// People paste bare hosts as often as full URLs.
	if !strings.Contains(trimmed, "://") {
		if !strings.HasPrefix(strings.ToLower(trimmed), "linear.app/") {
			return LinearRef{}, false
		}
		trimmed = "https://" + trimmed
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return LinearRef{}, false
	}
	host := strings.ToLower(parsed.Hostname())
	if host != "linear.app" && host != "www.linear.app" {
		return LinearRef{}, false
	}

	var segments []string
	for _, segment := range strings.Split(parsed.Path, "/") {
		if segment != "" {
			segments = append(segments, segment)
		}
	}

	for i, segment := range segments {
		kind := LinearRefKind(strings.ToLower(segment))
		switch kind {
		case LinearRefIssue, LinearRefProject, LinearRefDocument, LinearRefTeam, LinearRefInitiative, LinearRefReview:
		default:
			continue
		}
		if i == len(segments)-1 {
			return LinearRef{}, false
		}

		ref := LinearRef{Kind: kind, ID: segments[i+1], CommentID: commentIDFromURL(parsed)}
		if i > 0 {
			ref.Workspace = segments[0]
		}
		return ref, true
	}

	return LinearRef{}, false
}

func commentIDFromURL(parsed *url.URL) string {
	if value := strings.TrimSpace(parsed.Query().Get("commentId")); value != "" {
		return value
	}
	if fragment := strings.TrimSpace(parsed.Fragment); strings.HasPrefix(fragment, "comment-") {
		return strings.TrimPrefix(fragment, "comment-")
	}
	return ""
}

// NormalizeIssueRef accepts an issue identifier, a UUID or a linear.app issue URL
// and returns the value Linear's issue(id:) query understands.
func NormalizeIssueRef(ref string) (string, error) {
	return normalizeRef(ref, LinearRefIssue, "issue")
}

// NormalizeProjectRef accepts a project name, a UUID, a slug id or a linear.app
// project URL and returns the value Linear's project(id:) query understands.
func NormalizeProjectRef(ref string) (string, error) {
	return normalizeRef(ref, LinearRefProject, "project")
}

// NormalizeTeamRef accepts a team key, a UUID or a linear.app team URL and returns
// the value Linear's team(id:) query understands.
func NormalizeTeamRef(ref string) (string, error) {
	return normalizeRef(ref, LinearRefTeam, "team")
}

func normalizeRef(ref string, want LinearRefKind, label string) (string, error) {
	trimmed := strings.TrimSpace(ref)
	if trimmed == "" {
		return "", fmt.Errorf("%s reference cannot be empty", label)
	}

	parsed, ok := ParseLinearURL(trimmed)
	if !ok {
		return trimmed, nil
	}
	if parsed.Kind != want {
		return "", wrongRefKindError(trimmed, parsed, label)
	}
	if want == LinearRefIssue && !issueIdentifierPattern.MatchString(parsed.ID) && !uuidPattern.MatchString(parsed.ID) {
		return "", fmt.Errorf("%s does not contain an issue identifier", trimmed)
	}
	return parsed.ID, nil
}

func wrongRefKindError(raw string, parsed LinearRef, want string) error {
	if parsed.Kind == LinearRefReview {
		return fmt.Errorf("%s is a Linear review URL. It points at a pull request, and Linear's API cannot map one back to %s. Pass the GitHub pull request URL instead, or find the issue with: linctl issue search %q",
			raw, withArticle(want), reviewSearchTerms(parsed.ID))
	}
	return fmt.Errorf("%s is %s URL, not %s URL", raw, withArticle("Linear "+string(parsed.Kind)), withArticle(want))
}

// reviewSearchTerms turns the slug of a review URL into words worth searching for,
// dropping the trailing slug id. Review slugs come from the pull request title,
// which usually echoes the issue title closely enough for full-text search.
func reviewSearchTerms(slug string) string {
	words := strings.Split(slug, "-")
	if n := len(words); n > 1 && slugIDPattern.MatchString(words[n-1]) {
		words = words[:n-1]
	}
	return strings.Join(words, " ")
}

func withArticle(noun string) string {
	if noun == "" {
		return noun
	}
	if strings.ContainsRune("aeiouAEIOU", rune(noun[0])) {
		return "an " + noun
	}
	return "a " + noun
}

// ResolveIssueRef resolves a user-supplied issue reference. Identifiers and UUIDs
// pass through untouched, linear.app issue URLs are parsed locally, and GitHub pull
// request URLs are looked up through the attachment that links them to an issue.
func (c *Client) ResolveIssueRef(ctx context.Context, ref string) (string, error) {
	trimmed := strings.TrimSpace(ref)
	if prURL, ok := canonicalGitHubPullRequestURL(trimmed); ok {
		return c.issueRefFromPullRequestURL(ctx, prURL)
	}
	return NormalizeIssueRef(trimmed)
}

// ResolveCommentRef resolves a comment UUID or a linear.app comment URL to a comment
// UUID. Web app links carry only an 8-character prefix, so those cost a lookup.
func (c *Client) ResolveCommentRef(ctx context.Context, ref string) (string, error) {
	trimmed := strings.TrimSpace(ref)
	if trimmed == "" {
		return "", fmt.Errorf("comment reference cannot be empty")
	}

	parsed, ok := ParseLinearURL(trimmed)
	if !ok {
		return trimmed, nil
	}
	if parsed.Kind != LinearRefIssue {
		return "", wrongRefKindError(trimmed, parsed, "issue")
	}
	if parsed.CommentID == "" {
		return "", fmt.Errorf("%s does not point at a comment. Use Linear's 'Copy link' on the comment itself", trimmed)
	}
	if uuidPattern.MatchString(parsed.CommentID) {
		return parsed.CommentID, nil
	}
	return c.findCommentByIDPrefix(ctx, parsed.ID, parsed.CommentID)
}

func (c *Client) findCommentByIDPrefix(ctx context.Context, issueRef string, prefix string) (string, error) {
	after := ""
	for {
		comments, err := c.GetIssueComments(ctx, issueRef, 100, after, "")
		if err != nil {
			return "", err
		}
		for _, comment := range comments.Nodes {
			if strings.HasPrefix(comment.ID, prefix) {
				return comment.ID, nil
			}
		}
		if !comments.PageInfo.HasNextPage || comments.PageInfo.EndCursor == "" {
			return "", fmt.Errorf("no comment starting with %q found on issue %s", prefix, issueRef)
		}
		after = comments.PageInfo.EndCursor
	}
}

func (c *Client) issueRefFromPullRequestURL(ctx context.Context, prURL string) (string, error) {
	query := `
		query AttachmentsForURL($url: String!) {
			attachmentsForURL(url: $url, first: 10) {
				nodes {
					issue {
						identifier
					}
				}
			}
		}
	`

	var response struct {
		AttachmentsForURL struct {
			Nodes []struct {
				Issue *struct {
					Identifier string `json:"identifier"`
				} `json:"issue"`
			} `json:"nodes"`
		} `json:"attachmentsForURL"`
	}

	if err := c.Execute(ctx, query, map[string]interface{}{"url": prURL}, &response); err != nil {
		return "", err
	}

	for _, node := range response.AttachmentsForURL.Nodes {
		if node.Issue != nil && node.Issue.Identifier != "" {
			return node.Issue.Identifier, nil
		}
	}

	return "", fmt.Errorf("no Linear issue is linked to %s", prURL)
}

// canonicalGitHubPullRequestURL recognises a GitHub pull request URL and rewrites it
// to the form Linear stores on the issue attachment.
func canonicalGitHubPullRequestURL(raw string) (string, bool) {
	if !strings.Contains(raw, "://") {
		return "", false
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", false
	}
	host := strings.ToLower(parsed.Hostname())
	if host != "github.com" && host != "www.github.com" {
		return "", false
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 4 || parts[2] != "pull" {
		return "", false
	}
	if _, err := strconv.Atoi(parts[3]); err != nil {
		return "", false
	}

	return fmt.Sprintf("https://github.com/%s/%s/pull/%s", parts[0], parts[1], parts[3]), true
}
