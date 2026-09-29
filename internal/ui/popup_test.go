package ui

import (
	"strings"
	"testing"
)

const miniLogoRow = "█▄█ █▀█ █▄█ █▀  █▀▄ ██▄"

// popupScreens is one model per screen, at the given size.
func popupScreens(t *testing.T, width, height int) map[string]Model {
	t.Helper()
	screens := map[string]Model{
		"main":     sizedModel(width, height),
		"response": responseModel(t, width, height, `{"access_token":"abc.def.ghi","ok":true}`),
	}
	for name, screen := range map[string]Screen{
		"history": ScreenHistory,
		"body":    ScreenPayloadEditor,
		"auth":    ScreenAuthEditor,
	} {
		m := sizedModel(width, height)
		m.Screen = screen
		screens[name] = m
	}
	return screens
}

func TestPopupDrawsOnlyThePopupsFrame(t *testing.T) {
	for name, m := range popupScreens(t, 120, 40) {
		framed, _ := m.renderScreen()
		m.Popup = true
		frameless, _ := m.renderScreen()

		// The screen's own box is the one that goes: the buttons and the
		// URL field inside it keep their borders.
		if got, want := strings.Count(frameless, "╭"), strings.Count(framed, "╭")-1; got != want {
			t.Errorf("%s: %d boxes in the popup, want %d:\n%s", name, got, want, stripANSI(frameless))
		}

		// The border is hidden, not removed, so nothing moves.
		framedLines, framelessLines := strings.Split(framed, "\n"), strings.Split(frameless, "\n")
		if len(framedLines) != len(framelessLines) {
			t.Fatalf("%s: %d lines in the popup, %d without", name, len(framelessLines), len(framedLines))
		}
		for i := range framedLines {
			if a, b := len([]rune(stripANSI(framedLines[i]))), len([]rune(stripANSI(framelessLines[i]))); a != b {
				t.Errorf("%s: line %d is %d cells in the popup, %d without", name, i, b, a)
			}
		}
	}
}

func TestPopupShowsTheMiniLogoInTheCorner(t *testing.T) {
	m := sizedModel(120, 40)
	m.Screen = ScreenHistory
	m.Popup = true

	lines := strings.Split(stripANSI(m.View()), "\n")
	if last := strings.TrimRight(lines[len(lines)-1], " "); !strings.HasSuffix(last, miniLogoRow) {
		t.Errorf("the mini logo isn't in the bottom-right corner:\n%s", strings.Join(lines, "\n"))
	}

	m.Popup = false
	if strings.Contains(stripANSI(m.View()), miniLogoRow) {
		t.Error("the mini logo is shown outside the popup")
	}
}

func TestPopupSkipsTheMiniLogoNextToTheFullOne(t *testing.T) {
	m := sizedModel(120, 40)
	m.Popup = true

	view := stripANSI(m.View())
	if !strings.Contains(view, "██████╗") {
		t.Fatal("the main screen should show the full logo at 120x40")
	}
	if strings.Contains(view, miniLogoRow) {
		t.Errorf("both logos are on the main screen:\n%s", view)
	}
}

// The mini logo is only ever drawn over blank rows: on a screen that needs
// the whole height it is left out rather than covering anything.
func TestPopupMiniLogoNeverCoversTheScreen(t *testing.T) {
	for _, size := range [][2]int{{120, 40}, {100, 30}, {80, 24}, {60, 18}} {
		for name, m := range popupScreens(t, size[0], size[1]) {
			m.Popup = true
			bare, _ := m.renderScreen()
			stamped := m.View()

			bareLines, stampedLines := strings.Split(bare, "\n"), strings.Split(stamped, "\n")
			if len(stampedLines) != len(bareLines) {
				t.Fatalf("%s at %dx%d: the mini logo changed the height", name, size[0], size[1])
			}
			for i := range bareLines {
				if w := len([]rune(stripANSI(stampedLines[i]))); w > size[0] {
					t.Errorf("%s at %dx%d: line %d is %d cells wide", name, size[0], size[1], i, w)
				}
				if bareLines[i] != stampedLines[i] && strings.TrimSpace(stripANSI(bareLines[i])) != "" {
					t.Errorf("%s at %dx%d: the mini logo covered line %d: %q",
						name, size[0], size[1], i, stripANSI(bareLines[i]))
				}
			}
		}
	}
}
