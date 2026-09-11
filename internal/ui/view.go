package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const (
	// Bounds for the URL field's width. Past the maximum a long URL
	// scrolls inside the field instead of stretching the main box past the
	// edge of the terminal.
	urlMinWidth = 42
	urlMaxWidth = 64

	// Columns the main screen spends around the URL text: the main box's
	// border and padding, the URL box's border and padding, the "URL: "
	// prompt and the cursor.
	urlChrome = 16

	// Columns the main box spends on its border and side padding.
	mainBoxChrome = 6
)

func (m Model) View() string {

	switch m.Screen {
	case ScreenResponse:
		if m.Response != nil {
			return m.renderResponseScreen()
		}
	case ScreenTokenPicker:
		return m.renderTokenPickerScreen()
	case ScreenHistory:
		return m.renderHistoryScreen()
	case ScreenPayloadEditor:
		return m.renderPayloadScreen()
	case ScreenAuthEditor:
		return m.renderAuthscreen()
	}

	return m.renderMainScreen()
}

// renderMainScreen draws the main box, giving up the logo and then some
// padding when the terminal is too small for the full layout - which used
// to be taller than a 30-row terminal and lose its top border.
func (m Model) renderMainScreen() string {
	layouts := []struct {
		logo    bool
		padding int // vertical; the sides always keep 2 columns
	}{
		{logo: true, padding: 2},
		{logo: false, padding: 2},
		{logo: false, padding: 1},
		{logo: false, padding: 0},
	}

	var box string
	for _, layout := range layouts {
		box = m.renderMainBox(layout.logo, layout.padding)
		if m.Height <= 0 || (lipgloss.Height(box) <= m.Height && lipgloss.Width(box) <= m.Width) {
			break
		}
	}

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) renderMainBox(showLogo bool, padding int) string {
	help := m.Help
	help.Width = maxInt(m.Width-mainBoxChrome, 0)

	var rows []string
	if showLogo {
		rows = append(rows, RenderLogo(), "")
	}
	rows = append(rows,
		RenderMethodButtons(m.SelectedMethod, m.MainFocus == FocusMethod),
		"",
		m.renderURLInput(),
		"",
		RenderPayloadButtons(m.PayloadMenu, m.MainFocus == FocusPayloadMenu),
		"",
		m.renderMainStatus(),
		"",
		help.View(m.mainHelp()),
	)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(padding, 2)

	return boxStyle.Render(lipgloss.JoinVertical(lipgloss.Center, rows...))
}

// renderURLInput draws the URL field, showing its cursor only while it has
// focus. The box has a fixed width: the placeholder shown for an empty URL
// is shorter than the field, and the box used to shrink around it.
func (m Model) renderURLInput() string {
	input := m.URLInput
	borderColor := lipgloss.Color("#4ECDC4")
	if m.MainFocus != FocusURL {
		input.Blur()
		borderColor = lipgloss.Color("#874BFD")
	}

	const padding = 1
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, padding).
		Width(lipgloss.Width(input.Prompt) + input.Width + 1 + 2*padding). // +1 for the cursor
		Render(input.View())
}

// urlInputWidth sizes the URL field to the terminal, within bounds that
// keep the main box in proportion with the logo.
func (m Model) urlInputWidth() int {
	if m.Width <= 0 {
		return urlMinWidth
	}
	preferred := clampInt(m.Width-48, urlMinWidth, urlMaxWidth)
	return maxInt(minInt(preferred, m.Width-urlChrome), 10)
}

// renderMainStatus shows the spinner while a request is in flight, and
// otherwise a one-line summary of what will be sent along with it - most
// usefully whether the bearer token has expired, before you get the 401.
func (m Model) renderMainStatus() string {
	if m.Loading {
		return fmt.Sprintf("Loading %s", m.Spinner.View())
	}

	var parts []string

	if body := m.PayloadArea.Value(); strings.TrimSpace(body) != "" {
		switch {
		case Methods[m.SelectedMethod] == "GET":
			parts = append(parts, EditorWarnStyle.Render("body not sent with GET"))
		case !inspectBody(body).Valid:
			parts = append(parts, EditorBadStyle.Render("body: invalid JSON"))
		default:
			parts = append(parts, EditorMetaStyle.Render("body: "+formatBytes(len(body))))
		}
	}

	if token := strings.TrimSpace(m.AuthInput.Value()); token != "" {
		parts = append(parts, authSummary(token, time.Now()))
	}

	return strings.Join(parts, EditorMetaStyle.Render(" · "))
}

// authSummary describes the bearer token in a few words for the main
// screen.
func authSummary(token string, now time.Time) string {
	jwt, ok := decodeJWT(token)
	switch {
	case !ok:
		return EditorMetaStyle.Render("auth: token set")
	case jwt.ExpiresAt.IsZero():
		return EditorMetaStyle.Render("auth: JWT")
	case jwt.Expired(now):
		return EditorBadStyle.Render("auth: JWT expired " + relativeTime(jwt.ExpiresAt, now))
	case jwt.ExpiresAt.Sub(now) < 5*time.Minute:
		return EditorWarnStyle.Render("auth: JWT expires " + relativeTime(jwt.ExpiresAt, now))
	default:
		return EditorMetaStyle.Render("auth: JWT, expires " + relativeTime(jwt.ExpiresAt, now))
	}
}

// mainHelp lists the shortcuts for the focused control: the arrows choose
// a method on the method row but move the cursor in the URL, and q types a
// q in the URL instead of quitting.
func (m Model) mainHelp() helpBindings {
	move := MainKeys.Up
	move.SetHelp("↑/↓", "move")
	choose := MainKeys.Choose
	enter := MainKeys.Enter

	switch m.MainFocus {
	case FocusURL:
		return helpBindings{move, enter, MainKeys.History, MainKeys.ForceQuit}
	case FocusPayloadMenu:
		choose.SetHelp("←/→", "body/auth")
		enter.SetHelp("enter", "open")
	default:
		choose.SetHelp("←/→", "method")
	}
	return helpBindings{choose, move, enter, MainKeys.History, MainKeys.Quit}
}

func (m Model) renderHistoryScreen() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4ECDC4")).
		Padding(1, 2).
		Width(m.historyBoxWidth())

	boxedContent := boxStyle.Render(m.HistoryList.View())
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, boxedContent)
}

// historyBoxWidth is the width of the history frame, border excluded. The
// list is sized to exactly fit inside it: it used to be sized to the whole
// terminal inside a fixed 60-column box, which re-wrapped and centered every
// entry into a jumble.
func (m Model) historyBoxWidth() int {
	preferred := clampInt(m.Width-10, 50, 100)
	return maxInt(minInt(preferred, m.Width-2), 20)
}

func (m Model) historyListSize() (int, int) {
	const chromeRows = 5 // borders, vertical padding, and a spare row
	return m.historyBoxWidth() - 4, maxInt(m.Height-chromeRows, 5)
}
