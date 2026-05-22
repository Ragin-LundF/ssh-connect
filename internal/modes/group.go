package modes

import (
	"fmt"
	"os"
	"strings"

	"ssh_connect/internal/cli"
	"ssh_connect/internal/config"
	"ssh_connect/internal/ui"
)

const createGroupChoice = "Create new group..."

func AddGroup(opts cli.Options) error {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg = config.File{Group: map[string]config.Group{}}
		} else {
			return err
		}
	}

	groupName, err := promptGroupName(cfg)
	if err != nil {
		if err == ui.ErrCancelled {
			return nil
		}
		return err
	}

	config.EnsureGroup(&cfg, groupName)
	if err := config.Save(opts.ConfigPath, cfg); err != nil {
		return err
	}

	return ui.ShowMessage("Group Added", fmt.Sprintf("Group '%s' was saved.", groupName))
}

func selectOrCreateGroup(cfg *config.File, title string) (string, error) {
	names := config.GroupNames(*cfg)
	choices := append([]string{}, names...)
	choices = append(choices, createGroupChoice)

	idx, err := ui.SelectIndex(title, "Choose a group for the server", choices)
	if err != nil {
		return "", err
	}

	if choices[idx] == createGroupChoice {
		groupName, err := promptGroupName(*cfg)
		if err != nil {
			return "", err
		}
		return config.EnsureGroup(cfg, groupName), nil
	}

	return config.EnsureGroup(cfg, choices[idx]), nil
}

func promptGroupName(cfg config.File) (string, error) {
	for {
		name, err := ui.PromptInput("Add Group", "Group name:", true)
		if err != nil {
			return "", err
		}
		name = strings.TrimSpace(name)
		if !config.IsValidGroupName(name) {
			if err := ui.ShowMessage("Invalid Group Name", "Use letters, numbers, spaces, '_' and '-'."); err != nil {
				return "", err
			}
			continue
		}

		exists := false
		for _, groupName := range config.GroupNames(cfg) {
			if strings.EqualFold(groupName, name) {
				exists = true
				break
			}
		}
		if exists {
			if err := ui.ShowMessage("Group Exists", fmt.Sprintf("Group '%s' already exists.", name)); err != nil {
				return "", err
			}
			continue
		}
		return name, nil
	}
}
