package ui

import (
	"testing"
	"time"
)

func TestDecodeJWT(t *testing.T) {
	exp := time.Now().Add(time.Hour).Truncate(time.Second)
	iat := time.Now().Add(-time.Hour).Truncate(time.Second)

	token := testJWT(t, map[string]any{
		"sub": "1234567890",
		"iss": "https://auth.example.com",
		"aud": "api://gaufre",
		"iat": float64(iat.Unix()),
		"exp": float64(exp.Unix()),
	})

	jwt, ok := decodeJWT(token)
	if !ok {
		t.Fatal("decodeJWT returned false for a valid token")
	}

	if jwt.Algorithm != "HS256" {
		t.Errorf("Algorithm = %q, want HS256", jwt.Algorithm)
	}
	if jwt.Subject != "1234567890" {
		t.Errorf("Subject = %q", jwt.Subject)
	}
	if jwt.Issuer != "https://auth.example.com" {
		t.Errorf("Issuer = %q", jwt.Issuer)
	}
	if jwt.Audience != "api://gaufre" {
		t.Errorf("Audience = %q", jwt.Audience)
	}
	if !jwt.ExpiresAt.Equal(exp) {
		t.Errorf("ExpiresAt = %v, want %v", jwt.ExpiresAt, exp)
	}
	if jwt.Expired(time.Now()) {
		t.Error("token expiring in an hour reported as expired")
	}
}

func TestDecodeJWT_ArrayAudience(t *testing.T) {
	token := testJWT(t, map[string]any{"aud": []any{"one", "two"}})

	jwt, ok := decodeJWT(token)
	if !ok {
		t.Fatal("decodeJWT returned false")
	}
	if jwt.Audience != "one, two" {
		t.Errorf("Audience = %q, want %q", jwt.Audience, "one, two")
	}
}

func TestDecodeJWT_Expired(t *testing.T) {
	token := testJWT(t, map[string]any{"exp": float64(time.Now().Add(-time.Hour).Unix())})

	jwt, ok := decodeJWT(token)
	if !ok {
		t.Fatal("decodeJWT returned false")
	}
	if !jwt.Expired(time.Now()) {
		t.Error("token that expired an hour ago not reported as expired")
	}
}

func TestDecodeJWT_Rejects(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"empty", ""},
		{"opaque", "sk_live_abc123"},
		{"two segments", "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0"},
		{"not base64", "!!!.???.###"},
		{"not json", "aGVsbG8.d29ybGQ.c2ln"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, ok := decodeJWT(tt.token); ok {
				t.Errorf("decodeJWT(%q) = true, want false", tt.token)
			}
		})
	}
}

func TestRelativeTime(t *testing.T) {
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		t    time.Time
		want string
	}{
		{now.Add(30 * time.Second), "in 30s"},
		{now.Add(42 * time.Minute), "in 42m"},
		{now.Add(2*time.Hour + 15*time.Minute), "in 2h15m"},
		{now.Add(72 * time.Hour), "in 3d"},
		{now.Add(-45 * time.Minute), "45m ago"},
		{now.Add(-48 * time.Hour), "2d ago"},
	}

	for _, tt := range tests {
		if got := relativeTime(tt.t, now); got != tt.want {
			t.Errorf("relativeTime(%v) = %q, want %q", tt.t, got, tt.want)
		}
	}
}
