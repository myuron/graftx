package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// helpEntry は1つのキーバインドの表記と説明を表す。
type helpEntry struct {
	keys string // キー表記
	desc string // 説明
}

// helpEntries はヘルプモーダルに表示するキーバインド一覧を返す。
func helpEntries() []helpEntry {
	return []helpEntry{
		{"j / k", "Move cursor down / up"},
		{"h / l", "Go to parent / enter directory"},
		{"gg / G", "Jump to top / bottom"},
		{"Tab", "Switch left / right pane"},
		{"Shift+Tab", "Switch pane / preview"},
		{"Space", "Toggle selection"},
		{"Ctrl+a", "Select all"},
		{"Ctrl+r", "Invert selection"},
		{"y", "Yank (mark copy source)"},
		{"p / P", "Paste / overwrite paste"},
		{"d / D", "Move to trash / delete permanently"},
		{"a", "Create new (trailing / for directory)"},
		{"r", "Rename"},
		{"/", "Search"},
		{"n / N", "Next / previous search result"},
		{"f", "Filter"},
		{".", "Toggle hidden files"},
		{"s", "Select repository"},
		{"?", "Show this help"},
		{"Esc", "Clear selection / filter"},
		{"q", "Quit"},
	}
}

// renderHelpModal はキーバインド一覧の中央モーダルを描画する。
func (a *App) renderHelpModal() string {
	entries := helpEntries()

	// キー表記カラムの最大幅を求めて説明をそろえる。
	keyWidth := 0
	for _, e := range entries {
		if w := runewidth.StringWidth(e.keys); w > keyWidth {
			keyWidth = w
		}
	}

	title := lipgloss.NewStyle().Bold(true).Render(" Keybindings ")
	lines := make([]string, 0, len(entries)+1)
	lines = append(lines, title, "")
	for _, e := range entries {
		pad := strings.Repeat(" ", keyWidth-runewidth.StringWidth(e.keys))
		lines = append(lines, "  "+e.keys+pad+"  "+e.desc)
	}
	lines = append(lines, "", " any key / Esc to close ")

	content := strings.Join(lines, "\n")
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorderActive).
		Padding(0, 1)
	return boxStyle.Render(content)
}
