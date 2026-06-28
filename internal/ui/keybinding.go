package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/myuron/graftx/internal/pane"
)

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

// enterTextInput はテキスト入力モードに入り、入力欄を初期化してフォーカスする。
func (a *App) enterTextInput(mode InputMode, initial string) tea.Cmd {
	a.inputMode = mode
	a.input.SetValue(initial)
	a.input.CursorEnd()
	a.Status = ""
	return a.input.Focus()
}

// cursorDown はカーソルを1つ下に移動する。
func (a *App) cursorDown() {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil {
		return
	}
	p.MoveDown()
}

// cursorUp はカーソルを1つ上に移動する。
func (a *App) cursorUp() {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil {
		return
	}
	p.MoveUp()
}

// enterDir はカーソル行のディレクトリに入る。
func (a *App) enterDir() {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil {
		return
	}
	if err := p.EnterDir(); err != nil {
		a.Status = err.Error()
	}
}

// parentDir は親ディレクトリに戻る。
func (a *App) parentDir() {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil {
		return
	}
	if err := p.ParentDir(); err != nil {
		a.Status = err.Error()
	}
}

// toggleFocus はペイン間（左右）のフォーカスを切り替える（横移動）。
// 縦位置（ペイン/プレビュー）は維持する。SourcePaneが未選択の場合は移動しない。
func (a *App) toggleFocus() {
	a.resetGPending()
	if a.SourcePane != nil {
		a.FocusLeft = !a.FocusLeft
	}
	a.Status = ""
}

// toggleFocusVertical はペインとプレビュー間（上下）のフォーカスを切り替える（縦移動）。
// 左右位置は維持する。
func (a *App) toggleFocusVertical() {
	a.resetGPending()
	a.previewFocused = !a.previewFocused
	a.Status = ""
}

// handlePreviewKey はプレビューフォーカス時のキー入力を処理する。
func (a *App) handlePreviewKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		a.toggleFocus()
	case "shift+tab":
		a.toggleFocusVertical()
	case "j", "down":
		a.resetGPending()
		a.previewScrollDown()
	case "k", "up":
		a.resetGPending()
		a.previewScrollUp()
	case "g":
		a.handlePreviewG()
	case "G":
		a.resetGPending()
		a.previewScroll = a.previewMaxScroll()
	case "esc":
		a.resetGPending()
		a.previewFocused = false
		a.previewScroll = 0
		a.Status = ""
	case "q", "ctrl+c":
		return a, a.quit()
	}
	return a, nil
}

// handlePreviewG はプレビューフォーカス時の'g'キーを処理する。ggで先頭へ。
func (a *App) handlePreviewG() {
	if a.gPending {
		a.gPending = false
		a.previewScroll = 0
		return
	}
	a.gPending = true
}

// selectRepo はリポジトリ選択ポップアップを開く。
func (a *App) selectRepo() tea.Cmd {
	a.resetGPending()
	return a.openRepoSelector()
}

// quit はアプリケーションを終了する。
func (a *App) quit() tea.Cmd {
	a.resetGPending()
	a.ExitReason = ExitReasonQuit
	return tea.Quit
}

// handleG は'g'キーを処理する。ggで先頭ジャンプ。
func (a *App) handleG() {
	if a.gPending {
		// gg: 先頭にジャンプ
		a.gPending = false
		p := a.focusedPane()
		if p == nil {
			return
		}
		p.MoveToTop()
		a.Status = ""
		return
	}
	a.gPending = true
}

// moveToBottom はカーソルを末尾に移動する。
func (a *App) moveToBottom() {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil {
		return
	}
	p.MoveToBottom()
}

// toggleSelect はカーソル行の選択をトグルする。
func (a *App) toggleSelect() {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil {
		return
	}
	p.ToggleSelect()
}

// selectAll は全エントリを選択する。
func (a *App) selectAll() {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil {
		return
	}
	p.SelectAll()
	a.Status = fmt.Sprintf("%d selected", len(p.Selected))
}

// invertSelection は選択を反転する。
func (a *App) invertSelection() {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil {
		return
	}
	p.InvertSelection()
	a.Status = fmt.Sprintf("%d selected", len(p.Selected))
}

// handleEsc はEscキーを処理する。フィルタ解除または選択解除。
func (a *App) handleEsc() {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil {
		return
	}
	// フィルタがかかっている場合はフィルタ解除
	if p.FilterQuery != "" {
		if err := p.ClearFilter(); err != nil {
			a.Status = fmt.Sprintf("refresh error: %v", err)
			return
		}
		a.Status = "filter cleared"
		return
	}
	// 選択がある場合は選択解除
	if len(p.Selected) > 0 {
		p.ClearSelection()
		a.Status = "selection cleared"
		return
	}
	a.Status = ""
}

