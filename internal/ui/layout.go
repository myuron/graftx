// Package ui はBubble Teaを使ったTUIの描画とイベント処理を行うパッケージ。
package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/myuron/graftx/internal/fs"
	"github.com/myuron/graftx/internal/pane"
	"github.com/myuron/graftx/internal/selector"
)

// ExitReason はプログラム終了後の理由を表す型。
type ExitReason int

const (
	// ExitReasonNone は未設定。
	ExitReasonNone ExitReason = iota
	// ExitReasonQuit はユーザーによる終了。
	ExitReasonQuit
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
// Bubble Teaのtea.Modelインターフェースを実装する。
type App struct {
	SourcePane     *pane.Pane             // 左ペイン（コピー元）、未選択時はnil
	DestPane       *pane.Pane             // 右ペイン（コピー先）
	Selector       selector.CommandRunner // リポジトリ選択
	FS             fs.FileSystem          // ファイルシステム
	FocusLeft      bool                   // trueなら左ペインにフォーカス
	Status         string                 // ステータスバーに表示するメッセージ
	ExitReason     ExitReason             // プログラム終了理由
	YankBuf        *pane.YankBuffer       // ヤンクバッファ
	gPending       bool                   // 'g'キーが押された状態
	inputMode      InputMode              // 入力モード
	searchQuery    string                 // 現在の検索クエリ
	searchFwd      bool                   // 検索方向（trueなら前方）
	width          int                    // 端末の幅
	height         int                    // 端末の高さ
	input          textinput.Model        // 入力欄（検索/フィルタ/作成/リネーム）
	pendingTargets []string               // 削除確認時のターゲットパスのスナップショット

	repoList           []string        // リポジトリ一覧（全件）
	filteredRepoList   []string        // フィルタ済みリポジトリ一覧
	repoSelectorCursor int             // リポジトリ選択カーソル位置
	repoFilterQuery    string          // リポジトリフィルタクエリ
	filterInput        textinput.Model // セレクタフィルタ入力
}

// NewApp は新しいAppを作成する。
// destDirは右ペイン（コピー先）の初期ディレクトリ。
func NewApp(destDir string, fsys fs.FileSystem, sel selector.CommandRunner) (*App, error) {
	destPane, err := pane.NewPane(destDir, fsys)
	if err != nil {
		return nil, err
	}

	// 入力欄の初期化（プロンプトはステータスバー側で描画するため空にする）
	input := textinput.New()
	input.Prompt = ""
	filterInput := textinput.New()
	filterInput.Prompt = ""

	return &App{
		DestPane:    destPane,
		Selector:    sel,
		FS:          fsys,
		FocusLeft:   false,
		ExitReason:  ExitReasonNone,
		input:       input,
		filterInput: filterInput,
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
