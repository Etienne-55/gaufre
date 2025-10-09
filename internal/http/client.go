package http

import (
	"io"
	"fmt"
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

// MakeRequest will handle all HTTP methods (for future expansion)
func MakeRequest(req *types.Request) *types.Response {
	// TODO: Implement POST, PUT, DELETE, etc.
	// For now, just delegate to GET
	if req.Method == "GET" {
		return MakeGetRequest(req.URL)
	}
	
	return &types.Response{
		Error: fmt.Errorf("method %s not implemented yet", req.Method),
	}
}

