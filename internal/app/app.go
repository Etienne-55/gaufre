package app

import (
	"gaufre/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the application
func Run() error {
	p := tea.NewProgram(ui.NewModel())
	_, err := p.Run()
	return err
}

