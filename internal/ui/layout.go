// Package ui はgocuiを使ったTUIの描画とイベント処理を行うパッケージ。
package ui

import (
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/mattn/go-runewidth"
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
)

// View名の定数。
const (
	viewLeft           = "left"
	viewRight          = "right"
	viewStatus         = "status"
	viewInput          = "input"
	viewSelectorFilter = "selector_filter"
	viewSelectorList   = "selector_list"
)

// InputMode はステータスバーの入力モードを表す型。
type InputMode int

const (
	// InputModeNone は通常モード。
	InputModeNone InputMode = iota
	// InputModeSearch は検索モード（/）。
	InputModeSearch
	// InputModeSearchBackward は後方検索モード（?）。
	InputModeSearchBackward
	// InputModeFilter はフィルタモード（f）。
	InputModeFilter
	// InputModeCreate は新規作成モード（a）。
	InputModeCreate
	// InputModeRename はリネームモード（r）。
	InputModeRename
	// InputModeConfirmDelete は削除確認モード（d）。
	InputModeConfirmDelete
	// InputModeConfirmForceDelete は完全削除確認モード（D）。
	InputModeConfirmForceDelete
	// InputModeConfirmPaste はペースト同名確認モード。
	InputModeConfirmPaste
	// InputModeSelectRepo はリポジトリ選択ポップアップモード。
	InputModeSelectRepo
)

