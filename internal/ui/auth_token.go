package ui

import (
	"encoding/json"
	"regexp"
	"strings"
	"unicode"
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
		// Keys are visited in sorted order: Go's random map order used to
		// pick a different token from the same paste from one run to the
		// next.
		keys := sortedKeys(v)
		for _, name := range tokenFieldNames {
			for _, key := range keys {
				if !strings.EqualFold(key, name) {
					continue
				}
				if s, ok := v[key].(string); ok && strings.TrimSpace(s) != "" {
					return normalizeBareToken(s)
				}
			}
		}
		for _, key := range keys {
			if found := searchTokenField(v[key], depth+1); found != "" {
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

// normalizeBareToken strips an Authorization header name, a Bearer prefix,
// surrounding quotes/punctuation and any whitespace that a line-wrapped copy
// introduced, without touching the token's own characters.
func normalizeBareToken(text string) string {
	token := strings.TrimSpace(text)
	// Header names are case-insensitive, and curl -v prints HTTP/2 ones in
	// lower case.
	token = strings.TrimSpace(trimPrefixFold(token, "authorization:"))
	token = trimPrefixFold(token, "bearer ")
	token = strings.Trim(token, surroundingJunk)

	// Remove whitespace introduced by copying a token that was displayed
	// across several lines, and any other control character: none can be
	// sent in a header, and escape sequences would reach the terminal.
	token = strings.Join(strings.Fields(token), "")
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, token)
}

// trimPrefixFold removes prefix from the start of s, ignoring case.
func trimPrefixFold(s, prefix string) string {
	if len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix) {
		return s[len(prefix):]
	}
	return s
}
