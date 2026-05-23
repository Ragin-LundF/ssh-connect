# ssh_connect

`ssh_connect` is a terminal UI for organizing and connecting to SSH servers,
built in Go with [`go-tui`](https://github.com/grindlemire/go-tui).
It helps you keep multiple systems grouped and easy to browse from one place,
then launches the matching `ssh` command for the server you select.

## Features

- Interactive home screen with a grouped server overview tree
- TUI dialogs for adding servers, creating groups, confirming deletes, reading help, and browsing the main menu
- In-app actions without restarting the program:
  - connect to the selected server
  - add a server
  - add a group
  - delete a server
  - open the main menu
  - show inline help
  - quit
- Missing-config flow: the interactive mode can prompt you to create your first server when the config file does not exist
- Grouped TOML config using `[group.<name>.server.<alias>]`
- Group-level SSH key fallback via `group_certificate`
- Legacy config migration:
  - top-level `server.<alias>` entries are folded into the `Default` group
  - legacy `servers = ["server.<alias>"]` references are migrated into embedded group server tables
- `--dry-run` to preview the SSH command without executing it
- `--debug-ui` to trace UI key events and screen transitions to stderr

## Keyboard Shortcuts

### Home screen

| Key | Action |
|-----|--------|
| `↑` / `k` | Move selection up |
| `↓` / `j` | Move selection down |
| `←` | Jump to the previous non-empty group |
| `→` / `Tab` / `l` | Jump to the next non-empty group |
| `Enter` | Connect to selected server |
| `A` | Add a new server |
| `G` | Add a new group |
| `D` | Delete selected server |
| `M` | Open main menu |
| `?` / `h` | Show help |
| `Q` / `Esc` | Quit |

Group navigation always jumps to the first server in the next or previous populated group. Empty groups are shown in the overview, but skipped when cycling with the keyboard.

### Dialogs

| Key | Action |
|-----|--------|
| `Enter` | Confirm / close |
| `Esc` | Cancel / close |

The main menu offers: connect, add server, add group, delete selected server, return to the server list, help, and quit.

## CLI Flags

```text
Usage: ssh_connect [--dry-run] [--config <path>] [--debug-ui] [--init | --add | --delete | --help]

Options:
  --config <path>       Use a custom TOML config file (default: ssh_connect_server.toml)
  --dry-run             Show the SSH command after selection, but do not execute it
  --debug-ui            Print UI key/focus debug logs to stderr
  --init                Create an example config file and exit
  --add                 Add a new server entry using TUI prompts
  --delete              Delete a server entry using TUI prompts
  -h, --help            Show the help message
```

Without an explicit mode, `ssh_connect` starts in the interactive home view.

Only one mode flag (`--init`, `--add`, `--delete`, `--help`) may be given at a time.

## Config Format

Config is a TOML file organized by groups. Each server lives under `[group.<group>.server.<alias>]`:

```toml
[group.Default]
name = "Default"

[group.Default.server.my_box]
name = "My Box"
ip = "192.168.1.10"
user = "alice"
certificate = "~/.ssh/id_ed25519"   # optional

[group.production]
name = "Production"
group_certificate = "~/.ssh/prod_shared.pem" # optional fallback for servers in this group

[group.production.server.app_prod]
name = "App Production"
ip = "203.0.113.10"
user = "deploy"
```

Notes:

- `certificate` is optional. When omitted, the app falls back to the group's `group_certificate`, if present.
- If neither `certificate` nor `group_certificate` is set, SSH uses its normal default key lookup / agent behavior.
- If no group is explicitly chosen, the server is placed into `Default`.
- Group names may contain letters, numbers, spaces, `_`, and `-`.
- Server aliases may contain letters, numbers, `_`, and `-`.

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
    group.go                   add-group helpers
    init.go                    create example config
    help.go                    static help text
  config/
    model.go                   TOML data models
    store.go                   load / save / normalization / validation
  ssh/
    command.go                 SSH command assembly and exec
  ui/
    debug.go                   debug logger (SetDebug / debugf)
    screen.go                  TUI app runner (interactiveScreen / runUI)
    layout.go                  shared layout helpers (newSection, buildScreenRoot, renderList)
    widgets.go                 reusable dialogs (SelectIndex, PromptInput, Confirm, ShowMessage)
    home.go                    home screen entrypoint and HomeAction type
    home_state.go              home selection and navigation state
    home_overview.go           grouped overview tree rendering
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

# Equivalent direct Go run
go run .

# Show help
./ssh_connect.sh --help

# Create an example config
./ssh_connect.sh --init --config ./my_servers.toml

# Preview the SSH command without connecting
# (after selecting a configured server in the UI)
./ssh_connect.sh --dry-run

# Debug UI key events
./ssh_connect.sh --debug-ui
```

## Running Tests

```bash
go test ./...
```
