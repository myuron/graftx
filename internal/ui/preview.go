package ui

import (
	"bytes"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/myuron/graftx/internal/pane"
)

// isBinary はデータがバイナリかどうかをNULバイトの有無で簡易判定する。
func isBinary(data []byte) bool {
	return bytes.IndexByte(data, 0) >= 0
}

// previewContent はペインのカーソル行に対応するプレビュー内容を全行返す。
// テキストファイルは内容、ディレクトリは中の一覧、バイナリや読み込み失敗は
// その旨のメッセージを返す。各行は幅widthに切り詰める。
func (a *App) previewContent(p *pane.Pane, width int) []string {
	if p == nil || len(p.Entries) == 0 {
		return []string{"(プレビューなし)"}
	}

	entry := p.Entries[p.Cursor]
	path := filepath.Join(p.Dir, entry.Name)

	if entry.IsDir {
		entries, err := a.FS.ReadDir(path)
		if err != nil {
			return []string{"(読み込みエラー)"}
		}
		if len(entries) == 0 {
			return []string{"(空のディレクトリ)"}
		}
		lines := make([]string, 0, len(entries))
		for _, e := range entries {
			name := e.Name
			if e.IsDir {
				name += "/"
			}
			lines = append(lines, truncateToWidth(name, width))
		}
		return lines
	}

	data, err := a.FS.ReadFile(path)
	if err != nil {
		return []string{"(読み込みエラー)"}
	}
	if isBinary(data) {
		return []string{"(バイナリファイル)"}
	}
	if len(data) == 0 {
		return []string{"(空のファイル)"}
	}

	raw := strings.Split(string(data), "\n")
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		// CRLF対応とタブ展開で表示崩れを防ぐ
		line = strings.TrimSuffix(line, "\r")
		line = strings.ReplaceAll(line, "\t", "    ")
		lines = append(lines, truncateToWidth(line, width))
	}
	return lines
}

// previewMaxScroll は直近の描画情報から最大スクロールオフセットを返す。
func (a *App) previewMaxScroll() int {
	max := a.previewTotalLines - a.previewVisibleLines
	if max < 0 {
		return 0
	}
	return max
}

// previewScrollDown はプレビューを1行下にスクロールする。
func (a *App) previewScrollDown() {
	if a.previewScroll < a.previewMaxScroll() {
		a.previewScroll++
	}
}

// previewScrollUp はプレビューを1行上にスクロールする。
func (a *App) previewScrollUp() {
	if a.previewScroll > 0 {
		a.previewScroll--
	}
}

// renderPreview はフォーカス中ペインのプレビュー領域を枠付きで描画する。
// focusedがtrueの場合は枠をアクティブ色にする。
func (a *App) renderPreview(p *pane.Pane, outerW, outerH int, focused bool) string {
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

	// タイトル1行ぶんを差し引いた高さに本文を収める
	bodyHeight := innerH - 1
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	content := a.previewContent(p, innerW)

	// スクロール状態を直近の描画情報として保存し、範囲外をクランプする
	a.previewTotalLines = len(content)
	a.previewVisibleLines = bodyHeight
	if a.previewScroll > a.previewMaxScroll() {
		a.previewScroll = a.previewMaxScroll()
	}
	if a.previewScroll < 0 {
		a.previewScroll = 0
	}

	// 可視範囲を切り出す
	start := a.previewScroll
	end := start + bodyHeight
	if end > len(content) {
		end = len(content)
	}
	visible := content[start:end]

	normalStyle := lipgloss.NewStyle().Width(innerW)
	rendered := make([]string, 0, bodyHeight)
	for _, line := range visible {
		rendered = append(rendered, normalStyle.Render(line))
	}
	for len(rendered) < bodyHeight {
		rendered = append(rendered, normalStyle.Render(""))
	}

	titleLine := lipgloss.NewStyle().Bold(true).MaxWidth(innerW).Render(a.previewTitle())
	content2 := titleLine + "\n" + strings.Join(rendered, "\n")
	return boxStyle.Render(content2)
}

// previewTitle はプレビュー領域のタイトル文字列を返す。
// スクロール可能な場合は位置インジケータを付与する。
func (a *App) previewTitle() string {
	if a.previewMaxScroll() > 0 {
		return " Preview ▲▼ "
	}
	return " Preview "
}
