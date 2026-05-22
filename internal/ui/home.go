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
		tui.WithDirection(tui.Row),
		tui.WithGap(1),
		tui.WithFlexGrow(1),
	)

	serverSection, serverList := newSection("Available Servers")
	groupSection, groupList := newStaticSection("Available Groups")
	body.AddChild(serverSection)
	body.AddChild(groupSection)

	hint := "Enter: Connect | A: Add Server | G: Add Group | D: Delete | M: Menu | Left/Right/Tab: Switch Pane | Up/Down: Move | ?: Help | Q: Quit"
	root := buildScreenRoot("🔌 SSH Connect", fmt.Sprintf("Config: %s", configPath), body, hint)

	action := HomeQuit
	selectedGroup := 0
	selectedServer := 0
	focusGroups := false

	groupIndexByName := map[string]int{}
	for idx, group := range groupNames {
		groupIndexByName[group] = idx
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

	serverRows := func() []string {
		rows := make([]string, 0, len(normalizedItems))
		prevGroup := ""
		for _, item := range normalizedItems {
			sep := "│"
			if prevGroup != "" && prevGroup != item.Group {
				sep = "╎"
			}
			prevGroup = item.Group
			rows = append(rows, fmt.Sprintf("%s %s", sep, item.Label))
		}
		return rows
	}

	render := func() {
		groupRows := buildCompactGroupRows(normalizedItems, groupNames)
		debugf("home render groups rows=%v selectedGroup=%d focusGroups=%t", groupRows, selectedGroup, focusGroups)
		renderHomeGroupPane(groupList, groupRows, selectedGroup, focusGroups, "No groups")

		rows := serverRows()
		if selectedServer >= len(rows) {
			selectedServer = 0
		}
		renderHomeList(serverList, rows, selectedServer, !focusGroups, "No servers found. Press 'A' to add one.")
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
			if focusGroups {
				if selectedGroup > 0 {
					selectedGroup--
					syncServerFromGroup()
				}
				render()
				return
			}
			if selectedServer > 0 {
				selectedServer--
				syncGroupFromServer()
			}
			render()
		}

		moveDown := func() {
			if focusGroups {
				if selectedGroup < len(groupNames)-1 {
					selectedGroup++
					syncServerFromGroup()
				}
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
			focusGroups = false
			render()
			return true
		case tui.KeyRight:
			focusGroups = true
			syncGroupFromServer()
			render()
			return true
		case tui.KeyTab:
			focusGroups = !focusGroups
			if focusGroups {
				syncGroupFromServer()
			}
			render()
			return true
		case tui.KeyEnter:
			if focusGroups {
				syncServerFromGroup()
				debugf("home focus switched to servers from groups pane")
				focusGroups = false
				render()
				return true
			}
			entryIdx := selectedEntryIndex()
			if entryIdx < 0 {
				debugf("home enter ignored: no server in selected group")
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
			focusGroups = true
			syncGroupFromServer()
			render()
			return true
		case 'h':
			if focusGroups {
				focusGroups = false
				render()
				return true
			}
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

func renderHomeList(container *tui.Element, items []string, selected int, active bool, emptyMessage string) {
	container.RemoveAllChildren()
	if len(items) == 0 {
		container.AddChild(tui.New(
			tui.WithText("  ◌ "+emptyMessage),
			tui.WithTextStyle(tui.NewStyle().Foreground(tui.BrightBlack)),
			tui.WithWrap(false),
		))
		return
	}

	for idx, item := range items {
		prefix := "  ◌ "
		style := tui.NewStyle().Foreground(tui.White)
		if idx == selected {
			if active {
				prefix = "  ◉ "
				style = tui.NewStyle().Foreground(tui.BrightGreen).Background(tui.BrightBlack).Bold()
			} else {
				prefix = "  ◉ "
				style = tui.NewStyle().Foreground(tui.BrightCyan)
			}
		}
		container.AddChild(tui.New(
			tui.WithText(prefix+item),
			tui.WithTextStyle(style),
			tui.WithWrap(false),
		))
	}
}

func renderHomeGroupPane(container *tui.Element, items []string, selected int, active bool, emptyMessage string) {
	container.RemoveAllChildren()
	if len(items) == 0 {
		container.AddChild(tui.New(
			tui.WithText("  - "+emptyMessage),
			tui.WithTextStyle(tui.NewStyle().Foreground(tui.BrightBlack)),
			tui.WithWrap(false),
		))
		return
	}

	lines := make([]string, 0, len(items))
	for idx, item := range items {
		prefix := "  "
		if idx == selected {
			if active {
				prefix = "> "
			} else {
				prefix = "* "
			}
		}
		lines = append(lines, prefix+item)
	}

	style := tui.NewStyle().Foreground(tui.White)
	if active {
		style = tui.NewStyle().Foreground(tui.BrightGreen)
	}

	container.AddChild(tui.New(
		tui.WithText(strings.Join(lines, "\n")),
		tui.WithTextStyle(style),
		tui.WithWrap(false),
	))
}

func buildCompactGroupRows(items []HomeServerItem, groups []string) []string {
	serverCountByGroup := map[string]int{}
	for _, item := range items {
		serverCountByGroup[item.Group]++
	}

	rows := make([]string, 0, len(groups))
	for _, group := range groups {
		count := serverCountByGroup[group]
		row := fmt.Sprintf("%s (%d)", group, count)
		rows = append(rows, row)
	}
	return rows
}
