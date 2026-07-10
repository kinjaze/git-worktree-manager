package cli

import (
	"fmt"

	"github.com/kinjaze/git-worktree-manager/internal/core"
	"github.com/kinjaze/git-worktree-manager/internal/jsonapi"
	"github.com/spf13/cobra"
)

func newUpdateCommand(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "update <id-or-path>",
		Short: "Update a worktree from its recorded source branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tr := state.translator()
			manager := core.NewManager(state.gitRunner(), state.metadataStore())
			result, err := manager.Update(cmd.Context(), args[0])
			if err != nil {
				return handleError(err, tr, state.jsonOutput)
			}
			if state.jsonOutput {
				return printJSON(jsonapi.Success(jsonapi.StatusUpdated, result))
			}
			fmt.Println(tr.T("update.success", args[0], "recorded source"))
			return nil
		},
	}
}
