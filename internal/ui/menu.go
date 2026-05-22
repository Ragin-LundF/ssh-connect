package ui

import (
	tui "github.com/grindlemire/go-tui"
)

// MenuAction represents the choice made in the main menu.
type MenuAction int

const (
	MenuConnect MenuAction = iota
	MenuAdd
	MenuAddGroup
	MenuDelete
	MenuList
	MenuHelp
	MenuQuit
	MenuBack
)

// SelectMainMenu displays the main-menu dialog and returns the chosen action.
// Escape maps to MenuBack.
func SelectMainMenu() (MenuAction, error) {
	debugf("open main menu dialog")

	choices := []string{
		"Connect to selected server",
		"Add server",
		"Add group",
		"Delete selected server",
		"Back to server list",
		"Help",
		"Quit",
	}

	section, listContainer := newSection("Actions")
	selected := 0
	renderList(listContainer, choices, selected)

	root := buildScreenRoot("Main Menu", "", section, "Up/Down: Move | Enter: Select | G: Add Group | Esc: Back")
	action := MenuBack

	if err := runUI(root, "main-menu", func(ke tui.KeyEvent) bool {
		switch ke.Key {
		case tui.KeyUp:
			if selected > 0 {
				selected--
				renderList(listContainer, choices, selected)
			}
			return true
		case tui.KeyDown:
			if selected < len(choices)-1 {
				selected++
				renderList(listContainer, choices, selected)
			}
			return true
		case tui.KeyEnter:
			action = menuActionFromIndex(selected)
			ke.App().Stop()
			return true
		case tui.KeyEscape:
			action = MenuBack
			ke.App().Stop()
			return true
		}

		r, ok := lowerRune(ke)
		if !ok {
			return false
		}

		switch r {
		case 'k':
			if selected > 0 {
				selected--
				renderList(listContainer, choices, selected)
			}
			return true
		case 'j':
			if selected < len(choices)-1 {
				selected++
				renderList(listContainer, choices, selected)
			}
			return true
		case 'g':
			action = MenuAddGroup
			ke.App().Stop()
			return true
		default:
			return false
		}
	}); err != nil {
		return MenuBack, err
	}

	return action, nil
}

func menuActionFromIndex(idx int) MenuAction {
	switch idx {
	case 0:
		return MenuConnect
	case 1:
		return MenuAdd
	case 2:
		return MenuAddGroup
	case 3:
		return MenuDelete
	case 4:
		return MenuList
	case 5:
		return MenuHelp
	case 6:
		return MenuQuit
	default:
		return MenuBack
	}
}
