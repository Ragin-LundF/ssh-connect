package config

type Server struct {
	Name        string `toml:"name"`
	IP          string `toml:"ip"`
	User        string `toml:"user"`
	Certificate string `toml:"certificate,omitempty"`
}

type Group struct {
	Name             string            `toml:"name,omitempty"`
	Server           map[string]Server `toml:"server,omitempty"`
	Servers          []string          `toml:"servers,omitempty"`
	GroupCertificate string            `toml:"group_certificate,omitempty"`
}

type File struct {
	Group  map[string]Group  `toml:"group"`
	Server map[string]Server `toml:"server,omitempty"`
}

type ServerEntry struct {
	Key       string
	GroupKey  string
	GroupName string
	Server    Server
}

const DefaultGroupName = "Default"
