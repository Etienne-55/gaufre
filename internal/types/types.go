package types

import "time"

type Response struct {
	Body         string
	StatusCode   int
	Headers      map[string][]string
	ResponseTime time.Duration
	Error        error
}
