package ui

import (
	"strings"
	"testing"
	"time"

	"gaufre/internal/http"
	"gaufre/internal/types"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// pressCmd is press for a single key, keeping the command it returned.
func pressCmd(m Model, k string) (Model, tea.Cmd) {
	next, cmd := m.Update(keyMsg(k))
	return next.(Model), cmd
}

func quits(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

// assertFits checks a screen's rendering stays inside the terminal. Unlike
// assertFrameIntact it allows boxes inside the frame (the main screen's
// buttons are boxes too).
func assertFits(t *testing.T, out string, width, height int) {
	t.Helper()

	lines := strings.Split(out, "\n")
	if len(lines) > height {
		t.Errorf("rendered %d lines, terminal height is %d", len(lines), height)
	}
	for i, line := range lines {
		if w := len([]rune(stripANSI(line))); w > width {
			t.Errorf("line %d is %d cells wide, terminal width is %d", i, w, width)
		}
	}
	// The first thing on screen must be the box's top border, not a row
	// from its middle.
	for _, line := range lines {
		plain := strings.TrimSpace(stripANSI(line))
		if plain == "" {
			continue
		}
		if !strings.HasPrefix(plain, "╭") {
			t.Errorf("the top border isn't on screen:\n%s", stripANSI(out))
		}
		break
	}
}

func TestURLBarTypesLettersThatAreShortcutsElsewhere(t *testing.T) {
	m := sizedModel(100, 30)
	m.URLInput.SetValue("http://api.test/search?")
	m.URLInput.CursorEnd()

	m = press(m, "down") // method selector -> URL
	for _, k := range []string{"q", "=", "g", "o"} {
		var cmd tea.Cmd
		m, cmd = pressCmd(m, k)
		if quits(cmd) {
			t.Fatalf("typing %q in the URL bar quit the app", k)
		}
	}

	if got := m.URLInput.Value(); got != "http://api.test/search?q=go" {
		t.Errorf("URL = %q, want the typed query appended", got)
	}
}

func TestMainScreenQuitKeys(t *testing.T) {
	m := sizedModel(100, 30)

	if _, cmd := pressCmd(m, "q"); !quits(cmd) {
		t.Error("q on the method selector should quit")
	}

	url := press(m, "down")
	if _, cmd := pressCmd(url, "ctrl+c"); !quits(cmd) {
		t.Error("ctrl+c should quit even while typing a URL")
	}
}

func TestURLBarEditing(t *testing.T) {
	m := sizedModel(100, 30)
	m.URLInput.SetValue("http://api.test/users")
	m.URLInput.CursorEnd()

	m = press(m, "down", "left", "left", "left", "left", "left", "x")
	if got := m.URLInput.Value(); got != "http://api.test/xusers" {
		t.Errorf("URL = %q, want an x inserted at the cursor", got)
	}

	// Multi-byte characters used to be sliced mid-rune by a byte cursor.
	m.URLInput.SetValue("http://api.test/café/1")
	m.URLInput.CursorEnd()
	m = press(m, "left", "left", "left", "x")
	if got := m.URLInput.Value(); got != "http://api.test/cafxé/1" {
		t.Errorf("URL = %q, want x inserted before the é", got)
	}
}

func TestMainScreenArrowsDependOnFocus(t *testing.T) {
	m := sizedModel(100, 30)
	m.SelectedMethod = 0

	m = press(m, "right", "right")
	if m.SelectedMethod != 2 {
		t.Errorf("SelectedMethod = %d, want 2 after two rights on the method row", m.SelectedMethod)
	}
	m = press(m, "right", "right", "right")
	if m.SelectedMethod != len(Methods)-1 {
		t.Errorf("SelectedMethod = %d, want it to stop at the last method", m.SelectedMethod)
	}

	m = press(m, "down", "down", "right")
	if m.PayloadMenu != 1 {
		t.Errorf("PayloadMenu = %d, want Auth selected", m.PayloadMenu)
	}
	if m.SelectedMethod != len(Methods)-1 {
		t.Error("arrows on the Body/Auth row changed the method")
	}
}

func TestResponseArrivingOffTheMainScreenIsShown(t *testing.T) {
	for _, screen := range []Screen{ScreenHistory, ScreenPayloadEditor, ScreenAuthEditor} {
		m := sizedModel(100, 30)
		m.Loading = true
		m.Screen = screen

		next, _ := m.Update(http.ResponseMsg{Response: &types.Response{Body: `{"ok":true}`, StatusCode: 200}})
		m = next.(Model)

		if m.Loading {
			t.Errorf("screen %v: Loading still set after the response arrived, every later send would be ignored", screen)
		}
		if m.Screen != ScreenResponse {
			t.Errorf("screen %v: Screen = %v, want the response", screen, m.Screen)
		}
		if m.PayloadArea.Focused() || m.AuthInput.Focused() {
			t.Errorf("screen %v: an editor kept focus behind the response", screen)
		}
	}
}

func TestSpinnerStopsWhenNotLoading(t *testing.T) {
	m := sizedModel(100, 30)

	if _, cmd := m.Update(spinner.TickMsg{}); cmd != nil {
		t.Error("spinner kept ticking with no request in flight")
	}

	m.Loading = true
	m.Screen = ScreenHistory
	if _, cmd := m.Update(spinner.TickMsg{}); cmd == nil {
		t.Error("spinner stopped ticking while loading on another screen")
	}
}

func TestMainScreenFitsTerminal(t *testing.T) {
	for _, size := range [][2]int{{120, 40}, {100, 30}, {80, 24}, {60, 18}} {
		m := sizedModel(size[0], size[1])
		assertFits(t, m.View(), size[0], size[1])

		// A long URL scrolls inside its field instead of widening the box.
		m.URLInput.SetValue("http://localhost:3000/api/v1/users?" + strings.Repeat("filter=value&", 20))
		assertFits(t, press(m, "down").View(), size[0], size[1])
	}

	// The logo is kept whenever there is room for it.
	if view := stripANSI(sizedModel(100, 30).View()); !strings.Contains(view, "██████╗") {
		t.Error("the logo was dropped on a 100x30 terminal")
	}
}

func TestMainScreenSummarisesWhatWillBeSent(t *testing.T) {
	m := sizedModel(100, 30)
	m.SelectedMethod = 0 // GET
	m.PayloadArea.SetValue(`{"a":1}`)
	m.AuthInput.SetValue(testJWT(t, map[string]any{"exp": float64(time.Now().Add(-3 * time.Minute).Unix())}))

	view := stripANSI(m.View())
	for _, want := range []string{"body not sent with GET", "auth: JWT expired"} {
		if !strings.Contains(view, want) {
			t.Errorf("main screen is missing %q", want)
		}
	}

	m.SelectedMethod = 1 // POST
	if view := stripANSI(m.View()); !strings.Contains(view, "body: 7 B") {
		t.Errorf("main screen should show the body size for a POST:\n%s", view)
	}
}

func TestButtonsKeepTheirSizeWhenFocused(t *testing.T) {
	for i := range Methods {
		focused, idle := RenderMethodButtons(i, true), RenderMethodButtons(i, false)
		if stripANSI(focused) == "" || len(strings.Split(focused, "\n")) != len(strings.Split(idle, "\n")) ||
			len([]rune(stripANSI(strings.Split(focused, "\n")[1]))) != len([]rune(stripANSI(strings.Split(idle, "\n")[1]))) {
			t.Errorf("method row changes size when focus enters it (selected %d)", i)
		}
	}
}

func TestHistorySelectionRespectsFilter(t *testing.T) {
	m := sizedModel(100, 30)
	m.History = []types.HistoryItem{
		{Method: "GET", URL: "http://api.test/users"},
		{Method: "POST", URL: "http://api.test/login"},
		{Method: "DELETE", URL: "http://api.test/users/1"},
	}
	items := make([]list.Item, len(m.History))
	for i, h := range m.History {
		items[i] = h
	}
	m.HistoryList.SetItems(items)
	m.Screen = ScreenHistory

	// Filtering down to "login" leaves one item, at filtered index 0 -
	// which is "users" in the unfiltered history.
	m.HistoryList.SetFilterText("login")
	m.HistoryList.SetFilterState(list.FilterApplied)

	m = press(m, "enter")
	if got := m.URLInput.Value(); got != "http://api.test/login" {
		t.Errorf("loaded %q, want the filtered entry http://api.test/login", got)
	}
	if Methods[m.SelectedMethod] != "POST" {
		t.Errorf("method = %s, want POST", Methods[m.SelectedMethod])
	}
}

func TestHistoryEscClearsFilterFirst(t *testing.T) {
	m := sizedModel(100, 30)
	m.HistoryList.SetItems([]list.Item{types.HistoryItem{Method: "GET", URL: "http://api.test/a"}})
	m.Screen = ScreenHistory
	m.HistoryList.SetFilterText("a")
	m.HistoryList.SetFilterState(list.FilterApplied)

	m = press(m, "esc")
	if m.Screen != ScreenHistory {
		t.Fatal("esc left the history instead of clearing the filter")
	}
	if m.HistoryList.FilterState() != list.Unfiltered {
		t.Errorf("filter state = %v, want cleared", m.HistoryList.FilterState())
	}

	m = press(m, "esc")
	if m.Screen != ScreenMain {
		t.Error("second esc should leave the history")
	}
}

func TestHistoryScreenFitsTerminal(t *testing.T) {
	for _, size := range [][2]int{{120, 40}, {100, 30}, {70, 20}} {
		m := sizedModel(size[0], size[1])
		m.HistoryList.SetItems([]list.Item{
			types.HistoryItem{Method: "GET", URL: "http://localhost:3000/api/v1/users?page=2&limit=50&sort=created_at&order=desc", Response: 200},
			types.HistoryItem{Method: "POST", URL: "http://localhost:3000/api/v1/auth/login", Response: 401},
		})
		m.Screen = ScreenHistory
		assertFrameIntact(t, m.View(), size[0], size[1])
	}
}

func TestURLBoxKeepsItsWidthWhenEmpty(t *testing.T) {
	m := sizedModel(100, 30)
	m.MainFocus = FocusURL

	filled := lipgloss.Width(m.renderURLInput())
	m.URLInput.SetValue("")
	if empty := lipgloss.Width(m.renderURLInput()); empty != filled {
		t.Errorf("URL box is %d wide when empty, %d with a URL", empty, filled)
	}

	m.MainFocus = FocusMethod
	if blurred := lipgloss.Width(m.renderURLInput()); blurred != filled {
		t.Errorf("URL box is %d wide without focus, %d with it", blurred, filled)
	}
}
