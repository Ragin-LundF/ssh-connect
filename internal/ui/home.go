package ui

import (
	"errors"
	"fmt"

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

	model := newHomeScreenModel(items, groups)
	controller := newHomeController(model)

	body := tui.New(
		tui.WithDisplay(tui.DisplayFlex),
		tui.WithDirection(tui.Column),
		tui.WithGap(1),
		tui.WithFlexGrow(1),
	)

	overviewSection, overviewList := newSection("Server Overview")
	body.AddChild(overviewSection)
	controller.attachOverviewList(overviewList)

	hint := "Enter: Connect | A: Add Server | G: Add Group | D: Delete | M: Menu | Left/Right/Tab: Cycle Groups | Up/Down: Move | ?: Help | Q: Quit"
	root := buildScreenRoot("🔌 SSH Connect", fmt.Sprintf("Config: %s", configPath), body, hint)

	controller.render()

	if err := runUI(root, "home", controller.handleKey); err != nil {
		return HomeQuit, -1, err
	}

	selected := controller.selectedEntryIndex()
	debugf("close home view action=%d selected=%d", controller.action, selected)
	if controller.action == HomeQuit {
		return HomeQuit, -1, ErrCancelled
	}
	return controller.action, selected, nil
}
