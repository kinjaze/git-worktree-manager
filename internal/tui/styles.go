package tui

import "github.com/charmbracelet/lipgloss"

var titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
var selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
var mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
var errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
