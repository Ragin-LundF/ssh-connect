package config

import "testing"

func TestIsValidAlias(t *testing.T) {
	valid := []string{"app_prod", "demo-1", "A1"}
	invalid := []string{"", "bad alias", "bad.dot", "bad/slash"}

	for _, v := range valid {
		if !IsValidAlias(v) {
			t.Fatalf("expected valid alias: %s", v)
		}
	}

	for _, v := range invalid {
		if IsValidAlias(v) {
			t.Fatalf("expected invalid alias: %s", v)
		}
	}
}

func TestIsValidGroupName(t *testing.T) {
	valid := []string{"Default", "Prod Team", "group-1", "Ops_1"}
	invalid := []string{"", "*bad", "bad.dot", "bad/slash"}

	for _, v := range valid {
		if !IsValidGroupName(v) {
			t.Fatalf("expected valid group name: %s", v)
		}
	}

	for _, v := range invalid {
		if IsValidGroupName(v) {
			t.Fatalf("expected invalid group name: %s", v)
		}
	}
}

func TestToEntriesMigratesLegacyServersToDefaultGroup(t *testing.T) {
	cfg := File{
		Server: map[string]Server{
			"legacy": {Name: "Legacy", IP: "192.0.2.10", User: "alice"},
		},
	}

	entries := ToEntries(cfg)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].GroupName != DefaultGroupName {
		t.Fatalf("expected group %q, got %q", DefaultGroupName, entries[0].GroupName)
	}
	if entries[0].Key != "legacy" {
		t.Fatalf("expected alias legacy, got %s", entries[0].Key)
	}
}
