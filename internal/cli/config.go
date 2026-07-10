package cli

import (
	"fmt"

	"github.com/qinbin/git-worktree-manager/internal/config"
	"github.com/qinbin/git-worktree-manager/internal/jsonapi"
	"github.com/spf13/cobra"
)

func newConfigCommand(state *appState) *cobra.Command {
	cmd := &cobra.Command{Use: "config", Short: "Read or update gwt configuration"}
	cmd.AddCommand(&cobra.Command{
		Use:   "get language",
		Short: "Print configured language",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tr := state.translator()
			if args[0] != "language" {
				return handleError(fmt.Errorf("only language is supported"), tr, state.jsonOutput)
			}
			cfg, err := config.NewStore(state.configPath).Load()
			if err != nil {
				return handleError(err, tr, state.jsonOutput)
			}
			if state.jsonOutput {
				return printJSON(jsonapi.Success("config", map[string]string{"language": cfg.Language}))
			}
			fmt.Println(tr.T("config.language", cfg.Language))
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "set language <en|zh>",
		Short: "Set configured language",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			tr := state.translator()
			if args[0] != "language" {
				return handleError(fmt.Errorf("only language is supported"), tr, state.jsonOutput)
			}
			language := config.NormalizeLanguage(args[1])
			if args[1] != "en" && args[1] != "zh" {
				return handleError(fmt.Errorf("language must be en or zh"), tr, state.jsonOutput)
			}
			cfg := config.Default()
			cfg.Language = language
			if err := config.NewStore(state.configPath).Save(cfg); err != nil {
				return handleError(err, tr, state.jsonOutput)
			}
			if state.jsonOutput {
				return printJSON(jsonapi.Success("config", map[string]string{"language": language}))
			}
			fmt.Println(tr.T("config.language.updated", language))
			return nil
		},
	})
	return cmd
}
