package app

import (
	"fmt"

	"ssh_connect/internal/cli"
	"ssh_connect/internal/modes"
	"ssh_connect/internal/ui"
)

func Run(opts cli.Options) error {
	ui.SetDebug(opts.DebugUI)

	switch opts.Mode {
	case cli.ModeHelp:
		fmt.Print(modes.HelpText)
		return nil
	case cli.ModeInit:
		return modes.Init(opts)
	case cli.ModeAdd:
		return modes.Add(opts)
	case cli.ModeDelete:
		return modes.Delete(opts)
	case cli.ModeConnect:
		return modes.Connect(opts)
	default:
		return fmt.Errorf("unsupported mode: %s", opts.Mode)
	}
}
