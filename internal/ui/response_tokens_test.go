package ui

import (
	"testing"
	"time"

	"gaufre/internal/types"
)

func sources(candidates []tokenCandidate) []string {
	out := make([]string, len(candidates))
	for i, c := range candidates {
		out[i] = c.Source
	}
	return out
}

func TestFindResponseTokens_LoginResponse(t *testing.T) {
	access := testJWT(t, map[string]any{"sub": "alice"})
	refresh := testJWT(t, map[string]any{"sub": "alice", "typ": "refresh"})

	resp := &types.Response{Body: `{
		"token_type": "Bearer",
		"expires_in": 3600,
		"refresh_token": "` + refresh + `",
		"access_token": "` + access + `"
	}`}

	got := findResponseTokens(resp)
	if len(got) != 2 {
		t.Fatalf("found %d tokens (%v), want 2", len(got), sources(got))
	}
	if got[0].Source != "access_token" || got[0].Token != access {
		t.Errorf("first candidate = %+v, want access_token first", got[0])
	}
	if got[1].Source != "refresh_token" {
		t.Errorf("second candidate = %q, want refresh_token", got[1].Source)
	}
}

func TestFindResponseTokens_Sources(t *testing.T) {
	jwt := testJWT(t, map[string]any{"sub": "bob"})

	tests := []struct {
		name string
		resp *types.Response
		want []string
	}{
		{
			name: "nested field gets its path",
			resp: &types.Response{Body: `{"data":{"session":{"accessToken":"opaque-token-value"}}}`},
			want: []string{"data.session.accessToken"},
		},
		{
			name: "jwt under an arbitrary key",
			resp: &types.Response{Body: `{"user":{"id":1,"credential":"` + jwt + `"}}`},
			want: []string{"user.credential"},
		},
		{
			name: "array index in path",
			resp: &types.Response{Body: `{"sessions":[{"token":"abc123"},{"token":"def456"}]}`},
			want: []string{"sessions[0].token", "sessions[1].token"},
		},
		{
			name: "json string body",
			resp: &types.Response{Body: `"` + jwt + `"`},
			want: []string{"body"},
		},
		{
			name: "plain text body",
			resp: &types.Response{Body: "your token is " + jwt + " enjoy"},
			want: []string{"body"},
		},
		{
			name: "authorization header",
			resp: &types.Response{
				Body:    `{"ok":true}`,
				Headers: map[string][]string{"Authorization": {"Bearer sk_live_abc123"}},
			},
			want: []string{"header Authorization"},
		},
		{
			name: "jwt in a cookie",
			resp: &types.Response{
				Body:    `{"ok":true}`,
				Headers: map[string][]string{"Set-Cookie": {"session=" + jwt + "; Path=/; HttpOnly"}},
			},
			want: []string{"header Set-Cookie"},
		},
		{
			name: "same token in body and header is listed once",
			resp: &types.Response{
				Body:    `{"access_token":"` + jwt + `"}`,
				Headers: map[string][]string{"X-Auth-Token": {jwt}},
			},
			want: []string{"access_token"},
		},
		{
			name: "token-ish keys need a token-ish value",
			resp: &types.Response{Body: `{"token_type":"Bearer","csrf_token":"short","refresh_token_expires_in":86400}`},
			want: nil,
		},
		{
			name: "no tokens",
			resp: &types.Response{Body: `{"id":1,"name":"alice","tags":["a","b"]}`},
			want: nil,
		},
		{
			name: "failed request",
			resp: &types.Response{Error: errTest},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sources(findResponseTokens(tt.resp))
			if len(got) != len(tt.want) {
				t.Fatalf("sources = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("sources = %v, want %v", got, tt.want)
					break
				}
			}
		})
	}
}

func TestFindResponseTokens_StripsBearerPrefix(t *testing.T) {
	got := findResponseTokens(&types.Response{Body: `{"token":"Bearer abc123"}`})
	if len(got) != 1 || got[0].Token != "abc123" {
		t.Errorf("got %+v, want the bare token abc123", got)
	}
}

func TestDescribeToken(t *testing.T) {
	now := time.Now()

	live := testJWT(t, map[string]any{"sub": "alice", "exp": float64(now.Add(time.Hour).Unix())})
	if got, _ := describeToken(live, now); got != "JWT · HS256 · sub alice · exp in 59m" && got != "JWT · HS256 · sub alice · exp in 1h0m" {
		t.Errorf("describeToken(live) = %q", got)
	}

	expired := testJWT(t, map[string]any{"exp": float64(now.Add(-2 * time.Hour).Unix())})
	if got, style := describeToken(expired, now); got != "JWT · HS256 · expired 2h0m ago" || style.GetForeground() != ColorBad {
		t.Errorf("describeToken(expired) = %q", got)
	}

	if got, _ := describeToken("sk_live_abc123", now); got != "opaque · 14 chars" {
		t.Errorf("describeToken(opaque) = %q", got)
	}
}
