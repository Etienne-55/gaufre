package ui

import (
	"gaufre/internal/types"
	tea "github.com/charmbracelet/bubbletea"
)


type Model struct {
	URL      string
	Response *types.Response
	Loading  bool
	Cursor   int
	Width int
	Height int
	SelectedMethod int
	SelectURL bool
}

func NewModel() Model {
	defaultURL := "http://localhost:8080"
	return Model{
		URL:    defaultURL,
		Cursor: len(defaultURL),
		SelectedMethod: 0,
		SelectURL: false,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

