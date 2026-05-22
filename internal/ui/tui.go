package ui

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"

	tui "github.com/grindlemire/go-tui"
)

var ErrCancelled = errors.New("cancelled")

var debugLogger *log.Logger

// SetDebug toggles verbose UI interaction logs.
func SetDebug(enabled bool) {
	if enabled {
		debugLogger = log.New(os.Stderr, "[ui] ", log.LstdFlags|log.Lmicroseconds)
		debugf("debug logging enabled")
		return
	}
	debugLogger = nil
}

func debugf(format string, args ...interface{}) {
	if debugLogger == nil {
		return
	}
	debugLogger.Printf(format, args...)
}

type HomeAction int

const (
	HomeConnect HomeAction = iota
	HomeAdd
	HomeDelete
	HomeMenu
	HomeHelp
	HomeQuit
)

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

func SelectIndex(title, hint string, items []string) (int, error) {
	if len(items) == 0 {
		return -1, fmt.Errorf("no items available")
	}

	debugf("open select screen title=%q items=%d", title, len(items))

	section, listContainer := newSection("Options")
	selected := 0
	renderList(listContainer, items, selected)

	root := buildScreenRoot(title, "", section, hint+" | Enter: Select | Esc: Cancel")

	confirmed := -1
	if err := runUI(root, "select-index", func(ke tui.KeyEvent) bool {
		switch ke.Key {
		case tui.KeyUp:
			if selected > 0 {
				selected--
				renderList(listContainer, items, selected)
			}
			return true
		case tui.KeyDown:
			if selected < len(items)-1 {
				selected++
				renderList(listContainer, items, selected)
			}
			return true
		case tui.KeyEnter:
			debugf("select screen activated index=%d", selected)
			confirmed = selected
			ke.App().Stop()
			return true
		case tui.KeyEscape:
			debugf("select screen cancelled")
			confirmed = -1
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
				renderList(listContainer, items, selected)
			}
			return true
		case 'j':
			if selected < len(items)-1 {
				selected++
				renderList(listContainer, items, selected)
			}
			return true
		default:
			return false
		}
	}); err != nil {
		return -1, err
	}
	debugf("close select screen title=%q selected=%d", title, confirmed)
	if confirmed < 0 {
		return -1, ErrCancelled
	}
	return confirmed, nil
}

func PromptInput(title, prompt string, required bool) (string, error) {
	debugf("open input screen title=%q required=%t", title, required)

	panel := tui.New(
		tui.WithDisplay(tui.DisplayFlex),
		tui.WithDirection(tui.Column),
		tui.WithGap(1),
		tui.WithPadding(1),
		tui.WithBorder(tui.BorderRounded),
		tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Blue)),
	)
	panel.AddChild(tui.New(tui.WithText(prompt), tui.WithWrap(true), tui.WithTextStyle(tui.NewStyle().Foreground(tui.White))))

	inputBox := tui.New(
		tui.WithBorder(tui.BorderSingle),
		tui.WithPadding(0),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.Cyan)),
	)
	errorLabel := tui.New(tui.WithText(""), tui.WithTextStyle(tui.NewStyle().Foreground(tui.BrightRed)))
	panel.AddChild(inputBox)
	panel.AddChild(errorLabel)

	root := buildScreenRoot(title, "", panel, "Type text | Enter: Confirm | Backspace: Delete | Esc: Cancel")

	buf := []rune{}
	value := ""
	refreshInput := func() {
		if len(buf) == 0 {
			inputBox.SetText(" ")
			inputBox.SetTextStyle(tui.NewStyle().Foreground(tui.BrightBlack))
			return
		}
		inputBox.SetText(string(buf))
		inputBox.SetTextStyle(tui.NewStyle().Foreground(tui.Cyan))
	}
	refreshInput()

	if err := runUI(root, "prompt-input", func(ke tui.KeyEvent) bool {
		switch ke.Key {
		case tui.KeyEscape:
			debugf("input screen cancelled title=%q", title)
			value = ""
			ke.App().Stop()
			return true
		case tui.KeyBackspace:
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
				errorLabel.SetText("")
				refreshInput()
			}
			return true
		case tui.KeyEnter:
			text := strings.TrimSpace(string(buf))
			if required && text == "" {
				debugf("input screen blocked empty required title=%q", title)
				errorLabel.SetText("Value is required.")
				return true
			}
			debugf("input screen submit title=%q value_len=%d", title, len(text))
			value = text
			ke.App().Stop()
			return true
		case tui.KeyRune:
			if ke.Rune == '\n' || ke.Rune == '\r' {
				return true
			}
			if unicode.IsControl(ke.Rune) {
				return true
			}
			buf = append(buf, ke.Rune)
			errorLabel.SetText("")
			refreshInput()
			return true
		default:
			return false
		}
	}); err != nil {
		return "", err
	}
	debugf("close input screen title=%q has_value=%t", title, value != "")
	if value == "" && required {
		return "", ErrCancelled
	}
	return value, nil
}

