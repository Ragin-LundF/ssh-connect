package ui

import (
	"errors"
	"fmt"
	"strings"

	tui "github.com/grindlemire/go-tui"
)

// ErrCancelled is returned when the user dismisses a dialog without confirming.
var ErrCancelled = errors.New("cancelled")

// HomeAction represents the navigation choice made on the home screen.
type HomeAction int

const (
	HomeConnect HomeAction = iota
	HomeAdd
	HomeAddGroup
	HomeDelete
	HomeMenu
	HomeHelp
	HomeQuit
)

// HomeServerItem is a server row rendered in the home screen.
type HomeServerItem struct {
	EntryIndex int
	Label      string
	Group      string
}

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

// SelectServerHome shows the main server-list screen.
// It returns the chosen action, the index of the selected server (-1 if none),
// and ErrCancelled when the user quits without selecting.
func SelectServerHome(configPath string, items []HomeServerItem, groups []string) (HomeAction, int, error) {
	debugf("open home view config=%s servers=%d", configPath, len(items))

	normalizedItems := make([]HomeServerItem, 0, len(items))
	for _, item := range items {
		item.Group = normalizeGroupName(item.Group)
		normalizedItems = append(normalizedItems, item)
	}

	groupNames := make([]string, 0, len(groups)+len(normalizedItems))
	seenGroups := map[string]struct{}{}

	// Keep groups in the same order as the first matching server in the left pane.
	for _, item := range normalizedItems {
		if _, exists := seenGroups[item.Group]; exists {
			continue
		}
		seenGroups[item.Group] = struct{}{}
		groupNames = append(groupNames, item.Group)
	}

	for _, group := range groups {
		trimmed := normalizeGroupName(group)
		if _, exists := seenGroups[trimmed]; exists {
			continue
		}
		seenGroups[trimmed] = struct{}{}
		groupNames = append(groupNames, trimmed)
	}
	if len(groupNames) == 0 {
		groupNames = []string{"Default"}
	}

	body := tui.New(
		tui.WithDisplay(tui.DisplayFlex),
		tui.WithDirection(tui.Column),
		tui.WithGap(1),
		tui.WithFlexGrow(1),
	)

	overviewSection, overviewList := newSection("Server Overview")
	body.AddChild(overviewSection)

	hint := "Enter: Connect | A: Add Server | G: Add Group | D: Delete | M: Menu | Left/Right/Tab: Cycle Groups | Up/Down: Move | ?: Help | Q: Quit"
	root := buildScreenRoot("🔌 SSH Connect", fmt.Sprintf("Config: %s", configPath), body, hint)

	action := HomeQuit
	selectedGroup := 0
	selectedServer := -1
	if len(normalizedItems) > 0 {
		selectedServer = 0
	}

	groupIndexByName := map[string]int{}
	for idx, group := range groupNames {
		groupIndexByName[group] = idx
	}

	serverCountByGroup := map[string]int{}
	for _, item := range normalizedItems {
		serverCountByGroup[item.Group]++
	}

	groupForServer := func(serverIdx int) string {
		if serverIdx < 0 || serverIdx >= len(normalizedItems) {
			return ""
		}
		return normalizedItems[serverIdx].Group
	}

	firstServerForGroup := func(group string) int {
		for idx, item := range normalizedItems {
			if item.Group == group {
				return idx
			}
		}
		return -1
	}

	syncGroupFromServer := func() {
		group := groupForServer(selectedServer)
		if group == "" {
			return
		}
		if idx, exists := groupIndexByName[group]; exists {
			selectedGroup = idx
		}
	}

	syncServerFromGroup := func() {
		if selectedGroup < 0 || selectedGroup >= len(groupNames) {
			return
		}
		serverIdx := firstServerForGroup(groupNames[selectedGroup])
		if serverIdx >= 0 {
			selectedServer = serverIdx
		}
	}

	if len(normalizedItems) > 0 {
		syncGroupFromServer()
	}
	debugf("home groups initialized names=%v selectedGroup=%d selectedServer=%d", groupNames, selectedGroup, selectedServer)

	clampSelectedServer := func() {
		if len(normalizedItems) == 0 {
			selectedServer = -1
			return
		}
		if selectedServer < 0 {
			selectedServer = 0
		}
		if selectedServer >= len(normalizedItems) {
			selectedServer = len(normalizedItems) - 1
		}
	}

	var render func()

	advanceGroup := func(step int) {
		nextGroup := nextGroupIndexWithServers(groupNames, serverCountByGroup, selectedGroup, step)
		if nextGroup >= 0 {
			selectedGroup = nextGroup
			syncServerFromGroup()
		}
		render()
	}

	render = func() {
		clampSelectedServer()
		if len(normalizedItems) > 0 {
			syncGroupFromServer()
		}

		rows := buildHomeOverviewRows(normalizedItems, groupNames)
		debugf("home render overview rows=%d selectedGroup=%d selectedServer=%d", len(rows), selectedGroup, selectedServer)
		renderHomeOverviewList(overviewList, rows, selectedServer, selectedGroup, "No servers found. Press 'A' to add one.")
		scrollHomeOverviewToSelection(overviewList, rows, selectedServer)
	}

	selectedEntryIndex := func() int {
		if len(normalizedItems) == 0 || selectedServer < 0 || selectedServer >= len(normalizedItems) {
			return -1
		}
		return normalizedItems[selectedServer].EntryIndex
	}

	render()

	if err := runUI(root, "home", func(ke tui.KeyEvent) bool {
		moveUp := func() {
			if selectedServer > 0 {
				selectedServer--
				syncGroupFromServer()
			}
			render()
		}

		moveDown := func() {
			if len(normalizedItems) > 0 && selectedServer < 0 {
				selectedServer = 0
				syncGroupFromServer()
				render()
				return
			}
			if selectedServer < len(normalizedItems)-1 {
				selectedServer++
				syncGroupFromServer()
			}
			render()
		}

		switch ke.Key {
		case tui.KeyUp:
			moveUp()
			return true
		case tui.KeyDown:
			moveDown()
			return true
		case tui.KeyLeft:
			advanceGroup(-1)
			return true
		case tui.KeyRight:
			advanceGroup(1)
			return true
		case tui.KeyTab:
			advanceGroup(1)
			return true
		case tui.KeyEnter:
			entryIdx := selectedEntryIndex()
			if entryIdx < 0 {
				debugf("home enter ignored: no server selected")
				return true
			}
			debugf("home connect selected index=%d", entryIdx)
			action = HomeConnect
			ke.App().Stop()
			return true
		case tui.KeyEscape:
			debugf("home action quit (esc)")
			action = HomeQuit
			ke.App().Stop()
			return true
		default:
			// Fall through to rune-based bindings below.
		}

		r, ok := lowerRune(ke)
		if !ok {
			return false
		}

		switch r {
		case 'k':
			moveUp()
			return true
		case 'j':
			moveDown()
			return true
		case 'l':
			advanceGroup(1)
			return true
		case 'h':
			debugf("home action help")
			action = HomeHelp
			ke.App().Stop()
			return true
		case 'a':
			debugf("home action add")
			action = HomeAdd
			ke.App().Stop()
			return true
		case 'g':
			debugf("home action add group")
			action = HomeAddGroup
			ke.App().Stop()
			return true
		case 'd':
			entryIdx := selectedEntryIndex()
			if entryIdx < 0 {
				debugf("home delete ignored: no server")
				return true
			}
			debugf("home action delete index=%d", entryIdx)
			action = HomeDelete
			ke.App().Stop()
			return true
		case 'm':
			debugf("home action menu")
			action = HomeMenu
			ke.App().Stop()
			return true
		case '?':
			debugf("home action help")
			action = HomeHelp
			ke.App().Stop()
			return true
		case 'q':
			debugf("home action quit")
			action = HomeQuit
			ke.App().Stop()
			return true
		default:
			return false
		}
	}); err != nil {
		return HomeQuit, -1, err
	}

	selected := selectedEntryIndex()
	debugf("close home view action=%d selected=%d", action, selected)
	if action == HomeQuit {
		return HomeQuit, -1, ErrCancelled
	}
	return action, selected, nil
}

