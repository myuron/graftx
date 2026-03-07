// Package pane はペインの状態管理を行うパッケージ。
// gocuiに依存しない。
package pane

import (
	"path/filepath"
	"strings"

	"github.com/myuron/graftx/internal/fs"
)

// Pane は各ペインの状態を管理する構造体。
type Pane struct {
	Dir         string        // 現在のディレクトリの絶対パス
	Entries     []fs.Entry    // 現在のディレクトリのエントリ一覧（フィルタ後）
	Cursor      int           // カーソル位置（0始まり）
	Selected    map[int]bool  // 選択中のエントリのインデックス
	FS          fs.FileSystem // ファイルシステム操作のインターフェース
	ShowHidden  bool          // 隠しファイルを表示するか
	FilterQuery string        // フィルタクエリ（空なら全表示）
}

// YankBuffer はヤンクしたエントリを保持する構造体。
type YankBuffer struct {
	Entries []string // ヤンクしたエントリのフルパス一覧
	SrcDir  string   // ヤンク元のディレクトリパス
}

// NewPane は新しいPaneを作成する。
func NewPane(dir string, fsys fs.FileSystem) (*Pane, error) {
	p := &Pane{
		Dir:      dir,
		Selected: make(map[int]bool),
		FS:       fsys,
	}
	if err := p.Refresh(); err != nil {
		return nil, err
	}
	return p, nil
}

// Refresh は現在のディレクトリのエントリ一覧を再取得する。
// ShowHiddenがfalseの場合、ドットで始まるエントリを除外する。
// FilterQueryが設定されている場合、名前に含まれないエントリを除外する。
func (p *Pane) Refresh() error {
	entries, err := p.FS.ReadDir(p.Dir)
	if err != nil {
		return err
	}

	// フィルタリング
	filtered := make([]fs.Entry, 0, len(entries))
	for _, e := range entries {
		// 隠しファイルフィルタ
		if !p.ShowHidden && strings.HasPrefix(e.Name, ".") {
			continue
		}
		// クエリフィルタ
		if p.FilterQuery != "" && !strings.Contains(strings.ToLower(e.Name), strings.ToLower(p.FilterQuery)) {
			continue
		}
		filtered = append(filtered, e)
	}

	p.Entries = filtered
	p.Selected = make(map[int]bool)
	p.clampCursor()
	return nil
}

// ClearFilter はフィルタをクリアしてリフレッシュする。
func (p *Pane) ClearFilter() error {
	p.FilterQuery = ""
	return p.Refresh()
}

// MoveDown はカーソルを1つ下に移動する。
func (p *Pane) MoveDown() {
	if p.Cursor < len(p.Entries)-1 {
		p.Cursor++
	}
}

// MoveUp はカーソルを1つ上に移動する。
func (p *Pane) MoveUp() {
	if p.Cursor > 0 {
		p.Cursor--
	}
}

// MoveToTop はカーソルを先頭に移動する。
func (p *Pane) MoveToTop() {
	p.Cursor = 0
}

// MoveToBottom はカーソルを末尾に移動する。
func (p *Pane) MoveToBottom() {
	if len(p.Entries) == 0 {
		p.Cursor = 0
		return
	}
	p.Cursor = len(p.Entries) - 1
}

// EnterDir はカーソル行がディレクトリの場合、そのディレクトリに入る。
// ファイルの場合は何もしない。
func (p *Pane) EnterDir() error {
	if len(p.Entries) == 0 {
		return nil
	}
	entry := p.Entries[p.Cursor]
	if !entry.IsDir {
		return nil
	}

	p.Dir = filepath.Join(p.Dir, entry.Name)
	p.Cursor = 0
	p.Selected = make(map[int]bool)
	return p.Refresh()
}

// ParentDir は親ディレクトリに遷移する。
// 直前にいたディレクトリにカーソルを復元する。
func (p *Pane) ParentDir() error {
	parent := filepath.Dir(p.Dir)
	if parent == p.Dir {
		// ルートディレクトリの場合は何もしない
		return nil
	}

	prevName := filepath.Base(p.Dir)
	p.Dir = parent
	p.Cursor = 0
	p.Selected = make(map[int]bool)
	if err := p.Refresh(); err != nil {
		return err
	}

	// 直前にいたディレクトリにカーソルを復元
	for i, e := range p.Entries {
		if e.Name == prevName {
			p.Cursor = i
			break
		}
	}
	return nil
}

// ChangeDir は指定されたディレクトリに遷移する。
func (p *Pane) ChangeDir(dir string) error {
	p.Dir = dir
	p.Cursor = 0
	p.Selected = make(map[int]bool)
	return p.Refresh()
}

// ToggleSelect はカーソル行のエントリの選択をトグルする。
func (p *Pane) ToggleSelect() {
	if len(p.Entries) == 0 {
		return
	}
	if p.Selected[p.Cursor] {
		delete(p.Selected, p.Cursor)
	} else {
		p.Selected[p.Cursor] = true
	}
}

// SelectAll は全エントリを選択する。
func (p *Pane) SelectAll() {
	for i := range p.Entries {
		p.Selected[i] = true
	}
}

// InvertSelection は選択状態を反転する。
func (p *Pane) InvertSelection() {
	for i := range p.Entries {
		if p.Selected[i] {
			delete(p.Selected, i)
		} else {
			p.Selected[i] = true
		}
	}
}

// ClearSelection は選択を全解除する。
func (p *Pane) ClearSelection() {
	p.Selected = make(map[int]bool)
}

// Yank は選択中のエントリ（なければカーソル行）をYankBufferに格納する。
func (p *Pane) Yank() *YankBuffer {
	if len(p.Entries) == 0 {
		return &YankBuffer{}
	}

	yb := &YankBuffer{
		SrcDir: p.Dir,
	}

	if len(p.Selected) > 0 {
		for i := range p.Entries {
			if p.Selected[i] {
				yb.Entries = append(yb.Entries, filepath.Join(p.Dir, p.Entries[i].Name))
			}
		}
	} else {
		yb.Entries = []string{filepath.Join(p.Dir, p.Entries[p.Cursor].Name)}
	}

	return yb
}

// clampCursor はカーソルをエントリ数の範囲内にクランプする。
func (p *Pane) clampCursor() {
	if len(p.Entries) == 0 {
		p.Cursor = 0
		return
	}
	if p.Cursor >= len(p.Entries) {
		p.Cursor = len(p.Entries) - 1
	}
	if p.Cursor < 0 {
		p.Cursor = 0
	}
}
