package modes

import (
	"fmt"
	"os"
	"strings"

	"ssh_connect/internal/cli"
	"ssh_connect/internal/config"
	"ssh_connect/internal/ui"
)

func Add(opts cli.Options) error {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg = config.File{Server: map[string]config.Server{}, Group: map[string]config.Group{}}
			if err := config.Save(opts.ConfigPath, cfg); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	alias, err := ui.PromptInput("Add Server", "Alias (server.<alias>):", true)
	if err != nil {
		if err == ui.ErrCancelled {
			return nil
		}
		return err
	}
	if !config.IsValidAlias(alias) {
		return fmt.Errorf("alias may only contain letters, numbers, '_' and '-'")
	}
	if _, exists := cfg.Server[alias]; exists {
		return fmt.Errorf("a server with alias '%s' already exists", alias)
	}

	name, err := ui.PromptInput("Add Server", "Display name:", true)
	if err != nil {
		return cancellationToNil(err)
	}
	ip, err := ui.PromptInput("Add Server", "IP address or hostname:", true)
	if err != nil {
		return cancellationToNil(err)
	}
	user, err := ui.PromptInput("Add Server", "SSH user:", true)
	if err != nil {
		return cancellationToNil(err)
	}
	cert, err := ui.PromptInput("Add Server", "Certificate path (optional):", false)
	if err != nil {
		return cancellationToNil(err)
	}

	cfg.Server[alias] = config.Server{
		Name:        strings.TrimSpace(name),
		IP:          strings.TrimSpace(ip),
		User:        strings.TrimSpace(user),
		Certificate: strings.TrimSpace(cert),
	}

	if err := config.Save(opts.ConfigPath, cfg); err != nil {
		return err
	}

	return ui.ShowMessage("Server Added", fmt.Sprintf("Server '%s' was saved.", alias))
}

func cancellationToNil(err error) error {
	if err == ui.ErrCancelled {
		return nil
	}
	return err
}
