package http

import (
	"bytes"
	"gaufre/internal/types"
	"io"
	"net/http"
	"strings"
	"time"
)


func MakePostRequest(url string, jsonBody string, authToken string) *types.Response {
	start := time.Now()
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	payload := bytes.NewBufferString(jsonBody)
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return &types.Response{
			Error: err,
			ResponseTime: time.Since(start),
		}
	}
	req.Header.Set("Content-type", "application/json")

	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(authToken))
	}

	resp, err := client.Do(req)
	if err != nil {
		return &types.Response{
			Error: err,
			ResponseTime: time.Since(start),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &types.Response{
			StatusCode: resp.StatusCode,
			Error: err,
			ResponseTime: time.Since(start),
		}
	}

	return &types.Response{
		Body: string(body),
		StatusCode: resp.StatusCode,
		Headers: resp.Header,
		ResponseTime: time.Since(start),
	}
}

