package ui

// MenuAction represents the choice made in the main menu.
type MenuAction int

const (
	MenuConnect MenuAction = iota
	MenuAdd
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
		"Delete selected server",
		"Back to server list",
		"Help",
		"Quit",
	}

	idx, err := SelectIndex("Main Menu", "Choose an action", choices)
	if err != nil {
		if err == ErrCancelled {
			return MenuBack, nil
		}
		return MenuBack, err
	}

	switch idx {
	case 0:
		return MenuConnect, nil
	case 1:
		return MenuAdd, nil
	case 2:
		return MenuDelete, nil
	case 3:
		return MenuList, nil
	case 4:
		return MenuHelp, nil
	case 5:
		return MenuQuit, nil
	default:
		return MenuBack, nil
	}
}
