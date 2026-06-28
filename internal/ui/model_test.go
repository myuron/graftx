package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/myuron/graftx/internal/fs"
	"github.com/myuron/graftx/internal/pane"
)

// keyRunes はルーン入力のキーメッセージを生成する。
func keyRunes(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

// stubFS は削除系テスト用の最小のFileSystem実装。
type stubFS struct {
	entries []fs.Entry
	trashed []string
	removed []string
}

func (s *stubFS) ReadDir(string) ([]fs.Entry, error) { return s.entries, nil }
func (s *stubFS) Copy(string, string) error          { return nil }
func (s *stubFS) Remove(p string) error              { s.removed = append(s.removed, p); return nil }
func (s *stubFS) Trash(p string) error               { s.trashed = append(s.trashed, p); return nil }
func (s *stubFS) Rename(string, string) error        { return nil }
func (s *stubFS) Create(string, bool) error          { return nil }

func newDestApp() *App {
	app, _ := NewApp("/dst", &stubFS{}, nil)
	app.DestPane = &pane.Pane{
		Dir:      "/dst",
		Entries:  []fs.Entry{{Name: "apple"}, {Name: "banana"}, {Name: "cherry"}},
		Cursor:   0,
		Selected: map[int]bool{},
	}
	app.width = 80
	app.height = 24
	return app
}

func TestUpdate_WindowSizeで幅高さを保存(t *testing.T) {
	app, _ := NewApp("/dst", &stubFS{}, nil)
	m, _ := app.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	got := m.(*App)
	if got.width != 100 || got.height != 40 {
		t.Errorf("width/height = %d/%d, want 100/40", got.width, got.height)
	}
}

func TestUpdate_jでカーソル下移動(t *testing.T) {
	app := newDestApp()
	app.Update(keyRunes("j"))
	if app.DestPane.Cursor != 1 {
		t.Errorf("Cursor = %d, want 1", app.DestPane.Cursor)
	}
}

func TestUpdate_スペースでマークのみトグルしカーソルは移動しない(t *testing.T) {
	app := newDestApp()
	if app.DestPane.Cursor != 0 {
		t.Fatalf("初期Cursor = %d, want 0", app.DestPane.Cursor)
	}

	// スペースでカーソル行をトグル選択
	app.Update(tea.KeyMsg{Type: tea.KeySpace})

	if !app.DestPane.Selected[0] {
		t.Error("スペースでカーソル行が選択されるべき")
	}
	if app.DestPane.Cursor != 0 {
		t.Errorf("スペース後のCursor = %d, want 0（移動しないべき）", app.DestPane.Cursor)
	}

	// 再度スペースで選択解除（カーソルは依然移動しない）
	app.Update(tea.KeyMsg{Type: tea.KeySpace})
	if app.DestPane.Selected[0] {
		t.Error("再度のスペースで選択解除されるべき")
	}
	if app.DestPane.Cursor != 0 {
		t.Errorf("2回目スペース後のCursor = %d, want 0", app.DestPane.Cursor)
	}
}

func TestUpdate_Tabでフォーカス切替(t *testing.T) {
	app := newDestApp()
	if app.FocusLeft {
		t.Fatal("初期状態は右フォーカスのはず")
	}
	app.Update(tea.KeyMsg{Type: tea.KeyTab})
	if !app.FocusLeft {
		t.Error("Tab後は左フォーカスになるべき")
	}
}

func TestUpdate_qで終了コマンド(t *testing.T) {
	app := newDestApp()
	_, cmd := app.Update(keyRunes("q"))
	if app.ExitReason != ExitReasonQuit {
		t.Errorf("ExitReason = %d, want %d", app.ExitReason, ExitReasonQuit)
	}
	if cmd == nil {
		t.Fatal("qで終了コマンドが返るべき")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("返されたコマンドはtea.Quitであるべき")
	}
}

func TestUpdate_検索モード遷移と確定(t *testing.T) {
	app := newDestApp()

	// '/' で検索モードに入る
	app.Update(keyRunes("/"))
	if app.inputMode != InputModeSearch {
		t.Fatalf("inputMode = %d, want InputModeSearch", app.inputMode)
	}

	// "ban" と入力
	for _, ch := range []string{"b", "a", "n"} {
		app.Update(keyRunes(ch))
	}

	// Enterで確定 → banana(idx1)にジャンプ
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.inputMode != InputModeNone {
		t.Errorf("確定後のinputMode = %d, want InputModeNone", app.inputMode)
	}
	if app.searchQuery != "ban" {
		t.Errorf("searchQuery = %q, want \"ban\"", app.searchQuery)
	}
	if app.DestPane.Cursor != 1 {
		t.Errorf("Cursor = %d, want 1 (banana)", app.DestPane.Cursor)
	}
}

func TestUpdate_削除確認_yで実行(t *testing.T) {
	stub := &stubFS{entries: []fs.Entry{}}
	app, _ := NewApp("/dst", stub, nil)
	app.DestPane = &pane.Pane{
		Dir:      "/dst",
		Entries:  []fs.Entry{{Name: "apple"}},
		Cursor:   0,
		Selected: map[int]bool{},
		FS:       stub,
	}
	app.width, app.height = 80, 24

	// 'd' でゴミ箱確認モード
	app.Update(keyRunes("d"))
	if app.inputMode != InputModeConfirmDelete {
		t.Fatalf("inputMode = %d, want InputModeConfirmDelete", app.inputMode)
	}

	// 'y' で実行
	app.Update(keyRunes("y"))
	if app.inputMode != InputModeNone {
		t.Errorf("実行後のinputMode = %d, want InputModeNone", app.inputMode)
	}
	if len(stub.trashed) != 1 || stub.trashed[0] != "/dst/apple" {
		t.Errorf("trashed = %v, want [/dst/apple]", stub.trashed)
	}
}

func TestUpdate_削除確認_nでキャンセル(t *testing.T) {
	app := newDestApp()
	app.Update(keyRunes("d"))
	app.Update(keyRunes("n"))
	if app.inputMode != InputModeNone {
		t.Errorf("キャンセル後のinputMode = %d, want InputModeNone", app.inputMode)
	}
	if app.Status != "cancelled" {
		t.Errorf("Status = %q, want \"cancelled\"", app.Status)
	}
}

func TestView_サイズ未設定なら空文字(t *testing.T) {
	app, _ := NewApp("/dst", &stubFS{}, nil)
	if app.View() != "" {
		t.Error("width/height未設定ではViewは空文字を返すべき")
	}
}

func TestView_基本描画でエントリとステータスを含む(t *testing.T) {
	app := newDestApp()
	out := app.View()
	if out == "" {
		t.Fatal("サイズ設定済みではViewは空でないべき")
	}
	// エントリ名・ディレクトリ名・デフォルトステータスが描画されている
	for _, want := range []string{"apple", "banana", "/dst", "[q]Quit"} {
		if !strings.Contains(out, want) {
			t.Errorf("View出力に %q が含まれていない", want)
		}
	}
}

func TestView_セレクタモーダル描画(t *testing.T) {
	app := newDestApp()
	app.inputMode = InputModeSelectRepo
	app.filteredRepoList = []string{"/repo/alpha", "/repo/beta"}
	out := app.View()
	for _, want := range []string{"Filter", "/repo/alpha", "/repo/beta"} {
		if !strings.Contains(out, want) {
			t.Errorf("セレクタモーダル出力に %q が含まれていない", want)
		}
	}
}
