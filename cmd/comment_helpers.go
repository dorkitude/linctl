package cmd

import (
	"strings"

	"github.com/dorkitude/linctl/pkg/api"
)

func commentAuthorName(comment *api.Comment) string {
	if comment == nil || comment.User == nil {
		return "System"
	}
	if name := strings.TrimSpace(comment.User.Name); name != "" {
		return name
	}
	if email := strings.TrimSpace(comment.User.Email); email != "" {
		return email
	}
	return "System"
}
