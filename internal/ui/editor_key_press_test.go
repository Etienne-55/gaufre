package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "ctrl+f":
		return tea.KeyMsg{Type: tea.KeyCtrlF}
	case "ctrl+d":
		return tea.KeyMsg{Type: tea.KeyCtrlD}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func press(m Model, keys ...string) Model {
	for _, k := range keys {
		next, _ := m.Update(keyMsg(k))
		m = next.(Model)
	}
	return m
}

func TestOpeningEditorsFocusesTheField(t *testing.T) {
	m := sizedModel(100, 30)

	// method selector -> URL -> Body/Auth menu, then open Body.
	body := press(m, "down", "down", "enter")
	if body.Screen != ScreenPayloadEditor {
		t.Fatalf("Screen = %v, want ScreenPayloadEditor", body.Screen)
	}
	if !body.PayloadArea.Focused() {
		t.Error("body editor opened without focusing the textarea")
	}

	auth := press(m, "down", "down", "right", "enter")
	if auth.Screen != ScreenAuthEditor {
		t.Fatalf("Screen = %v, want ScreenAuthEditor", auth.Screen)
	}
	if !auth.AuthInput.Focused() {
		t.Error("token editor opened without focusing the input")
	}
}

func TestEditorsAcceptLettersThatAreGlobalShortcutsElsewhere(t *testing.T) {
	m := sizedModel(100, 30)

	body := press(m, "down", "down", "enter", "q", "u", "e", "r", "y")
	if got := body.PayloadArea.Value(); got != "query" {
		t.Errorf("body value = %q, want %q (q must type, not quit)", got, "query")
	}
	if body.Screen != ScreenPayloadEditor {
		t.Errorf("typing in the body editor left the screen: %v", body.Screen)
	}

	auth := press(m, "down", "down", "right", "enter", "q", "1")
	if got := auth.AuthInput.Value(); got != "q1" {
		t.Errorf("token value = %q, want %q", got, "q1")
	}
}

func TestPayloadEditorFormatsJSON(t *testing.T) {
	m := sizedModel(100, 30)
	m.Screen = ScreenPayloadEditor
	m.PayloadArea.Focus()
	m.PayloadArea.SetValue(`{"a":1,"b":2}`)

	m = press(m, "ctrl+f")

	want := "{\n  \"a\": 1,\n  \"b\": 2\n}"
	if got := m.PayloadArea.Value(); got != want {
		t.Errorf("body after ctrl+f =\n%q\nwant:\n%q", got, want)
	}
	if m.Status != "formatted" {
		t.Errorf("Status = %q, want %q", m.Status, "formatted")
	}
}

func TestPayloadEditorReportsFormatFailure(t *testing.T) {
	m := sizedModel(100, 30)
	m.Screen = ScreenPayloadEditor
	m.PayloadArea.Focus()
	m.PayloadArea.SetValue(`{"a": }`)

	m = press(m, "ctrl+f")

	if got := m.PayloadArea.Value(); got != `{"a": }` {
		t.Errorf("invalid body was modified: %q", got)
	}
	if !strings.HasPrefix(m.Status, "can't format") {
		t.Errorf("Status = %q, want a format failure message", m.Status)
	}
}

func TestPayloadEditorIndentsWithTab(t *testing.T) {
	m := sizedModel(100, 30)
	m.Screen = ScreenPayloadEditor
	m.PayloadArea.Focus()

	m = press(m, "tab", "a")

	if got := m.PayloadArea.Value(); got != "  a" {
		t.Errorf("body = %q, want %q", got, "  a")
	}
}

func TestEditorsClearAndReturn(t *testing.T) {
	m := sizedModel(100, 30)
	m.Screen = ScreenPayloadEditor
	m.PayloadArea.Focus()
	m.PayloadArea.SetValue("{}")

	m = press(m, "ctrl+d")
	if got := m.PayloadArea.Value(); got != "" {
		t.Errorf("body after ctrl+d = %q, want empty", got)
	}

	m = press(m, "esc")
	if m.Screen != ScreenMain || m.MainFocus != FocusURL {
		t.Errorf("after esc: screen=%v focus=%v, want main/URL", m.Screen, m.MainFocus)
	}
	if m.PayloadArea.Focused() {
		t.Error("textarea still focused after leaving the editor")
	}
	if m.Status != "" {
		t.Errorf("Status = %q, want it cleared on exit", m.Status)
	}

	m.Screen = ScreenAuthEditor
	m.AuthInput.Focus()
	m.AuthInput.SetValue("tok")
	m = press(m, "ctrl+d")
	if got := m.AuthInput.Value(); got != "" {
		t.Errorf("token after ctrl+d = %q, want empty", got)
	}

	m = press(m, "esc")
	if m.Screen != ScreenMain || m.MainFocus != FocusPayloadMenu {
		t.Errorf("after esc: screen=%v focus=%v, want main/payload menu", m.Screen, m.MainFocus)
	}
	if m.AuthInput.Focused() {
		t.Error("token input still focused after leaving the editor")
	}
}

func TestStatusMessageExpiry(t *testing.T) {
	m := sizedModel(100, 30)
	m, _ = m.setStatus("copied")

	if m.Status != "copied" {
		t.Fatalf("Status = %q", m.Status)
	}

	// A stale tick from an earlier message must not clear a newer one.
	stale := statusExpiredMsg{seq: m.StatusSeq - 1}
	next, _ := m.Update(stale)
	if next.(Model).Status != "copied" {
		t.Error("a stale expiry cleared the current status message")
	}

	next, _ = m.Update(statusExpiredMsg{seq: m.StatusSeq})
	if got := next.(Model).Status; got != "" {
		t.Errorf("Status = %q after its own expiry, want empty", got)
	}
}
