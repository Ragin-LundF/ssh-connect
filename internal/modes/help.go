package modes

const HelpText = `Usage: ssh_connect [--dry-run] [--config <path>]

Without an explicit mode, ssh_connect starts in the interactive home view.
Available shortcuts there:
  Left/Right, Tab, H/L  Switch between server list and group list
  Group list  Selects a group and jumps to its first server
  Enter  Connect to selected server
  A      Add server
  G      Add group
  D      Delete selected server
  M      Open main menu
  ?      Show in-app help
  Q/Esc  Quit

Servers are organized in groups. If no group is chosen, 'Default' is used.
You can create groups from the main menu or while adding a server.

Options:
  --dry-run   Show the SSH command after selection, but do not execute it.
  --debug-ui  Print UI key/focus debug logs to stderr.
  --config    Use a custom TOML config file.
  --init      Create an example config file and exit.
  --add       Add a new server entry using TUI prompts.
  --delete    Delete a server entry using TUI prompts.
  -h, --help  Show this help message.
`
