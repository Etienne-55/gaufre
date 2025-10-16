package types

import (
	"time"
)


type HistoryItem struct {
	Method string
	URL string
	Response int
	Timestamp time.Time
	Payload string 
}

func (h HistoryItem) FilterValue() string { return h.URL + " " + h.Method }
func (h HistoryItem) Title() string       { return h.Method + " " + h.URL }
func (h HistoryItem) Description() string {
	return ""
}

