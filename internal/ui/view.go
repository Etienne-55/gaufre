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

	methods := RenderButtons(m.SelectedMethod)
	urlInput := RenderURLInput(m.URL, m.Cursor, m.SelectURL) 

	payloadButton := ""
	if m.SelectedMethod == 1 || m.SelectedMethod == 2 {
		buttonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#874BFD")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(0, 2)
		payloadButton = buttonStyle.Render("Edit Payload")
	} else {
		buttonStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("0")). 
			Padding(0, 2)
		payloadButton = buttonStyle.Render("            ")
	}


	loading := ""
	if m.Loading {
		loading = fmt.Sprintf("Loading %s", m.Spinner.View())	
	}

	help := HelpStyle.Render("←→: choose method | Enter: send | q: quit")

	content := lipgloss.JoinVertical(lipgloss.Center,
		RenderLogo(),
		"",
		methods,
		"",
		urlInput,
		"",
		payloadButton,
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
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2)

	availableWidth := m.Width - 10
	availableHeight := m.Height - 8

	response := m.renderResponse()

	styledResponse := lipgloss.NewStyle().
		Width(availableWidth).
		Height(availableHeight).
		AlignVertical(lipgloss.Top).
		Render(response)


	help := HelpStyle.Render("Press Esc to go back | q: quit")

	content := lipgloss.JoinVertical(lipgloss.Center,
		styledResponse,
		"",
		help,
	)

	boxedContent := boxStyle.Render(content)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, boxedContent)
}
 
func (m Model) renderPayloadScreen() string {
	payloadWithCursor := m.Payload[:m.PayloadCursor] + "|" + m.Payload[m.PayloadCursor:]

	payloadStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#282828")).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4ECDC4")).
		Width(50).
		Height(20).
		AlignHorizontal(lipgloss.Left).
		AlignVertical(lipgloss.Top)

	help := HelpStyle.Render("Edit JSON | ↑: back | q: quit")

	content := lipgloss.JoinVertical(lipgloss.Center,
		TitleStyle.Render("JSON Payload"),
		"",
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

