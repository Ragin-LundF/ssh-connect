package ui

import (
	"fmt"
	"strings"
	"unicode"

	tui "github.com/grindlemire/go-tui"
)

// SelectIndex shows a generic scrollable list and returns the chosen index.
// Returns ErrCancelled if the user presses Escape.
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

// PromptInput shows a text-input dialog and returns the trimmed value.
// When required is true, empty submissions are rejected.
// Returns ErrCancelled if the user presses Escape.
func PromptInput(title, prompt string, required bool) (string, error) {
	debugf("open input screen title=%q required=%t", title, required)

	panel := tui.New(
		tui.WithDisplay(tui.DisplayFlex),
		tui.WithDirection(tui.Column),
		tui.WithGap(1),
		tui.WithPadding(1),
		tui.WithBorder(tui.BorderRounded),
		tui.WithBorderStyle(tui.NewStyle().Foreground(tui.BrightCyan)),
	)
	panel.AddChild(tui.New(tui.WithText("▸ "+prompt), tui.WithWrap(true), tui.WithTextStyle(tui.NewStyle().Foreground(tui.White))))

	inputBox := tui.New(
		tui.WithBorder(tui.BorderSingle),
		tui.WithPadding(0),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.BrightGreen)),
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
		inputBox.SetTextStyle(tui.NewStyle().Foreground(tui.BrightGreen))
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

// Confirm shows a yes/no dialog. Returns (true, nil) when confirmed.
func Confirm(title, question string) (bool, error) {
	debugf("open confirm screen title=%q", title)

	panel := tui.New(
		tui.WithDisplay(tui.DisplayFlex),
		tui.WithDirection(tui.Column),
		tui.WithGap(1),
		tui.WithPadding(1),
		tui.WithBorder(tui.BorderRounded),
		tui.WithBorderStyle(tui.NewStyle().Foreground(tui.BrightYellow)),
	)
	panel.AddChild(tui.New(tui.WithText("⚠ "+question), tui.WithWrap(true), tui.WithTextStyle(tui.NewStyle().Foreground(tui.BrightYellow))))

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

// ShowMessage displays a scrollable informational message.
// The user closes it with Enter or Escape.
func ShowMessage(title, message string) error {
	debugf("open message screen title=%q", title)

	panel, content := newSection("Message")
	renderMessageLines(content, message)

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

func renderMessageLines(container *tui.Element, message string) {
	container.RemoveAllChildren()
	for idx, line := range splitMessageLines(message) {
		prefix := "  "
		if idx == 0 {
			prefix = "ℹ "
		}
		container.AddChild(tui.New(
			tui.WithText(prefix+line),
			tui.WithWrap(true),
			tui.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
		))
	}
}

func splitMessageLines(message string) []string {
	lines := strings.Split(message, "\n")
	if len(lines) == 0 {
		return []string{""}
	}
	for idx, line := range lines {
		if line == "" {
			lines[idx] = " "
		}
	}
	return lines
}
