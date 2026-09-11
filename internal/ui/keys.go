package ui

import "github.com/charmbracelet/bubbles/key"

// payloadKeyMap describes the shortcuts available in the request body
// editor. Using key.Binding rather than bare strings lets the help footer
// render itself from the same source of truth that the handlers match on,
// so the two can't drift apart.
type payloadKeyMap struct {
	Format key.Binding
	Indent key.Binding
	Copy   key.Binding
	Paste  key.Binding
	Clear  key.Binding
	Back   key.Binding
}

func (k payloadKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Format, k.Indent, k.Copy, k.Clear, k.Back}
}

func (k payloadKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Format, k.Indent},
		{k.Copy, k.Paste, k.Clear},
		{k.Back},
	}
}

// authKeyMap describes the shortcuts available in the bearer token editor.
type authKeyMap struct {
	Paste key.Binding
	Copy  key.Binding
	Clear key.Binding
	Back  key.Binding
}

func (k authKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Paste, k.Copy, k.Clear, k.Back}
}

func (k authKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Paste, k.Copy},
		{k.Clear, k.Back},
	}
}

// tokenPickerKeyMap describes the shortcuts available in the picker that
// lists the tokens found in a response.
type tokenPickerKeyMap struct {
	Up   key.Binding
	Down key.Binding
	Use  key.Binding
	Copy key.Binding
	Back key.Binding
}

func (k tokenPickerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Use, k.Copy, k.Back}
}

func (k tokenPickerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.Use, k.Copy, k.Back},
	}
}

var (
	PayloadKeys = payloadKeyMap{
		Format: key.NewBinding(key.WithKeys("ctrl+f"), key.WithHelp("ctrl+f", "format")),
		Indent: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "indent")),
		Copy:   key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "copy")),
		Paste:  key.NewBinding(key.WithKeys("ctrl+v"), key.WithHelp("ctrl+v", "paste")),
		Clear:  key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "clear")),
		Back:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	}

	AuthKeys = authKeyMap{
		Paste: key.NewBinding(key.WithKeys("ctrl+v"), key.WithHelp("ctrl+v", "paste token")),
		Copy:  key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "copy")),
		Clear: key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "clear")),
		Back:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	}

	TokenPickerKeys = tokenPickerKeyMap{
		Up:   key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down: key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Use:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "use as auth token")),
		Copy: key.NewBinding(key.WithKeys("c", "ctrl+c"), key.WithHelp("c", "copy")),
		Back: key.NewBinding(key.WithKeys("esc", "t"), key.WithHelp("esc", "back")),
	}
)
