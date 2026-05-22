package config

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

func Load(path string) (File, error) {
	var cfg File
	if _, err := os.Stat(path); err != nil {
		return File{}, err
	}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return File{}, fmt.Errorf("failed to parse config: %w", err)
	}
	normalize(&cfg)
	return cfg, nil
}

func Save(path string, cfg File) error {
	normalize(&cfg)

	// Persist only grouped servers and keep legacy fields empty.
	cfg.Server = nil
	for key, group := range cfg.Group {
		group.Servers = nil
		cfg.Group[key] = group
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

func ExampleConfig() File {
	return File{
		Group: map[string]Group{
			DefaultGroupName: {
				Name: DefaultGroupName,
				Server: map[string]Server{
					"app_prod": {
						Name:        "App Production",
						IP:          "203.0.113.10",
						User:        "deploy",
						Certificate: "/Users/youruser/.ssh/app_prod.pem",
					},
				},
			},
			"production": {
				Name:             "Production",
				GroupCertificate: "/Users/youruser/.ssh/prod_shared.pem",
				Server: map[string]Server{
					"db_prod": {
						Name: "DB Production",
						IP:   "203.0.113.20",
						User: "dbadmin",
					},
				},
			},
		},
	}
}

func ToEntries(cfg File) []ServerEntry {
	normalize(&cfg)

	entries := make([]ServerEntry, 0)
	for groupKey, group := range cfg.Group {
		groupName := strings.TrimSpace(group.Name)
		if groupName == "" {
			groupName = groupKey
		}
		for alias, server := range group.Server {
			if strings.TrimSpace(server.IP) == "" || strings.TrimSpace(server.User) == "" {
				continue
			}
			if strings.TrimSpace(server.Name) == "" {
				server.Name = alias
			}
			if strings.TrimSpace(server.Certificate) == "" {
				server.Certificate = strings.TrimSpace(group.GroupCertificate)
			}
			entries = append(entries, ServerEntry{Key: alias, GroupKey: groupKey, GroupName: groupName, Server: server})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].GroupName != entries[j].GroupName {
			if entries[i].GroupName == DefaultGroupName {
				return true
			}
			if entries[j].GroupName == DefaultGroupName {
				return false
			}
			return entries[i].GroupName < entries[j].GroupName
		}
		return entries[i].Server.Name < entries[j].Server.Name
	})
	return entries
}

func GroupNames(cfg File) []string {
	normalize(&cfg)
	names := make([]string, 0, len(cfg.Group))
	for key, group := range cfg.Group {
		name := strings.TrimSpace(group.Name)
		if name == "" {
			name = key
		}
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		if names[i] == DefaultGroupName {
			return true
		}
		if names[j] == DefaultGroupName {
			return false
		}
		return names[i] < names[j]
	})
	return names
}

func EnsureGroup(cfg *File, name string) string {
	if cfg.Group == nil {
		cfg.Group = map[string]Group{}
	}
	if cfg.Server == nil {
		cfg.Server = map[string]Server{}
	}
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		trimmed = DefaultGroupName
	}
	for key, group := range cfg.Group {
		groupName := strings.TrimSpace(group.Name)
		if groupName == "" {
			groupName = key
		}
		if strings.EqualFold(groupName, trimmed) {
			if group.Server == nil {
				group.Server = map[string]Server{}
				cfg.Group[key] = group
			}
			return key
		}
	}
	cfg.Group[trimmed] = Group{Name: trimmed, Server: map[string]Server{}}
	return trimmed
}

func IsValidGroupName(name string) bool {
	if strings.TrimSpace(name) == "" {
		return false
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_' || r == '-' || r == ' ':
		default:
			return false
		}
	}
	return true
}

func IsValidAlias(alias string) bool {
	if alias == "" {
		return false
	}
	for _, r := range alias {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_' || r == '-':
		default:
			return false
		}
	}
	return true
}

func normalize(cfg *File) {
	if cfg.Group == nil {
		cfg.Group = map[string]Group{}
	}
	if cfg.Server == nil {
		cfg.Server = map[string]Server{}
	}

	defaultKey := EnsureGroup(cfg, DefaultGroupName)

	for groupKey, group := range cfg.Group {
		if strings.TrimSpace(group.Name) == "" {
			group.Name = groupKey
		}
		if group.Server == nil {
			group.Server = map[string]Server{}
		}

		// Migrate legacy group server references like "server.alias".
		for _, ref := range group.Servers {
			alias := strings.TrimPrefix(strings.TrimSpace(ref), "server.")
			if alias == "" {
				continue
			}
			if _, exists := group.Server[alias]; exists {
				continue
			}
			if srv, exists := cfg.Server[alias]; exists {
				group.Server[alias] = srv
			}
		}
		cfg.Group[groupKey] = group
	}

	defaultGroup := cfg.Group[defaultKey]
	if defaultGroup.Server == nil {
		defaultGroup.Server = map[string]Server{}
	}

	for alias, srv := range cfg.Server {
		migrated := false
		for groupKey, group := range cfg.Group {
			if _, exists := group.Server[alias]; exists {
				migrated = true
				cfg.Group[groupKey] = group
				break
			}
		}
		if !migrated {
			defaultGroup.Server[alias] = srv
		}
	}
	cfg.Group[defaultKey] = defaultGroup
}
