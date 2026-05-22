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
	HomeDelete
	HomeMenu
	HomeHelp
	HomeQuit
)

// SelectServerHome shows the main server-list screen.
// It returns the chosen action, the index of the selected server (-1 if none),
// and ErrCancelled when the user quits without selecting.
func SelectServerHome(configPath string, items []string) (HomeAction, int, error) {
	debugf("open home view config=%s servers=%d", configPath, len(items))

	hasServers := len(items) > 0
	listItems := items
	if !hasServers {
		listItems = []string{"No servers found. Press A to add one."}
	}

	section, listContainer := newSection("Servers")
	renderList(listContainer, listItems, 0)

	hint := "Enter: Connect | A: Add | D: Delete | M: Menu | H: Help | Up/Down: Move | Q/Esc: Quit"
	root := buildScreenRoot("SSH Connect", fmt.Sprintf("Config: %s", configPath), section, hint)

	action := HomeQuit
	selected := 0

	if err := runUI(root, "home", func(ke tui.KeyEvent) bool {
		switch ke.Key {
		case tui.KeyUp:
			if hasServers && selected > 0 {
				selected--
				renderList(listContainer, listItems, selected)
			}
			return true
		case tui.KeyDown:
			if hasServers && selected < len(listItems)-1 {
				selected++
				renderList(listContainer, listItems, selected)
			}
			return true
		case tui.KeyEnter:
			if !hasServers {
				debugf("home enter ignored: no servers")
				return true
			}
			debugf("home connect selected index=%d", selected)
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
			if hasServers && selected > 0 {
				selected--
				renderList(listContainer, listItems, selected)
			}
			return true
		case 'j':
			if hasServers && selected < len(listItems)-1 {
				selected++
				renderList(listContainer, listItems, selected)
			}
			return true
		case 'a':
			debugf("home action add")
			action = HomeAdd
			ke.App().Stop()
			return true
		case 'd':
			if !hasServers {
				debugf("home delete ignored: no servers")
				return true
			}
			debugf("home action delete index=%d", selected)
			action = HomeDelete
			ke.App().Stop()
			return true
		case 'm':
			debugf("home action menu")
			action = HomeMenu
			ke.App().Stop()
			return true
		case 'h':
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

	if !hasServers {
		selected = -1
	}
	debugf("close home view action=%d selected=%d", action, selected)
	if action == HomeQuit {
		return HomeQuit, -1, ErrCancelled
	}
	return action, selected, nil
}
