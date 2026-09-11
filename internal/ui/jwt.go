package ui

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// JWT holds the decoded, human-interesting parts of a JSON Web Token. Only
// the claims the auth editor displays are pulled out; Claims keeps the full
// payload for anything else.
type JWT struct {
	Algorithm string
	Type      string
	Subject   string
	Issuer    string
	Audience  string
	IssuedAt  time.Time
	ExpiresAt time.Time
	Claims    map[string]any
}

// Expired reports whether the token carries an exp claim that is in the past.
func (j JWT) Expired(now time.Time) bool {
	return !j.ExpiresAt.IsZero() && now.After(j.ExpiresAt)
}

// decodeJWT parses a JWT without verifying its signature - this is a client
// for poking at APIs, so the point is to show what a token claims, not to
// trust it. It returns false for anything that isn't a well-formed JWT.
func decodeJWT(token string) (JWT, bool) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 3 {
		return JWT{}, false
	}

	header, err := decodeJWTSegment(parts[0])
	if err != nil {
		return JWT{}, false
	}
	claims, err := decodeJWTSegment(parts[1])
	if err != nil {
		return JWT{}, false
	}

	j := JWT{
		Algorithm: stringClaim(header, "alg"),
		Type:      stringClaim(header, "typ"),
		Subject:   stringClaim(claims, "sub"),
		Issuer:    stringClaim(claims, "iss"),
		Audience:  audienceClaim(claims),
		IssuedAt:  timeClaim(claims, "iat"),
		ExpiresAt: timeClaim(claims, "exp"),
		Claims:    claims,
	}

	return j, true
}

func decodeJWTSegment(segment string) (map[string]any, error) {
	// JWT segments are base64url without padding, but be forgiving about
	// tokens that were copied with padding intact.
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(segment, "="))
	if err != nil {
		return nil, err
	}

	var out map[string]any
	if err := json.Unmarshal(decoded, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func stringClaim(claims map[string]any, name string) string {
	if v, ok := claims[name].(string); ok {
		return v
	}
	return ""
}

// audienceClaim handles aud being either a single string or an array, both
// of which are legal per RFC 7519.
func audienceClaim(claims map[string]any) string {
	switch v := claims["aud"].(type) {
	case string:
		return v
	case []any:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, ", ")
	}
	return ""
}

// timeClaim reads a NumericDate claim (seconds since the epoch), which
// encoding/json hands us as a float64.
func timeClaim(claims map[string]any, name string) time.Time {
	switch v := claims[name].(type) {
	case float64:
		return time.Unix(int64(v), 0)
	case json.Number:
		if n, err := v.Int64(); err == nil {
			return time.Unix(n, 0)
		}
	}
	return time.Time{}
}

// relativeTime renders a duration the way a human would say it: "in 42m",
// "3h ago". Used for the exp/iat claims in the auth editor.
func relativeTime(t, now time.Time) string {
	d := t.Sub(now)
	suffix := "ago"
	if d >= 0 {
		suffix = ""
	} else {
		d = -d
	}

	var magnitude string
	switch {
	case d < time.Minute:
		magnitude = fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		magnitude = fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		magnitude = fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
	default:
		magnitude = fmt.Sprintf("%dd", int(d.Hours()/24))
	}

	if suffix == "" {
		return "in " + magnitude
	}
	return magnitude + " " + suffix
}
