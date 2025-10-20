package ui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)
 

func (m Model) View() string {

	if m.ShowResponse && m.Response != nil {
		return m.renderResponseScreen()
	}

	if m.ShowHistory {
		return m.renderHistoryScreen()
	}

	if m.SelectPayload {
		return m.renderPayloadScreen()
	}

	if m.SelectAuth {
		return m.renderAuthscreen()
	}

	focused := !m.SelectURL && !m.SelectPayloadMenu && !m.SelectPayload && !m.SelectAuth
	focused2 := !m.SelectURL  && !focused
	methods := RenderMethodButtons(m.SelectedMethod, focused)
	payloadOptios := RenderPayloadButtons(m.PayloadMenu, focused2)
	urlInput := RenderURLInput(m.URL, m.Cursor, m.SelectURL) 

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
 
func (m Model) renderPayloadScreen() string {
	payloadWithCursor := m.Payload[:m.PayloadCursor] + "|" + m.Payload[m.PayloadCursor:]

	payloadStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#24283b")).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4ECDC4")).
		Width(50).
		Height(20).
		AlignHorizontal(lipgloss.Left).
		AlignVertical(lipgloss.Top)

	help := HelpStyle.Render("Edit JSON | Esc: back | q: quit")

	content := lipgloss.JoinVertical(lipgloss.Center,
		payloadStyle.Render(payloadWithCursor),
		"",
		help,
	)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(2).
		Width(60).
		AlignHorizontal(lipgloss.Center)

	boxedContent := boxStyle.Render(content)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, boxedContent)
}

func (m Model) renderAuthscreen() string {
	payloadWithCursor := m.AuthToken[:m.AuthTokenCursor] + "|" + m.AuthToken[m.AuthTokenCursor:]

	payloadStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#24283b")).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4ECDC4")).
		Width(50).
		Height(20).
		AlignHorizontal(lipgloss.Left).
		AlignVertical(lipgloss.Top)

	help := HelpStyle.Render("Edit Token | Esc: back | q: quit")

	content := lipgloss.JoinVertical(lipgloss.Center,
		payloadStyle.Render(payloadWithCursor),
		"",
		help,
	)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
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

	help := HelpStyle.Render("↑↓/j/k: scroll | Esc: back | q: quit")

	content := lipgloss.JoinVertical(lipgloss.Center,
		m.Viewport.View(),
		"",
		help,
	)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2)

	boxedContent := boxStyle.Render(content)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, boxedContent)
}

func (m Model) renderResponseContent() string {
	return m.renderResponse()
}

