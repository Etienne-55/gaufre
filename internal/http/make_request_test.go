package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDo_MethodsAndAuth(t *testing.T) {
	tests := []struct {
		method string
		body   string
	}{
		{"GET", ""},
		{"POST", `{"a":1}`},
		{"PUT", `{"a":1}`},
		{"DELETE", ""},
		{"DELETE", `{"ids":[1,2]}`}, // DELETE with a body should still be sent
	}

	for _, tt := range tests {
		t.Run(tt.method+"/body="+tt.body, func(t *testing.T) {
			var gotMethod, gotAuth, gotBody, gotContentType string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotAuth = r.Header.Get("Authorization")
				gotContentType = r.Header.Get("Content-Type")
				b, _ := io.ReadAll(r.Body)
				gotBody = string(b)
				w.WriteHeader(http.StatusTeapot)
				w.Write([]byte(`{"ok":true}`))
			}))
			defer server.Close()

			resp := Do(tt.method, server.URL, tt.body, "  my-token  ")

			if resp.Error != nil {
				t.Fatalf("unexpected error: %v", resp.Error)
			}
			if gotMethod != tt.method {
				t.Errorf("method = %q, want %q", gotMethod, tt.method)
			}
			if resp.StatusCode != http.StatusTeapot {
				t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusTeapot)
			}
			if resp.Body != `{"ok":true}` {
				t.Errorf("Body = %q", resp.Body)
			}

			// The Bearer token must be applied and trimmed for every verb,
			// not just POST.
			if gotAuth != "Bearer my-token" {
				t.Errorf("Authorization header = %q, want %q", gotAuth, "Bearer my-token")
			}

			if tt.body != "" {
				if gotBody != tt.body {
					t.Errorf("request body = %q, want %q", gotBody, tt.body)
				}
				if gotContentType != "application/json" {
					t.Errorf("Content-Type = %q, want application/json", gotContentType)
				}
			} else if gotBody != "" {
				t.Errorf("request body = %q, want empty", gotBody)
			}
		})
	}
}

func TestDo_GetIgnoresBody(t *testing.T) {
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	resp := Do("GET", server.URL, `{"should":"not be sent"}`, "")
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	if gotBody != "" {
		t.Errorf("GET request body = %q, want empty", gotBody)
	}
}

func TestDo_NoAuthHeaderWhenTokenEmpty(t *testing.T) {
	var gotAuth string
	sawAuthHeader := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth, sawAuthHeader = r.Header.Get("Authorization"), r.Header.Get("Authorization") != ""
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	Do("GET", server.URL, "", "")
	if sawAuthHeader {
		t.Errorf("Authorization header = %q, want none", gotAuth)
	}
}

func TestDo_TransportError(t *testing.T) {
	resp := Do("GET", "http://127.0.0.1:0", "", "")
	if resp.Error == nil {
		t.Fatal("expected an error for an unreachable URL, got nil")
	}
}

func TestDo_InvalidURL(t *testing.T) {
	resp := Do("GET", "://not-a-url", "", "")
	if resp.Error == nil {
		t.Fatal("expected an error for an invalid URL, got nil")
	}
}
