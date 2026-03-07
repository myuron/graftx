package ui

import (
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
)

// filterRepoList はrepoFilterQueryでリポジトリ一覧をフィルタする。
// 大文字小文字を無視したサブストリングマッチ。
func (a *App) filterRepoList() {
	a.repoSelectorCursor = 0
	if a.repoFilterQuery == "" {
		a.filteredRepoList = make([]string, len(a.repoList))
		copy(a.filteredRepoList, a.repoList)
		return
	}

	query := strings.ToLower(a.repoFilterQuery)
	a.filteredRepoList = nil
	for _, repo := range a.repoList {
		if strings.Contains(strings.ToLower(repo), query) {
			a.filteredRepoList = append(a.filteredRepoList, repo)
		}
	}
}

// selectorMoveCursorDown はリポジトリ選択カーソルを1つ下に移動する。
func (a *App) selectorMoveCursorDown() {
	if len(a.filteredRepoList) == 0 {
		return
	}
	if a.repoSelectorCursor < len(a.filteredRepoList)-1 {
		a.repoSelectorCursor++
	}
}

// selectorMoveCursorUp はリポジトリ選択カーソルを1つ上に移動する。
func (a *App) selectorMoveCursorUp() {
	if a.repoSelectorCursor > 0 {
		a.repoSelectorCursor--
	}
}

// openRepoSelector はリポジトリ選択ポップアップを開く。
func (a *App) openRepoSelector(g *gocui.Gui, v *gocui.View) error {
	repos, err := a.Selector.ListRepositories()
	if err != nil {
		a.Status = fmt.Sprintf("failed to list repositories: %v", err)
		return nil
	}

	a.repoList = repos
	a.repoFilterQuery = ""
	a.repoSelectorCursor = 0
	a.filterRepoList()
	a.inputMode = InputModeSelectRepo

	return a.setupSelectorKeybindings(g)
}

// closeRepoSelector はリポジトリ選択ポップアップを閉じる。
func (a *App) closeRepoSelector(g *gocui.Gui) {
	a.inputMode = InputModeNone
	a.repoList = nil
	a.filteredRepoList = nil
	a.repoSelectorCursor = 0
	a.repoFilterQuery = ""

	_ = g.DeleteView(viewSelectorFilter)
	_ = g.DeleteView(viewSelectorList)
	g.DeleteKeybindings(viewSelectorFilter)
	g.DeleteKeybindings(viewSelectorList)
	a.selectorKeybindingsInited = false
}

// selectorConfirm は選択確定処理。
func (a *App) selectorConfirm(g *gocui.Gui, v *gocui.View) error {
	if len(a.filteredRepoList) == 0 {
		return nil
	}

	selected := a.filteredRepoList[a.repoSelectorCursor]
	a.closeRepoSelector(g)

	if err := a.SetSourceDir(selected); err != nil {
		a.Status = fmt.Sprintf("failed to set source directory: %v", err)
		return nil
	}

	a.FocusLeft = true
	a.Status = ""
	return nil
}

// selectorCancel はキャンセル処理。
func (a *App) selectorCancel(g *gocui.Gui, v *gocui.View) error {
	a.closeRepoSelector(g)
	a.Status = ""
	return nil
}

// selectorCursorDown はリスト内カーソル下移動ハンドラ。
func (a *App) selectorCursorDown(g *gocui.Gui, v *gocui.View) error {
	a.selectorMoveCursorDown()
	return nil
}

// selectorCursorUp はリスト内カーソル上移動ハンドラ。
func (a *App) selectorCursorUp(g *gocui.Gui, v *gocui.View) error {
	a.selectorMoveCursorUp()
	return nil
}

// selectorFilterEditor はセレクタフィルタ入力のカスタムエディタ。
func (a *App) selectorFilterEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		if len(a.repoFilterQuery) > 0 {
			// UTF-8対応: runeスライスで処理
			runes := []rune(a.repoFilterQuery)
			a.repoFilterQuery = string(runes[:len(runes)-1])
		}
	case ch != 0 && key == 0:
		a.repoFilterQuery += string(ch)
	case key == gocui.KeySpace:
		a.repoFilterQuery += " "
	default:
		return
	}
	a.filterRepoList()
}

