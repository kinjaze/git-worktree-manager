package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/qinbin/git-worktree-manager/internal/config"
	"github.com/qinbin/git-worktree-manager/internal/core"
	gitpkg "github.com/qinbin/git-worktree-manager/internal/git"
	"github.com/qinbin/git-worktree-manager/internal/i18n"
)

type operationProgressMsg struct {
	step  int
	total int
	label string
}

type operationDoneMsg struct {
	message string
	err     error
}

func operationCmd(message string, progressChan chan operationProgressMsg, run func(progress core.ProgressFunc) error) tea.Cmd {
	return func() tea.Msg {
		err := run(func(step int, total int, label string) {
			progressChan <- operationProgressMsg{step: step, total: total, label: label}
		})
		close(progressChan)
		return operationDoneMsg{message: message, err: err}
	}
}

func waitProgress(progressChan chan operationProgressMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-progressChan
		if !ok {
			return nil
		}
		return msg
	}
}

func (m *model) startOperation(message string, total int) chan operationProgressMsg {
	m.err = ""
	m.message = ""
	m.conflict = nil
	m.loadingMessage = message
	m.progressStep = 0
	m.progressTotal = total
	m.progressLabels = make([]string, total)
	m.progressChan = make(chan operationProgressMsg, total)
	m.screen = screenLoading
	return m.progressChan
}

func (m model) handleDashboardKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < len(m.records)-1 {
			m.selected++
		}
	case "enter":
		if _, ok := m.current(); ok {
			m.screen = screenDetail
		}
	case "c":
		m.screen = screenCreate
		m.focusActiveInput()
		m.refreshSuggestions()
	case "u":
		if _, ok := m.current(); ok {
			m.screen = screenConfirmUpdate
		}
	case "m":
		if _, ok := m.current(); ok {
			m.screen = screenConfirmMergeBack
		}
	case "d":
		if _, ok := m.current(); ok {
			m.screen = screenConfirmRemove
		}
	case "r":
		m.refresh()
	case "s":
		m.screen = screenSettings
	}
	return m, nil
}

func (m model) handleDetailKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.screen = screenDashboard
	case "u":
		m.screen = screenConfirmUpdate
	case "m":
		m.screen = screenConfirmMergeBack
	case "d":
		m.screen = screenConfirmRemove
	}
	return m, nil
}

func (m model) handleCreateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenDashboard
	case "tab":
		m.nextField()
	case "shift+tab":
		m.prevField()
	case "up":
		if len(m.suggestions) > 0 {
			m.suggestionIndex = (m.suggestionIndex - 1 + len(m.suggestions)) % len(m.suggestions)
			return m, nil
		}
		m.prevField()
	case "down":
		if len(m.suggestions) > 0 {
			m.suggestionIndex = (m.suggestionIndex + 1) % len(m.suggestions)
			return m, nil
		}
		m.nextField()
	case "enter":
		if m.acceptSuggestion() {
			return m, nil
		}
		options := core.CreateOptions{Name: m.inputValue(fieldName), Repo: m.inputValue(fieldRepo), Source: m.inputValue(fieldSource), Branch: m.inputValue(fieldBranch), Path: m.inputValue(fieldPath)}
		progressChan := m.startOperation("Creating worktree...", 5)
		return m, tea.Batch(waitProgress(progressChan), operationCmd("created", progressChan, func(progress core.ProgressFunc) error {
			_, err := m.manager.CreateWithProgress(m.ctx, options, progress)
			return err
		}))
	default:
		var cmd tea.Cmd
		m.inputs[m.active], cmd = m.inputs[m.active].Update(msg)
		m.refreshSuggestions()
		return m, cmd
	}
	m.refreshSuggestions()
	return m, nil
}

func (m model) handleConfirmUpdateKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.screen = screenDashboard
	case "y":
		record, ok := m.current()
		if !ok {
			m.screen = screenDashboard
			break
		}
		progressChan := m.startOperation("Updating worktree...", 4)
		return m, tea.Batch(waitProgress(progressChan), operationCmd("updated", progressChan, func(progress core.ProgressFunc) error {
			_, err := m.manager.UpdateWithProgress(m.ctx, record.ID, progress)
			return err
		}))
	}
	return m, nil
}

func (m model) handleConfirmMergeBackKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.screen = screenDashboard
	case "y":
		record, ok := m.current()
		if !ok {
			m.screen = screenDashboard
			break
		}
		progressChan := m.startOperation("Merging back worktree...", 7)
		return m, tea.Batch(waitProgress(progressChan), operationCmd("merged back", progressChan, func(progress core.ProgressFunc) error {
			_, err := m.manager.MergeBackWithProgress(m.ctx, record.ID, progress)
			return err
		}))
	}
	return m, nil
}

