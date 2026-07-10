package cli

import (
	"fmt"

	"github.com/kinjaze/git-worktree-manager/internal/core"
	"github.com/kinjaze/git-worktree-manager/internal/jsonapi"
	"github.com/spf13/cobra"
)

func newCreateCommand(state *appState) *cobra.Command {
	var source string
	var branch string
	var path string
	cmd := &cobra.Command{
		Use:   "create <worktree-name>",
		Short: "Create a managed Git worktree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tr := state.translator()
			manager := core.NewManager(state.gitRunner(), state.metadataStore())
			result, err := manager.Create(cmd.Context(), core.CreateOptions{Name: args[0], Repo: state.repo, Source: source, Branch: branch, Path: path})
			if err != nil {
				return handleError(err, tr, state.jsonOutput)
			}
			if state.jsonOutput {
				return printJSON(jsonapi.Success(jsonapi.StatusCreated, result))
			}
			fmt.Println(tr.T("create.success", result.Record.Name, result.Record.Path))
			return nil
		},
	}
	cmd.Flags().StringVar(&source, "source", "", "source remote branch, for example origin/main")
	cmd.Flags().StringVar(&branch, "branch", "", "user-defined worktree branch")
	cmd.Flags().StringVar(&path, "path", "", "new worktree path")
	return cmd
}
