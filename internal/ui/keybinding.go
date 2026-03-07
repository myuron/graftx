package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/myuron/graftx/internal/pane"
)

// paneViews はペインビュー名の一覧。
var paneViews = []string{viewLeft, viewRight}

// bindPaneKey はペインビュー共通のキーバインドを登録するヘルパー。
func (a *App) bindPaneKey(g *gocui.Gui, key interface{}, mod gocui.Modifier, handler func(*gocui.Gui, *gocui.View) error) error {
	for _, view := range paneViews {
		if err := g.SetKeybinding(view, key, mod, handler); err != nil {
			return err
		}
	}
	return nil
}

// setKeybindings は全キーバインドを登録する。
// ペイン操作のキーバインドは左右ペインビューにのみ登録し、入力ビューと競合しないようにする。
func (a *App) setKeybindings(g *gocui.Gui) error {
	// ペインビュー共通のキーバインド
	if err := a.bindPaneKey(g, 'j', gocui.ModNone, a.cursorDown); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, 'k', gocui.ModNone, a.cursorUp); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, 'l', gocui.ModNone, a.enterDir); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, 'h', gocui.ModNone, a.parentDir); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, gocui.KeyTab, gocui.ModNone, a.toggleFocus); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, 's', gocui.ModNone, a.selectRepo); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, 'q', gocui.ModNone, a.quit); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, 'g', gocui.ModNone, a.handleG); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, 'G', gocui.ModNone, a.moveToBottom); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, gocui.KeySpace, gocui.ModNone, a.toggleSelect); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, gocui.KeyCtrlA, gocui.ModNone, a.selectAll); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, gocui.KeyCtrlR, gocui.ModNone, a.invertSelection); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, gocui.KeyEsc, gocui.ModNone, a.handleEsc); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, 'y', gocui.ModNone, a.yank); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, 'p', gocui.ModNone, a.paste); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, 'P', gocui.ModNone, a.pasteOverwrite); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, 'd', gocui.ModNone, a.trashFile); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, 'D', gocui.ModNone, a.deleteFile); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, 'a', gocui.ModNone, a.createFile); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, 'r', gocui.ModNone, a.renameFile); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, '/', gocui.ModNone, a.searchForward); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, 'n', gocui.ModNone, a.searchNext); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, 'N', gocui.ModNone, a.searchPrev); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, 'f', gocui.ModNone, a.filterMode); err != nil {
		return err
	}
	if err := a.bindPaneKey(g, '.', gocui.ModNone, a.toggleHidden); err != nil {
		return err
	}

	// 確認モード用のy/nキーバインド（ステータスバーに表示される確認ダイアログ用）
	// 通常モードではyはyank、nはsearchNextとして動作するが、
	// 確認モード中はpaneにフォーカスがあるのでpane keybindingsのハンドラが呼ばれる。
	// trashFile/deleteFileの確認はpaneビューでy/nとして既に処理される。

	return nil
}

// resetGPending はgPendingをリセットする。
func (a *App) resetGPending() {
	a.gPending = false
}

// focusedPane はフォーカス中のペインを返す。
// 左ペインが未設定の場合はnilを返す。
func (a *App) focusedPane() *pane.Pane {
	if a.FocusLeft {
		return a.SourcePane
	}
	return a.DestPane
}

// oppositePane はフォーカスしていない方のペインを返す。
func (a *App) oppositePane() *pane.Pane {
	if a.FocusLeft {
		return a.DestPane
	}
	return a.SourcePane
}

// cursorDown はカーソルを1つ下に移動する。
func (a *App) cursorDown(g *gocui.Gui, v *gocui.View) error {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil {
		return nil
	}
	p.MoveDown()
	return nil
}

// cursorUp はカーソルを1つ上に移動する。
func (a *App) cursorUp(g *gocui.Gui, v *gocui.View) error {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil {
		return nil
	}
	p.MoveUp()
	return nil
}