// setupSelectorKeybindings はセレクタ用キーバインドを登録する。
func (a *App) setupSelectorKeybindings(g *gocui.Gui) error {
	if a.selectorKeybindingsInited {
		return nil
	}
	a.selectorKeybindingsInited = true

	// フィルタビュー用キーバインド
	if err := g.SetKeybinding(viewSelectorFilter, gocui.KeyEnter, gocui.ModNone, a.selectorConfirm); err != nil {
		return err
	}
	if err := g.SetKeybinding(viewSelectorFilter, gocui.KeyEsc, gocui.ModNone, a.selectorCancel); err != nil {
		return err
	}
	if err := g.SetKeybinding(viewSelectorFilter, gocui.KeyCtrlC, gocui.ModNone, a.selectorCancel); err != nil {
		return err
	}
	if err := g.SetKeybinding(viewSelectorFilter, gocui.KeyCtrlJ, gocui.ModNone, a.selectorCursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding(viewSelectorFilter, gocui.KeyCtrlK, gocui.ModNone, a.selectorCursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding(viewSelectorFilter, gocui.KeyCtrlN, gocui.ModNone, a.selectorCursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding(viewSelectorFilter, gocui.KeyCtrlP, gocui.ModNone, a.selectorCursorUp); err != nil {
		return err
	}

	return nil
}

// renderSelectorPopup はリポジトリ選択ポップアップを描画する。
func (a *App) renderSelectorPopup(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// ポップアップサイズ計算（60%幅 x 70%高さ、最小40x10）
	w := maxX * 60 / 100
	if w < 40 {
		w = 40
	}
	if w > maxX-2 {
		w = maxX - 2
	}
	h := maxY * 70 / 100
	if h < 10 {
		h = 10
	}
	if h > maxY-2 {
		h = maxY - 2
	}

	// 中央配置
	x0 := (maxX - w) / 2
	y0 := (maxY - h) / 2
	x1 := x0 + w
	y1 := y0 + h

	// フィルタ入力ビュー（1行）
	filterY1 := y0 + 2
	if v, err := g.SetView(viewSelectorFilter, x0, y0, x1, filterY1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Filter "
		v.Editable = true
		v.Editor = gocui.EditorFunc(a.selectorFilterEditor)
	}

	// リストビュー
	if v, err := g.SetView(viewSelectorList, x0, filterY1, x1, y1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Repositories "
		v.Highlight = true
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack
	}

	// フィルタビューの内容更新
	filterView, err := g.View(viewSelectorFilter)
	if err != nil {
		return err
	}
	filterView.Clear()
	_, _ = fmt.Fprint(filterView, a.repoFilterQuery)

	// リストビューの内容更新
	listView, err := g.View(viewSelectorList)
	if err != nil {
		return err
	}
	listView.Clear()
	for _, repo := range a.filteredRepoList {
		_, _ = fmt.Fprintln(listView, repo)
	}

	// リストビューのカーソル同期
	a.syncSelectorScroll(listView)

	// ビューを前面に表示
	if _, err := g.SetViewOnTop(viewSelectorFilter); err != nil {
		return err
	}
	if _, err := g.SetViewOnTop(viewSelectorList); err != nil {
		return err
	}

	// フィルタビューにフォーカス
	if _, err := g.SetCurrentView(viewSelectorFilter); err != nil {
		return err
	}

	return nil
}

// syncSelectorScroll はセレクタリストビューのスクロール位置を同期する。
func (a *App) syncSelectorScroll(v *gocui.View) {
	_, viewHeight := v.Size()
	if viewHeight <= 0 {
		return
	}

	oy := 0
	cy := a.repoSelectorCursor
	if a.repoSelectorCursor >= viewHeight {
		oy = a.repoSelectorCursor - viewHeight + 1
		cy = viewHeight - 1
	}

	v.SetOrigin(0, oy)
	v.SetCursor(0, cy)
}
