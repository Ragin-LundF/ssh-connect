package modes

import "fmt"

import "ssh_connect/internal/cli"

func HelpText() string {
	return fmt.Sprintf(`Usage: ssh_connect [--dry-run] [--config <path>] [--debug-ui] [--init | --add | --delete | --help]

Without an explicit mode, ssh_connect starts in the interactive home view.

Interactive shortcuts:
  Up/Down or J/K        Move selection between servers
  Left/Right or Tab     Jump to the next/previous non-empty group
  Enter                 Connect to the selected server
  A                     Add a server
  G                     Add a group
  D                     Delete the selected server
  M                     Open the main menu
  ?                     Show help
  Q or Esc              Quit

Servers are organized in groups. If no group is chosen, 'Default' is used.
You can create groups from the main menu or while adding a server.

Options:
  --config <path>       Use a custom TOML config file (default: %s).
  --dry-run             Show the SSH command after selection, but do not execute it.
  --debug-ui            Print UI key/focus debug logs to stderr.
  --init                Create an example config file and exit.
  --add                 Add a new server entry using TUI prompts.
  --delete              Delete a server entry using TUI prompts.
  -h, --help            Show this help message.
`, cli.DefaultConfigPath)
}
