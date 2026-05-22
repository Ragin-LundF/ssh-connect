package ssh

import (
	"fmt"
	"os"
	"os/exec"

	"ssh_connect/internal/config"
)

func BuildCommand(server config.Server) ([]string, error) {
	target := fmt.Sprintf("%s@%s", server.User, server.IP)
	if server.Certificate == "" {
		return []string{"ssh", target}, nil
	}
	if _, err := os.Stat(server.Certificate); err != nil {
		return nil, fmt.Errorf("certificate file not found: %s", server.Certificate)
	}
	return []string{"ssh", "-o", "IdentityFile=" + server.Certificate, target}, nil
}

func Exec(args []string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