// yank はフォーカス中のペインからヤンクする。
func (a *App) yank() {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil {
		return
	}
	// 同じペインで再度yを押した場合はヤンク取り消し
	if a.YankBuf != nil && len(a.YankBuf.Entries) > 0 && a.YankBuf.SrcDir == p.Dir {
		a.YankBuf = nil
		p.ClearSelection()
		a.Status = "yank cancelled"
		return
	}
	// 選択なしの場合はカーソル行を選択してからヤンク（*マーク表示のため）
	if len(p.Selected) == 0 && len(p.Entries) > 0 {
		p.Selected[p.Cursor] = true
	}
	a.YankBuf = p.Yank()
	a.Status = fmt.Sprintf("%d item(s) yanked", len(a.YankBuf.Entries))
}

// paste はヤンクバッファの内容をフォーカス中のペインにコピーする。
// 同名ファイルが存在する場合はスキップしてステータスに表示する。
func (a *App) paste() {
	a.resetGPending()
	if a.YankBuf == nil || len(a.YankBuf.Entries) == 0 {
		a.Status = "yank buffer is empty"
		return
	}

	dst := a.focusedPane()
	if dst == nil {
		a.Status = "no target pane for paste"
		return
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
			return
		}
		copied++
	}

	if err := dst.Refresh(); err != nil {
		a.Status = fmt.Sprintf("refresh error: %v", err)
		return
	}

	// コピー元ペインの選択状態とヤンクバッファをクリア
	a.clearYankSourceSelection()
	a.YankBuf = nil

	if skipped > 0 {
		a.Status = fmt.Sprintf("%d copied, %d skipped (duplicate)", copied, skipped)
	} else {
		a.Status = fmt.Sprintf("%d item(s) copied", copied)
	}
}

// pasteOverwrite はヤンクバッファの内容をフォーカス中のペインに上書きコピーする。
func (a *App) pasteOverwrite() {
	a.resetGPending()
	if a.YankBuf == nil || len(a.YankBuf.Entries) == 0 {
		a.Status = "yank buffer is empty"
		return
	}

	dst := a.focusedPane()
	if dst == nil {
		a.Status = "no target pane for paste"
		return
	}

	for _, src := range a.YankBuf.Entries {
		name := filepath.Base(src)
		dstPath := filepath.Join(dst.Dir, name)
		// コピー元と同一パスへの上書きはコピー元を破壊しうるため防ぐ
		if filepath.Clean(src) == filepath.Clean(dstPath) {
			a.Status = "copy error: source and destination are the same"
			return
		}
		if err := a.FS.Copy(src, dstPath); err != nil {
			a.Status = fmt.Sprintf("copy error: %v", err)
			return
		}
	}

	if err := dst.Refresh(); err != nil {
		a.Status = fmt.Sprintf("refresh error: %v", err)
		return
	}

	// コピー元ペインの選択状態とヤンクバッファをクリア
	a.clearYankSourceSelection()
	count := len(a.YankBuf.Entries)
	a.YankBuf = nil

	a.Status = fmt.Sprintf("%d item(s) overwritten", count)
}

// trashFile はカーソル行または選択中のエントリをゴミ箱に移動する確認モードに入る。
func (a *App) trashFile() {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil || len(p.Entries) == 0 {
		return
	}

	a.pendingTargets = a.getTargetPaths(p)
	a.inputMode = InputModeConfirmDelete
	a.Status = fmt.Sprintf("move to trash? (%d item(s)) [y/n]", len(a.pendingTargets))
}

// deleteFile はカーソル行または選択中のエントリを完全削除する確認モードに入る。
func (a *App) deleteFile() {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil || len(p.Entries) == 0 {
		return
	}

	a.pendingTargets = a.getTargetPaths(p)
	a.inputMode = InputModeConfirmForceDelete
	a.Status = fmt.Sprintf("permanently delete? (%d item(s)) [y/n]", len(a.pendingTargets))
}

// executeConfirm は確認モードの実行を行う（y押下時）。
func (a *App) executeConfirm() {
	p := a.focusedPane()
	if p == nil {
		a.inputMode = InputModeNone
		a.pendingTargets = nil
		a.Status = ""
		return
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
		return
	}

	if lastErr != nil {
		a.Status = fmt.Sprintf("%d processed, error: %v", count, lastErr)
	} else if mode == InputModeConfirmDelete {
		a.Status = fmt.Sprintf("%d item(s) moved to trash", count)
	} else {
		a.Status = fmt.Sprintf("%d item(s) permanently deleted", count)
	}
}

// createFile は新規ファイル/ディレクトリ作成モードに入る。
func (a *App) createFile() tea.Cmd {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil {
		return nil
	}
	return a.enterTextInput(InputModeCreate, "")
}

// renameFile はリネームモードに入る。
func (a *App) renameFile() tea.Cmd {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil || len(p.Entries) == 0 {
		return nil
	}
	return a.enterTextInput(InputModeRename, p.Entries[p.Cursor].Name)
}

// searchForward は前方検索モードに入る。
func (a *App) searchForward() tea.Cmd {
	a.resetGPending()
	a.searchFwd = true
	return a.enterTextInput(InputModeSearch, "")
}

// toggleHelp はキーバインドヘルプの表示/非表示を切り替える。
func (a *App) toggleHelp() {
	a.resetGPending()
	a.showHelp = !a.showHelp
	a.Status = ""
}

