package ui

import (
	"os"
	"testing"
)

// TestMain points the home directory at a temporary one. The UI saves the
// request history whenever a response arrives, and the tests used to write
// theirs into the real history, pushing out the user's own requests.
func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "gaufre-ui-test")
	if err != nil {
		panic(err)
	}
	os.Setenv("HOME", home)
	os.Setenv("USERPROFILE", home) // Windows equivalent used by os.UserHomeDir
	os.Unsetenv("XDG_CONFIG_HOME")

	code := m.Run()
	os.RemoveAll(home)
	os.Exit(code)
}
