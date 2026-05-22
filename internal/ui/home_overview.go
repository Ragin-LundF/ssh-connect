package ui

import (
	"fmt"

	tui "github.com/grindlemire/go-tui"
)

type homeOverviewRowKind int

const (
	homeOverviewRowGroup homeOverviewRowKind = iota
	homeOverviewRowServer
	homeOverviewRowEmpty
)

type homeOverviewRow struct {
	kind        homeOverviewRowKind
	groupIndex  int
	serverIndex int
	text        string
}

func renderHomeOverviewList(container *tui.Element, rows []homeOverviewRow, selectedServer int, selectedGroup int, emptyMessage string) {
	container.RemoveAllChildren()
	if len(rows) == 0 {
		container.AddChild(tui.New(
			tui.WithText("  ◌ "+emptyMessage),
			tui.WithTextStyle(tui.NewStyle().Foreground(tui.BrightBlack)),
			tui.WithWrap(false),
		))
		return
	}

	for _, row := range rows {
		container.AddChild(newHomeOverviewRowElement(row, selectedServer, selectedGroup))
	}
}

func newHomeOverviewRowElement(row homeOverviewRow, selectedServer int, selectedGroup int) *tui.Element {
	text := row.text
	style := tui.NewStyle().Foreground(tui.White)

	switch row.kind {
	case homeOverviewRowGroup:
		prefix := "▸"
		style = tui.NewStyle().Foreground(tui.BrightCyan).Bold()
		if row.groupIndex == selectedGroup {
			prefix = "▾"
			style = tui.NewStyle().Foreground(tui.BrightCyan).Background(tui.BrightBlack).Bold()
		}
		text = prefix + " " + text
	case homeOverviewRowServer:
		if row.serverIndex == selectedServer {
			style = tui.NewStyle().Foreground(tui.BrightGreen).Background(tui.BrightBlack).Bold()
		}
	case homeOverviewRowEmpty:
		style = tui.NewStyle().Foreground(tui.BrightBlack).Dim()
	}

	return tui.New(
		tui.WithText(text),
		tui.WithTextStyle(style),
		tui.WithWrap(false),
	)
}

func buildHomeOverviewRows(items []HomeServerItem, groups []string) []homeOverviewRow {
	serverIndexesByGroup := buildServerIndexesByGroup(items)
	rows := make([]homeOverviewRow, 0, len(groups)+len(items))

	for groupIndex, group := range groups {
		serverIndexes := serverIndexesByGroup[group]
		rows = append(rows, newHomeGroupRow(group, groupIndex, len(serverIndexes)))
		rows = append(rows, buildHomeServerRows(items, groupIndex, serverIndexes)...)
	}

	return rows
}

func buildServerIndexesByGroup(items []HomeServerItem) map[string][]int {
	serverIndexesByGroup := map[string][]int{}
	for idx, item := range items {
		serverIndexesByGroup[item.Group] = append(serverIndexesByGroup[item.Group], idx)
	}
	return serverIndexesByGroup
}

func newHomeGroupRow(group string, groupIndex int, serverCount int) homeOverviewRow {
	noun := "servers"
	if serverCount == 1 {
		noun = "server"
	}

	return homeOverviewRow{
		kind:        homeOverviewRowGroup,
		groupIndex:  groupIndex,
		serverIndex: -1,
		text:        fmt.Sprintf("%s (%d %s)", group, serverCount, noun),
	}
}

func buildHomeServerRows(items []HomeServerItem, groupIndex int, serverIndexes []int) []homeOverviewRow {
	if len(serverIndexes) == 0 {
		return []homeOverviewRow{{
			kind:        homeOverviewRowEmpty,
			groupIndex:  groupIndex,
			serverIndex: -1,
			text:        "  └─ No servers",
		}}
	}

	rows := make([]homeOverviewRow, 0, len(serverIndexes))
	for idx, serverIndex := range serverIndexes {
		branch := "├─"
		if idx == len(serverIndexes)-1 {
			branch = "└─"
		}
		rows = append(rows, homeOverviewRow{
			kind:        homeOverviewRowServer,
			groupIndex:  groupIndex,
			serverIndex: serverIndex,
			text:        fmt.Sprintf("  %s %s", branch, items[serverIndex].Label),
		})
	}
	return rows
}

func scrollHomeOverviewToSelection(container *tui.Element, rows []homeOverviewRow, selectedServer int) {
	if selectedServer < 0 {
		return
	}

	selectedRow, anchorRow := findHomeSelectionRows(rows, selectedServer)
	if selectedRow < 0 {
		return
	}

	_, viewportHeight := container.ViewportSize()
	if viewportHeight <= 0 {
		return
	}

	_, scrollY := container.ScrollOffset()
	switch {
	case anchorRow < scrollY:
		container.ScrollTo(0, anchorRow)
	case selectedRow >= scrollY+viewportHeight:
		container.ScrollTo(0, selectedRow-viewportHeight+1)
	}
}

func findHomeSelectionRows(rows []homeOverviewRow, selectedServer int) (selectedRow int, anchorRow int) {
	for idx, row := range rows {
		if row.serverIndex != selectedServer {
			continue
		}

		selectedRow = idx
		anchorRow = idx
		if idx > 0 {
			prevRow := rows[idx-1]
			if prevRow.kind == homeOverviewRowGroup && prevRow.groupIndex == row.groupIndex {
				anchorRow = idx - 1
			}
		}
		return selectedRow, anchorRow
	}

	return -1, -1
}

func nextGroupIndexWithServers(groups []string, serverCountByGroup map[string]int, current int, step int) int {
	if len(groups) == 0 || step == 0 {
		return -1
	}

	if current < 0 || current >= len(groups) {
		current = 0
	}

	for offset := 1; offset <= len(groups); offset++ {
		idx := current + (offset * step)
		idx %= len(groups)
		if idx < 0 {
			idx += len(groups)
		}
		if serverCountByGroup[groups[idx]] > 0 {
			return idx
		}
	}

	return -1
}
