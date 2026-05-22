# ssh_connect

`ssh_connect` is a terminal UI tool for managing and connecting to SSH servers,
written in Go with [`go-tui`](https://github.com/grindlemire/go-tui).

## Features

- Dark-themed home screen with highlighted server selection
- Modal-style dialogs (bordered, scrollable) for add / confirm / message / menu
- Default startup shows the current server list
- In-app actions without any CLI round-trip:
  - connect to a server
  - add a server
  - delete a server
  - open the main menu
  - show inline help
  - quit
- Missing-config flow: prompts to add a server when the config file does not exist
- TOML-based config (`[server.<alias>]`), optional per-server certificate
- `--dry-run` to preview the SSH command without executing it
- `--debug-ui` to trace UI key-events and screen transitions to stderr

## Keyboard Shortcuts

### Home screen

| Key | Action |
|-----|--------|
| `↑` / `k` | Move selection up |
| `↓` / `j` | Move selection down |
| `Enter` | Connect to selected server |
| `A` | Add a new server |
| `D` | Delete selected server |
| `M` | Open main menu |
| `H` | Show help |
| `Q` / `Esc` | Quit |

### All dialogs

| Key | Action |
|-----|--------|
| `Enter` | Confirm / close |
| `Esc` | Cancel / close |
| `Ctrl+C` | Force quit |

## CLI Flags

```
ssh_connect [flags]

Flags:
  --config <path>   TOML config file to use (default: ssh_connect_server.toml)
  --dry-run         Print the SSH command instead of executing it
  --debug-ui        Write UI key/screen debug logs to stderr
  --init            Create an example config file and exit
  --add             Open the add-server dialog and exit
  --delete          Open the delete-server dialog and exit
  -h, --help        Print help text and exit
```

Only one mode flag (`--init`, `--add`, `--delete`, `--help`) may be given at a time.

## Config Format

Config is a TOML file. Each server lives under `[server.<alias>]`:

```toml
[server.my_box]
name        = "My Box"
ip          = "192.168.1.10"
user        = "alice"
certificate = "~/.ssh/id_ed25519"   # optional
```

The `certificate` field is optional; when omitted, SSH uses the default key agent.

## Project Structure

```
main.go                        entrypoint
internal/
  app/app.go                   mode dispatch
  cli/options.go               flag parsing
  modes/
    connect.go                 home-screen loop and connect flow
    add.go                     add-server flow
    delete.go                  delete-server flow
    init.go                    create example config
    help.go                    static help text
  config/
    model.go                   TOML data models
    store.go                   load / save / validation
  ssh/
    command.go                 SSH command assembly and exec
  ui/
    debug.go                   debug logger (SetDebug / debugf)
    screen.go                  TUI app runner (interactiveScreen / runUI)
    layout.go                  shared layout helpers (newSection, buildScreenRoot, renderList)
    widgets.go                 reusable dialogs (SelectIndex, PromptInput, Confirm, ShowMessage)
    home.go                    home screen and HomeAction type
    menu.go                    main-menu screen and MenuAction type
ssh_connect.sh                 convenience wrapper: go run . "$@"
```

## Requirements

- Go `1.26.0` or later (toolchain `go1.26.3` pinned in `go.mod`)
- `ssh` available in `$PATH`

## Quick Start

```bash
# Run interactively (no build step needed)
./ssh_connect.sh

# Show help
./ssh_connect.sh --help

# Create an example config
./ssh_connect.sh --init --config ./my_servers.toml

# Preview the SSH command without connecting
./ssh_connect.sh --dry-run

# Debug UI key events
./ssh_connect.sh --debug-ui
```

## Running Tests

```bash
go test ./...
```
