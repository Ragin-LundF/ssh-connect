package ui

import (
	"reflect"
	"testing"
)

func TestBuildHomeOverviewRowsGroupsServersTogether(t *testing.T) {
	items := []HomeServerItem{
		{Label: "prod-web-01", Group: "Prod"},
		{Label: "prod-db-01", Group: "Prod"},
		{Label: "qa-db-01", Group: "QA"},
	}
	groups := []string{"Prod", "QA", "Default"}

	rows := buildHomeOverviewRows(items, groups)
	got := make([]string, 0, len(rows))
	gotKinds := make([]homeOverviewRowKind, 0, len(rows))
	gotServerIndexes := make([]int, 0, len(rows))
	for _, row := range rows {
		got = append(got, row.text)
		gotKinds = append(gotKinds, row.kind)
		gotServerIndexes = append(gotServerIndexes, row.serverIndex)
	}

	want := []string{
		"Prod (2 servers)",
		"  ├─ prod-web-01",
		"  └─ prod-db-01",
		"QA (1 server)",
		"  └─ qa-db-01",
		"Default (0 servers)",
		"  └─ No servers",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("overview rows mismatch\n got: %#v\nwant: %#v", got, want)
	}

	wantKinds := []homeOverviewRowKind{
		homeOverviewRowGroup,
		homeOverviewRowServer,
		homeOverviewRowServer,
		homeOverviewRowGroup,
		homeOverviewRowServer,
		homeOverviewRowGroup,
		homeOverviewRowEmpty,
	}
	if !reflect.DeepEqual(gotKinds, wantKinds) {
		t.Fatalf("overview row kinds mismatch\n got: %#v\nwant: %#v", gotKinds, wantKinds)
	}

	wantServerIndexes := []int{-1, 0, 1, -1, 2, -1, -1}
	if !reflect.DeepEqual(gotServerIndexes, wantServerIndexes) {
		t.Fatalf("overview row server indexes mismatch\n got: %#v\nwant: %#v", gotServerIndexes, wantServerIndexes)
	}
}

func TestNextGroupIndexWithServersWrapsAndSkipsEmptyGroups(t *testing.T) {
	groups := []string{"Default", "Prod", "QA", "Ops"}
	serverCountByGroup := map[string]int{
		"Default": 0,
		"Prod":    2,
		"QA":      0,
		"Ops":     1,
	}

	if got := nextGroupIndexWithServers(groups, serverCountByGroup, 1, 1); got != 3 {
		t.Fatalf("expected next non-empty group after Prod to be Ops, got %d", got)
	}

	if got := nextGroupIndexWithServers(groups, serverCountByGroup, 3, 1); got != 1 {
		t.Fatalf("expected wrap from Ops back to Prod, got %d", got)
	}

	if got := nextGroupIndexWithServers(groups, serverCountByGroup, 1, -1); got != 3 {
		t.Fatalf("expected previous non-empty group before Prod to wrap to Ops, got %d", got)
	}

	if got := nextGroupIndexWithServers(groups, map[string]int{"Default": 0}, 0, 1); got != -1 {
		t.Fatalf("expected -1 when no group contains servers, got %d", got)
	}
}
