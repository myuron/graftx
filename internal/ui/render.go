package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/myuron/graftx/internal/pane"
)

// 配色定数。
var (
	colorCursorBg     = lipgloss.Color("6")   // シアン（カーソル行背景、旧gocui.ColorCyan）
	colorCursorFg     = lipgloss.Color("0")   // 黒（カーソル行前景、旧gocui.ColorBlack）
	colorBorderActive = lipgloss.Color("2")   // 緑（アクティブペイン枠）
	colorBorderIdle   = lipgloss.Color("240") // グレー（非アクティブペイン枠）
)

// highlightMatch は検索クエリにマッチする部分をANSIカラーでハイライトする。
// 大文字小文字を無視してマッチし、最初のマッチ部分を黄色背景＋黒文字で装飾する。
// クエリが空またはマッチなしの場合はそのまま返す。
// マルチバイト文字を分割しないようrune境界で処理する。
func highlightMatch(name, query string) string {
	if query == "" {
		return name
	}
	orig := []rune(name)
	lower := []rune(strings.ToLower(name))
	q := []rune(strings.ToLower(query))
	idx := runeSliceIndex(lower, q)
	if idx < 0 {
		return name
	}
	end := idx + len(q)
	// ToLowerでrune数が変わる稀なケースに備えて範囲を検証する
	if end > len(orig) {
		return name
	}
	// 元の文字列の大文字小文字を保持してハイライト
	return string(orig[:idx]) + "\x1b[30;43m" + string(orig[idx:end]) + "\x1b[0m" + string(orig[end:])
}

