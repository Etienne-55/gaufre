package ui

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"gaufre/internal/types"

	"github.com/charmbracelet/lipgloss"
)

// tokenCandidate is a value found in a response that could be used as the
// bearer token for the next request - typically the access_token of a login
// response.
type tokenCandidate struct {
	// Source says where the value was found: a JSON path such as
	// "data.access_token", "header Authorization", or "body" for a token in
	// a non-JSON body.
	Source string
	Token  string

	// rank orders candidates so the one most likely to be wanted comes
	// first: well-known token fields, then JWTs and token headers, then
	// anything else whose key merely mentions a token.
	rank int
}

const (
	rankKnownField = 0
	rankJWT        = 100
	rankTokenish   = 200

	// minTokenishLength keeps short values under token-ish keys (such as
	// "token_type": "Bearer") out of the list.
	minTokenishLength = 16
)

// findResponseTokens collects every value in a response that looks like a
// usable bearer token, best candidate first. It is what lets the response
// viewer hand a login response's token straight to the auth editor instead
// of making you select it with the mouse and paste it.
func findResponseTokens(resp *types.Response) []tokenCandidate {
	if resp == nil || resp.Error != nil {
		return nil
	}

	var found []tokenCandidate
	add := func(source, token string, rank int) {
		token = normalizeBareToken(token)
		if token == "" {
			return
		}
		found = append(found, tokenCandidate{Source: sanitizeForTerminal(source), Token: token, rank: rank})
	}

	var doc any
	if err := json.Unmarshal([]byte(resp.Body), &doc); err == nil {
		walkJSONForTokens(doc, "", "", 0, add)
	} else {
		for _, jwt := range jwtPattern.FindAllString(resp.Body, -1) {
			add("body", jwt, rankJWT)
		}
	}

	for _, name := range sortedKeys(resp.Headers) {
		for _, value := range resp.Headers[name] {
			if isTokenHeader(name) {
				add("header "+name, value, rankJWT)
				continue
			}
			// Tokens also turn up inside other headers, most often a
			// Set-Cookie carrying a JWT session.
			for _, jwt := range jwtPattern.FindAllString(value, -1) {
				add("header "+name, jwt, rankJWT)
			}
		}
	}

	sort.SliceStable(found, func(i, j int) bool { return found[i].rank < found[j].rank })
	return dedupeTokens(found)
}

// walkJSONForTokens visits every string in a decoded JSON document and
// reports the ones that look like tokens, along with their path. Object keys
// are visited in sorted order so the result doesn't depend on Go's map
// iteration order.
func walkJSONForTokens(node any, path, key string, depth int, add func(source, token string, rank int)) {
	const maxDepth = 12
	if depth > maxDepth {
		return
	}

	switch v := node.(type) {
	case map[string]any:
		for _, k := range sortedKeys(v) {
			walkJSONForTokens(v[k], joinJSONPath(path, k), k, depth+1, add)
		}
	case []any:
		for i, item := range v {
			walkJSONForTokens(item, fmt.Sprintf("%s[%d]", path, i), key, depth+1, add)
		}
	case string:
		source := path
		if source == "" {
			source = "body"
		}
		if rank, ok := tokenRank(key, v); ok {
			add(source, v, rank)
		}
	}
}

// tokenRank decides whether a JSON string value is worth offering as a
// token, and how prominently.
func tokenRank(key, value string) (int, bool) {
	for i, name := range tokenFieldNames {
		if strings.EqualFold(key, name) && strings.TrimSpace(value) != "" {
			return rankKnownField + i, true
		}
	}

	if _, ok := decodeJWT(normalizeBareToken(value)); ok {
		return rankJWT, true
	}

	lower := strings.ToLower(key)
	if (strings.Contains(lower, "token") || strings.Contains(lower, "jwt")) &&
		!strings.HasSuffix(lower, "type") &&
		len(value) >= minTokenishLength && !strings.ContainsAny(value, " \t\r\n") &&
		!strings.Contains(value, "://") { // "token_endpoint": "https://..."
		return rankTokenish, true
	}

	return 0, false
}

// isTokenHeader reports whether a response header is one APIs use to hand
// back a token directly. Rate-limit headers only count tokens
// (x-ratelimit-remaining-tokens: 149984, anthropic-ratelimit-tokens-reset,
// ...), and used to be offered - and, alone, set - as the bearer token.
func isTokenHeader(name string) bool {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "tokens") || strings.Contains(lower, "ratelimit") ||
		strings.Contains(lower, "rate-limit") {
		return false
	}
	return lower == "authorization" || strings.Contains(lower, "token")
}

func joinJSONPath(path, key string) string {
	if path == "" {
		return key
	}
	return path + "." + key
}

// dedupeTokens drops repeated tokens (the same JWT in the body and a
// header, say), keeping the best-ranked occurrence.
func dedupeTokens(candidates []tokenCandidate) []tokenCandidate {
	seen := make(map[string]bool, len(candidates))
	out := candidates[:0]
	for _, c := range candidates {
		if seen[c.Token] {
			continue
		}
		seen[c.Token] = true
		out = append(out, c)
	}
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// describeToken is the one-line summary shown next to a candidate in the
// token picker: what kind of token it is, who it's for and when it expires.
func describeToken(token string, now time.Time) (string, lipgloss.Style) {
	jwt, ok := decodeJWT(token)
	if !ok {
		return fmt.Sprintf("opaque · %d chars", len(token)), EditorMetaStyle
	}

	parts := []string{"JWT"}
	if jwt.Algorithm != "" {
		parts = append(parts, jwt.Algorithm)
	}
	if jwt.Subject != "" {
		parts = append(parts, "sub "+jwt.Subject)
	}

	style := EditorValueStyle
	if !jwt.ExpiresAt.IsZero() {
		note := "exp " + relativeTime(jwt.ExpiresAt, now)
		switch {
		case jwt.Expired(now):
			style = EditorBadStyle
			note = "expired " + relativeTime(jwt.ExpiresAt, now)
		case jwt.ExpiresAt.Sub(now) < 5*time.Minute:
			style = EditorWarnStyle
		}
		parts = append(parts, note)
	}

	return strings.Join(parts, " · "), style
}
