package cli

import (
	"fmt"

	"github.com/qinbin/git-worktree-manager/internal/core"
	"github.com/qinbin/git-worktree-manager/internal/jsonapi"
	"github.com/spf13/cobra"
)

func newMergeBackCommand(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "merge-back <id-or-path>",
		Short: "Merge a worktree branch back to the recorded target branch using --no-ff",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tr := state.translator()
			manager := core.NewManager(state.gitRunner(), state.metadataStore())
			result, err := manager.MergeBack(cmd.Context(), args[0])
			if err != nil {
				return handleError(err, tr, state.jsonOutput)
			}
			if state.jsonOutput {
				return printJSON(jsonapi.Success(jsonapi.StatusMergedBack, result))
			}
			fmt.Println(tr.T("mergeBack.success", result.WorktreeBranch, result.TargetBranch))
			return nil
		},
	}
}
