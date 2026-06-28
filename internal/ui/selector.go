package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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

// handleSelectorKey はリポジトリ選択ポップアップモードのキー入力を処理する。
func (a *App) handleSelectorKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		a.selectorConfirm()
		return a, nil
	case "esc", "ctrl+c":
		a.selectorCancel()
		return a, nil
	case "ctrl+j", "ctrl+n", "down":
		a.selectorMoveCursorDown()
		return a, nil
	case "ctrl+k", "ctrl+p", "up":
		a.selectorMoveCursorUp()
		return a, nil
	}
	// それ以外はフィルタ入力欄に委譲し、フィルタを更新する
	var cmd tea.Cmd
	a.filterInput, cmd = a.filterInput.Update(msg)
	a.repoFilterQuery = a.filterInput.Value()
	a.filterRepoList()
	return a, cmd
}

// openRepoSelector はリポジトリ選択ポップアップを開く。
func (a *App) openRepoSelector() tea.Cmd {
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
	a.filterInput.SetValue("")
	return a.filterInput.Focus()
}

// closeRepoSelector はリポジトリ選択ポップアップを閉じる。
func (a *App) closeRepoSelector() {
	a.inputMode = InputModeNone
	a.repoList = nil
	a.filteredRepoList = nil
	a.repoSelectorCursor = 0
	a.repoFilterQuery = ""
	a.filterInput.Blur()
	a.filterInput.SetValue("")
}

// selectorConfirm は選択確定処理。
func (a *App) selectorConfirm() {
	if len(a.filteredRepoList) == 0 {
		return
	}

	selected := a.filteredRepoList[a.repoSelectorCursor]
	a.closeRepoSelector()

	if err := a.SetSourceDir(selected); err != nil {
		a.Status = fmt.Sprintf("failed to set source directory: %v", err)
		return
	}

	a.FocusLeft = true
	a.Status = ""
}

// selectorCancel はキャンセル処理。
func (a *App) selectorCancel() {
	a.closeRepoSelector()
	a.Status = ""
}
