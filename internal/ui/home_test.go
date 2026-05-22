package ui

import (
	"strings"
	"testing"
)

func TestBuildCompactGroupRowsShowsGroupName(t *testing.T) {
	items := []HomeServerItem{
		{Label: "prod-web-01", Group: "Prod"},
		{Label: "qa-db-01", Group: "QA"},
	}
	groups := []string{"Prod", "QA", "Default"}

	rows := buildCompactGroupRows(items, groups)
	if len(rows) != len(groups) {
		t.Fatalf("expected %d rows, got %d", len(groups), len(rows))
	}

	for idx, group := range groups {
		if !strings.Contains(rows[idx], group) {
			t.Fatalf("row %d does not include group %q: %q", idx, group, rows[idx])
		}
		if strings.Contains(rows[idx], "[") || strings.Contains(rows[idx], "]") {
			t.Fatalf("row %d should not include square brackets: %q", idx, rows[idx])
		}
	}
}
