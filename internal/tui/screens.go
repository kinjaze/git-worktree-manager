package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kinjaze/git-worktree-manager/internal/core"
	"github.com/kinjaze/git-worktree-manager/internal/i18n"
	"github.com/kinjaze/git-worktree-manager/internal/metadata"
)

type screen int

const (
	screenDashboard screen = iota
	screenDetail
	screenCreate
	screenConfirmUpdate
	screenConfirmMergeBack
	screenConfirmRemove
	screenSettings
	screenConflict
	screenLoading
)

type field int

const (
	fieldName field = iota
	fieldRepo
	fieldSource
	fieldBranch
	fieldPath
	fieldCount
)

type model struct {
	ctx             context.Context
	manager         core.Manager
	tr              i18n.Translator
	configPath      string
	initialRepo     string
	screen          screen
	selected        int
	records         []metadata.Record
	message         string
	err             string
	active          field
	inputs          []textinput.Model
	suggestions     []string
	suggestionIndex int
	spinner         spinner.Model
	loadingMessage  string
	progressStep    int
	progressTotal   int
	progressLabels  []string
	progressChan    chan operationProgressMsg
	conflict        map[string]any
}

func newModel(ctx context.Context, manager core.Manager, tr i18n.Translator, configPath string, initialRepo string) model {
	spin := spinner.New()
	spin.Spinner = spinner.Line
	m := model{ctx: ctx, manager: manager, tr: tr, configPath: configPath, initialRepo: initialRepo, inputs: newCreateInputs(initialRepo), spinner: spin}
	m.inputs[fieldName].Focus()
	m.refresh()
	m.refreshSuggestions()
	return m
}

func newCreateInputs(initialRepo string) []textinput.Model {
	labels := []string{"Worktree name", "Source repo", "Source branch", "Worktree branch", "Worktree path"}
	inputs := make([]textinput.Model, fieldCount)
	for i, label := range labels {
		input := textinput.New()
		input.Placeholder = label
		input.CharLimit = 512
		input.Width = 80
		inputs[i] = input
	}
	inputs[fieldRepo].SetValue(initialRepo)
	inputs[fieldSource].SetValue("origin/main")
	return inputs
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case operationProgressMsg:
		m.progressStep = msg.step
		m.progressTotal = msg.total
		if len(m.progressLabels) != msg.total {
			m.progressLabels = make([]string, msg.total)
		}
		if msg.step > 0 && msg.step <= msg.total {
			m.progressLabels[msg.step-1] = msg.label
		}
		return m, waitProgress(m.progressChan)
	case operationDoneMsg:
		m.handleOperationResult(msg.err, msg.message)
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	switch m.screen {
	case screenDetail:
		return m.viewDetail()
	case screenCreate:
		return m.viewCreate()
	case screenConfirmUpdate:
		return m.viewConfirmUpdate()
	case screenConfirmMergeBack:
		return m.viewConfirmMergeBack()
	case screenConfirmRemove:
		return m.viewConfirmRemove()
	case screenSettings:
		return m.viewSettings()
	case screenConflict:
		return m.viewConflict()
	case screenLoading:
		return m.viewLoading()
	default:
		return m.viewDashboard()
	}
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if key == "ctrl+c" || key == "q" && m.screen == screenDashboard {
		return m, tea.Quit
	}
	switch m.screen {
	case screenDashboard:
		return m.handleDashboardKey(key)
	case screenDetail:
		return m.handleDetailKey(key)
	case screenCreate:
		return m.handleCreateKey(msg)
	case screenConfirmUpdate:
		return m.handleConfirmUpdateKey(key)
	case screenConfirmMergeBack:
		return m.handleConfirmMergeBackKey(key)
	case screenConfirmRemove:
		return m.handleConfirmRemoveKey(key)
	case screenSettings:
		return m.handleSettingsKey(key)
	case screenConflict:
		if key == "esc" || key == "enter" || key == "r" {
			m.screen = screenDashboard
			m.refresh()
		}
	case screenLoading:
		return m, nil
	}
	return m, nil
}

