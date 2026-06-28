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
		{"j / k", "カーソルを上下に移動"},
		{"h / l", "親ディレクトリへ / ディレクトリに入る"},
		{"gg / G", "先頭 / 末尾へ移動"},
		{"Tab", "左右ペインを切り替え"},
		{"Shift+Tab", "ペイン / プレビューを切り替え"},
		{"Space", "選択をトグル"},
		{"Ctrl+a", "全選択"},
		{"Ctrl+r", "選択を反転"},
		{"y", "ヤンク（コピー元を指定）"},
		{"p / P", "ペースト / 上書きペースト"},
		{"d / D", "ゴミ箱へ移動 / 完全削除"},
		{"a", "新規作成（末尾に / でディレクトリ）"},
		{"r", "リネーム"},
		{"/", "検索"},
		{"n / N", "次 / 前の検索結果へ"},
		{"f", "フィルタ"},
		{".", "隠しファイルの表示切替"},
		{"s", "リポジトリを選択"},
		{"?", "このヘルプを表示"},
		{"Esc", "選択 / フィルタを解除"},
		{"q", "終了（Quit）"},
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
