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
	if cfg.Server == nil {
		cfg.Server = map[string]Server{}
	}
	if cfg.Group == nil {
		cfg.Group = map[string]Group{}
	}
	return cfg, nil
}

func Save(path string, cfg File) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

func ExampleConfig() File {
	return File{
		Server: map[string]Server{
			"app_prod": {
				Name:        "App Production",
				IP:          "203.0.113.10",
				User:        "deploy",
				Certificate: "/Users/youruser/.ssh/app_prod.pem",
			},
			"db_prod": {
				Name:        "DB Production",
				IP:          "203.0.113.20",
				User:        "dbadmin",
				Certificate: "/Users/youruser/.ssh/db_prod.pem",
			},
		},
		Group: map[string]Group{
			"production": {
				Name:             "Production",
				Servers:          []string{"server.app_prod", "server.db_prod"},
				GroupCertificate: "/Users/youruser/.ssh/prod_shared.pem",
			},
		},
	}
}

func ToEntries(cfg File) []ServerEntry {
	entries := make([]ServerEntry, 0, len(cfg.Server))
	for key, server := range cfg.Server {
		if strings.TrimSpace(server.IP) == "" || strings.TrimSpace(server.User) == "" {
			continue
		}
		if strings.TrimSpace(server.Name) == "" {
			server.Name = key
		}
		entries = append(entries, ServerEntry{Key: key, Server: server})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Server.Name < entries[j].Server.Name
	})
	return entries
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
