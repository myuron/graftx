package ui

import (
	"testing"

	"github.com/myuron/graftx/internal/fs"
	"github.com/myuron/graftx/internal/pane"
)

func TestYank_トグルで取り消し(t *testing.T) {
	entries := []fs.Entry{
		{Name: "file1.txt", IsDir: false},
		{Name: "file2.txt", IsDir: false},
	}
	p := &pane.Pane{
		Dir:      "/src",
		Entries:  entries,
		Selected: map[int]bool{0: true},
	}
	app := &App{
		SourcePane: p,
		FocusLeft:  true,
	}

	// 1回目のy: ヤンクされる
	if err := app.yank(nil, nil); err != nil {
		t.Fatal(err)
	}
	if app.YankBuf == nil || len(app.YankBuf.Entries) == 0 {
		t.Fatal("1回目のyでYankBufが空")
	}
	if len(app.YankBuf.Entries) != 1 {
		t.Fatalf("YankBuf.Entries = %d, want 1", len(app.YankBuf.Entries))
	}

	// 2回目のy（同じペイン）: ヤンク取り消し
	if err := app.yank(nil, nil); err != nil {
		t.Fatal(err)
	}
	if app.YankBuf != nil && len(app.YankBuf.Entries) > 0 {
		t.Error("2回目のyでYankBufがクリアされていない")
	}
	if len(p.Selected) != 0 {
		t.Errorf("2回目のyでSelectedがクリアされていない: %v", p.Selected)
	}
}

func TestYank_別ペインからは取り消しではなく新規ヤンク(t *testing.T) {
	srcEntries := []fs.Entry{
		{Name: "src1.txt", IsDir: false},
	}
	dstEntries := []fs.Entry{
		{Name: "dst1.txt", IsDir: false},
	}
	srcPane := &pane.Pane{
		Dir:      "/src",
		Entries:  srcEntries,
		Selected: map[int]bool{0: true},
	}
	dstPane := &pane.Pane{
		Dir:      "/dst",
		Entries:  dstEntries,
		Selected: map[int]bool{},
	}
	app := &App{
		SourcePane: srcPane,
		DestPane:   dstPane,
		FocusLeft:  true,
	}

	// 左ペインでヤンク
	if err := app.yank(nil, nil); err != nil {
		t.Fatal(err)
	}
	if app.YankBuf == nil || len(app.YankBuf.Entries) != 1 {
		t.Fatal("左ペインでのヤンクが失敗")
	}

	// 右ペインに切り替えてy → 取り消しではなく新規ヤンク
	app.FocusLeft = false
	if err := app.yank(nil, nil); err != nil {
		t.Fatal(err)
	}
	if app.YankBuf == nil || len(app.YankBuf.Entries) == 0 {
		t.Fatal("別ペインからのyでYankBufが空")
	}
	if app.YankBuf.SrcDir != "/dst" {
		t.Errorf("YankBuf.SrcDir = %q, want /dst", app.YankBuf.SrcDir)
	}
}
