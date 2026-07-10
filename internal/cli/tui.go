package cli

import (
	"github.com/qinbin/git-worktree-manager/internal/core"
	uitool "github.com/qinbin/git-worktree-manager/internal/tui"
	"github.com/spf13/cobra"
)

func newTUICommand(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Open the interactive worktree manager",
		RunE: func(cmd *cobra.Command, args []string) error {
			tr := state.translator()
			manager := core.NewManager(state.gitRunner(), state.metadataStore())
			return uitool.Run(cmd.Context(), manager, tr, state.configPath, state.repo)
		},
	}
}
