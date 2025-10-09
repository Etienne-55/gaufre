package http

import (
	// "fmt"
	"gaufre/internal/types"
	"io"
	"net/http"
	"time"
)


// func MakeGetRequest(url string) *types.Response {
// 	start := time.Now()
//
// 	client := &http.Client{
// 		Timeout: 10 * time.Second,
// 	}
//
// 	resp, err := client.Get(url)
// 	if err != nil {
// 		return &types.Response{
// 			Error:        err,
// 			ResponseTime: time.Since(start),
// 		}
// 	}
// 	defer resp.Body.Close()
//
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return &types.Response{
// 			StatusCode:   resp.StatusCode,
// 			Error:        err,
// 			ResponseTime: time.Since(start),
// 		}
// 	}
//
// 	return &types.Response{
// 		Body:         string(body),
// 		StatusCode:   resp.StatusCode,
// 		Headers:      resp.Header,
// 		ResponseTime: time.Since(start),
// 	}
//
// }
// func MakeRequest(req *types.Request) *types.Response {
// 	if req.Method == "GET" {
// 		return MakeGetRequest(req.URL)
// 	}
//
// 	return &types.Response{
// 		Error: fmt.Errorf("method %s not implemented yet", req.Method),
// 	}
// }

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

