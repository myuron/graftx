package ui

import (
	"testing"

	"github.com/jroimartin/gocui"
	"github.com/myuron/graftx/internal/pane"
)

// newTestGui はテスト用のgocui.Guiインスタンスを返す。
// gocui.NewGuiが失敗する環境ではテストをスキップする。
func newTestGui(t *testing.T) *gocui.Gui {
	t.Helper()
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		t.Skip("gocui.NewGui failed (no terminal):", err)
	}
	return g
}

func TestHighlight_フォーカスペインのみハイライト_右フォーカス(t *testing.T) {
	g := newTestGui(t)
	defer g.Close()

	app := &App{
		DestPane:  &pane.Pane{},
		FocusLeft: false,
	}
	app.gui = g
	g.SetManager(app)

	// Layoutを1回実行してビューを作成
	if err := app.Layout(g); err != nil {
		t.Fatal(err)
	}

	leftView, err := g.View(viewLeft)
	if err != nil {
		t.Fatal(err)
	}
	rightView, err := g.View(viewRight)
	if err != nil {
		t.Fatal(err)
	}

	// 右フォーカス時: 右ペインのみHighlightがtrue
	if leftView.Highlight {
		t.Error("左ペインのHighlightがtrue、右フォーカス時はfalseであるべき")
	}
	if !rightView.Highlight {
		t.Error("右ペインのHighlightがfalse、右フォーカス時はtrueであるべき")
	}
}

func TestHighlight_フォーカスペインのみハイライト_左フォーカス(t *testing.T) {
	g := newTestGui(t)
	defer g.Close()

	app := &App{
		DestPane:   &pane.Pane{},
		SourcePane: &pane.Pane{},
		FocusLeft:  true,
	}
	app.gui = g
	g.SetManager(app)

	if err := app.Layout(g); err != nil {
		t.Fatal(err)
	}

	leftView, err := g.View(viewLeft)
	if err != nil {
		t.Fatal(err)
	}
	rightView, err := g.View(viewRight)
	if err != nil {
		t.Fatal(err)
	}

	// 左フォーカス時: 左ペインのみHighlightがtrue
	if !leftView.Highlight {
		t.Error("左ペインのHighlightがfalse、左フォーカス時はtrueであるべき")
	}
	if rightView.Highlight {
		t.Error("右ペインのHighlightがtrue、左フォーカス時はfalseであるべき")
	}
}

func TestHighlight_フォーカス切替で動的に変化(t *testing.T) {
	g := newTestGui(t)
	defer g.Close()

	app := &App{
		DestPane:   &pane.Pane{},
		SourcePane: &pane.Pane{},
		FocusLeft:  false,
	}
	app.gui = g
	g.SetManager(app)

	// 最初は右フォーカス
	if err := app.Layout(g); err != nil {
		t.Fatal(err)
	}

	leftView, _ := g.View(viewLeft)
	rightView, _ := g.View(viewRight)

	if leftView.Highlight {
		t.Error("初期状態: 左ペインのHighlightがtrue")
	}
	if !rightView.Highlight {
		t.Error("初期状態: 右ペインのHighlightがfalse")
	}

	// フォーカスを左に切り替え
	app.FocusLeft = true
	if err := app.Layout(g); err != nil {
		t.Fatal(err)
	}

	if !leftView.Highlight {
		t.Error("左フォーカス後: 左ペインのHighlightがfalse")
	}
	if rightView.Highlight {
		t.Error("左フォーカス後: 右ペインのHighlightがtrue")
	}
}
