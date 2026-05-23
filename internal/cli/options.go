package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Mode string

const (
	ModeConnect Mode = "connect"
	ModeAdd     Mode = "add"
	ModeDelete  Mode = "delete"
	ModeInit    Mode = "init"
	ModeHelp    Mode = "help"
)

type Options struct {
	ConfigPath string
	DryRun     bool
	DebugUI    bool
	Mode       Mode
}

const DefaultConfigFilename = "ssh_connect_server.toml"

var DefaultConfigPath = defaultConfigPath()

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return DefaultConfigFilename
	}

	return filepath.Join(home, DefaultConfigFilename)
}

func ParseArgs(args []string) (Options, error) {
	opts := Options{
		ConfigPath: DefaultConfigPath,
		Mode:       ModeConnect,
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dry-run":
			opts.DryRun = true
		case "--debug-ui":
			opts.DebugUI = true
		case "--config":
			i++
			if i >= len(args) {
				return Options{}, errors.New("missing value for --config")
			}
			opts.ConfigPath = args[i]
		case "--add":
			if err := setMode(&opts, ModeAdd); err != nil {
				return Options{}, err
			}
		case "--delete":
			if err := setMode(&opts, ModeDelete); err != nil {
				return Options{}, err
			}
		case "--init":
			if err := setMode(&opts, ModeInit); err != nil {
				return Options{}, err
			}
		case "-h", "--help":
			if err := setMode(&opts, ModeHelp); err != nil {
				return Options{}, err
			}
		default:
			return Options{}, fmt.Errorf("unknown option: %s", args[i])
		}
	}

	return opts, nil
}

func setMode(opts *Options, mode Mode) error {
	if opts.Mode != ModeConnect {
		return errors.New("use only one mode at a time: --init, --add, --delete, --help")
	}
	opts.Mode = mode
	return nil
}
