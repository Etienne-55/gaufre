package ui

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// bodyStatus describes the request body for the editor's status line.
type bodyStatus struct {
	Empty   bool
	Valid   bool
	Problem string // populated when Valid is false
	Lines   int
	Bytes   int
}

// inspectBody reports whether the body currently parses as JSON, and where
// it breaks if it doesn't, so the editor can show that live while typing.
func inspectBody(body string) bodyStatus {
	status := bodyStatus{
		Lines: strings.Count(body, "\n") + 1,
		Bytes: len(body),
	}

	if strings.TrimSpace(body) == "" {
		status.Empty = true
		return status
	}

	var out any
	if err := json.Unmarshal([]byte(body), &out); err != nil {
		status.Problem = describeJSONError(body, err)
		return status
	}

	status.Valid = true
	return status
}

// describeJSONError turns encoding/json's byte-offset errors into a
// line:column message, which is far more useful while editing.
func describeJSONError(body string, err error) string {
	var offset int64 = -1

	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	switch {
	case errors.As(err, &syntaxErr):
		offset = syntaxErr.Offset
	case errors.As(err, &typeErr):
		offset = typeErr.Offset
	}

	if offset < 0 {
		return err.Error()
	}

	line, col := lineAndColumn(body, offset)
	return fmt.Sprintf("line %d, col %d: %s", line, col, err.Error())
}

func lineAndColumn(body string, offset int64) (line, col int) {
	if offset > int64(len(body)) {
		offset = int64(len(body))
	}

	line, col = 1, 1
	for _, r := range body[:offset] {
		if r == '\n' {
			line++
			col = 1
			continue
		}
		col++
	}
	return line, col
}

// formatJSONBody pretty-prints the body with the two-space indentation the
// editor uses. It reports an error (leaving the body untouched) when the
// body isn't valid JSON.
func formatJSONBody(body string) (string, error) {
	if strings.TrimSpace(body) == "" {
		return body, nil
	}

	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(body), "", "  "); err != nil {
		return body, err
	}

	return buf.String(), nil
}