func Confirm(title, question string) (bool, error) {
	debugf("open confirm screen title=%q", title)

	panel := tui.New(
		tui.WithDisplay(tui.DisplayFlex),
		tui.WithDirection(tui.Column),
		tui.WithGap(1),
		tui.WithPadding(1),
		tui.WithBorder(tui.BorderRounded),
		tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Blue)),
	)
	panel.AddChild(tui.New(tui.WithText(question), tui.WithWrap(true), tui.WithTextStyle(tui.NewStyle().Foreground(tui.White))))

	root := buildScreenRoot(title, "", panel, "Y/Enter: Yes | N/Esc: No")

	confirmed := false
	if err := runUI(root, "confirm", func(ke tui.KeyEvent) bool {
		switch ke.Key {
		case tui.KeyEnter:
			debugf("confirm screen accepted key=Enter title=%q", title)
			confirmed = true
			ke.App().Stop()
			return true
		case tui.KeyEscape:
			debugf("confirm screen rejected key=Esc title=%q", title)
			confirmed = false
			ke.App().Stop()
			return true
		}
		r, ok := lowerRune(ke)
		if !ok {
			return false
		}
		switch r {
		case 'y':
			debugf("confirm screen accepted key=y title=%q", title)
			confirmed = true
			ke.App().Stop()
			return true
		case 'n':
			debugf("confirm screen rejected key=n title=%q", title)
			confirmed = false
			ke.App().Stop()
			return true
		default:
			return false
		}
	}); err != nil {
		return false, err
	}
	debugf("close confirm screen title=%q confirmed=%t", title, confirmed)
	return confirmed, nil
}

func ShowMessage(title, message string) error {
	debugf("open message screen title=%q", title)

	panel := tui.New(
		tui.WithDisplay(tui.DisplayFlex),
		tui.WithDirection(tui.Column),
		tui.WithGap(1),
		tui.WithPadding(1),
		tui.WithBorder(tui.BorderRounded),
		tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Blue)),
		tui.WithScrollable(tui.ScrollVertical),
		tui.WithOverflow(tui.OverflowHidden),
	)
	panel.AddChild(tui.New(tui.WithText(message), tui.WithWrap(true), tui.WithTextStyle(tui.NewStyle().Foreground(tui.White))))

	root := buildScreenRoot(title, "", panel, "Enter/Esc: Close")

	return runUI(root, "message", func(ke tui.KeyEvent) bool {
		switch ke.Key {
		case tui.KeyEnter:
			debugf("message screen close key=Enter title=%q", title)
			ke.App().Stop()
			return true
		case tui.KeyEscape:
			debugf("message screen close key=Esc title=%q", title)
			ke.App().Stop()
			return true
		default:
			return false
		}
	})
}

func newSection(title string) (*tui.Element, *tui.Element) {
	section := tui.New(
		tui.WithDisplay(tui.DisplayFlex),
		tui.WithDirection(tui.Column),
		tui.WithGap(1),
		tui.WithPadding(1),
		tui.WithBorder(tui.BorderRounded),
		tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Blue)),
		tui.WithFlexGrow(1),
		tui.WithOverflow(tui.OverflowHidden),
	)

	titleLine := tui.New(
		tui.WithText(title),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.Cyan).Bold()),
	)
	section.AddChild(titleLine)

	content := tui.New(
		tui.WithDisplay(tui.DisplayFlex),
		tui.WithDirection(tui.Column),
		tui.WithFlexGrow(1),
		tui.WithScrollable(tui.ScrollVertical),
		tui.WithOverflow(tui.OverflowHidden),
	)
	section.AddChild(content)

	return section, content
}

func buildScreenRoot(title, subtitle string, body *tui.Element, footer string) *tui.Element {
	root := tui.New(
		tui.WithDisplay(tui.DisplayFlex),
		tui.WithDirection(tui.Column),
		tui.WithGap(1),
		tui.WithPadding(1),
		tui.WithBorder(tui.BorderRounded),
		tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Blue)),
		tui.WithBackground(tui.NewStyle().Background(tui.Black)),
	)

	root.AddChild(tui.New(
		tui.WithText(title),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.Cyan).Bold()),
	))

	if subtitle != "" {
		root.AddChild(tui.New(
			tui.WithText(subtitle),
			tui.WithTextStyle(tui.NewStyle().Foreground(tui.Blue)),
		))
	}

	root.AddChild(body)
	root.AddChild(tui.New(
		tui.WithText(footer),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.Cyan).Dim()),
	))

	return root
}

func renderList(container *tui.Element, items []string, selected int) {
	container.RemoveAllChildren()
	for idx, item := range items {
		lineStyle := tui.NewStyle().Foreground(tui.White)
		prefix := "  "
		if idx == selected {
			prefix = "> "
			lineStyle = tui.NewStyle().Foreground(tui.White).Background(tui.Blue).Bold()
		}
		container.AddChild(tui.New(
			tui.WithText(prefix+item),
			tui.WithTextStyle(lineStyle),
			tui.WithWrap(false),
		))
	}
}

func lowerRune(ke tui.KeyEvent) (rune, bool) {
	if ke.Key != tui.KeyRune {
		return 0, false
	}
	return unicode.ToLower(ke.Rune), true
}

type interactiveScreen struct {
	root       *tui.Element
	view       string
	keyHandler func(tui.KeyEvent) bool
}

func (s *interactiveScreen) Render(_ *tui.App) *tui.Element {
	return s.root
}

func (s *interactiveScreen) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnStop(tui.Rune('c').Ctrl(), func(ke tui.KeyEvent) {
			debugf("force stop view=%s key=ctrl+c", s.view)
			ke.App().Stop()
		}),
		tui.On(tui.AnyKey, func(ke tui.KeyEvent) {
			if s.keyHandler != nil {
				_ = s.keyHandler(ke)
			}
		}),
	}
}

func runUI(root *tui.Element, view string, keyHandler func(tui.KeyEvent) bool) error {
	app, err := tui.NewApp(
		tui.WithRootComponent(&interactiveScreen{root: root, view: view, keyHandler: keyHandler}),
		tui.WithLegacyKeyboard(),
	)
	if err != nil {
		return err
	}
	defer app.Close()

	debugf("run view=%s", view)
	return app.Run()
}
