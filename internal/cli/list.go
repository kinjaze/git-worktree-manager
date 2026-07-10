package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kinjaze/git-worktree-manager/internal/core"
	"github.com/kinjaze/git-worktree-manager/internal/jsonapi"
	"github.com/spf13/cobra"
)

func newListCommand(state *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List managed worktrees",
		RunE: func(cmd *cobra.Command, args []string) error {
			tr := state.translator()
			manager := core.NewManager(state.gitRunner(), state.metadataStore())
			result, err := manager.List(cmd.Context())
			if err != nil {
				return handleError(err, tr, state.jsonOutput)
			}
			if state.jsonOutput {
				return printJSON(jsonapi.Success(jsonapi.StatusListed, result))
			}
			if len(result.Worktrees) == 0 {
				fmt.Println(tr.T("list.empty"))
				return nil
			}
			writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(writer, "ID\tNAME\tBRANCH\tSOURCE\tTARGET\tSTATUS\tPATH")
			for _, record := range result.Worktrees {
				fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", record.ID, record.Name, record.WorktreeBranch, record.SourceRemoteBranch, record.TargetLocalBranch, record.Status, record.Path)
			}
			return writer.Flush()
		},
	}
}
