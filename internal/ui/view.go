package ui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {

	switch m.Screen {
	case ScreenResponse:
		if m.Response != nil {
			return m.renderResponseScreen()
		}
	case ScreenHistory:
		return m.renderHistoryScreen()
	case ScreenPayloadEditor:
		return m.renderPayloadScreen()
	case ScreenAuthEditor:
		return m.renderAuthscreen()
	}

	focused := m.MainFocus == FocusMethod
	focused2 := m.MainFocus == FocusPayloadMenu
	methods := RenderMethodButtons(m.SelectedMethod, focused)
	payloadOptios := RenderPayloadButtons(m.PayloadMenu, focused2)
	urlInput := RenderURLInput(m.URL, m.Cursor, m.MainFocus == FocusURL)

	loading := ""
	if m.Loading {
		loading = fmt.Sprintf("Loading %s", m.Spinner.View())
	}

	help := HelpStyle.Render("←→: choose method | Enter: send | Tab: history | q: quit")

	content := lipgloss.JoinVertical(lipgloss.Center,
		RenderLogo(),
		"",
		methods,
		"",
		urlInput,
		"",
		payloadOptios,
		"",
		loading,
		"",
		help,
	)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(2)

	boxedContent := boxStyle.Render(content)

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, boxedContent)
}

func (m Model) renderHistoryScreen() string {

	content := lipgloss.JoinVertical(lipgloss.Center,
		m.HistoryList.View(),
	)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4ECDC4")).
		Padding(2).
		Width(60).
		AlignHorizontal(lipgloss.Center)

	boxedContent := boxStyle.Render(content)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, boxedContent)
}

func (m Model) renderResponseScreen() string {
	if !m.ViewportReady {
		return "Loading..."
	}
	help := HelpStyle.Render("↑↓/j/k: scroll | ←→: horizontal scroll | Esc: back | q: quit")
	content := lipgloss.JoinVertical(lipgloss.Left,
		m.Viewport.View(),
		"",
		help,
	)

	boxWidth := m.Width - 10
	if boxWidth < 80 {
		boxWidth = 80
	}
	if boxWidth > 150 {
		boxWidth = 150
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Width(boxWidth).
		Padding(1, 2)

	boxedContent := boxStyle.Render(content)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, boxedContent)
}

func (m Model) renderResponseContent() string {
	return m.renderResponse()
}
