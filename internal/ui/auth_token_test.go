package ui

import (
	"strings"
	"testing"
)

func TestCleanAuthToken(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "bare token",
			in:   "  eyJhbGciOi.abc-123_XYZ  \n",
			want: "eyJhbGciOi.abc-123_XYZ",
		},
		{
			name: "strips bearer prefix",
			in:   "Bearer sk_live_abc123",
			want: "sk_live_abc123",
		},
		{
			name: "strips authorization header line",
			in:   "Authorization: Bearer sk_live_abc123",
			want: "sk_live_abc123",
		},
		{
			name: "keeps base64 punctuation in opaque tokens",
			in:   "sk_live_51H8xYz+GhI/QrSt=",
			want: "sk_live_51H8xYz+GhI/QrSt=",
		},
		{
			name: "rejoins a token copied across wrapped lines",
			in:   "eyJhbGciOiJIUzI1NiIs\nInR5cCI6IkpXVCJ9.abc\n.def",
			want: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.abc.def",
		},
		{
			name: "json token field",
			in:   `{"token":"abc123"}`,
			want: "abc123",
		},
		{
			name: "json access_token wins over other fields",
			in:   `{"token_type":"Bearer","access_token":"abc123","expires_in":3600}`,
			want: "abc123",
		},
		{
			name: "nested json token",
			in:   `{"data":{"user":{"id":1},"access_token":"nested-token"}}`,
			want: "nested-token",
		},
		{
			name: "pretty printed json response",
			in:   "{\n  \"access_token\": \"abc.def.ghi\",\n  \"expires_in\": 3600\n}",
			want: "abc.def.ghi",
		},
		{
			name: "jwt embedded in surrounding text",
			in:   `curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.sig" https://api.test`,
			want: "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.sig",
		},
		{
			name: "quoted value",
			in:   `"abc123",`,
			want: "abc123",
		},
		{
			name: "empty",
			in:   "   \n ",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanAuthToken(tt.in); got != tt.want {
				t.Errorf("cleanAuthToken(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// The previous implementation filtered every character outside the JWT
// alphabet, which silently corrupted opaque tokens and left the literal word
// "token" glued to the front of a value pulled out of a JSON body.
func TestCleanAuthToken_RegressionAgainstCharsetFilter(t *testing.T) {
	got := cleanAuthToken(`{"token":"abc123"}`)
	if strings.HasPrefix(got, "token") {
		t.Errorf("cleanAuthToken kept the field name in the value: %q", got)
	}

	if got := cleanAuthToken("opaque+token/with=padding"); got != "opaque+token/with=padding" {
		t.Errorf("cleanAuthToken stripped valid token characters: %q", got)
	}
}
