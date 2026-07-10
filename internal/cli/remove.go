package cli

import (
	"fmt"

	"github.com/qinbin/git-worktree-manager/internal/core"
	"github.com/qinbin/git-worktree-manager/internal/jsonapi"
	"github.com/spf13/cobra"
)

func newRemoveCommand(state *appState) *cobra.Command {
	var force bool
	var metadataOnly bool
	cmd := &cobra.Command{
		Use:   "remove <id-or-path>",
		Short: "Remove a managed worktree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tr := state.translator()
			manager := core.NewManager(state.gitRunner(), state.metadataStore())
			result, err := manager.Remove(cmd.Context(), core.RemoveOptions{Selector: args[0], Force: force, MetadataOnly: metadataOnly})
			if err != nil {
				return handleError(err, tr, state.jsonOutput)
			}
			if state.jsonOutput {
				return printJSON(jsonapi.Success(jsonapi.StatusRemoved, result))
			}
			fmt.Println(tr.T("remove.success", result.Path))
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "force Git worktree removal")
	cmd.Flags().BoolVar(&metadataOnly, "metadata-only", false, "remove only the metadata record")
	return cmd
}