// App はTUIアプリケーション全体の状態を管理する構造体。
type App struct {
	SourcePane                *pane.Pane             // 左ペイン（コピー元）、未選択時はnil
	DestPane                  *pane.Pane             // 右ペイン（コピー先）
	Selector                  selector.CommandRunner // リポジトリ選択
	FS                        fs.FileSystem          // ファイルシステム
	FocusLeft                 bool                   // trueなら左ペインにフォーカス
	Status                    string                 // ステータスバーに表示するメッセージ
	ExitReason                ExitReason             // MainLoop終了理由
	YankBuf                   *pane.YankBuffer       // ヤンクバッファ
	gPending                  bool                   // 'g'キーが押された状態
	inputMode                 InputMode              // 入力モード
	inputBuf                  string                 // 入力バッファ
	searchQuery               string                 // 現在の検索クエリ
	searchFwd                 bool                   // 検索方向（trueなら前方）
	gui                       *gocui.Gui             // gocuiインスタンスへの参照
	pendingTargets            []string               // 削除確認時のターゲットパスのスナップショット
	inputKeybindingsInited    bool                   // 入力ビューのキーバインド登録済みフラグ
	repoList                  []string               // リポジトリ一覧（全件）
	filteredRepoList          []string               // フィルタ済みリポジトリ一覧
	repoSelectorCursor        int                    // リポジトリ選択カーソル位置
	repoFilterQuery           string                 // リポジトリフィルタクエリ
	selectorKeybindingsInited bool                   // セレクタキーバインド登録済みフラグ
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

// SetSourceDir はSourcePaneを指定ディレクトリで初期化または変更する。
func (a *App) SetSourceDir(dir string) error {
	if a.SourcePane == nil {
		srcPane, err := pane.NewPane(dir, a.FS)
		if err != nil {
			return err
		}
		a.SourcePane = srcPane
		return nil
	}
	return a.SourcePane.ChangeDir(dir)
}

// SetGui はgocuiのGuiインスタンスを設定し、Manager登録・GUI設定・キーバインド登録を行う。
func (a *App) SetGui(g *gocui.Gui) error {
	a.gui = g
	g.SetManager(a)
	g.Highlight = true
	g.SelFgColor = gocui.ColorGreen
	g.SelBgColor = gocui.ColorDefault
	g.Cursor = false
	return a.setKeybindings(g)
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

	// リポジトリ選択ポップアップモード
	if a.inputMode == InputModeSelectRepo {
		// ポップアップモード中はペインビューのcurrentViewを設定しない
		if err := a.renderSelectorPopup(g); err != nil {
			return err
		}
	} else {
		// ポップアップビューが残っていれば削除
		_ = g.DeleteView(viewSelectorFilter)
		_ = g.DeleteView(viewSelectorList)

		// 入力モード時は入力ビューを表示
		if a.isTextInputMode() {
			if v, err := g.SetView(viewInput, 0, maxY-2, maxX-1, maxY); err != nil {
				if err != gocui.ErrUnknownView {
					return err
				}
				v.Frame = false
				v.Editable = true
				v.Editor = gocui.EditorFunc(a.inputEditor)
			}
			if _, err := g.SetCurrentView(viewInput); err != nil {
				return err
			}
		} else {
			// 入力ビューが存在したら削除
			_ = g.DeleteView(viewInput)

			// フォーカス中のビューをcurrentViewに設定
			if a.FocusLeft {
				if _, err := g.SetCurrentView(viewLeft); err != nil {
					return err
				}
			} else {
				if _, err := g.SetCurrentView(viewRight); err != nil {
					return err
				}
			}
		}
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
	// カレントペインのみハイライトを有効にする
	v.Highlight = a.FocusLeft

	if a.SourcePane == nil {
		v.Title = " Source "
		w, h := v.Size()
		msg := "[s]select repository"
		msgWidth := runewidth.StringWidth(msg)
		padX := 0
		if w > msgWidth {
			padX = (w - msgWidth) / 2
		}
		padY := 0
		if h > 1 {
			padY = h / 2
		}
		for i := 0; i < padY; i++ {
			_, _ = fmt.Fprintln(v, "")
		}
		_, _ = fmt.Fprintf(v, "%s%s\n", strings.Repeat(" ", padX), msg)
		return nil
	}

	v.Title = fmt.Sprintf(" %s ", a.SourcePane.Dir)
	a.renderEntries(v, a.SourcePane)
	return a.syncScroll(v, a.SourcePane)
}

// renderRightPane は右ペインの内容を描画する。
func (a *App) renderRightPane(g *gocui.Gui) error {
	v, err := g.View(viewRight)
	if err != nil {
		return err
	}
	v.Clear()
	// カレントペインのみハイライトを有効にする
	v.Highlight = !a.FocusLeft

	v.Title = fmt.Sprintf(" %s ", a.DestPane.Dir)
	a.renderEntries(v, a.DestPane)
	return a.syncScroll(v, a.DestPane)
}

// highlightMatch は検索クエリにマッチする部分をANSIカラーでハイライトする。
// 大文字小文字を無視してマッチし、最初のマッチ部分を黄色背景＋黒文字で装飾する。
// クエリが空またはマッチなしの場合はそのまま返す。
func highlightMatch(name, query string) string {
	if query == "" {
		return name
	}
	lower := strings.ToLower(name)
	q := strings.ToLower(query)
	idx := strings.Index(lower, q)
	if idx < 0 {
		return name
	}
	// 元の文字列の大文字小文字を保持してハイライト
	return name[:idx] + "\x1b[30;43m" + name[idx:idx+len(query)] + "\x1b[0m" + name[idx+len(query):]
}

// renderEntries はペインのエントリ一覧をViewに描画する。
// 選択済みエントリには "* " プレフィックス、ディレクトリには "/" サフィックスを付与する。
// 各行はビュー幅いっぱいまでスペースで埋めてハイライトが横幅全体に表示されるようにする。
func (a *App) renderEntries(v *gocui.View, p *pane.Pane) {
	viewWidth, _ := v.Size()

	for i, entry := range p.Entries {
		prefix := "  "
		if p.Selected[i] {
			prefix = "* "
		}

		suffix := ""
		if entry.IsDir {
			suffix = "/"
		}

		icon := iconForEntry(entry) + " "

		// ANSIコードを含まない元テキストで幅を計算
		text := prefix + icon + entry.Name + suffix
		textWidth := runewidth.StringWidth(text)
		pad := 0
		if viewWidth > textWidth {
			pad = viewWidth - textWidth
		}

		// 検索クエリがある場合はマッチ部分をハイライト
		displayName := entry.Name
		if a.searchQuery != "" {
			displayName = highlightMatch(entry.Name, a.searchQuery)
		}
		displayText := prefix + icon + displayName + suffix

		_, _ = fmt.Fprintf(v, "%s%s\n", displayText, strings.Repeat(" ", pad))
	}
}

// syncScroll はViewのカーソル位置とスクロール位置をPaneのカーソルに同期させる。
func (a *App) syncScroll(v *gocui.View, p *pane.Pane) error {
	_, viewHeight := v.Size()
	if viewHeight <= 0 {
		return nil
	}

	// スクロールオフセットを計算
	oy := 0
	cy := p.Cursor
	if p.Cursor >= viewHeight {
		oy = p.Cursor - viewHeight + 1
		cy = viewHeight - 1
	}

	if err := v.SetOrigin(0, oy); err != nil {
		return err
	}
	return v.SetCursor(0, cy)
}

// inputPrefix は入力モードごとのプレフィックス文字列を返す。
func (a *App) inputPrefix() string {
	switch a.inputMode {
	case InputModeSearch:
		return "/:"
	case InputModeSearchBackward:
		return "?:"
	case InputModeFilter:
		return "filter: "
	case InputModeCreate:
		return "create new (end with / for dir): "
	case InputModeRename:
		return "rename: "
	}
	return ""
}

// isTextInputMode はテキスト入力を受け付けるモードかを返す。
func (a *App) isTextInputMode() bool {
	switch a.inputMode {
	case InputModeSearch, InputModeSearchBackward, InputModeFilter, InputModeCreate, InputModeRename:
		return true
	}
	return false
}

// renderStatus はステータスバーを描画する。
func (a *App) renderStatus(g *gocui.Gui) error {
	// テキスト入力モード時は入力ビューに描画
	if a.isTextInputMode() {
		v, err := g.View(viewInput)
		if err != nil {
			return nil // 入力ビューがまだ作られていない場合はスキップ
		}
		v.Clear()
		_, _ = fmt.Fprint(v, a.inputPrefix()+a.inputBuf)
		return nil
	}

	v, err := g.View(viewStatus)
	if err != nil {
		return err
	}
	v.Clear()

	if a.Status != "" {
		_, _ = fmt.Fprint(v, a.Status)
	} else {
		focus := "right"
		if a.FocusLeft {
			focus = "left"
		}
		_, _ = fmt.Fprintf(v, " [%s] [q]Quit [Tab]Switch [h/j/k/l]Move", focus)
	}

	return nil
}
