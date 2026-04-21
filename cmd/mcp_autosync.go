package cmd

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/dorkitude/linctl/pkg/auth"
	"github.com/dorkitude/linctl/pkg/mcpcache"
	"github.com/spf13/cobra"
)

const autoMCPSyncSkipEnv = "LINCTL_SKIP_AUTO_MCP_SYNC"

func maybeStartAutoMCPSync(cmd *cobra.Command) {
	if cmd == nil {
		return
	}
	if strings.TrimSpace(os.Getenv(autoMCPSyncSkipEnv)) == "1" {
		return
	}
	if cmd.Name() == "sync" {
		return
	}
	if _, err := auth.GetAuthHeader(); err != nil {
		return
	}

	cache, err := mcpcache.Load()
	if err == nil && !mcpcache.IsStale(cache, mcpCacheMaxAge) {
		return
	}

	exePath, err := os.Executable()
	if err != nil {
		return
	}

	bg := exec.Command(exePath, "mcp", "sync", "--quiet")
	bg.Env = append(os.Environ(), autoMCPSyncSkipEnv+"=1")
	bg.Stdout = io.Discard
	bg.Stderr = io.Discard
	bg.Stdin = nil
	_ = bg.Start()
}
