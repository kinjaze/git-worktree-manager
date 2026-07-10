package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kinjaze/git-worktree-manager/internal/core"
	"github.com/kinjaze/git-worktree-manager/internal/i18n"
)

func Run(ctx context.Context, manager core.Manager, tr i18n.Translator, configPath string, initialRepo string) error {
	model := newModel(ctx, manager, tr, configPath, defaultSourceRepo(ctx, manager, initialRepo))
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err := program.Run()
	return err
}
