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

func TestBuildHomeGroupNamesPreservesServerOrderAndAppendsExtraGroups(t *testing.T) {
	items := []HomeServerItem{
		{Label: "prod-web-01", Group: " Prod "},
		{Label: "ops-bastion", Group: "Ops"},
		{Label: "prod-db-01", Group: "Prod"},
	}

	got := buildHomeGroupNames(normalizeHomeItems(items), []string{"Default", "Ops", "  QA  ", ""})
	want := []string{"Prod", "Ops", "Default", "QA"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("group names mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func TestNewHomeScreenModelInitialSelectionAndAdvanceGroup(t *testing.T) {
	items := []HomeServerItem{
		{EntryIndex: 10, Label: "prod-web-01", Group: "Prod"},
		{EntryIndex: 11, Label: "prod-db-01", Group: "Prod"},
		{EntryIndex: 12, Label: "ops-bastion", Group: "Ops"},
	}

	model := newHomeScreenModel(items, []string{"Default", "Prod", "QA", "Ops"})

	if model.selectedServer != 0 {
		t.Fatalf("expected initial selected server to be first server, got %d", model.selectedServer)
	}
	if model.selectedGroup != 0 {
		t.Fatalf("expected initial selected group to match first populated group, got %d", model.selectedGroup)
	}
	if model.groupNames[model.selectedGroup] != "Prod" {
		t.Fatalf("expected initial selected group to be Prod, got %q", model.groupNames[model.selectedGroup])
	}
	if got := model.entryIndex(); got != 10 {
		t.Fatalf("expected first entry index to be selected, got %d", got)
	}

	model.advanceGroup(1)
	if model.groupNames[model.selectedGroup] != "Ops" {
		t.Fatalf("expected advanceGroup to skip empty groups and move to Ops, got %q", model.groupNames[model.selectedGroup])
	}
	if model.selectedServer != 2 {
		t.Fatalf("expected advanceGroup to sync to first server in Ops, got %d", model.selectedServer)
	}
	if got := model.entryIndex(); got != 12 {
		t.Fatalf("expected selected entry index to update with group navigation, got %d", got)
	}

	model.advanceGroup(-1)
	if model.groupNames[model.selectedGroup] != "Prod" {
		t.Fatalf("expected reverse advanceGroup to wrap back to Prod, got %q", model.groupNames[model.selectedGroup])
	}
	if model.selectedServer != 0 {
		t.Fatalf("expected reverse advanceGroup to sync to first Prod server, got %d", model.selectedServer)
	}
}
