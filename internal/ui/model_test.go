package ui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/myuron/graftx/internal/fs"
	"github.com/myuron/graftx/internal/pane"
)

// errStub はテスト用のダミーエラー。
var errStub = errors.New("stub error")

// keyRunes はルーン入力のキーメッセージを生成する。
func keyRunes(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

// stubFS は削除・コピー系テスト用の最小のFileSystem実装。
type stubFS struct {
	entries []fs.Entry
	trashed []string
	removed []string
	copies  [][2]string       // [src, dst] の記録
	files   map[string][]byte // パスごとのファイル内容（プレビュー用）
}

func (s *stubFS) ReadDir(string) ([]fs.Entry, error) { return s.entries, nil }
func (s *stubFS) ReadFile(p string) ([]byte, error) {
	if data, ok := s.files[p]; ok {
		return data, nil
	}
	return nil, errStub
}
func (s *stubFS) Copy(src, dst string) error {
	s.copies = append(s.copies, [2]string{src, dst})
	return nil
}
func (s *stubFS) Remove(p string) error       { s.removed = append(s.removed, p); return nil }
func (s *stubFS) Trash(p string) error        { s.trashed = append(s.trashed, p); return nil }
func (s *stubFS) Rename(string, string) error { return nil }
func (s *stubFS) Create(string, bool) error   { return nil }

// stubRunner はリポジトリ選択テスト用のCommandRunner実装。
type stubRunner struct {
	repos []string
	err   error
}

func (r *stubRunner) SelectRepository() (string, error)   { return "", nil }
func (r *stubRunner) ListRepositories() ([]string, error) { return r.repos, r.err }

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

func TestUpdate_Tabで左右ペインを切り替え(t *testing.T) {
	app := newDestApp()
	app.SourcePane = &pane.Pane{
		Dir:      "/src",
		Entries:  []fs.Entry{{Name: "x"}},
		Cursor:   0,
		Selected: map[int]bool{},
	}
	if app.FocusLeft {
		t.Fatal("初期状態は右（Dest）フォーカスのはず")
	}
	// Destペイン → Sourceペイン（横移動）
	app.Update(tea.KeyMsg{Type: tea.KeyTab})
	if !app.FocusLeft {
		t.Errorf("Tabで左ペインになるべき: FocusLeft=%v", app.FocusLeft)
	}
	// Sourceペイン → Destペイン
	app.Update(tea.KeyMsg{Type: tea.KeyTab})
	if app.FocusLeft {
		t.Errorf("Tabで右ペインに戻るべき: FocusLeft=%v", app.FocusLeft)
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

func TestUpdate_ヘルプ_クエスチョンで表示(t *testing.T) {
	app := newDestApp()

	// '?' でヘルプを表示する
	app.Update(keyRunes("?"))
	if !app.showHelp {
		t.Fatal("? 押下後はshowHelpがtrueであるべき")
	}
}

func TestUpdate_ヘルプ_任意キーで閉じる(t *testing.T) {
	app := newDestApp()
	app.showHelp = true

	// ヘルプ表示中に任意キーで閉じる
	app.Update(keyRunes("j"))
	if app.showHelp {
		t.Error("任意キー押下後はshowHelpがfalseであるべき")
	}
}

func TestUpdate_ヘルプ表示中はカーソルが動かない(t *testing.T) {
	app := newDestApp()
	app.DestPane.Cursor = 0
	app.showHelp = true

	// ヘルプを閉じるだけでカーソルは移動しない
	app.Update(keyRunes("j"))
	if app.DestPane.Cursor != 0 {
		t.Errorf("Cursor = %d, want 0 (ヘルプを閉じるだけで移動しない)", app.DestPane.Cursor)
	}
}

func TestRender_ヘルプ表示時はキー一覧を含む(t *testing.T) {
	app := newDestApp()
	app.width, app.height = 80, 24
	app.showHelp = true

	out := app.render()
	if !strings.Contains(out, "Keybindings") {
		t.Errorf("ヘルプにタイトル Keybindings が含まれない:\n%s", out)
	}
	if !strings.Contains(out, "Quit") {
		t.Errorf("ヘルプに q の説明が含まれない:\n%s", out)
	}
}

func TestPasteOverwrite_同一パスへのコピーを防ぐ(t *testing.T) {
	stub := &stubFS{}
	app, _ := NewApp("/dst", stub, nil)
	app.DestPane = &pane.Pane{Dir: "/dst", Entries: []fs.Entry{{Name: "apple"}}, Selected: map[int]bool{}, FS: stub}
	app.width, app.height = 80, 24
	// ヤンク元と貼り付け先が同一ディレクトリの同名ファイル
	app.YankBuf = &pane.YankBuffer{Entries: []string{"/dst/apple"}, SrcDir: "/dst"}

	app.pasteOverwrite()

	if len(stub.copies) != 0 {
		t.Errorf("同一パスではCopyを呼ばないべき: %v", stub.copies)
	}
	if app.Status != "copy error: source and destination are the same" {
		t.Errorf("Status = %q", app.Status)
	}
}

func TestUpdate_sでセレクタを開き非同期で一覧反映(t *testing.T) {
	app := newDestApp()
	app.Selector = &stubRunner{repos: []string{"/r/a", "/r/b"}}

	// 's' でセレクタを開く（この時点では一覧はまだ空・ローディング表示）
	app.Update(keyRunes("s"))
	if app.inputMode != InputModeSelectRepo {
		t.Fatalf("inputMode = %d, want InputModeSelectRepo", app.inputMode)
	}
	if len(app.repoList) != 0 {
		t.Errorf("開いた直後はrepoListは空のはず: %v", app.repoList)
	}

	// 非同期取得結果が届くと一覧へ反映される
	app.Update(repoListMsg{repos: []string{"/r/a", "/r/b"}})
	if len(app.filteredRepoList) != 2 {
		t.Errorf("filteredRepoList len = %d, want 2", len(app.filteredRepoList))
	}
}

func TestUpdate_セレクタ取得エラーで閉じる(t *testing.T) {
	app := newDestApp()
	app.inputMode = InputModeSelectRepo
	app.Update(repoListMsg{err: errStub})
	if app.inputMode != InputModeNone {
		t.Errorf("エラー時はセレクタを閉じるべき: inputMode = %d", app.inputMode)
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
