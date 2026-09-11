package ui

import (
	"strings"
	"testing"
	"time"

	"gaufre/internal/types"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// withColors makes lipgloss emit colors for the rest of the test. Without a
// terminal it strips them, which would make color assertions pass vacuously.
func withColors(t *testing.T) {
	t.Helper()
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })
}

func historyModel(items ...types.HistoryItem) Model {
	m := sizedModel(100, 30)
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}
	m.History = items
	m.HistoryList.SetItems(listItems)
	m.Screen = ScreenHistory
	return m
}

func TestHistoryColorsStatusCodes(t *testing.T) {
	withColors(t)

	when := time.Date(2026, 9, 11, 10, 52, 0, 0, time.Local)
	m := historyModel(
		types.HistoryItem{Method: "GET", URL: "http://api.test/ok", Response: 200, Timestamp: when},
		types.HistoryItem{Method: "GET", URL: "http://api.test/moved", Response: 302, Timestamp: when},
		types.HistoryItem{Method: "GET", URL: "http://api.test/missing", Response: 404, Timestamp: when},
		types.HistoryItem{Method: "POST", URL: "http://api.test/broken", Response: 503, Timestamp: when},
	)
	out := m.View()

	for status, color := range map[string]lipgloss.Color{
		"200": ColorOK, "302": ColorWarn, "404": ColorBad, "503": ColorBad,
	} {
		if want := lipgloss.NewStyle().Foreground(color).Render(status); !strings.Contains(out, want) {
			t.Errorf("status %s isn't drawn in its class color", status)
		}
	}

	// Coloring must not change the text itself.
	plain := stripANSI(out)
	for _, want := range []string{"│ GET http://api.test/ok", "│ 200 · Sep 11 10:52", "404 · Sep 11 10:52", "POST http://api.test/broken"} {
		if !strings.Contains(plain, want) {
			t.Errorf("history is missing %q:\n%s", want, plain)
		}
	}
}

func TestHistoryDimsStatusCodesWhileFilterIsEmpty(t *testing.T) {
	withColors(t)

	m := historyModel(types.HistoryItem{Method: "GET", URL: "http://api.test/missing", Response: 404})
	m = press(m, "/")
	if m.HistoryList.FilterState() != list.Filtering {
		t.Fatal("/ didn't open the filter")
	}

	if strings.Contains(m.View(), lipgloss.NewStyle().Foreground(ColorBad).Render("404")) {
		t.Error("status stays colored while the list is dimmed for an empty filter")
	}
}

func TestHistoryEntryWithoutStatus(t *testing.T) {
	m := historyModel(types.HistoryItem{Method: "GET", URL: "http://api.test/legacy"})
	if plain := stripANSI(m.View()); !strings.Contains(plain, "│ —") {
		t.Errorf("entry without a status should show a dash:\n%s", plain)
	}
}

func TestStatusCodeColor(t *testing.T) {
	tests := []struct {
		code  int
		color lipgloss.Color
		ok    bool
	}{
		{0, "", false},
		{101, "", false},
		{200, ColorOK, true},
		{204, ColorOK, true},
		{301, ColorWarn, true},
		{401, ColorBad, true},
		{500, ColorBad, true},
	}
	for _, tt := range tests {
		color, ok := statusCodeColor(tt.code)
		if color != tt.color || ok != tt.ok {
			t.Errorf("statusCodeColor(%d) = %q, %v; want %q, %v", tt.code, color, ok, tt.color, tt.ok)
		}
	}
}

func TestHistoryItemText(t *testing.T) {
	when := time.Date(2026, 9, 11, 10, 52, 0, 0, time.Local)
	tests := []struct {
		item types.HistoryItem
		want string
	}{
		{types.HistoryItem{Response: 200, Timestamp: when}, "200 · Sep 11 10:52"},
		{types.HistoryItem{Response: 404}, "404"},
		{types.HistoryItem{Timestamp: when}, "— · Sep 11 10:52"},
		{types.HistoryItem{}, "—"},
	}
	for _, tt := range tests {
		if got := tt.item.Description(); got != tt.want {
			t.Errorf("Description() = %q, want %q", got, tt.want)
		}
	}

	// Filter matches are highlighted by position in the title, so the two
	// have to be the same string.
	item := types.HistoryItem{Method: "POST", URL: "http://api.test/login"}
	if item.FilterValue() != item.Title() {
		t.Errorf("FilterValue() = %q, Title() = %q; highlights would land on the wrong characters", item.FilterValue(), item.Title())
	}
}