func normalizeGroupName(group string) string {
	trimmed := strings.TrimSpace(group)
	if trimmed == "" {
		return "Default"
	}
	return trimmed
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

		container.AddChild(tui.New(
			tui.WithText(text),
			tui.WithTextStyle(style),
			tui.WithWrap(false),
		))
	}
}

func buildHomeOverviewRows(items []HomeServerItem, groups []string) []homeOverviewRow {
	serverIndexesByGroup := map[string][]int{}
	for idx, item := range items {
		serverIndexesByGroup[item.Group] = append(serverIndexesByGroup[item.Group], idx)
	}

	rows := make([]homeOverviewRow, 0, len(groups)+len(items))
	for groupIndex, group := range groups {
		serverIndexes := serverIndexesByGroup[group]
		count := len(serverIndexes)
		noun := "servers"
		if count == 1 {
			noun = "server"
		}

		rows = append(rows, homeOverviewRow{
			kind:        homeOverviewRowGroup,
			groupIndex:  groupIndex,
			serverIndex: -1,
			text:        fmt.Sprintf("%s (%d %s)", group, count, noun),
		})

		if count == 0 {
			rows = append(rows, homeOverviewRow{
				kind:        homeOverviewRowEmpty,
				groupIndex:  groupIndex,
				serverIndex: -1,
				text:        "  └─ No servers",
			})
			continue
		}

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
	}
	return rows
}

func scrollHomeOverviewToSelection(container *tui.Element, rows []homeOverviewRow, selectedServer int) {
	if selectedServer < 0 {
		return
	}

	selectedRow := -1
	anchorRow := -1
	for idx, row := range rows {
		if row.serverIndex == selectedServer {
			selectedRow = idx
			anchorRow = idx
			if idx > 0 {
				prevRow := rows[idx-1]
				if prevRow.kind == homeOverviewRowGroup && prevRow.groupIndex == row.groupIndex {
					anchorRow = idx - 1
				}
			}
			break
		}
	}
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
