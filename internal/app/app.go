package app

import (
	"gaufre/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// Run starts Gaufre. popup is set when it runs inside a tmux popup, whose
// border then frames the app in place of Gaufre's own.
func Run(popup bool) error {
	model := ui.NewModel()
	model.Popup = popup

	// The alternate screen gives Gaufre the whole terminal and restores
	// what was there on exit, instead of leaving the last frame behind.
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