func (m *model) refresh() {
	result, err := m.manager.List(m.ctx)
	if err != nil {
		m.err = err.Error()
		return
	}
	m.err = ""
	m.records = result.Worktrees
	if m.selected >= len(m.records) {
		m.selected = len(m.records) - 1
	}
	if m.selected < 0 {
		m.selected = 0
	}
}

func (m model) current() (metadata.Record, bool) {
	if len(m.records) == 0 || m.selected < 0 || m.selected >= len(m.records) {
		return metadata.Record{}, false
	}
	return m.records[m.selected], true
}

func (m model) viewDashboard() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(m.tr.T("tui.title")))
	b.WriteString("\n\n")
	if m.err != "" {
		b.WriteString(errorStyle.Render(m.err))
		b.WriteString("\n")
	}
	if m.message != "" {
		b.WriteString(m.message)
		b.WriteString("\n")
	}
	if len(m.records) == 0 {
		b.WriteString(m.tr.T("tui.empty"))
		b.WriteString("\n")
	} else {
		b.WriteString(mutedStyle.Render(fmt.Sprintf("  %-18s %-24s %-18s %-12s %s", "NAME", "BRANCH", "SOURCE", "STATUS", "PATH")))
		b.WriteString("\n")
		b.WriteString(mutedStyle.Render(fmt.Sprintf("  %-18s %-24s %-18s %-12s %s", strings.Repeat("-", 18), strings.Repeat("-", 24), strings.Repeat("-", 18), strings.Repeat("-", 12), strings.Repeat("-", 24))))
		b.WriteString("\n")
		for i, record := range m.records {
			line := fmt.Sprintf("%-18s %-24s %-18s %-12s %s", truncate(record.Name, 18), truncate(record.WorktreeBranch, 24), truncate(record.SourceRemoteBranch, 18), truncate(record.Status, 12), record.Path)
			if i == m.selected {
				line = selectedStyle.Render("> " + line)
			} else {
				line = "  " + line
			}
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(m.tr.T("tui.footer")))
	return b.String()
}

func truncate(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	if max <= 1 {
		return string(runes[:max])
	}
	return string(runes[:max-1]) + "…"
}

func (m model) viewDetail() string {
	record, ok := m.current()
	if !ok {
		return m.viewDashboard()
	}
	return fmt.Sprintf("%s\n\nName: %s\nBranch: %s\nSource: %s\nTarget: %s\nPath: %s\nStatus: %s\n\nSuggested actions\n u  Update from source\n m  Merge back with --no-ff\n d  Delete worktree\n\n%s", titleStyle.Render(record.Name), record.Name, record.WorktreeBranch, record.SourceRemoteBranch, record.TargetLocalBranch, record.Path, record.Status, mutedStyle.Render(m.tr.T("tui.back")))
}

func (m model) viewCreate() string {
	labels := []string{"Worktree name", "Source repo", "Source branch", "Worktree branch", "Worktree path"}
	var fields []string
	for i, label := range labels {
		prefix := "  "
		extra := ""
		if field(i) == m.active {
			prefix = selectedStyle.Render("> ")
			extra = "\n  " + mutedStyle.Render(m.cliHint(field(i)))
			if suggestions := m.viewSuggestions(); suggestions != "" {
				extra += "\n" + suggestions
			}
		}
		fields = append(fields, fmt.Sprintf("%s%s:\n  %s%s", prefix, label, m.inputs[i].View(), extra))
	}
	preview := fmt.Sprintf("git -C %s fetch %s\ngit -C %s worktree add -b %s %s %s", m.inputValue(fieldRepo), sourceRemote(m.inputValue(fieldSource)), m.inputValue(fieldRepo), m.inputValue(fieldBranch), m.inputValue(fieldPath), m.inputValue(fieldSource))
	return fmt.Sprintf("%s\n\n%s\n\n%s\n%s\n\nTab next field  ↑/↓ select suggestion  Enter accept suggestion/submit  Esc back", titleStyle.Render("Create worktree"), strings.Join(fields, "\n\n"), m.tr.T("tui.commandPreview"), mutedStyle.Render(preview))
}

