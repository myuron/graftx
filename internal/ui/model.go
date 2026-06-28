package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Init はtea.Modelインターフェースの実装。初期コマンドを返す。
func (a *App) Init() tea.Cmd {
	return textinput.Blink
}

// Update はtea.Modelインターフェースの実装。
// メッセージを受け取り状態を更新する。
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil
	case repoListMsg:
		return a.handleRepoList(msg)
	case tea.KeyMsg:
		return a.handleKey(msg)
	default:
		// カーソル点滅など非キー系メッセージをアクティブな入力欄に流す
		return a.forwardToInput(msg)
	}
}

// forwardToInput は現在のモードに応じて非キーメッセージを入力欄へ委譲する。
func (a *App) forwardToInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch {
	case a.inputMode == InputModeSelectRepo:
		a.filterInput, cmd = a.filterInput.Update(msg)
	case a.isTextInputMode():
		a.input, cmd = a.input.Update(msg)
	}
	return a, cmd
}

// handleKey はキー入力を現在のモードに応じてルーティングする。
func (a *App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case a.inputMode == InputModeSelectRepo:
		return a.handleSelectorKey(msg)
	case a.isTextInputMode():
		return a.handleInputKey(msg)
	case a.inputMode == InputModeConfirmDelete || a.inputMode == InputModeConfirmForceDelete:
		return a.handleConfirmKey(msg)
	case a.previewFocused:
		return a.handlePreviewKey(msg)
	default:
		return a.handleNormalKey(msg)
	}
}

// handleNormalKey は通常モードのVimバインドを処理する。
func (a *App) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j":
		a.cursorDown()
	case "k":
		a.cursorUp()
	case "l":
		a.enterDir()
	case "h":
		a.parentDir()
	case "tab":
		a.toggleFocus()
	case "shift+tab":
		a.toggleFocusVertical()
	case "s":
		return a, a.selectRepo()
	case "q":
		return a, a.quit()
	case "g":
		a.handleG()
	case "G":
		a.moveToBottom()
	case " ":
		a.toggleSelect()
	case "ctrl+a":
		a.selectAll()
	case "ctrl+r":
		a.invertSelection()
	case "esc", "ctrl+c":
		a.handleEsc()
	case "y":
		a.yank()
	case "p":
		a.paste()
	case "P":
		a.pasteOverwrite()
	case "d":
		a.trashFile()
	case "D":
		a.deleteFile()
	case "a":
		return a, a.createFile()
	case "r":
		return a, a.renameFile()
	case "/":
		return a, a.searchForward()
	case "?":
		return a, a.searchBackward()
	case "n":
		a.searchNext()
	case "N":
		a.searchPrev()
	case "f":
		return a, a.filterMode()
	case ".":
		a.toggleHidden()
	}
	return a, nil
}

// handleInputKey はテキスト入力モードのキー入力を処理する。
func (a *App) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		a.inputSubmit()
		return a, nil
	case "esc", "ctrl+c":
		a.inputCancel()
		return a, nil
	}
	var cmd tea.Cmd
	a.input, cmd = a.input.Update(msg)
	return a, cmd
}

// handleConfirmKey は確認モード（y/n）のキー入力を処理する。
func (a *App) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		a.executeConfirm()
	case "n", "esc", "ctrl+c":
		a.inputMode = InputModeNone
		a.pendingTargets = nil
		a.Status = "cancelled"
	}
	return a, nil
}

// View はtea.Modelインターフェースの実装。画面全体を文字列として返す。
func (a *App) View() string {
	return a.render()
}