// enterDir はカーソル行のディレクトリに入る。
func (a *App) enterDir(g *gocui.Gui, v *gocui.View) error {
	a.resetGPending()
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
	a.resetGPending()
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
	a.resetGPending()
	a.FocusLeft = !a.FocusLeft
	a.Status = ""
	return nil
}

// selectRepo はリポジトリ選択ポップアップを開く。
func (a *App) selectRepo(g *gocui.Gui, v *gocui.View) error {
	a.resetGPending()
	return a.openRepoSelector(g, v)
}

// quit はアプリケーションを終了する。
func (a *App) quit(g *gocui.Gui, v *gocui.View) error {
	a.resetGPending()
	a.ExitReason = ExitReasonQuit
	return gocui.ErrQuit
}

// handleG は'g'キーを処理する。ggで先頭ジャンプ。
func (a *App) handleG(g *gocui.Gui, v *gocui.View) error {
	if a.gPending {
		// gg: 先頭にジャンプ
		a.gPending = false
		p := a.focusedPane()
		if p == nil {
			return nil
		}
		p.MoveToTop()
		a.Status = ""
		return nil
	}
	a.gPending = true
	return nil
}

// moveToBottom はカーソルを末尾に移動する。
func (a *App) moveToBottom(g *gocui.Gui, v *gocui.View) error {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil {
		return nil
	}
	p.MoveToBottom()
	return nil
}

// toggleSelect は選択をトグルしてカーソルを下に移動する。
func (a *App) toggleSelect(g *gocui.Gui, v *gocui.View) error {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil {
		return nil
	}
	p.ToggleSelect()
	p.MoveDown()
	return nil
}

// selectAll は全エントリを選択する。
func (a *App) selectAll(g *gocui.Gui, v *gocui.View) error {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil {
		return nil
	}
	p.SelectAll()
	a.Status = fmt.Sprintf("%d件選択", len(p.Selected))
	return nil
}

// invertSelection は選択を反転する。
func (a *App) invertSelection(g *gocui.Gui, v *gocui.View) error {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil {
		return nil
	}
	p.InvertSelection()
	a.Status = fmt.Sprintf("%d件選択", len(p.Selected))
	return nil
}

// handleEsc はEscキーを処理する。選択解除。
func (a *App) handleEsc(g *gocui.Gui, v *gocui.View) error {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil {
		return nil
	}
	// フィルタがかかっている場合はフィルタ解除
	if p.FilterQuery != "" {
		if err := p.ClearFilter(); err != nil {
			a.Status = fmt.Sprintf("refresh error: %v", err)
			return nil
		}
		a.Status = "フィルタ解除"
		return nil
	}
	// 選択がある場合は選択解除
	if len(p.Selected) > 0 {
		p.ClearSelection()
		a.Status = "選択解除"
		return nil
	}
	a.Status = ""
	return nil
}

// yank はフォーカス中のペインからヤンクする。
func (a *App) yank(g *gocui.Gui, v *gocui.View) error {
	a.resetGPending()
	// 確認モード中の'y'入力を処理
	if a.inputMode == InputModeConfirmDelete || a.inputMode == InputModeConfirmForceDelete {
		return a.executeConfirm(g)
	}
	p := a.focusedPane()
	if p == nil {
		return nil
	}
	// 同じペインで再度yを押した場合はヤンク取り消し
	if a.YankBuf != nil && len(a.YankBuf.Entries) > 0 && a.YankBuf.SrcDir == p.Dir {
		a.YankBuf = nil
		p.ClearSelection()
		a.Status = "yank cancelled"
		return nil
	}
	// 選択なしの場合はカーソル行を選択してからヤンク（*マーク表示のため）
	if len(p.Selected) == 0 && len(p.Entries) > 0 {
		p.Selected[p.Cursor] = true
	}
	a.YankBuf = p.Yank()
	a.Status = fmt.Sprintf("%d item(s) yanked", len(a.YankBuf.Entries))
	return nil
}

