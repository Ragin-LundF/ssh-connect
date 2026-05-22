# ssh_connect

`ssh_connect` is a modular Go application with a terminal UI built on `tui-go`.

## Features

- Dark-themed home screen with highlighted server selection
- Modal-style dialogs (centered, bordered) for add/confirm/message/menu
- Default startup shows the current server list
- In-app actions (no CLI round-trip needed):
  - connect
  - add server
  - delete server
  - help
  - quit
- Missing config flow: prompts to create/add a server
- TOML-based config (`[server.<alias>]`)
- `--dry-run` to preview SSH command safely
- `--debug-ui` to log UI key/focus/debug flow to stderr

## Keyboard Shortcuts (Home)

- `Enter` - connect to selected server
- `A` - add server
- `D` - delete selected server
- `M` - open main menu
- `H` - open help dialog
- `Q` / `Esc` - quit

## Debugging UI Rendering

Use the debug flag to trace UI actions and key handling:

```bash
./ssh_connect.sh --debug-ui
```

Note: `tui-go` is vendored in `vendor/` with a small lifecycle patch to ensure
clean shutdown/restart between dialog screens.

## Project Structure

- `cmd/ssh_connect/main.go` - entrypoint
- `internal/app/app.go` - mode routing
- `internal/cli/options.go` - argument parsing
- `internal/modes/connect.go` - home view and connect flow
- `internal/modes/add.go` - add flow
- `internal/modes/delete.go` - delete flow
- `internal/modes/init.go` - create sample config
- `internal/modes/help.go` - help text
- `internal/config/model.go` - config models
- `internal/config/store.go` - load/save/validation
- `internal/ui/tui.go` - themed TUI components, modal dialogs, debug hooks
- `internal/ssh/command.go` - SSH command creation/execution
- `ssh_connect.sh` - wrapper to run Go app

## Requirements

- Go `1.26.3` (toolchain pinned in `go.mod`)
- `ssh` available in `PATH`

## Quick Start

```bash
cd /Users/ragin/IdeaProjects/ssh_connect
./ssh_connect.sh
```

Help:

```bash
./ssh_connect.sh --help
```

Create sample config:

```bash
./ssh_connect.sh --init --config ./my_servers.toml
```

Run tests:

```bash
go test ./...
```
