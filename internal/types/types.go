package types

import "time"


type Request struct {
	Method string
	URL	string
	Headers map[string]string
	Body string
}

type Response struct {
	Body string
	StatusCode int
	Headers map[string][]string
	ResponseTime time.Duration
	Error error
}