// paste はヤンクバッファの内容をフォーカス中のペインにコピーする。
// 同名ファイルが存在する場合はスキップしてステータスに表示する。
func (a *App) paste(g *gocui.Gui, v *gocui.View) error {
	a.resetGPending()
	if a.YankBuf == nil || len(a.YankBuf.Entries) == 0 {
		a.Status = "yank buffer is empty"
		return nil
	}

	dst := a.focusedPane()
	if dst == nil {
		a.Status = "no target pane for paste"
		return nil
	}

	copied := 0
	skipped := 0
	for _, src := range a.YankBuf.Entries {
		name := filepath.Base(src)
		dstPath := filepath.Join(dst.Dir, name)

		// ファイルシステムレベルで同名ファイルの存在チェック
		if _, err := os.Stat(dstPath); err == nil {
			skipped++
			continue
		}

		if err := a.FS.Copy(src, dstPath); err != nil {
			a.Status = fmt.Sprintf("copy error: %v", err)
			return nil
		}
		copied++
	}

	if err := dst.Refresh(); err != nil {
		a.Status = fmt.Sprintf("refresh error: %v", err)
		return nil
	}

	// コピー元ペインの選択状態とヤンクバッファをクリア
	a.clearYankSourceSelection()
	a.YankBuf = nil

	if skipped > 0 {
		a.Status = fmt.Sprintf("%d copied, %d skipped (duplicate)", copied, skipped)
	} else {
		a.Status = fmt.Sprintf("%d item(s) copied", copied)
	}
	return nil
}

// pasteOverwrite はヤンクバッファの内容をフォーカス中のペインに上書きコピーする。
func (a *App) pasteOverwrite(g *gocui.Gui, v *gocui.View) error {
	a.resetGPending()
	if a.YankBuf == nil || len(a.YankBuf.Entries) == 0 {
		a.Status = "yank buffer is empty"
		return nil
	}

	dst := a.focusedPane()
	if dst == nil {
		a.Status = "no target pane for paste"
		return nil
	}

	for _, src := range a.YankBuf.Entries {
		name := filepath.Base(src)
		dstPath := filepath.Join(dst.Dir, name)
		if err := a.FS.Copy(src, dstPath); err != nil {
			a.Status = fmt.Sprintf("copy error: %v", err)
			return nil
		}
	}

	if err := dst.Refresh(); err != nil {
		a.Status = fmt.Sprintf("refresh error: %v", err)
		return nil
	}

	// コピー元ペインの選択状態とヤンクバッファをクリア
	a.clearYankSourceSelection()
	count := len(a.YankBuf.Entries)
	a.YankBuf = nil

	a.Status = fmt.Sprintf("%d item(s) overwritten", count)
	return nil
}

// trashFile はカーソル行または選択中のエントリをゴミ箱に移動する確認モードに入る。
func (a *App) trashFile(g *gocui.Gui, v *gocui.View) error {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil || len(p.Entries) == 0 {
		return nil
	}

	a.pendingTargets = a.getTargetPaths(p)
	a.inputMode = InputModeConfirmDelete
	a.Status = fmt.Sprintf("ゴミ箱に移動しますか？ (%d件) [y/n]", len(a.pendingTargets))
	return nil
}

// deleteFile はカーソル行または選択中のエントリを完全削除する確認モードに入る。
func (a *App) deleteFile(g *gocui.Gui, v *gocui.View) error {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil || len(p.Entries) == 0 {
		return nil
	}

	a.pendingTargets = a.getTargetPaths(p)
	a.inputMode = InputModeConfirmForceDelete
	a.Status = fmt.Sprintf("完全削除しますか？ (%d件) [y/n]", len(a.pendingTargets))
	return nil
}

