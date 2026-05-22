package ui

import (
	tui "github.com/grindlemire/go-tui"
)

// Theme encapsulates the modern UI color scheme and styling.
type Theme struct {
	// Primary colors
	Primary       tui.Color
	PrimaryBright tui.Color
	Accent        tui.Color
	AccentBright  tui.Color

	// Status colors
	Success    tui.Color
	Warning    tui.Color
	Error      tui.Color
	Info       tui.Color

	// Text colors
	TextPrimary   tui.Color
	TextSecondary tui.Color
	TextMuted     tui.Color

	// Background
	Background tui.Color
}

// ModernTheme is a sleek, modern color scheme.
var ModernTheme = Theme{
	// Modern gradient colors
	Primary:       tui.BrightMagenta,
	PrimaryBright: tui.BrightCyan,
	Accent:        tui.BrightCyan,
	AccentBright:  tui.BrightGreen,

	// Status colors with modern feel
	Success:    tui.BrightGreen,
	Warning:    tui.BrightYellow,
	Error:      tui.BrightRed,
	Info:       tui.BrightBlue,

	// Text hierarchy
	TextPrimary:   tui.White,
	TextSecondary: tui.BrightWhite,
	TextMuted:     tui.BrightBlack,

	// Dark background
	Background: tui.Black,
}

// Styles are commonly used style combinations using the modern theme.
var Styles = struct {
	// Border styles
	PrimaryBorder      tui.Style
	AccentBorder       tui.Style
	SuccessBorder      tui.Style
	WarningBorder      tui.Style
	ErrorBorder        tui.Style

	// Text styles
	Title              tui.Style
	Subtitle           tui.Style
	Label              tui.Style
	HeaderLabel        tui.Style
	MutedText          tui.Style
	SuccessText        tui.Style
	WarningText        tui.Style
	ErrorText          tui.Style

	// Selection/Highlight styles
	SelectedItem       tui.Style
	UnselectedItem     tui.Style
}{
	// Border styles
	PrimaryBorder:     tui.NewStyle().Foreground(ModernTheme.Primary),
	AccentBorder:      tui.NewStyle().Foreground(ModernTheme.Accent),
	SuccessBorder:     tui.NewStyle().Foreground(ModernTheme.Success),
	WarningBorder:     tui.NewStyle().Foreground(ModernTheme.Warning),
	ErrorBorder:       tui.NewStyle().Foreground(ModernTheme.Error),

	// Text styles
	Title:              tui.NewStyle().Foreground(ModernTheme.Primary).Bold(),
	Subtitle:          tui.NewStyle().Foreground(ModernTheme.Accent).Dim(),
	Label:             tui.NewStyle().Foreground(ModernTheme.TextPrimary),
	HeaderLabel:       tui.NewStyle().Foreground(ModernTheme.AccentBright).Bold(),
	MutedText:         tui.NewStyle().Foreground(ModernTheme.TextMuted),
	SuccessText:       tui.NewStyle().Foreground(ModernTheme.Success),
	WarningText:       tui.NewStyle().Foreground(ModernTheme.Warning),
	ErrorText:         tui.NewStyle().Foreground(ModernTheme.Error),

	// Selection/Highlight styles
	SelectedItem:      tui.NewStyle().Foreground(ModernTheme.AccentBright).Background(tui.BrightBlack).Bold(),
	UnselectedItem:    tui.NewStyle().Foreground(ModernTheme.TextPrimary),
}

// GetBorderStyle returns the appropriate border style based on context.
func GetBorderStyle(severity string) tui.Style {
	switch severity {
	case "success":
		return Styles.SuccessBorder
	case "warning":
		return Styles.WarningBorder
	case "error":
		return Styles.ErrorBorder
	case "info":
		return tui.NewStyle().Foreground(ModernTheme.Info)
	default:
		return Styles.AccentBorder
	}
}

// GetTextStyle returns the appropriate text style based on context.
func GetTextStyle(severity string) tui.Style {
	switch severity {
	case "success":
		return Styles.SuccessText
	case "warning":
		return Styles.WarningText
	case "error":
		return Styles.ErrorText
	case "info":
		return tui.NewStyle().Foreground(ModernTheme.Info)
	default:
		return Styles.Label
	}
}



