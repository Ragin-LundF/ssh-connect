package config

type Server struct {
	Name        string `toml:"name"`
	IP          string `toml:"ip"`
	User        string `toml:"user"`
	Certificate string `toml:"certificate,omitempty"`
}

type Group struct {
	Name             string   `toml:"name,omitempty"`
	Servers          []string `toml:"servers,omitempty"`
	GroupCertificate string   `toml:"group_certificate,omitempty"`
}

type File struct {
	Server map[string]Server `toml:"server"`
	Group  map[string]Group  `toml:"group,omitempty"`
}

type ServerEntry struct {
	Key    string
	Server Server
}