// executeConfirm は確認モードの実行を行う（y押下時）。
func (a *App) executeConfirm(g *gocui.Gui) error {
	p := a.focusedPane()
	if p == nil {
		a.inputMode = InputModeNone
		a.pendingTargets = nil
		a.Status = ""
		return nil
	}

	targets := a.pendingTargets
	mode := a.inputMode
	a.inputMode = InputModeNone
	a.pendingTargets = nil

	var lastErr error
	count := 0
	for _, path := range targets {
		var err error
		if mode == InputModeConfirmDelete {
			err = a.FS.Trash(path)
		} else {
			err = a.FS.Remove(path)
		}
		if err != nil {
			lastErr = err
		} else {
			count++
		}
	}

	if err := p.Refresh(); err != nil {
		a.Status = fmt.Sprintf("refresh error: %v", err)
		return nil
	}

	if lastErr != nil {
		a.Status = fmt.Sprintf("%d件処理、エラー: %v", count, lastErr)
	} else if mode == InputModeConfirmDelete {
		a.Status = fmt.Sprintf("%d件をゴミ箱に移動", count)
	} else {
		a.Status = fmt.Sprintf("%d件を完全削除", count)
	}
	return nil
}

// createFile は新規ファイル/ディレクトリ作成モードに入る。
func (a *App) createFile(g *gocui.Gui, v *gocui.View) error {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil {
		return nil
	}
	a.inputMode = InputModeCreate
	a.inputBuf = ""
	a.Status = "新規作成（末尾/でディレクトリ）: "
	// 入力ビューのキーバインドを登録
	return a.setupInputKeybindings(g)
}

// renameFile はリネームモードに入る。
func (a *App) renameFile(g *gocui.Gui, v *gocui.View) error {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil || len(p.Entries) == 0 {
		return nil
	}
	a.inputMode = InputModeRename
	a.inputBuf = p.Entries[p.Cursor].Name
	a.Status = fmt.Sprintf("リネーム: %s", a.inputBuf)
	return a.setupInputKeybindings(g)
}

// searchForward は前方検索モードに入る。
func (a *App) searchForward(g *gocui.Gui, v *gocui.View) error {
	a.resetGPending()
	a.inputMode = InputModeSearch
	a.inputBuf = ""
	a.searchFwd = true
	a.Status = "/:"
	return a.setupInputKeybindings(g)
}

// searchNext は次の検索結果に移動する。
func (a *App) searchNext(g *gocui.Gui, v *gocui.View) error {
	a.resetGPending()
	// 確認モード中の'n'入力を処理（キャンセル）
	if a.inputMode == InputModeConfirmDelete || a.inputMode == InputModeConfirmForceDelete {
		a.inputMode = InputModeNone
		a.pendingTargets = nil
		a.Status = "キャンセル"
		return nil
	}
	if a.searchQuery == "" {
		return nil
	}
	p := a.focusedPane()
	if p == nil {
		return nil
	}

	a.doSearch(p, a.searchFwd)
	return nil
}

// searchPrev は前の検索結果に移動する。
func (a *App) searchPrev(g *gocui.Gui, v *gocui.View) error {
	a.resetGPending()
	if a.searchQuery == "" {
		return nil
	}
	p := a.focusedPane()
	if p == nil {
		return nil
	}

	a.doSearch(p, !a.searchFwd)
	return nil
}

// filterMode はフィルタモードに入る。
func (a *App) filterMode(g *gocui.Gui, v *gocui.View) error {
	a.resetGPending()
	a.inputMode = InputModeFilter
	a.inputBuf = ""
	a.Status = "フィルタ: "
	return a.setupInputKeybindings(g)
}

// toggleHidden は隠しファイルの表示/非表示を切り替える。
func (a *App) toggleHidden(g *gocui.Gui, v *gocui.View) error {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil {
		return nil
	}
	p.ShowHidden = !p.ShowHidden
	if err := p.Refresh(); err != nil {
		a.Status = fmt.Sprintf("refresh error: %v", err)
		return nil
	}
	if p.ShowHidden {
		a.Status = "隠しファイルを表示"
	} else {
		a.Status = "隠しファイルを非表示"
	}
	return nil
}

