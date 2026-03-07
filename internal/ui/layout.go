// Package ui はgocuiを使ったTUIの描画とイベント処理を行うパッケージ。
package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/myuron/graftx/internal/fs"
	"github.com/myuron/graftx/internal/pane"
	"github.com/myuron/graftx/internal/selector"
)

// ExitReason はMainLoop終了後の理由を表す型。
type ExitReason int

const (
	// ExitReasonNone は未設定。
	ExitReasonNone ExitReason = iota
	// ExitReasonQuit はユーザーによる終了。
	ExitReasonQuit
	// ExitReasonSelectRepo はリポジトリ選択のための一時終了。
	ExitReasonSelectRepo
)

// View名の定数。
const (
	viewLeft   = "left"
	viewRight  = "right"
	viewStatus = "status"
)

// App はTUIアプリケーション全体の状態を管理する構造体。
type App struct {
	SourcePane *pane.Pane             // 左ペイン（コピー元）、未選択時はnil
	DestPane   *pane.Pane             // 右ペイン（コピー先）
	YankBuffer *pane.YankBuffer       // ヤンクバッファ
	Selector   selector.CommandRunner // リポジトリ選択
	FS         fs.FileSystem          // ファイルシステム
	FocusLeft  bool                   // trueなら左ペインにフォーカス
	Status     string                 // ステータスバーに表示するメッセージ
	ExitReason ExitReason             // MainLoop終了理由
	gui        *gocui.Gui             // gocuiインスタンスへの参照
}

// NewApp は新しいAppを作成する。
// destDirは右ペイン（コピー先）の初期ディレクトリ。
func NewApp(destDir string, fsys fs.FileSystem, sel selector.CommandRunner) (*App, error) {
	destPane, err := pane.NewPane(destDir, fsys)
	if err != nil {
		return nil, err
	}

	return &App{
		DestPane:   destPane,
		Selector:   sel,
		FS:         fsys,
		FocusLeft:  false,
		ExitReason: ExitReasonNone,
	}, nil
}

// SetGui はgocuiのGuiインスタンスを設定し、Manager登録・GUI設定・キーバインド登録を行う。
func (a *App) SetGui(g *gocui.Gui) {
	a.gui = g
	g.SetManager(a)
	g.Highlight = true
	g.SelBgColor = gocui.ColorGreen
	g.SelFgColor = gocui.ColorBlack
	g.Cursor = false
	a.setKeybindings(g)
}

// Layout はgocui.Managerインターフェースの実装。
// 左右ペインとステータスバーのレイアウトを構築する。
func (a *App) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// 左ペイン
	if v, err := g.SetView(viewLeft, 0, 0, maxX/2-1, maxY-3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		// 初回作成時の設定
		v.Highlight = true
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack
	}

	// 右ペイン
	if v, err := g.SetView(viewRight, maxX/2, 0, maxX-1, maxY-3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		// 初回作成時の設定
		v.Highlight = true
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack
	}

	// ステータスバー
	if v, err := g.SetView(viewStatus, 0, maxY-2, maxX-1, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
	}

	// フォーカス中のビューをcurrentViewに設定
	if a.FocusLeft {
		g.SetCurrentView(viewLeft)
	} else {
		g.SetCurrentView(viewRight)
	}

	// 左ペインの描画
	if err := a.renderLeftPane(g); err != nil {
		return err
	}

	// 右ペインの描画
	if err := a.renderRightPane(g); err != nil {
		return err
	}

	// ステータスバーの描画
	if err := a.renderStatus(g); err != nil {
		return err
	}

	return nil
}

// renderLeftPane は左ペインの内容を描画する。
func (a *App) renderLeftPane(g *gocui.Gui) error {
	v, err := g.View(viewLeft)
	if err != nil {
		return err
	}
	v.Clear()

	if a.SourcePane == nil {
		v.Title = " Source "
		fmt.Fprintln(v, "")
		fmt.Fprintln(v, "  's' キーでリポジトリを選択")
		return nil
	}

	v.Title = fmt.Sprintf(" %s ", a.SourcePane.Dir)
	a.renderEntries(v, a.SourcePane)
	a.syncScroll(v, a.SourcePane)
	return nil
}

// renderRightPane は右ペインの内容を描画する。
func (a *App) renderRightPane(g *gocui.Gui) error {
	v, err := g.View(viewRight)
	if err != nil {
		return err
	}
	v.Clear()

	v.Title = fmt.Sprintf(" %s ", a.DestPane.Dir)
	a.renderEntries(v, a.DestPane)
	a.syncScroll(v, a.DestPane)
	return nil
}

// renderEntries はペインのエントリ一覧をViewに描画する。
// 選択済みエントリには "* " プレフィックス、ディレクトリには "/" サフィックスを付与する。
func (a *App) renderEntries(v *gocui.View, p *pane.Pane) {
	for i, entry := range p.Entries {
		prefix := "  "
		if p.Selected[i] {
			prefix = "* "
		}

		suffix := ""
		if entry.IsDir {
			suffix = "/"
		}

		fmt.Fprintf(v, "%s%s%s\n", prefix, entry.Name, suffix)
	}
}

// syncScroll はViewのカーソル位置とスクロール位置をPaneのカーソルに同期させる。
func (a *App) syncScroll(v *gocui.View, p *pane.Pane) {
	_, viewHeight := v.Size()
	if viewHeight <= 0 {
		return
	}

	// スクロールオフセットを計算
	oy := 0
	cy := p.Cursor
	if p.Cursor >= viewHeight {
		oy = p.Cursor - viewHeight + 1
		cy = viewHeight - 1
	}

	v.SetOrigin(0, oy)
	v.SetCursor(0, cy)
}

// renderStatus はステータスバーを描画する。
func (a *App) renderStatus(g *gocui.Gui) error {
	v, err := g.View(viewStatus)
	if err != nil {
		return err
	}
	v.Clear()

	if a.Status != "" {
		fmt.Fprint(v, a.Status)
	} else {
		focus := "right"
		if a.FocusLeft {
			focus = "left"
		}
		fmt.Fprintf(v, " [%s] q:終了 Tab:切替 h/j/k/l:移動", focus)
	}

	return nil
}
