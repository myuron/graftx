package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/myuron/graftx/internal/fs"
	"github.com/myuron/graftx/internal/pane"
)

// テスト中は色出力を決定的にするためANSIプロファイルを強制する。
func init() {
	lipgloss.SetColorProfile(termenv.ANSI)
}

func TestHighlightMatch_マッチ部分を黄色背景でハイライト(t *testing.T) {
	tests := []struct {
		name  string
		input string
		query string
		want  string
	}{
		{
			name:  "完全一致",
			input: "foo",
			query: "foo",
			want:  "\x1b[30;43mfoo\x1b[0m",
		},
		{
			name:  "部分一致",
			input: "foobar",
			query: "bar",
			want:  "foo\x1b[30;43mbar\x1b[0m",
		},
		{
			name:  "大文字小文字を無視",
			input: "FooBar",
			query: "foo",
			want:  "\x1b[30;43mFoo\x1b[0m" + "Bar",
		},
		{
			name:  "マッチなし",
			input: "hello",
			query: "xyz",
			want:  "hello",
		},
		{
			name:  "空クエリ",
			input: "hello",
			query: "",
			want:  "hello",
		},
		{
			name:  "複数マッチは最初のみ",
			input: "abcabc",
			query: "abc",
			want:  "\x1b[30;43mabc\x1b[0m" + "abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := highlightMatch(tt.input, tt.query)
			if got != tt.want {
				t.Errorf("highlightMatch(%q, %q) = %q, want %q", tt.input, tt.query, got, tt.want)
			}
		})
	}
}

// renderEntries はフォーカス中のペインでのみカーソル行をハイライトする。
func TestRenderEntries_フォーカス時のみカーソル行をハイライト(t *testing.T) {
	p := &pane.Pane{
		Entries:  []fs.Entry{{Name: "a.txt"}, {Name: "b.txt"}},
		Cursor:   0,
		Selected: map[int]bool{},
	}

	focused := (&App{}).renderEntries(p, true, 20, 2)
	if !strings.Contains(focused, "\x1b[") {
		t.Errorf("フォーカス時はカーソル行に色エスケープが含まれるべき: %q", focused)
	}

	idle := (&App{}).renderEntries(p, false, 20, 2)
	if strings.Contains(idle, "\x1b[") {
		t.Errorf("非フォーカス時は色エスケープを含まないべき: %q", idle)
	}
}

// renderEntries はスクロールオフセットを正しく計算し可視範囲を返す。
func TestRenderEntries_スクロールで可視範囲を切り替え(t *testing.T) {
	entries := []fs.Entry{
		{Name: "e0"}, {Name: "e1"}, {Name: "e2"}, {Name: "e3"}, {Name: "e4"},
	}
	p := &pane.Pane{
		Entries:  entries,
		Cursor:   4,
		Selected: map[int]bool{},
	}

	// 可視3行・カーソル4 → オフセット2、e2/e3/e4が見える
	out := (&App{}).renderEntries(p, true, 20, 3)
	if strings.Contains(out, "e0") || strings.Contains(out, "e1") {
		t.Errorf("スクロール後にe0/e1が表示されている: %q", out)
	}
	for _, name := range []string{"e2", "e3", "e4"} {
		if !strings.Contains(out, name) {
			t.Errorf("%s が可視範囲に表示されていない: %q", name, out)
		}
	}
}
