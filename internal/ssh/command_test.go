package ssh

import (
	"testing"

	"ssh_connect/internal/config"
)

func TestBuildCommandWithoutCertificate(t *testing.T) {
	args, err := BuildCommand(config.Server{User: "deploy", IP: "example.org"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(args) != 2 {
		t.Fatalf("unexpected arg length: %d", len(args))
	}
	if args[0] != "ssh" || args[1] != "deploy@example.org" {
		t.Fatalf("unexpected args: %v", args)
	}
}
