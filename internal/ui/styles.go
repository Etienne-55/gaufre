package ui

import "github.com/charmbracelet/lipgloss"


var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF00")).
			MarginBottom(1)

	InputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	ResponseStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1).
			MarginTop(1)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00"))

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)
)

func RenderMethodButtons(selected int, focused bool) string {
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	var buttons []string

	for i, method := range methods {
		if focused && i == selected {
			btn := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#874BFD")).
				Bold(true).
				Padding(0, 2).
				Render(method)
			buttons = append(buttons, btn)
		} else if !focused && i == selected {
			btn := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#00FF00")).
				Bold(true).
				Padding(0, 2).
				Render(method)
			buttons = append(buttons, btn)
		} else {
			btn := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#874BFD")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#874BFD")).
				Padding(0, 2).
				Render(method)
			buttons = append(buttons, btn)
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, buttons...)
}

func RenderURLInput(url string, cursor int, focused bool) string {
	urlLabel := "URL: "
	urlWithCursor := url[:cursor] + "|" + url[cursor:]

	var urlStyle lipgloss.Style
	if focused {
		urlStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			// Background(lipgloss.Color("#874BFD")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4ECDC4")).
			Padding(0, 1)
	} else {
		urlStyle = InputStyle.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(0, 1)
	}

	return urlStyle.Render(urlLabel + urlWithCursor)
}

func RenderPayloadButtons(selected int, focused bool) string {
	methods := []string{"Body", "Auth"}
	var buttons []string

	for i, method := range methods {
		if focused && i == selected {
			btn := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#874BFD")).
				Bold(true).
				Padding(0, 2).
				Render(method)
			buttons = append(buttons, btn)
		} else {
			btn := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#874BFD")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#874BFD")).
				Padding(0, 2).
				Render(method)
			buttons = append(buttons, btn)
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, buttons...)
}

