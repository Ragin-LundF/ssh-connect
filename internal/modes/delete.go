package modes

import (
	"fmt"

	"ssh_connect/internal/cli"
	"ssh_connect/internal/config"
	"ssh_connect/internal/ui"
)

func Delete(opts cli.Options) error {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return err
	}
	entries := config.ToEntries(cfg)
	if len(entries) == 0 {
		return fmt.Errorf("no valid server entries found in %s", opts.ConfigPath)
	}

	labels := make([]string, 0, len(entries))
	for _, entry := range entries {
		labels = append(labels, fmt.Sprintf("%s (%s@%s) | %s", entry.Server.Name, entry.Server.User, entry.Server.IP, entry.GroupName))
	}

	idx, err := ui.SelectIndex("Delete Server", "Choose a server to delete", labels)
	if err != nil {
		if err == ui.ErrCancelled {
			return nil
		}
		return err
	}

	return DeleteAlias(opts, entries[idx].GroupKey, entries[idx].Key)
}

func DeleteAlias(opts cli.Options, groupKey, alias string) error {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return err
	}

	group, groupExists := cfg.Group[groupKey]
	if !groupExists {
		return fmt.Errorf("group '%s' not found", groupKey)
	}

	server, exists := group.Server[alias]
	if !exists {
		return fmt.Errorf("server '%s' not found", alias)
	}

	groupName := group.Name
	if groupName == "" {
		groupName = groupKey
	}

	ok, err := ui.Confirm("Delete Server", fmt.Sprintf("Delete server '%s' (%s) from group '%s'?", alias, server.Name, groupName))
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	delete(group.Server, alias)
	cfg.Group[groupKey] = group
	if err := config.Save(opts.ConfigPath, cfg); err != nil {
		return err
	}

	return ui.ShowMessage("Server Deleted", fmt.Sprintf("Server '%s' was removed.", alias))
}
