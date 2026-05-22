package ui

import (
	tui "github.com/grindlemire/go-tui"
)

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