func (m model) handleConfirmRemoveKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.screen = screenDashboard
	case "y":
		record, ok := m.current()
		if !ok {
			m.screen = screenDashboard
			break
		}
		progressChan := m.startOperation("Removing worktree...", 5)
		return m, tea.Batch(waitProgress(progressChan), operationCmd("removed", progressChan, func(progress core.ProgressFunc) error {
			_, err := m.manager.RemoveWithProgress(m.ctx, core.RemoveOptions{Selector: record.ID}, progress)
			return err
		}))
	}
	return m, nil
}

func (m model) handleSettingsKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.screen = screenDashboard
	case "l":
		language := "zh"
		if m.tr.Language() == "zh" {
			language = "en"
		}
		cfg := config.Default()
		cfg.Language = language
		if err := config.NewStore(m.configPath).Save(cfg); err != nil {
			m.err = err.Error()
			break
		}
		m.tr = i18n.New(language)
		m.message = m.tr.T("config.language.updated", language)
	}
	return m, nil
}

func (m *model) handleOperationResult(err error, action string) {
	m.loadingMessage = ""
	if err != nil {
		m.message = ""
		m.conflict = nil
		if coreErr, ok := err.(core.Error); ok && coreErr.Data != nil {
			if data, ok := coreErr.Data.(map[string]any); ok {
				m.conflict = data
				m.screen = screenConflict
				return
			}
		}
		m.err = err.Error()
		m.screen = screenDashboard
		return
	}
	m.err = ""
	m.conflict = nil
	m.message = fmt.Sprintf("%s", action)
	if action == "created" {
		m.inputs = newCreateInputs(m.initialRepo)
		m.active = fieldName
		m.focusActiveInput()
		m.refreshSuggestions()
	}
	m.screen = screenDashboard
	m.refresh()
}

func (m *model) nextField() {
	if m.active < fieldPath {
		m.active++
	} else {
		m.active = fieldName
	}
	m.focusActiveInput()
}

func (m *model) prevField() {
	if m.active > fieldName {
		m.active--
	} else {
		m.active = fieldPath
	}
	m.focusActiveInput()
}

func (m *model) focusActiveInput() {
	for i := range m.inputs {
		if field(i) == m.active {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

func (m *model) refreshSuggestions() {
	m.suggestionIndex = 0
	switch m.active {
	case fieldRepo, fieldPath:
		m.suggestions = pathSuggestions(m.inputValue(m.active))
	case fieldSource:
		m.suggestions = branchSuggestions(m.ctx, m.manager.Git(), m.inputValue(fieldRepo), m.inputValue(fieldSource))
	default:
		m.suggestions = nil
	}
}

func (m *model) acceptSuggestion() bool {
	if len(m.suggestions) == 0 || m.suggestionIndex < 0 || m.suggestionIndex >= len(m.suggestions) {
		return false
	}
	m.inputs[m.active].SetValue(m.suggestions[m.suggestionIndex])
	m.inputs[m.active].CursorEnd()
	m.refreshSuggestions()
	return true
}

func pathSuggestions(value string) []string {
	if value == "" {
		value = "."
	}
	expanded := value
	if strings.HasPrefix(expanded, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			expanded = filepath.Join(home, strings.TrimPrefix(expanded, "~/"))
		}
	}
	dir := expanded
	prefix := ""
	if info, err := os.Stat(expanded); err == nil && !info.IsDir() {
		dir = filepath.Dir(expanded)
		prefix = filepath.Base(expanded)
	} else if err != nil {
		dir = filepath.Dir(expanded)
		prefix = filepath.Base(expanded)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var suggestions []string
	for _, entry := range entries {
		name := entry.Name()
		if prefix != "" && !strings.HasPrefix(name, prefix) {
			continue
		}
		path := filepath.Join(dir, name)
		if entry.IsDir() {
			path += string(os.PathSeparator)
		}
		suggestions = append(suggestions, path)
		if len(suggestions) >= 8 {
			break
		}
	}
	sort.Strings(suggestions)
	return suggestions
}

func branchSuggestions(ctx context.Context, runner gitpkg.Runner, repo string, value string) []string {
	if repo == "" {
		return nil
	}
	result, err := runner.Run(ctx, repo, "branch", "-r")
	if err != nil {
		return nil
	}
	var suggestions []string
	for _, line := range strings.Split(result.Stdout, "\n") {
		branch := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "*"))
		branch = strings.TrimSpace(branch)
		if branch == "" || strings.Contains(branch, " -> ") {
			continue
		}
		if value != "" && !strings.HasPrefix(branch, value) {
			continue
		}
		suggestions = append(suggestions, branch)
		if len(suggestions) >= 8 {
			break
		}
	}
	return suggestions
}
