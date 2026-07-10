package main

import (
	"os"

	"github.com/kinjaze/git-worktree-manager/internal/cli"
)

func main() {
	if err := cli.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
