package tui

import (
	"context"
	"os"
	"strings"

	"github.com/qinbin/git-worktree-manager/internal/core"
)

func defaultSourceRepo(ctx context.Context, manager core.Manager, explicitRepo string) string {
	if explicitRepo != "" {
		return explicitRepo
	}
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	result, err := manager.Git().Run(ctx, cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(result.Stdout)
}