func (m model) cliHint(field field) string {
	switch field {
	case fieldName:
		return "CLI: gwt create <worktree-name>"
	case fieldRepo:
		return "CLI: --repo"
	case fieldSource:
		return "CLI: --source"
	case fieldBranch:
		return "CLI: --branch"
	case fieldPath:
		return "CLI: --path"
	default:
		return ""
	}
}

func (m model) viewSuggestions() string {
	if len(m.suggestions) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("Suggestions:\n")
	for i, suggestion := range m.suggestions {
		line := "  " + suggestion
		if i == m.suggestionIndex {
			line = selectedStyle.Render("> " + suggestion)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}

func (m model) viewConfirmUpdate() string {
	record, _ := m.current()
	preview := fmt.Sprintf("git fetch %s\ngit merge %s", record.SourceRemote, record.SourceRemoteBranch)
	return fmt.Sprintf("%s\n\nWorktree: %s\nBranch: %s\nSource: %s\n\n%s\n%s\n\ny confirm  esc cancel", titleStyle.Render("Update worktree"), record.Name, record.WorktreeBranch, record.SourceRemoteBranch, m.tr.T("tui.commandPreview"), mutedStyle.Render(preview))
}

func (m model) viewConfirmMergeBack() string {
	record, _ := m.current()
	preview := fmt.Sprintf("git merge --no-ff %s", record.WorktreeBranch)
	return fmt.Sprintf("%s\n\nWorktree: %s\nSource branch: %s\nTarget branch: %s\n\n%s\n%s\n\ny confirm  esc cancel", titleStyle.Render("Merge back"), record.Name, record.WorktreeBranch, record.TargetLocalBranch, m.tr.T("tui.commandPreview"), mutedStyle.Render(preview))
}

func (m model) viewConfirmRemove() string {
	record, _ := m.current()
	return fmt.Sprintf("%s\n\nWorktree: %s\nPath: %s\nStatus: %s\n\ny delete  esc cancel", titleStyle.Render("Delete worktree"), record.Name, record.Path, record.Status)
}

func (m model) viewSettings() string {
	return fmt.Sprintf("%s\n\n%s: %s\n%s\n%s", titleStyle.Render(m.tr.T("tui.settings")), m.tr.T("tui.language"), m.tr.Language(), m.tr.T("tui.switchLanguage"), m.tr.T("tui.back"))
}

func (m model) viewConflict() string {
	return fmt.Sprintf("%s\n\n%v\n\n%s\n\nEnter back", errorStyle.Render(m.tr.T("error.mergeConflict")), m.conflict, m.tr.T("tui.conflict.nextSteps"))
}

func (m model) viewLoading() string {
	return fmt.Sprintf("%s %s\n\n%s %d/%d\n\n%s\n\nPlease wait. Git operation is still running.", m.spinner.View(), m.loadingMessage, m.progressBar(), m.progressStep, m.progressTotal, m.progressList())
}

func (m model) progressBar() string {
	if m.progressTotal <= 0 {
		return "[░░░░░░░░░░]"
	}
	width := 10
	filled := m.progressStep * width / m.progressTotal
	if filled > width {
		filled = width
	}
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
}

func (m model) progressList() string {
	if m.progressTotal <= 0 {
		return mutedStyle.Render("Preparing operation...")
	}
	var b strings.Builder
	for i := 0; i < m.progressTotal; i++ {
		label := "Step"
		if i < len(m.progressLabels) && m.progressLabels[i] != "" {
			label = m.progressLabels[i]
		}
		marker := "  "
		if i+1 < m.progressStep {
			marker = "✓ "
		} else if i+1 == m.progressStep {
			marker = "→ "
		}
		line := marker + label
		if i+1 == m.progressStep {
			line = selectedStyle.Render(line)
		} else if i+1 > m.progressStep {
			line = mutedStyle.Render(line)
		}
		b.WriteString(line)
		if i < m.progressTotal-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func (m model) inputValue(field field) string {
	return m.inputs[field].Value()
}

func sourceRemote(source string) string {
	parts := strings.SplitN(source, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		return "origin"
	}
	return parts[0]
}
