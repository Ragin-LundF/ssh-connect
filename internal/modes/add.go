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
			cfg = config.File{Group: map[string]config.Group{}}
			config.EnsureGroup(&cfg, config.DefaultGroupName)
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
	for _, entry := range config.ToEntries(cfg) {
		if entry.Key == alias {
			return fmt.Errorf("a server with alias '%s' already exists", alias)
		}
	}

	groupKey, err := selectOrCreateGroup(&cfg, "Add Server")
	if err != nil {
		return cancellationToNil(err)
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

	group := cfg.Group[groupKey]
	if group.Server == nil {
		group.Server = map[string]config.Server{}
	}
	group.Server[alias] = config.Server{
		Name:        strings.TrimSpace(name),
		IP:          strings.TrimSpace(ip),
		User:        strings.TrimSpace(user),
		Certificate: strings.TrimSpace(cert),
	}
	cfg.Group[groupKey] = group

	if err := config.Save(opts.ConfigPath, cfg); err != nil {
		return err
	}

	return ui.ShowMessage("Server Added", fmt.Sprintf("Server '%s' was saved in group '%s'.", alias, group.Name))
}

func cancellationToNil(err error) error {
	if err == ui.ErrCancelled {
		return nil
	}
	return err
}
