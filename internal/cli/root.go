package cli

import (
	"github.com/kinjaze/git-worktree-manager/internal/config"
	gitpkg "github.com/kinjaze/git-worktree-manager/internal/git"
	"github.com/kinjaze/git-worktree-manager/internal/i18n"
	"github.com/kinjaze/git-worktree-manager/internal/metadata"
	"github.com/spf13/cobra"
)

type appState struct {
	jsonOutput   bool
	lang         string
	configPath   string
	metadataPath string
	repo         string
}

func NewRootCommand() *cobra.Command {
	state := &appState{}
	cmd := &cobra.Command{
		Use:           "gwt",
		Short:         "Manage Git worktrees with CLI, JSON, and TUI workflows",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.PersistentFlags().BoolVar(&state.jsonOutput, "json", false, "print JSON output")
	cmd.PersistentFlags().StringVar(&state.lang, "lang", "", "language: en or zh")
	cmd.PersistentFlags().StringVar(&state.configPath, "config", "", "config file path")
	cmd.PersistentFlags().StringVar(&state.metadataPath, "metadata", "", "metadata file path")
	cmd.PersistentFlags().StringVar(&state.repo, "repo", "", "source repository path")
	cmd.AddCommand(newCreateCommand(state))
	cmd.AddCommand(newListCommand(state))
	cmd.AddCommand(newUpdateCommand(state))
	cmd.AddCommand(newMergeBackCommand(state))
	cmd.AddCommand(newRemoveCommand(state))
	cmd.AddCommand(newConfigCommand(state))
	cmd.AddCommand(newTUICommand(state))
	return cmd
}

func (s *appState) translator() i18n.Translator {
	language := s.lang
	if language == "" {
		cfg, err := config.NewStore(s.configPath).Load()
		if err == nil {
			language = cfg.Language
		}
	}
	return i18n.New(config.NormalizeLanguage(language))
}

func (s *appState) metadataStore() metadata.JSONStore {
	path := s.metadataPath
	if path == "" {
		defaultPath, err := config.DefaultMetadataPath()
		if err == nil {
			path = defaultPath
		}
	}
	return metadata.NewJSONStore(path)
}

func (s *appState) gitRunner() gitpkg.Runner {
	return gitpkg.NewRunner()
}
