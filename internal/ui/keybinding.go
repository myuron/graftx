package ui

import (
	"github.com/jroimartin/gocui"
	"github.com/myuron/graftx/internal/pane"
)

// setKeybindings は全キーバインドを登録する。
// 全キーバインドはグローバル（viewname=""）で登録する。
func (a *App) setKeybindings(g *gocui.Gui) error {
	// j: カーソル下移動
	if err := g.SetKeybinding("", 'j', gocui.ModNone, a.cursorDown); err != nil {
		return err
	}
	// k: カーソル上移動
	if err := g.SetKeybinding("", 'k', gocui.ModNone, a.cursorUp); err != nil {
		return err
	}
	// l: ディレクトリに入る
	if err := g.SetKeybinding("", 'l', gocui.ModNone, a.enterDir); err != nil {
		return err
	}
	// h: 親ディレクトリに戻る
	if err := g.SetKeybinding("", 'h', gocui.ModNone, a.parentDir); err != nil {
		return err
	}
	// Tab: フォーカス切替
	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, a.toggleFocus); err != nil {
		return err
	}
	// s: リポジトリ選択
	if err := g.SetKeybinding("", 's', gocui.ModNone, a.selectRepo); err != nil {
		return err
	}
	// q: 終了
	if err := g.SetKeybinding("", 'q', gocui.ModNone, a.quit); err != nil {
		return err
	}
	return nil
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

// selectRepo はリポジトリ選択のためにMainLoopを一時終了する。
func (a *App) selectRepo(g *gocui.Gui, v *gocui.View) error {
	a.ExitReason = ExitReasonSelectRepo
	return gocui.ErrQuit
}

// quit はアプリケーションを終了する。
func (a *App) quit(g *gocui.Gui, v *gocui.View) error {
	a.ExitReason = ExitReasonQuit
	return gocui.ErrQuit
}
