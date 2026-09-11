package ui

import (
	"fmt"
	"io"

	"gaufre/internal/types"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// historyDelegate draws history entries the way the list's default delegate
// does - same styles, selection bar and filter highlighting - but colors
// each entry's status code by class, so failed requests stand out while
// scanning the history.
type historyDelegate struct {
	list.DefaultDelegate
}

func newHistoryDelegate() historyDelegate {
	return historyDelegate{DefaultDelegate: list.NewDefaultDelegate()}
}

func (d historyDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	h, ok := item.(types.HistoryItem)
	if !ok || m.Width() <= 0 {
		return
	}

	s := &d.Styles
	textWidth := m.Width() - s.NormalTitle.GetPaddingLeft() - s.NormalTitle.GetPaddingRight()
	title := truncateString(h.Title(), textWidth)

	var (
		isSelected  = index == m.Index()
		emptyFilter = m.FilterState() == list.Filtering && m.FilterValue() == ""
		isFiltered  = m.FilterState() == list.Filtering || m.FilterState() == list.FilterApplied
	)

	// While the filter prompt is open and empty, everything is dimmed,
	// status codes included.
	if emptyFilter {
		fmt.Fprintf(w, "%s\n%s", s.DimmedTitle.Render(title),
			s.DimmedDesc.Render(truncateString(h.Description(), textWidth)))
		return
	}

	titleStyle, descStyle := s.NormalTitle, s.NormalDesc
	if isSelected && m.FilterState() != list.Filtering {
		titleStyle, descStyle = s.SelectedTitle, s.SelectedDesc
	}

	if isFiltered {
		unmatched := titleStyle.Inline(true)
		matched := unmatched.Inherit(s.FilterMatch)
		title = lipgloss.StyleRunes(title, m.MatchesForItem(index), matched, unmatched)
	}

	fmt.Fprintf(w, "%s\n%s", titleStyle.Render(title), renderHistoryDescription(h, descStyle, textWidth))
}

// renderHistoryDescription draws "404 · Sep 11 10:52" with the status code
// in its class color and the rest in the description style. The segments
// are styled separately and only then framed by the style's border and
// padding, since styling the whole line would recolor the status code.
func renderHistoryDescription(h types.HistoryItem, style lipgloss.Style, width int) string {
	inline := style.Inline(true)

	status := h.Status()
	statusStyle := inline
	if color, ok := statusCodeColor(h.Response); ok {
		statusStyle = inline.Foreground(color)
	}

	rest := ""
	if when := h.When(); when != "" {
		rest = " · " + when
	}
	rest = truncateString(rest, width-len([]rune(status)))

	line := statusStyle.Render(status)
	if rest != "" {
		line += inline.Render(rest)
	}
	return style.UnsetForeground().Render(line)
}

// statusCodeColor picks a color for a status code from the editors'
// palette: green for success, amber for redirects, red for client and
// server errors. Anything else (no response recorded, 1xx) keeps the
// description's own color.
func statusCodeColor(code int) (lipgloss.Color, bool) {
	switch {
	case code >= 400:
		return ColorBad, true
	case code >= 300:
		return ColorWarn, true
	case code >= 200:
		return ColorOK, true
	}
	return "", false
}