// searchNext は次の検索結果に移動する。
func (a *App) searchNext() {
	a.resetGPending()
	if a.searchQuery == "" {
		return
	}
	p := a.focusedPane()
	if p == nil {
		return
	}
	a.doSearch(p, a.searchFwd)
}

// searchPrev は前の検索結果に移動する。
func (a *App) searchPrev() {
	a.resetGPending()
	if a.searchQuery == "" {
		return
	}
	p := a.focusedPane()
	if p == nil {
		return
	}
	a.doSearch(p, !a.searchFwd)
}

// filterMode はフィルタモードに入る。
func (a *App) filterMode() tea.Cmd {
	a.resetGPending()
	return a.enterTextInput(InputModeFilter, "")
}

// toggleHidden は隠しファイルの表示/非表示を切り替える。
func (a *App) toggleHidden() {
	a.resetGPending()
	p := a.focusedPane()
	if p == nil {
		return
	}
	p.ShowHidden = !p.ShowHidden
	if err := p.Refresh(); err != nil {
		a.Status = fmt.Sprintf("refresh error: %v", err)
		return
	}
	if p.ShowHidden {
		a.Status = "show hidden files"
	} else {
		a.Status = "hide hidden files"
	}
}

// inputSubmit は入力モードの確定処理を行う。
func (a *App) inputSubmit() {
	input := a.input.Value()

	mode := a.inputMode
	a.inputMode = InputModeNone
	a.input.Blur()
	a.input.SetValue("")

	switch mode {
	case InputModeSearch:
		a.searchQuery = input
		// 確定時に検索方向を前方として保存し、後続のn/Nが正しい方向を使えるようにする
		a.searchFwd = true
		if input == "" {
			a.Status = ""
			return
		}
		p := a.focusedPane()
		if p == nil {
			return
		}
		a.doSearch(p, true)
	case InputModeFilter:
		p := a.focusedPane()
		if p == nil {
			return
		}
		if input == "" {
			if err := p.ClearFilter(); err != nil {
				a.Status = fmt.Sprintf("refresh error: %v", err)
				return
			}
			a.Status = "filter cleared"
		} else {
			p.FilterQuery = input
			if err := p.Refresh(); err != nil {
				a.Status = fmt.Sprintf("refresh error: %v", err)
				return
			}
			a.Status = fmt.Sprintf("filter: %s (%d item(s))", input, len(p.Entries))
		}
	case InputModeCreate:
		a.executeCreate(input)
	case InputModeRename:
		a.executeRename(input)
	}
}

// inputCancel は入力モードをキャンセルする。
func (a *App) inputCancel() {
	a.inputMode = InputModeNone
	a.input.Blur()
	a.input.SetValue("")
	a.Status = ""
}

// executeCreate は新規ファイル/ディレクトリの作成を実行する。
func (a *App) executeCreate(name string) {
	if name == "" {
		a.Status = "name is empty"
		return
	}

	p := a.focusedPane()
	if p == nil {
		return
	}

	isDir := strings.HasSuffix(name, "/")
	if isDir {
		name = strings.TrimSuffix(name, "/")
	}

	// パストラバーサル防止: ベース名のみ許可
	base := filepath.Base(name)
	if base != name || base == "." || base == ".." {
		a.Status = "invalid filename"
		return
	}

	path := filepath.Join(p.Dir, name)
	if err := a.FS.Create(path, isDir); err != nil {
		a.Status = fmt.Sprintf("create error: %v", err)
		return
	}

	if err := p.Refresh(); err != nil {
		a.Status = fmt.Sprintf("refresh error: %v", err)
		return
	}

	// 作成したエントリにカーソルを移動
	for i, e := range p.Entries {
		if e.Name == name {
			p.Cursor = i
			break
		}
	}

	if isDir {
		a.Status = fmt.Sprintf("directory created: %s", name)
	} else {
		a.Status = fmt.Sprintf("file created: %s", name)
	}
}

// executeRename はリネームを実行する。
func (a *App) executeRename(newName string) {
	if newName == "" {
		a.Status = "name is empty"
		return
	}

	p := a.focusedPane()
	if p == nil || len(p.Entries) == 0 {
		return
	}

	oldName := p.Entries[p.Cursor].Name
	if oldName == newName {
		a.Status = "name not changed"
		return
	}

	// パストラバーサル防止: ベース名のみ許可
	base := filepath.Base(newName)
	if base != newName || base == "." || base == ".." {
		a.Status = "invalid filename"
		return
	}

	oldPath := filepath.Join(p.Dir, oldName)
	newPath := filepath.Join(p.Dir, newName)
	if err := a.FS.Rename(oldPath, newPath); err != nil {
		a.Status = fmt.Sprintf("rename error: %v", err)
		return
	}

	if err := p.Refresh(); err != nil {
		a.Status = fmt.Sprintf("refresh error: %v", err)
		return
	}

	// リネームしたエントリにカーソルを移動
	for i, e := range p.Entries {
		if e.Name == newName {
			p.Cursor = i
			break
		}
	}

	a.Status = fmt.Sprintf("renamed: %s -> %s", oldName, newName)
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
	a.Status = fmt.Sprintf("not found: %s", a.searchQuery)
}
