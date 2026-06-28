package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/myuron/graftx/internal/fs"
	"github.com/myuron/graftx/internal/pane"
)

func TestIsBinary_NULバイトを含むかで判定(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{name: "テキスト", data: []byte("hello world"), want: false},
		{name: "空", data: []byte(""), want: false},
		{name: "NULバイトを含む", data: []byte{'a', 0x00, 'b'}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBinary(tt.data); got != tt.want {
				t.Errorf("isBinary(%v) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}

func TestPreviewLines_テキストファイルの内容を行で返す(t *testing.T) {
	fsys := &stubFS{files: map[string][]byte{
		"/repo/a.txt": []byte("hello\nworld\n"),
	}}
	app := &App{FS: fsys}
	p := &pane.Pane{
		Dir:      "/repo",
		Entries:  []fs.Entry{{Name: "a.txt"}},
		Cursor:   0,
		Selected: map[int]bool{},
	}

	lines := app.previewContent(p, 20)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "hello") || !strings.Contains(joined, "world") {
		t.Errorf("テキスト内容が含まれていない: %q", joined)
	}
}

func TestPreviewLines_バイナリは不可メッセージ(t *testing.T) {
	fsys := &stubFS{files: map[string][]byte{
		"/repo/bin": {0x00, 0x01, 0x02},
	}}
	app := &App{FS: fsys}
	p := &pane.Pane{
		Dir:      "/repo",
		Entries:  []fs.Entry{{Name: "bin"}},
		Cursor:   0,
		Selected: map[int]bool{},
	}

	joined := strings.Join(app.previewContent(p, 20), "\n")
	if !strings.Contains(joined, "バイナリ") {
		t.Errorf("バイナリ不可メッセージが含まれていない: %q", joined)
	}
}

func TestPreviewLines_ディレクトリは中の一覧を返す(t *testing.T) {
	fsys := &stubFS{entries: []fs.Entry{
		{Name: "sub", IsDir: true},
		{Name: "f.go"},
	}}
	app := &App{FS: fsys}
	p := &pane.Pane{
		Dir:      "/repo",
		Entries:  []fs.Entry{{Name: "dir", IsDir: true}},
		Cursor:   0,
		Selected: map[int]bool{},
	}

	joined := strings.Join(app.previewContent(p, 20), "\n")
	if !strings.Contains(joined, "sub/") || !strings.Contains(joined, "f.go") {
		t.Errorf("ディレクトリ一覧が含まれていない: %q", joined)
	}
}

func TestPreviewLines_読み込み失敗は不可メッセージ(t *testing.T) {
	app := &App{FS: &stubFS{}}
	p := &pane.Pane{
		Dir:      "/repo",
		Entries:  []fs.Entry{{Name: "missing"}},
		Cursor:   0,
		Selected: map[int]bool{},
	}

	joined := strings.Join(app.previewContent(p, 20), "\n")
	if joined == "" {
		t.Errorf("読み込み失敗時にメッセージが返るべき")
	}
}

func TestPreviewLines_空ペインでもパニックしない(t *testing.T) {
	app := &App{FS: &stubFS{}}
	p := &pane.Pane{Dir: "/repo", Entries: nil, Cursor: 0, Selected: map[int]bool{}}
	if lines := app.previewContent(p, 20); len(lines) == 0 {
		t.Errorf("空ペインでも1行以上返すべき")
	}
	if lines := app.previewContent(nil, 20); len(lines) == 0 {
		t.Errorf("nilペインでも1行以上返すべき")
	}
}

func TestPreviewScroll_範囲内でクランプされる(t *testing.T) {
	// 総行数10・可視4行 → 最大オフセットは6
	a := &App{previewTotalLines: 10, previewVisibleLines: 4}

	a.previewScrollDown()
	if a.previewScroll != 1 {
		t.Errorf("1回下スクロールでオフセット1のはず: %d", a.previewScroll)
	}
	for i := 0; i < 100; i++ {
		a.previewScrollDown()
	}
	if a.previewScroll != 6 {
		t.Errorf("最大オフセット6でクランプされるべき: %d", a.previewScroll)
	}
	for i := 0; i < 100; i++ {
		a.previewScrollUp()
	}
	if a.previewScroll != 0 {
		t.Errorf("上スクロールで0でクランプされるべき: %d", a.previewScroll)
	}
}

func TestPreviewScroll_可視範囲が総行数以上ならスクロールしない(t *testing.T) {
	a := &App{previewTotalLines: 3, previewVisibleLines: 5}
	a.previewScrollDown()
	if a.previewScroll != 0 {
		t.Errorf("内容が可視範囲に収まる場合はスクロールしない: %d", a.previewScroll)
	}
}

func TestToggleFocus_横方向に左右ペインを切り替える(t *testing.T) {
	a := &App{
		SourcePane: &pane.Pane{Selected: map[int]bool{}},
		DestPane:   &pane.Pane{Selected: map[int]bool{}},
		FocusLeft:  false,
	}

	a.toggleFocus() // Dest → Source
	if !a.FocusLeft {
		t.Fatalf("横移動で左ペインになるべき: FocusLeft=%v", a.FocusLeft)
	}
	a.toggleFocus() // Source → Dest
	if a.FocusLeft {
		t.Fatalf("横移動で右ペインに戻るべき: FocusLeft=%v", a.FocusLeft)
	}

	// 縦位置（プレビュー）は横移動で維持される
	a.previewFocused = true
	a.toggleFocus()
	if !a.FocusLeft || !a.previewFocused {
		t.Fatalf("横移動でプレビュー状態を維持すべき: FocusLeft=%v preview=%v", a.FocusLeft, a.previewFocused)
	}
}

func TestToggleFocus_Source未選択時は横移動しない(t *testing.T) {
	a := &App{DestPane: &pane.Pane{Selected: map[int]bool{}}, FocusLeft: false}

	a.toggleFocus()
	if a.FocusLeft {
		t.Fatalf("Source未選択時は左に移動しないべき: FocusLeft=%v", a.FocusLeft)
	}
}

func TestToggleFocusVertical_縦方向にペインとプレビューを切り替える(t *testing.T) {
	a := &App{DestPane: &pane.Pane{Selected: map[int]bool{}}, FocusLeft: false}

	a.toggleFocusVertical() // ペイン → プレビュー
	if !a.previewFocused {
		t.Fatalf("縦移動でプレビューにフォーカスすべき: preview=%v", a.previewFocused)
	}
	a.toggleFocusVertical() // プレビュー → ペイン
	if a.previewFocused {
		t.Fatalf("縦移動でペインに戻るべき: preview=%v", a.previewFocused)
	}

	// 左右位置は縦移動で維持される
	a.FocusLeft = true
	a.toggleFocusVertical()
	if !a.FocusLeft || !a.previewFocused {
		t.Fatalf("縦移動で左右位置を維持すべき: FocusLeft=%v preview=%v", a.FocusLeft, a.previewFocused)
	}
}

func TestUpdate_ShiftTabで縦移動(t *testing.T) {
	app := newDestApp()

	// Destペイン → Destプレビュー
	app.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if !app.previewFocused {
		t.Errorf("Shift+Tabでプレビューにフォーカスするべき: preview=%v", app.previewFocused)
	}
	// Destプレビュー → Destペイン
	app.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if app.previewFocused {
		t.Errorf("Shift+Tabでペインに戻るべき: preview=%v", app.previewFocused)
	}
}

func TestUpdate_プレビューフォーカス中のTabは横移動して維持(t *testing.T) {
	app := newDestApp()
	app.SourcePane = &pane.Pane{
		Dir:      "/src",
		Entries:  []fs.Entry{{Name: "x"}},
		Cursor:   0,
		Selected: map[int]bool{},
	}

	// Destプレビューにフォーカス
	app.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if app.FocusLeft || !app.previewFocused {
		t.Fatalf("Destプレビューになるべき: FocusLeft=%v preview=%v", app.FocusLeft, app.previewFocused)
	}
	// Tab で横移動。プレビュー状態は維持されSourceプレビューへ
	app.Update(tea.KeyMsg{Type: tea.KeyTab})
	if !app.FocusLeft || !app.previewFocused {
		t.Errorf("Tabで横移動しプレビューを維持すべき: FocusLeft=%v preview=%v", app.FocusLeft, app.previewFocused)
	}
}

func TestUpdate_プレビューフォーカス時にjでスクロールする(t *testing.T) {
	// 可視範囲を超える行数のファイルを用意する
	var sb strings.Builder
	for i := 0; i < 100; i++ {
		sb.WriteString("line\n")
	}
	fsys := &stubFS{
		entries: []fs.Entry{{Name: "big.txt"}},
		files:   map[string][]byte{"/dst/big.txt": []byte(sb.String())},
	}
	app, _ := NewApp("/dst", fsys, nil)
	app.DestPane = &pane.Pane{
		Dir:      "/dst",
		Entries:  []fs.Entry{{Name: "big.txt"}},
		Cursor:   0,
		Selected: map[int]bool{},
	}
	app.width = 80
	app.height = 24

	// プレビューにフォーカス（Destペイン → Destプレビュー）
	app.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if !app.previewFocused {
		t.Fatal("Shift+Tabでプレビューにフォーカスするべき")
	}
	// 描画で総行数・可視行数を確定させる
	app.View()
	if app.previewScroll != 0 {
		t.Fatalf("初期スクロールは0のはず: %d", app.previewScroll)
	}
	// j でスクロール
	app.Update(keyRunes("j"))
	if app.previewScroll != 1 {
		t.Errorf("jでスクロールオフセットが1になるべき: %d", app.previewScroll)
	}
	// k で戻る
	app.Update(keyRunes("k"))
	if app.previewScroll != 0 {
		t.Errorf("kでスクロールオフセットが0に戻るべき: %d", app.previewScroll)
	}
}

// render はフォーカス中ペインの下にプレビューを表示する。
func TestRender_フォーカス側にプレビューが表示される(t *testing.T) {
	fsys := &stubFS{
		entries: []fs.Entry{{Name: "apple"}},
		files: map[string][]byte{
			"/src/apple": []byte("SOURCE-CONTENT"),
		},
	}
	app, _ := NewApp("/dst", fsys, nil)
	app.SourcePane = &pane.Pane{
		Dir:      "/src",
		Entries:  []fs.Entry{{Name: "apple"}},
		Cursor:   0,
		Selected: map[int]bool{},
	}
	app.DestPane = &pane.Pane{
		Dir:      "/dst",
		Entries:  []fs.Entry{{Name: "apple"}},
		Cursor:   0,
		Selected: map[int]bool{},
	}
	app.width = 80
	app.height = 24
	app.FocusLeft = true

	out := app.View()
	if !strings.Contains(out, "Preview") {
		t.Errorf("プレビュー領域のタイトルが表示されるべき: %q", out)
	}
	if !strings.Contains(out, "SOURCE-CONTENT") {
		t.Errorf("フォーカス側のファイル内容が表示されるべき: %q", out)
	}
}