// setupInputKeybindings は入力ビュー用のキーバインドを登録する。
func (a *App) setupInputKeybindings(g *gocui.Gui) error {
	if a.inputKeybindingsInited {
		return nil
	}
	a.inputKeybindingsInited = true
	// Enter: 入力確定
	if err := g.SetKeybinding(viewInput, gocui.KeyEnter, gocui.ModNone, a.inputSubmit); err != nil {
		return err
	}
	// Esc: 入力キャンセル
	if err := g.SetKeybinding(viewInput, gocui.KeyEsc, gocui.ModNone, a.inputCancel); err != nil {
		return err
	}
	// Backspace: 1文字削除
	if err := g.SetKeybinding(viewInput, gocui.KeyBackspace2, gocui.ModNone, a.inputBackspace); err != nil {
		return err
	}
	if err := g.SetKeybinding(viewInput, gocui.KeyBackspace, gocui.ModNone, a.inputBackspace); err != nil {
		return err
	}
	return nil
}

// inputSubmit は入力モードの確定処理を行う。
func (a *App) inputSubmit(g *gocui.Gui, v *gocui.View) error {
	// 入力ビューの内容を取得
	input := strings.TrimSpace(v.Buffer())
	if input == "" && a.inputBuf != "" {
		input = a.inputBuf
	}

	mode := a.inputMode
	a.inputMode = InputModeNone
	a.inputBuf = ""
	g.DeleteKeybindings(viewInput)
	a.inputKeybindingsInited = false

	switch mode {
	case InputModeSearch, InputModeSearchBackward:
		a.searchQuery = input
		if input == "" {
			a.Status = ""
			return nil
		}
		p := a.focusedPane()
		if p == nil {
			return nil
		}
		a.doSearch(p, mode == InputModeSearch)
	case InputModeFilter:
		p := a.focusedPane()
		if p == nil {
			return nil
		}
		if input == "" {
			if err := p.ClearFilter(); err != nil {
				a.Status = fmt.Sprintf("refresh error: %v", err)
				return nil
			}
			a.Status = "フィルタ解除"
		} else {
			p.FilterQuery = input
			if err := p.Refresh(); err != nil {
				a.Status = fmt.Sprintf("refresh error: %v", err)
				return nil
			}
			a.Status = fmt.Sprintf("フィルタ: %s (%d件)", input, len(p.Entries))
		}
	case InputModeCreate:
		return a.executeCreate(input)
	case InputModeRename:
		return a.executeRename(input)
	}

	return nil
}

// inputCancel は入力モードをキャンセルする。
func (a *App) inputCancel(g *gocui.Gui, v *gocui.View) error {
	a.inputMode = InputModeNone
	a.inputBuf = ""
	a.Status = ""
	g.DeleteKeybindings(viewInput)
	a.inputKeybindingsInited = false
	return nil
}

// inputBackspace は入力バッファから1文字削除する。
func (a *App) inputBackspace(g *gocui.Gui, v *gocui.View) error {
	// gocuiのEditable viewが自動的にバックスペースを処理するので、
	// ステータス表示の更新のみ行う
	return nil
}

