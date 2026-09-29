# Gaufre

A terminal-based API client built with Go and Bubble Tea. 
Designed for developers who live in the terminal.
Works seamlessly with tmux and neovim workflows.


![](assets/screenshot.png)

## Features

- **Multiple HTTP methods** - GET, POST, PUT, DELETE
- **Request history** - Saved locally with fuzzy search
- **Authentication support** - Bearer token on every method
- **JWT inspection** - Pasted tokens are decoded in place: algorithm, subject, issuer, audience, and expiry with an expired warning
- **Use a token from the response** - Press `t` on a login response and its token becomes the Bearer token for the next request; when a response holds several (access, refresh, id token...), a picker shows each one decoded so you can choose
- **Smart token paste** - Paste a whole auth response (`{"access_token": "..."}`), a `Bearer ...` header, or a token copied across wrapped lines; Gaufre extracts the token
- **Request summary** - The main screen warns before you send when the bearer token has expired, or when a body won't be sent with a GET
- **Response details** - Status and reason phrase, time, size and content type stay pinned above the body while you scroll; toggle the headers with `h`
- **JSON syntax highlighting** - Response formatting
- **Live JSON validation** - The body editor reports validity and the line/column of the first error as you type
- **Body formatting** - Pretty-print the request body with `Ctrl+F`
- **Scrollable responses** - Handle large API responses
- **Request editor** - Multiline JSON body editing with line numbers
- **Persistent storage** - History saved between sessions

## Installation
```bash
git clone https://github.com/Etienne-55/gaufre.git
cd gaufre
go build -o gaufre
```

## Usage
```bash
./gaufre
```

### In a tmux popup
Gaufre can float over your panes in a tmux popup (tmux 3.3+). With `--popup`
it drops its own frame so the popup's border is the only one, and signs the
bottom-right corner with a small logo. Add to `~/.tmux.conf`:
```
bind-key e display-popup -E -b rounded -S fg=magenta -T " Gaufre " -w 80% -h 80% "/path/to/gaufre --popup"
```

### Keyboard Shortcuts

**Main Screen:**
- `↑↓` - Navigate between method selector, URL and the Body/Auth menu
- `←→` - Change HTTP method (or choose Body/Auth)
- `Enter` - Send request (or open the selected editor)
- `Tab` - Open history
- `q` - Quit (`Ctrl+C` while typing a URL)

**URL Input:**
- Type to edit the URL — every letter types, including `q`
- `←→`, `Home/End` - Move the cursor; `Alt+←/→` by word
- `Ctrl+W` - Delete the previous word; `Ctrl+U`/`Ctrl+K` delete to the start/end
- `Ctrl+V` - Paste (line breaks from a wrapped copy are removed)
- `↑` - Back to method selector
- `↓` - Go to request editor

**Request Editor:**
- `←→` - Switch between Body and Auth tabs
- `Enter` - Open selected editor
- `↑` - Back to main screen

**Body Editor:**
- Type to edit — the status line shows JSON validity live
- `Ctrl+F` - Format (pretty-print) the body
- `Tab` - Indent two spaces
- `Ctrl+V` - Paste from clipboard
- `Ctrl+C` - Copy the body
- `Ctrl+D` - Clear the body
- `Esc` - Back to main screen

**Auth Editor:**
- Type to edit — JWTs are decoded and summarised below the field
- `Ctrl+V` - Paste a token, extracting it from a JSON response, a `Bearer` header, or a line-wrapped copy
- `Ctrl+C` - Copy the token
- `Ctrl+D` - Clear the token
- `Esc` - Back to main screen

**History:**
- `/` - To filter requests 
- `↑↓` - Navigate
- `Enter` - Load request
- `Esc` - Clear the filter, then close history

**Response Viewer:**
- `↑↓` or `j/k` - Scroll; `Space`/`b` by page, `d`/`u` by half page
- `g`/`G` - Jump to the top/bottom
- `h` - Show or hide the response headers
- `t` - Use the token found in the response as the auth token (opens the token picker when there are several)
- `Esc` - Back to main screen

**Token Picker:**
- `↑↓` or `j/k` - Choose a token — each one shows where it was found and, for JWTs, its subject and expiry
- `Enter` - Use it as the auth token
- `c` - Copy it to the clipboard
- `Esc` - Back to the response

## Configuration

History is stored next to your other tools' configs:
- **Linux/macOS:** `~/.config/gaufre/history.json` (or `$XDG_CONFIG_HOME/gaufre/history.json` when that is set)
- **Windows:** `%USERPROFILE%\.config\gaufre\history.json`

A history from older versions (`~/.url_history.json`) is moved there the first time Gaufre starts.

Maximum history items: 100 (configurable in `internal/storage/history.go`)

## Technologies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components

## Development
```bash
# Run
go run .

# Build
go build -o gaufre
