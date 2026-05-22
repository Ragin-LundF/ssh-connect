package ui

import (
	"unicode"

	tui "github.com/grindlemire/go-tui"
)

// newSection creates a bordered, scrollable section with a title label.
// It returns the outer section element and the inner content container.
func newSection(title string) (*tui.Element, *tui.Element) {
	section := tui.New(
		tui.WithDisplay(tui.DisplayFlex),
		tui.WithDirection(tui.Column),
		tui.WithGap(1),
		tui.WithPadding(1),
		tui.WithBorder(tui.BorderRounded),
		tui.WithBorderStyle(tui.NewStyle().Foreground(tui.BrightCyan)),
		tui.WithFlexGrow(1),
		tui.WithOverflow(tui.OverflowHidden),
	)

	titleLine := tui.New(
		tui.WithText("▸ "+title+" ◂"),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.BrightCyan).Bold()),
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

// newStaticSection creates a bordered section with a non-scrollable content area.
// Useful for short lists where scroll state can hide content.
func newStaticSection(title string) (*tui.Element, *tui.Element) {
	section := tui.New(
		tui.WithDisplay(tui.DisplayFlex),
		tui.WithDirection(tui.Column),
		tui.WithGap(1),
		tui.WithPadding(1),
		tui.WithBorder(tui.BorderRounded),
		tui.WithBorderStyle(tui.NewStyle().Foreground(tui.BrightCyan)),
		tui.WithFlexGrow(1),
		tui.WithOverflow(tui.OverflowHidden),
	)

	titleLine := tui.New(
		tui.WithText("▸ "+title+" ◂"),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.BrightCyan).Bold()),
	)
	section.AddChild(titleLine)

	content := tui.New(
		tui.WithDisplay(tui.DisplayFlex),
		tui.WithDirection(tui.Column),
		tui.WithFlexGrow(1),
		tui.WithOverflow(tui.OverflowHidden),
	)
	section.AddChild(content)

	return section, content
}

// buildScreenRoot assembles the standard full-screen layout:
// title -> optional subtitle -> body -> footer hint bar.
func buildScreenRoot(title, subtitle string, body *tui.Element, footer string) *tui.Element {
	root := tui.New(
		tui.WithDisplay(tui.DisplayFlex),
		tui.WithDirection(tui.Column),
		tui.WithGap(1),
		tui.WithPadding(1),
		tui.WithBorder(tui.BorderRounded),
		tui.WithBorderStyle(tui.NewStyle().Foreground(tui.BrightMagenta)),
		tui.WithBackground(tui.NewStyle().Background(tui.Black)),
	)

	root.AddChild(tui.New(
		tui.WithText("╰─ "+title+" ─╯"),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.BrightMagenta).Bold()),
	))

	if subtitle != "" {
		root.AddChild(tui.New(
			tui.WithText("  └─ "+subtitle),
			tui.WithTextStyle(tui.NewStyle().Foreground(tui.BrightCyan).Dim()),
		))
	}

	root.AddChild(body)
	root.AddChild(tui.New(
		tui.WithText("───── "+footer+" ─────"),
		tui.WithTextStyle(tui.NewStyle().Foreground(tui.BrightGreen).Dim()),
	))

	return root
}

// renderList re-renders all items inside container, highlighting the selected row.
func renderList(container *tui.Element, items []string, selected int) {
	container.RemoveAllChildren()
	for idx, item := range items {
		lineStyle := tui.NewStyle().Foreground(tui.White)
		prefix := "  ◌ "
		if idx == selected {
			prefix = "  ◉ "
			lineStyle = tui.NewStyle().Foreground(tui.BrightGreen).Background(tui.BrightBlack).Bold()
		}
		container.AddChild(tui.New(
			tui.WithText(prefix+item),
			tui.WithTextStyle(lineStyle),
			tui.WithWrap(false),
		))
	}
}

// lowerRune extracts the rune from a KeyRune event, converted to lower-case.
// Returns (0, false) for any other key type.
func lowerRune(ke tui.KeyEvent) (rune, bool) {
	if ke.Key != tui.KeyRune {
		return 0, false
	}
	return unicode.ToLower(ke.Rune), true
}
