package types

import (
	"fmt"
	"time"
)


type HistoryItem struct {
	Method string
	URL string
	Response int
	Timestamp time.Time
}

func (h HistoryItem) FilterValue() string { return h.URL + " " + h.Method }
func (h HistoryItem) Title() string       { return h.Method + " " + h.URL }
func (h HistoryItem) Description() string {
	return "status: " + fmt.Sprintf("%d", h.Response) + " | " + h.Timestamp.Format("15:04:05")
}

