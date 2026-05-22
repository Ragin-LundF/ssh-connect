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