// runeSliceIndex はruneスライスsの中でsubが最初に現れる位置（rune単位）を返す。
// 見つからなければ-1を返す。
func runeSliceIndex(s, sub []rune) int {
	if len(sub) == 0 {
		return 0
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		match := true
		for j := range sub {
			if s[i+j] != sub[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// render は画面全体を文字列として描画する。
func (a *App) render() string {
	if a.width <= 0 || a.height <= 0 {
		return ""
	}

	// リポジトリ選択ポップアップモードは中央モーダルで全画面を占有する
	if a.inputMode == InputModeSelectRepo {
		return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, a.renderSelectorModal())
	}

	statusHeight := 1
	paneAreaHeight := a.height - statusHeight
	if paneAreaHeight < 3 {
		paneAreaHeight = 3
	}

	leftWidth := a.width / 2
	rightWidth := a.width - leftWidth

	// フォーカス側はペインとプレビューに縦分割し、非フォーカス側は全高で描画する。
	paneH, previewH := splitPreviewHeight(paneAreaHeight)

	var left, right string
	if a.FocusLeft {
		pane := a.renderPane(a.SourcePane, " Source ", true, leftWidth, paneH)
		preview := a.renderPreview(a.SourcePane, leftWidth, previewH)
		left = lipgloss.JoinVertical(lipgloss.Left, pane, preview)
		right = a.renderPane(a.DestPane, " Dest ", false, rightWidth, paneAreaHeight)
	} else {
		left = a.renderPane(a.SourcePane, " Source ", false, leftWidth, paneAreaHeight)
		pane := a.renderPane(a.DestPane, " Dest ", true, rightWidth, paneH)
		preview := a.renderPreview(a.DestPane, rightWidth, previewH)
		right = lipgloss.JoinVertical(lipgloss.Left, pane, preview)
	}

	panes := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	return panes + "\n" + a.renderStatus()
}

// splitPreviewHeight はフォーカス側の利用可能高さをペインとプレビューに配分する。
// プレビューは全体の約1/3を占め、双方が最低3行（枠込み）を確保できるよう調整する。
func splitPreviewHeight(total int) (paneH, previewH int) {
	previewH = total / 3
	if previewH < 3 {
		previewH = 3
	}
	paneH = total - previewH
	if paneH < 3 {
		paneH = 3
		previewH = total - paneH
		if previewH < 1 {
			previewH = 1
		}
	}
	return paneH, previewH
}

// renderPane は1つのペインを枠付きで描画する。
// pがnilの場合（左ペイン未選択時）はプレースホルダを表示する。
func (a *App) renderPane(p *pane.Pane, defaultTitle string, focused bool, outerW, outerH int) string {
	innerW := outerW - 2
	innerH := outerH - 2
	if innerW < 1 {
		innerW = 1
	}
	if innerH < 1 {
		innerH = 1
	}

	borderColor := colorBorderIdle
	if focused {
		borderColor = colorBorderActive
	}
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Width(innerW).
		Height(innerH)

	var title string
	var body string
	if p == nil {
		title = defaultTitle
		body = a.renderPlaceholder(innerW, innerH-1)
	} else {
		title = fmt.Sprintf(" %s ", p.Dir)
		body = a.renderEntries(p, focused, innerW, innerH-1)
	}

	titleLine := lipgloss.NewStyle().Bold(true).MaxWidth(innerW).Render(title)
	content := titleLine + "\n" + body
	return boxStyle.Render(content)
}

// renderPlaceholder は左ペイン未選択時の中央メッセージを描画する。
func (a *App) renderPlaceholder(w, h int) string {
	msg := "[s]select repository"
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, msg)
}

// renderEntries はペインのエントリ一覧を可視範囲ぶん描画する。
// 選択済みエントリには "* " プレフィックス、ディレクトリには "/" サフィックスを付与する。
// focusedがtrueの場合はカーソル行をハイライトする。
func (a *App) renderEntries(p *pane.Pane, focused bool, width, visible int) string {
	if visible < 1 {
		visible = 1
	}

	// スクロールオフセットの計算（旧syncScrollと同じロジック）
	oy := 0
	if p.Cursor >= visible {
		oy = p.Cursor - visible + 1
	}
	end := oy + visible
	if end > len(p.Entries) {
		end = len(p.Entries)
	}

	cursorStyle := lipgloss.NewStyle().Background(colorCursorBg).Foreground(colorCursorFg).Width(width)
	normalStyle := lipgloss.NewStyle().Width(width)

	lines := make([]string, 0, visible)
	for i := oy; i < end; i++ {
		entry := p.Entries[i]

		prefix := "  "
		if p.Selected[i] {
			prefix = "* "
		}
		suffix := ""
		if entry.IsDir {
			suffix = "/"
		}
		icon := iconForEntry(entry) + " "

		// 検索クエリがある場合はマッチ部分をハイライト
		displayName := entry.Name
		if a.searchQuery != "" {
			displayName = highlightMatch(entry.Name, a.searchQuery)
		}
		text := prefix + icon + displayName + suffix

		if focused && i == p.Cursor {
			lines = append(lines, cursorStyle.Render(text))
		} else {
			lines = append(lines, normalStyle.Render(text))
		}
	}

	// 残りを空行で埋める
	for len(lines) < visible {
		lines = append(lines, normalStyle.Render(""))
	}

	return strings.Join(lines, "\n")
}

// renderStatus はステータスバーを描画する。
func (a *App) renderStatus() string {
	if a.isTextInputMode() {
		return a.inputPrefix() + a.input.View()
	}

	if a.Status != "" {
		return a.Status
	}

	focus := "right"
	if a.FocusLeft {
		focus = "left"
	}
	return fmt.Sprintf(" [%s] [q]Quit [Tab]Switch [h/j/k/l]Move", focus)
}

// renderSelectorModal はリポジトリ選択ポップアップのモーダルを描画する。
func (a *App) renderSelectorModal() string {
	// モーダルサイズ計算（60%幅 x 70%高さ、最小40x10）
	w := a.width * 60 / 100
	if w < 40 {
		w = 40
	}
	if w > a.width-2 {
		w = a.width - 2
	}
	h := a.height * 70 / 100
	if h < 10 {
		h = 10
	}
	if h > a.height-2 {
		h = a.height - 2
	}

	innerW := w - 2
	innerH := h - 2
	if innerW < 1 {
		innerW = 1
	}
	if innerH < 1 {
		innerH = 1
	}

	// フィルタ行（タイトル + 入力内容）
	filterLine := lipgloss.NewStyle().Bold(true).MaxWidth(innerW).Render(" Filter ")
	inputLine := lipgloss.NewStyle().MaxWidth(innerW).Render(a.filterInput.View())

	// リスト表示領域（フィルタ2行ぶんを差し引く）
	listVisible := innerH - 2
	if listVisible < 1 {
		listVisible = 1
	}
	listBody := a.renderSelectorList(innerW, listVisible)

	content := filterLine + "\n" + inputLine + "\n" + listBody
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorderActive).
		Width(innerW).
		Height(innerH)
	return boxStyle.Render(content)
}

// renderSelectorList はリポジトリ一覧を可視範囲ぶん描画する。
func (a *App) renderSelectorList(width, visible int) string {
	if visible < 1 {
		visible = 1
	}

	// スクロールオフセットの計算
	oy := 0
	if a.repoSelectorCursor >= visible {
		oy = a.repoSelectorCursor - visible + 1
	}
	end := oy + visible
	if end > len(a.filteredRepoList) {
		end = len(a.filteredRepoList)
	}

	cursorStyle := lipgloss.NewStyle().Background(colorCursorBg).Foreground(colorCursorFg).Width(width)
	normalStyle := lipgloss.NewStyle().Width(width)

	lines := make([]string, 0, visible)
	for i := oy; i < end; i++ {
		text := truncateToWidth(a.filteredRepoList[i], width)
		if i == a.repoSelectorCursor {
			lines = append(lines, cursorStyle.Render(text))
		} else {
			lines = append(lines, normalStyle.Render(text))
		}
	}
	for len(lines) < visible {
		lines = append(lines, normalStyle.Render(""))
	}

	return strings.Join(lines, "\n")
}

// truncateToWidth は表示幅がwを超える場合に末尾を切り詰める。
func truncateToWidth(s string, w int) string {
	if runewidth.StringWidth(s) <= w {
		return s
	}
	return runewidth.Truncate(s, w, "")
}
