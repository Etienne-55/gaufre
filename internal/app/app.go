package app

import (
	"gaufre/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func Run() error {
	// The alternate screen gives Gaufre the whole terminal and restores
	// what was there on exit, instead of leaving the last frame behind.
	p := tea.NewProgram(ui.NewModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

