package ui

import (
	"github.com/jroimartin/gocui"
	"github.com/myuron/graftx/internal/pane"
)

// setKeybindings は全キーバインドを登録する。
// 全キーバインドはグローバル（viewname=""）で登録する。
func (a *App) setKeybindings(g *gocui.Gui) {
	// j: カーソル下移動
	g.SetKeybinding("", 'j', gocui.ModNone, a.cursorDown)
	// k: カーソル上移動
	g.SetKeybinding("", 'k', gocui.ModNone, a.cursorUp)
	// l: ディレクトリに入る
	g.SetKeybinding("", 'l', gocui.ModNone, a.enterDir)
	// h: 親ディレクトリに戻る
	g.SetKeybinding("", 'h', gocui.ModNone, a.parentDir)
	// Tab: フォーカス切替
	g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, a.toggleFocus)
	// q: 終了
	g.SetKeybinding("", 'q', gocui.ModNone, a.quit)
}

// focusedPane はフォーカス中のペインを返す。
// 左ペインが未設定の場合はnilを返す。
func (a *App) focusedPane() *pane.Pane {
	if a.FocusLeft {
		return a.SourcePane
	}
	return a.DestPane
}

// cursorDown はカーソルを1つ下に移動する。
func (a *App) cursorDown(g *gocui.Gui, v *gocui.View) error {
	p := a.focusedPane()
	if p == nil {
		return nil
	}
	p.MoveDown()
	return nil
}

// cursorUp はカーソルを1つ上に移動する。
func (a *App) cursorUp(g *gocui.Gui, v *gocui.View) error {
	p := a.focusedPane()
	if p == nil {
		return nil
	}
	p.MoveUp()
	return nil
}

// enterDir はカーソル行のディレクトリに入る。
func (a *App) enterDir(g *gocui.Gui, v *gocui.View) error {
	p := a.focusedPane()
	if p == nil {
		return nil
	}
	if err := p.EnterDir(); err != nil {
		a.Status = err.Error()
	}
	return nil
}

// parentDir は親ディレクトリに戻る。
func (a *App) parentDir(g *gocui.Gui, v *gocui.View) error {
	p := a.focusedPane()
	if p == nil {
		return nil
	}
	if err := p.ParentDir(); err != nil {
		a.Status = err.Error()
	}
	return nil
}

// toggleFocus はペイン間のフォーカスを切り替える。
func (a *App) toggleFocus(g *gocui.Gui, v *gocui.View) error {
	a.FocusLeft = !a.FocusLeft
	a.Status = ""
	return nil
}

// quit はアプリケーションを終了する。
func (a *App) quit(g *gocui.Gui, v *gocui.View) error {
	a.ExitReason = ExitReasonQuit
	return gocui.ErrQuit
}
