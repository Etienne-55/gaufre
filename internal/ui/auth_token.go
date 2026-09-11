package ui

import (
	"encoding/json"
	"regexp"
	"strings"
)

// tokenFieldNames are the JSON keys commonly used to carry a bearer token in
// an auth response body, checked in order of preference.
var tokenFieldNames = []string{
	"access_token", "accessToken",
	"id_token", "idToken",
	"token", "jwt",
	"auth_token", "authToken",
	"bearer", "accessJwt",
}

// jwtPattern matches a JWT anywhere inside a larger blob of text.
var jwtPattern = regexp.MustCompile(`eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]*`)

var surroundingJunk = "\"'`,;: \t\r\n"

// cleanAuthToken turns whatever was on the clipboard into a usable bearer
// token. It handles the three ways a token usually gets copied:
//
//  1. as a JSON response body ({"access_token": "..."}, possibly nested),
//     in which case the token field is pulled out;
//  2. as a JWT embedded in surrounding text or log output;
//  3. as the bare token, possibly with a "Bearer " prefix, quotes, or line
//     breaks introduced by copying out of a wrapped terminal.
//
// Unlike the previous implementation it does not strip every character
// outside the JWT alphabet, so opaque tokens containing '+', '/', '=' or '~'
// survive intact.
func cleanAuthToken(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ""
	}

	if token := tokenFromJSON(trimmed); token != "" {
		return token
	}

	if token := jwtPattern.FindString(trimmed); token != "" {
		return token
	}

	return normalizeBareToken(trimmed)
}

// tokenFromJSON looks for a token-ish field anywhere in a JSON document,
// including nested objects such as {"data": {"access_token": "..."}}.
func tokenFromJSON(text string) string {
	var doc any
	if err := json.Unmarshal([]byte(text), &doc); err != nil {
		return ""
	}
	return searchTokenField(doc, 0)
}

func searchTokenField(node any, depth int) string {
	const maxDepth = 6
	if depth > maxDepth {
		return ""
	}

	switch v := node.(type) {
	case map[string]any:
		// Prefer an exact, well-known field name at this level before
		// descending, so {"token": "..."} wins over a nested match.
		for _, name := range tokenFieldNames {
			for key, value := range v {
				if !strings.EqualFold(key, name) {
					continue
				}
				if s, ok := value.(string); ok && strings.TrimSpace(s) != "" {
					return normalizeBareToken(s)
				}
			}
		}
		for _, value := range v {
			if found := searchTokenField(value, depth+1); found != "" {
				return found
			}
		}
	case []any:
		for _, item := range v {
			if found := searchTokenField(item, depth+1); found != "" {
				return found
			}
		}
	}

	return ""
}

// normalizeBareToken strips a Bearer prefix, surrounding quotes/punctuation
// and any whitespace that a line-wrapped copy introduced, without touching
// the token's own characters.
func normalizeBareToken(text string) string {
	token := strings.TrimSpace(text)
	token = strings.TrimPrefix(token, "Authorization:")
	token = strings.TrimSpace(token)

	if len(token) >= len("bearer ") && strings.EqualFold(token[:len("bearer ")], "bearer ") {
		token = token[len("bearer "):]
	}

	token = strings.Trim(token, surroundingJunk)

	// Remove whitespace introduced by copying a token that was displayed
	// across several lines.
	return strings.Join(strings.Fields(token), "")
}
