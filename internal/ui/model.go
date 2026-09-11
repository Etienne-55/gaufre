package ui

import (
	"gaufre/internal/storage"
	"gaufre/internal/types"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Methods is the ordered list of HTTP methods the method selector cycles
// through. SelectedMethod is an index into it; kept as a single shared list
// so every screen that needs the method names (sending a request, matching
// a history entry back to its method, ...) stays in sync.
var Methods = []string{"GET", "POST", "PUT", "DELETE"}

// Screen identifies which full-screen view is currently active.
type Screen int

const (
	ScreenMain Screen = iota
	ScreenPayloadEditor
	ScreenAuthEditor
	ScreenHistory
	ScreenResponse
	ScreenTokenPicker
)

// MainFocus identifies which control on the main screen has focus. It is
// only meaningful while Screen == ScreenMain.
type MainFocus int

const (
	FocusMethod MainFocus = iota
	FocusURL
	FocusPayloadMenu
)

type Model struct {
	Screen    Screen
	MainFocus MainFocus

	// URLInput is the request URL field. Like the token field it is a
	// textinput, so editing gets proper cursor movement, word deletion and
	// horizontal scrolling for long URLs.
	URLInput textinput.Model

	Response *types.Response
	Loading  bool

	// ResponseTokens are the token-looking values found in Response, best
	// first, so the response viewer can offer them to the auth editor.
	// TokenCursor is the selected entry while the token picker is open.
	ResponseTokens []tokenCandidate
	TokenCursor    int

	Width  int
	Height int

	SelectedMethod int

	PayloadArea textarea.Model
	AuthInput   textinput.Model
	PayloadMenu int // 0 = Body, 1 = Auth

	Help help.Model

	// Status is a transient message shown in the editors' status line after
	// an action (copy, paste, format). StatusSeq lets a pending expiry tick
	// know whether it is clearing its own message or a newer one.
	Status    string
	StatusSeq int

	Spinner     spinner.Model
	History     []types.HistoryItem
	HistoryList list.Model
	Viewport    viewport.Model

	// ShowHeaders lists the response headers above the body. It is a
	// preference, so it carries over from one response to the next.
	ShowHeaders bool
}

// newURLInput builds the URL field. It stays focused for its whole life:
// keys only reach it while the URL has focus on the main screen, and the
// view hides its cursor the rest of the time.
func newURLInput(url string) textinput.Model {
	ti := textinput.New()
	ti.Prompt = "URL: "
	ti.Placeholder = "https://api.example.com/..."
	ti.CharLimit = 0
	ti.Width = urlMinWidth

	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(ColorMuted)
	ti.Cursor.SetMode(cursor.CursorStatic)

	ti.Focus()
	ti.SetValue(url)
	ti.CursorEnd()
	return ti
}

// newPayloadArea builds the multi-line JSON body editor.
func newPayloadArea() textarea.Model {
	ta := textarea.New()
	ta.Placeholder = "{\n  \"key\": \"value\"\n}"
	ta.ShowLineNumbers = true
	ta.Prompt = ""
	ta.CharLimit = 0
	ta.SetWidth(editorMinContentWidth)
	ta.SetHeight(editorMinBodyHeight)

	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.FocusedStyle.Text = lipgloss.NewStyle().Foreground(ColorText)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Foreground(ColorText)
	ta.FocusedStyle.LineNumber = lipgloss.NewStyle().Foreground(ColorMuted)
	ta.FocusedStyle.CursorLineNumber = lipgloss.NewStyle().Foreground(ColorAccent)
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(ColorMuted)
	ta.FocusedStyle.EndOfBuffer = lipgloss.NewStyle().Foreground(ColorBackground)
	ta.BlurredStyle = ta.FocusedStyle

	ta.Blur()
	return ta
}

// newAuthInput builds the bearer token field. A token is a single-line
// value, so this is a textinput (which scrolls horizontally) rather than a
// textarea: a multi-line widget both wrapped long tokens across lines and
// allowed newlines that would produce an invalid Authorization header.
func newAuthInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "paste a token with ctrl+v, or start typing"
	ti.Prompt = authPrompt
	ti.CharLimit = 0
	ti.Width = editorMinContentWidth - len(authPrompt) - 1

	ti.PromptStyle = lipgloss.NewStyle().Foreground(ColorMuted)
	ti.TextStyle = lipgloss.NewStyle().Foreground(ColorText)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(ColorMuted)

	ti.Blur()
	return ti
}

func NewModel() Model {
	defaultURL := "http://localhost:"

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	historyList := list.New([]list.Item{}, newHistoryDelegate(), 80, 20)
	historyList.Title = "Request History"
	historyList.SetFilteringEnabled(true)
	historyList.SetShowStatusBar(false)

	history := []types.HistoryItem{}
	if loaded, err := storage.LoadHistory(); err == nil {
		history = loaded
		items := make([]list.Item, len(history))
		for i, h := range history {
			items[i] = h
		}
		historyList.SetItems(items)
	}

	helpModel := help.New()
	helpModel.Styles.ShortKey = lipgloss.NewStyle().Foreground(ColorAccent)
	helpModel.Styles.ShortDesc = lipgloss.NewStyle().Foreground(ColorMuted)
	helpModel.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(ColorMuted)

	return Model{
		Screen:      ScreenMain,
		MainFocus:   FocusMethod,
		URLInput:    newURLInput(defaultURL),
		PayloadArea: newPayloadArea(),
		AuthInput:   newAuthInput(),
		Help:        helpModel,
		Spinner:     s,
		History:     history,
		HistoryList: historyList,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
