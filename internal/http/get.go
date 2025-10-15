package http

import (
	"io"
	"time"
	"net/http"
	"gaufre/internal/types"
)


func MakeGetRequest(url string) *types.Response {
	start := time.Now()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return &types.Response{
			Error:        err,
			ResponseTime: time.Since(start),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &types.Response{
			StatusCode:   resp.StatusCode,
			Error:        err,
			ResponseTime: time.Since(start),
		}
	}

	return &types.Response{
		Body:         string(body),
		StatusCode:   resp.StatusCode,
		Headers:      resp.Header,
		ResponseTime: time.Since(start),
	}
}

