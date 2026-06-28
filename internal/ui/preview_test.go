package ui

import (
	"strings"
	"testing"

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

	lines := app.previewLines(p, 20, 5)
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

	joined := strings.Join(app.previewLines(p, 20, 5), "\n")
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

	joined := strings.Join(app.previewLines(p, 20, 5), "\n")
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

	joined := strings.Join(app.previewLines(p, 20, 5), "\n")
	if joined == "" {
		t.Errorf("読み込み失敗時にメッセージが返るべき")
	}
}

func TestPreviewLines_空ペインでもパニックしない(t *testing.T) {
	app := &App{FS: &stubFS{}}
	p := &pane.Pane{Dir: "/repo", Entries: nil, Cursor: 0, Selected: map[int]bool{}}
	if lines := app.previewLines(p, 20, 5); len(lines) == 0 {
		t.Errorf("空ペインでも1行以上返すべき")
	}
	if lines := app.previewLines(nil, 20, 5); len(lines) == 0 {
		t.Errorf("nilペインでも1行以上返すべき")
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
