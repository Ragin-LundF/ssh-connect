package modes

import (
	"strings"
	"testing"

	"ssh_connect/internal/cli"
)

func TestHelpTextIncludesCLIArguments(t *testing.T) {
	help := HelpText()

	checks := []string{
		"Usage: ssh_connect",
		"--config <path>",
		"--dry-run",
		"--debug-ui",
		"--init",
		"--add",
		"--delete",
		"-h, --help",
		cli.DefaultConfigPath,
	}

	for _, check := range checks {
		if !strings.Contains(help, check) {
			t.Fatalf("expected help text to include %q\nhelp:\n%s", check, help)
		}
	}
}
