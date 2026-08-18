package api

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	issueIdentifierPattern = regexp.MustCompile(`^[A-Za-z0-9]+-[0-9]+$`)
	uuidPattern            = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

const (
	refIssue      = "issue"
	refProject    = "project"
	refDocument   = "document"
	refTeam       = "team"
	refInitiative = "initiative"
	refReview     = "review"
)

type linearRef struct {
	kind      string
	id        string
	commentID string
}

// parseLinearURL parses <workspace>/<kind>/<id> Linear web and desktop URLs.
func parseLinearURL(raw string) (linearRef, bool) {
	raw = strings.TrimSpace(raw)
	if !strings.Contains(raw, "://") {
		if !strings.HasPrefix(strings.ToLower(raw), "linear.app/") {
			return linearRef{}, false
		}
		raw = "https://" + raw
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return linearRef{}, false
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https", "linear":
	default:
		return linearRef{}, false
	}
	host := strings.ToLower(parsed.Hostname())
	if host != "linear.app" && host != "www.linear.app" {
		return linearRef{}, false
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 3 || parts[0] == "" || parts[2] == "" {
		return linearRef{}, false
	}

	kind := strings.ToLower(parts[1])
	switch kind {
	case refIssue, refProject, refDocument, refTeam, refInitiative, refReview:
	default:
		return linearRef{}, false
	}

	commentID := strings.TrimSpace(parsed.Query().Get("commentId"))
	if commentID == "" && strings.HasPrefix(parsed.Fragment, "comment-") {
		commentID = strings.TrimPrefix(parsed.Fragment, "comment-")
	}

	return linearRef{kind: kind, id: parts[2], commentID: commentID}, true
}

// NormalizeIssueRef accepts an issue identifier, UUID or Linear issue URL.
func NormalizeIssueRef(raw string) (string, error) {
	return normalizeRef(raw, refIssue)
}

// NormalizeProjectRef accepts a project name, UUID, slug ID or Linear project URL.
func NormalizeProjectRef(raw string) (string, error) {
	return normalizeRef(raw, refProject)
}

// NormalizeTeamRef accepts a team key, UUID or Linear team URL.
func NormalizeTeamRef(raw string) (string, error) {
	return normalizeRef(raw, refTeam)
}

func normalizeRef(raw, want string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("%s reference cannot be empty", want)
	}

	ref, ok := parseLinearURL(raw)
	if !ok {
		return raw, nil
	}
	if ref.kind != want {
		if ref.kind == refReview && want == refIssue {
			return "", fmt.Errorf("%s is a Linear review URL, not an issue URL. Pass the GitHub pull request URL instead", raw)
		}
		return "", fmt.Errorf("%s is a Linear %s URL, not a %s URL", raw, ref.kind, want)
	}
	if want == refIssue && !issueIdentifierPattern.MatchString(ref.id) && !uuidPattern.MatchString(ref.id) {
		return "", fmt.Errorf("%s does not contain an issue identifier", raw)
	}
	return ref.id, nil
}

func (c *Client) resolveIssueRef(ctx context.Context, raw string) (string, error) {
	if prURL, ok := canonicalGitHubPullRequestURL(strings.TrimSpace(raw)); ok {
		return c.issueRefFromPullRequestURL(ctx, prURL)
	}
	return NormalizeIssueRef(raw)
}

func (c *Client) resolveCommentRef(ctx context.Context, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("comment reference cannot be empty")
	}

	ref, ok := parseLinearURL(raw)
	if !ok {
		return raw, nil
	}
	if ref.kind != refIssue {
		return "", fmt.Errorf("%s is a Linear %s URL, not a comment URL", raw, ref.kind)
	}
	if ref.commentID == "" {
		return "", fmt.Errorf("%s does not point at a comment. Use Linear's 'Copy link' on the comment itself", raw)
	}
	if uuidPattern.MatchString(ref.commentID) {
		return ref.commentID, nil
	}
	return c.findCommentByIDPrefix(ctx, ref.id, ref.commentID)
}

func (c *Client) findCommentByIDPrefix(ctx context.Context, issueRef, prefix string) (string, error) {
	match := ""
	after := ""
	for {
		comments, err := c.GetIssueComments(ctx, issueRef, 100, after, "")
		if err != nil {
			return "", err
		}
		for _, comment := range comments.Nodes {
			if !strings.HasPrefix(comment.ID, prefix) || comment.ID == match {
				continue
			}
			if match != "" {
				return "", fmt.Errorf("comment prefix %q is ambiguous on issue %s", prefix, issueRef)
			}
			match = comment.ID
		}
		if !comments.PageInfo.HasNextPage {
			break
		}
		if comments.PageInfo.EndCursor == "" {
			return "", fmt.Errorf("comment lookup for issue %s returned no page cursor", issueRef)
		}
		after = comments.PageInfo.EndCursor
	}

	if match == "" {
		return "", fmt.Errorf("no comment starting with %q found on issue %s", prefix, issueRef)
	}
	return match, nil
}

func (c *Client) issueRefFromPullRequestURL(ctx context.Context, prURL string) (string, error) {
	query := `
		query AttachmentsForURL($url: String!) {
			attachmentsForURL(url: $url, first: 100) {
				nodes {
					issue {
						identifier
					}
				}
				pageInfo {
					hasNextPage
					endCursor
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
			PageInfo PageInfo `json:"pageInfo"`
		} `json:"attachmentsForURL"`
	}
	if err := c.Execute(ctx, query, map[string]interface{}{"url": prURL}, &response); err != nil {
		return "", err
	}
	if response.AttachmentsForURL.PageInfo.HasNextPage {
		return "", fmt.Errorf("too many Linear attachments match %s", prURL)
	}

	identifiers := make(map[string]struct{})
	for _, node := range response.AttachmentsForURL.Nodes {
		if node.Issue != nil && node.Issue.Identifier != "" {
			identifiers[node.Issue.Identifier] = struct{}{}
		}
	}
	if len(identifiers) > 1 {
		matches := make([]string, 0, len(identifiers))
		for identifier := range identifiers {
			matches = append(matches, identifier)
		}
		sort.Strings(matches)
		return "", fmt.Errorf("%s is linked to multiple Linear issues: %s", prURL, strings.Join(matches, ", "))
	}

	for identifier := range identifiers {
		return identifier, nil
	}
	return "", fmt.Errorf("no Linear issue is linked to %s", prURL)
}

func canonicalGitHubPullRequestURL(raw string) (string, bool) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" {
		return "", false
	}
	host := strings.ToLower(parsed.Hostname())
	if host != "github.com" && host != "www.github.com" {
		return "", false
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 4 || parts[0] == "" || parts[1] == "" || parts[2] != "pull" {
		return "", false
	}
	number, err := strconv.ParseUint(parts[3], 10, 64)
	if err != nil || number == 0 {
		return "", false
	}

	return fmt.Sprintf("https://github.com/%s/%s/pull/%d", parts[0], parts[1], number), true
}