// executeCreate は新規ファイル/ディレクトリの作成を実行する。
func (a *App) executeCreate(name string) error {
	if name == "" {
		a.Status = "名前が空です"
		return nil
	}

	p := a.focusedPane()
	if p == nil {
		return nil
	}

	isDir := strings.HasSuffix(name, "/")
	if isDir {
		name = strings.TrimSuffix(name, "/")
	}

	// パストラバーサル防止: ベース名のみ許可
	base := filepath.Base(name)
	if base != name || base == "." || base == ".." {
		a.Status = "無効なファイル名です"
		return nil
	}

	path := filepath.Join(p.Dir, name)
	if err := a.FS.Create(path, isDir); err != nil {
		a.Status = fmt.Sprintf("作成エラー: %v", err)
		return nil
	}

	if err := p.Refresh(); err != nil {
		a.Status = fmt.Sprintf("refresh error: %v", err)
		return nil
	}

	// 作成したエントリにカーソルを移動
	for i, e := range p.Entries {
		if e.Name == name {
			p.Cursor = i
			break
		}
	}

	if isDir {
		a.Status = fmt.Sprintf("ディレクトリ作成: %s", name)
	} else {
		a.Status = fmt.Sprintf("ファイル作成: %s", name)
	}
	return nil
}

// executeRename はリネームを実行する。
func (a *App) executeRename(newName string) error {
	if newName == "" {
		a.Status = "名前が空です"
		return nil
	}

	p := a.focusedPane()
	if p == nil || len(p.Entries) == 0 {
		return nil
	}

	oldName := p.Entries[p.Cursor].Name
	if oldName == newName {
		a.Status = "名前が変更されていません"
		return nil
	}

	// パストラバーサル防止: ベース名のみ許可
	base := filepath.Base(newName)
	if base != newName || base == "." || base == ".." {
		a.Status = "無効なファイル名です"
		return nil
	}

	oldPath := filepath.Join(p.Dir, oldName)
	newPath := filepath.Join(p.Dir, newName)
	if err := a.FS.Rename(oldPath, newPath); err != nil {
		a.Status = fmt.Sprintf("リネームエラー: %v", err)
		return nil
	}

	if err := p.Refresh(); err != nil {
		a.Status = fmt.Sprintf("refresh error: %v", err)
		return nil
	}

	// リネームしたエントリにカーソルを移動
	for i, e := range p.Entries {
		if e.Name == newName {
			p.Cursor = i
			break
		}
	}

	a.Status = fmt.Sprintf("リネーム: %s → %s", oldName, newName)
	return nil
}

// getTargetPaths は選択中またはカーソル行のエントリのフルパス一覧を返す。
func (a *App) getTargetPaths(p *pane.Pane) []string {
	if len(p.Entries) == 0 {
		return nil
	}
	if len(p.Selected) > 0 {
		var paths []string
		for i := range p.Entries {
			if p.Selected[i] {
				paths = append(paths, filepath.Join(p.Dir, p.Entries[i].Name))
			}
		}
		return paths
	}
	if p.Cursor < 0 || p.Cursor >= len(p.Entries) {
		return nil
	}
	return []string{filepath.Join(p.Dir, p.Entries[p.Cursor].Name)}
}

// clearYankSourceSelection はヤンク元ペインの選択状態をクリアする。
func (a *App) clearYankSourceSelection() {
	if a.YankBuf == nil {
		return
	}
	if a.SourcePane != nil && a.SourcePane.Dir == a.YankBuf.SrcDir {
		a.SourcePane.ClearSelection()
	}
	if a.DestPane != nil && a.DestPane.Dir == a.YankBuf.SrcDir {
		a.DestPane.ClearSelection()
	}
}

// doSearch はペイン内で検索を実行する。
func (a *App) doSearch(p *pane.Pane, forward bool) {
	query := strings.ToLower(a.searchQuery)
	n := len(p.Entries)
	if n == 0 {
		return
	}

	for i := 1; i <= n; i++ {
		var idx int
		if forward {
			idx = (p.Cursor + i) % n
		} else {
			idx = (p.Cursor - i + n) % n
		}
		if strings.Contains(strings.ToLower(p.Entries[idx].Name), query) {
			p.Cursor = idx
			a.Status = fmt.Sprintf("/%s", a.searchQuery)
			return
		}
	}
	a.Status = fmt.Sprintf("見つかりません: %s", a.searchQuery)
}
