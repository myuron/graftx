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

// previewLines はペインのカーソル行に対応するプレビュー内容を行配列で返す。
// テキストファイルは内容、ディレクトリは中の一覧、バイナリや読み込み失敗は
// その旨のメッセージを返す。最大height行・幅widthに切り詰める。
func (a *App) previewLines(p *pane.Pane, width, height int) []string {
	if height < 1 {
		height = 1
	}
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
		lines := make([]string, 0, len(entries))
		for _, e := range entries {
			name := e.Name
			if e.IsDir {
				name += "/"
			}
			lines = append(lines, truncateToWidth(name, width))
			if len(lines) >= height {
				break
			}
		}
		if len(lines) == 0 {
			return []string{"(空のディレクトリ)"}
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
	lines := make([]string, 0, height)
	for _, line := range raw {
		// CRLF対応とタブ展開で表示崩れを防ぐ
		line = strings.TrimSuffix(line, "\r")
		line = strings.ReplaceAll(line, "\t", "    ")
		lines = append(lines, truncateToWidth(line, width))
		if len(lines) >= height {
			break
		}
	}
	return lines
}

// renderPreview はフォーカス中ペインのプレビュー領域を枠付きで描画する。
func (a *App) renderPreview(p *pane.Pane, outerW, outerH int) string {
	innerW := outerW - 2
	innerH := outerH - 2
	if innerW < 1 {
		innerW = 1
	}
	if innerH < 1 {
		innerH = 1
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(colorBorderIdle).
		Width(innerW).
		Height(innerH)

	// タイトル1行ぶんを差し引いた高さに本文を収める
	bodyHeight := innerH - 1
	if bodyHeight < 1 {
		bodyHeight = 1
	}
	lines := a.previewLines(p, innerW, bodyHeight)

	normalStyle := lipgloss.NewStyle().Width(innerW)
	rendered := make([]string, 0, bodyHeight)
	for _, line := range lines {
		rendered = append(rendered, normalStyle.Render(line))
		if len(rendered) >= bodyHeight {
			break
		}
	}
	for len(rendered) < bodyHeight {
		rendered = append(rendered, normalStyle.Render(""))
	}

	titleLine := lipgloss.NewStyle().Bold(true).MaxWidth(innerW).Render(" Preview ")
	content := titleLine + "\n" + strings.Join(rendered, "\n")
	return boxStyle.Render(content)
}
