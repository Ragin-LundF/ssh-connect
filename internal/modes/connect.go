package modes

import (
	"fmt"
	"os"

	"ssh_connect/internal/cli"
	"ssh_connect/internal/config"
	"ssh_connect/internal/ssh"
	"ssh_connect/internal/ui"
)

func Connect(opts cli.Options) error {
	for {
		cfg, err := config.Load(opts.ConfigPath)
		if err != nil {
			if os.IsNotExist(err) {
				ok, err := ui.Confirm(
					"Config Missing",
					fmt.Sprintf("Config file '%s' was not found. Do you want to add a server now?", opts.ConfigPath),
				)
				if err != nil {
					return err
				}
				if !ok {
					return nil
				}
				if err := Add(opts); err != nil {
					return err
				}
				continue
			}
			return err
		}

		entries := config.ToEntries(cfg)
		homeItems := buildHomeItems(entries)
		groups := config.GroupNames(cfg)

		action, idx, err := ui.SelectServerHome(opts.ConfigPath, homeItems, groups)
		if err != nil {
			if err == ui.ErrCancelled {
				return nil
			}
			return err
		}

		switch action {
		case ui.HomeConnect:
			if idx < 0 || idx >= len(entries) {
				if err := ui.ShowMessage("No Servers", "Please add a server first."); err != nil {
					return err
				}
				continue
			}
			if err := connectToEntry(entries[idx], opts.DryRun); err != nil {
				return err
			}
			if opts.DryRun {
				continue
			}
			return nil
		case ui.HomeAdd:
			if err := Add(opts); err != nil {
				return err
			}
		case ui.HomeAddGroup:
			if err := AddGroup(opts); err != nil {
				return err
			}
		case ui.HomeDelete:
			if idx < 0 || idx >= len(entries) {
				continue
			}
			if err := DeleteAlias(opts, entries[idx].GroupKey, entries[idx].Key); err != nil {
				return err
			}
		case ui.HomeHelp:
			if err := ui.ShowMessage("Help", HelpText()); err != nil {
				return err
			}
		case ui.HomeMenu:
			menuAction, err := ui.SelectMainMenu()
			if err != nil {
				return err
			}
			if shouldQuit, err := executeMenuAction(menuAction, opts, entries, idx); err != nil {
				return err
			} else if shouldQuit {
				return nil
			}
		case ui.HomeQuit:
			return nil
		}
	}
}

func connectToEntry(entry config.ServerEntry, dryRun bool) error {
	args, err := ssh.BuildCommand(entry.Server)
	if err != nil {
		if dryRun {
			return ui.ShowMessage("Warning", err.Error())
		}
		return err
	}

	if dryRun {
		line := ""
		for _, part := range args {
			line += fmt.Sprintf(" %q", part)
		}
		return ui.ShowMessage("Dry Run", "Command will not be executed.\nWould run:"+line)
	}

	fmt.Printf("Connecting to %s (%s@%s)\n", entry.Server.Name, entry.Server.User, entry.Server.IP)
	return ssh.Exec(args)
}

func executeMenuAction(action ui.MenuAction, opts cli.Options, entries []config.ServerEntry, selected int) (bool, error) {
	switch action {
	case ui.MenuConnect:
		if selected < 0 || selected >= len(entries) {
			if err := ui.ShowMessage("No Servers", "Please add a server first."); err != nil {
				return false, err
			}
			return false, nil
		}
		if err := connectToEntry(entries[selected], opts.DryRun); err != nil {
			return false, err
		}
		if opts.DryRun {
			return false, nil
		}
		return true, nil
	case ui.MenuAdd:
		return false, Add(opts)
	case ui.MenuAddGroup:
		return false, AddGroup(opts)
	case ui.MenuDelete:
		if selected < 0 || selected >= len(entries) {
			if err := ui.ShowMessage("No Servers", "There is no server to delete."); err != nil {
				return false, err
			}
			return false, nil
		}
		return false, DeleteAlias(opts, entries[selected].GroupKey, entries[selected].Key)
	case ui.MenuHelp:
		return false, ui.ShowMessage("Help", HelpText())
	case ui.MenuQuit:
		return true, nil
	case ui.MenuList, ui.MenuBack:
		return false, nil
	default:
		return false, nil
	}
}

func buildHomeItems(entries []config.ServerEntry) []ui.HomeServerItem {
	const maxLabelWidth = 62

	items := make([]ui.HomeServerItem, 0, len(entries))
	for idx, entry := range entries {
		label := fmt.Sprintf("%s (%s@%s)", entry.Server.Name, entry.Server.User, entry.Server.IP)
		runes := []rune(label)
		if len(runes) > maxLabelWidth {
			label = string(runes[:maxLabelWidth-3]) + "..."
		}
		items = append(items, ui.HomeServerItem{
			EntryIndex: idx,
			Label:      label,
			Group:      entry.GroupName,
		})
	}
	return items
}
