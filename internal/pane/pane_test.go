package pane

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/myuron/graftx/internal/fs"
)

// mockFS はテスト用のFileSystemモック。
type mockFS struct {
	dirs map[string][]fs.Entry
}

func (m *mockFS) ReadDir(path string) ([]fs.Entry, error) {
	entries, ok := m.dirs[path]
	if !ok {
		return nil, fmt.Errorf("directory not found: %s", path)
	}
	return entries, nil
}

func (m *mockFS) Copy(src, dst string) error           { return nil }
func (m *mockFS) Remove(path string) error             { return nil }
func (m *mockFS) Trash(path string) error              { return nil }
func (m *mockFS) Rename(old, new string) error         { return nil }
func (m *mockFS) Create(path string, isDir bool) error { return nil }

// テスト用のモックFSを作成するヘルパー。
func newTestFS() *mockFS {
	return &mockFS{
		dirs: map[string][]fs.Entry{
			"/project": {
				{Name: "docs", IsDir: true},
				{Name: "src", IsDir: true},
				{Name: "README.md", IsDir: false},
				{Name: "go.mod", IsDir: false},
			},
			"/project/src": {
				{Name: "main.go", IsDir: false},
				{Name: "util.go", IsDir: false},
			},
			"/project/docs": {
				{Name: "design.md", IsDir: false},
			},
			"/": {
				{Name: "project", IsDir: true},
			},
		},
	}
}

func TestNewPane_初期化(t *testing.T) {
	mock := newTestFS()
	p, err := NewPane("/project", mock)
	if err != nil {
		t.Fatalf("NewPane() error = %v", err)
	}

	if p.Dir != "/project" {
		t.Errorf("Dir = %q, want %q", p.Dir, "/project")
	}
	if len(p.Entries) != 4 {
		t.Errorf("Entries count = %d, want 4", len(p.Entries))
	}
	if p.Cursor != 0 {
		t.Errorf("Cursor = %d, want 0", p.Cursor)
	}
}

func TestMoveDown_カーソル下移動(t *testing.T) {
	mock := newTestFS()
	p, _ := NewPane("/project", mock)

	p.MoveDown()
	if p.Cursor != 1 {
		t.Errorf("MoveDown後 Cursor = %d, want 1", p.Cursor)
	}

	p.MoveDown()
	p.MoveDown()
	if p.Cursor != 3 {
		t.Errorf("3回MoveDown後 Cursor = %d, want 3", p.Cursor)
	}

	// 末尾でクランプされるか
	p.MoveDown()
	if p.Cursor != 3 {
		t.Errorf("末尾超過後 Cursor = %d, want 3", p.Cursor)
	}
}

func TestMoveUp_カーソル上移動(t *testing.T) {
	mock := newTestFS()
	p, _ := NewPane("/project", mock)

	// 先頭でクランプされるか
	p.MoveUp()
	if p.Cursor != 0 {
		t.Errorf("先頭でMoveUp後 Cursor = %d, want 0", p.Cursor)
	}

	p.Cursor = 2
	p.MoveUp()
	if p.Cursor != 1 {
		t.Errorf("MoveUp後 Cursor = %d, want 1", p.Cursor)
	}
}

func TestEnterDir_ディレクトリに入る(t *testing.T) {
	mock := newTestFS()
	p, _ := NewPane("/project", mock)

	// カーソルを"src"ディレクトリに合わせる（index 1）
	p.Cursor = 1
	if err := p.EnterDir(); err != nil {
		t.Fatalf("EnterDir() error = %v", err)
	}

	if p.Dir != "/project/src" {
		t.Errorf("Dir = %q, want %q", p.Dir, "/project/src")
	}
	if p.Cursor != 0 {
		t.Errorf("Cursor = %d, want 0", p.Cursor)
	}
	if len(p.Entries) != 2 {
		t.Errorf("Entries count = %d, want 2", len(p.Entries))
	}
}

func TestEnterDir_ファイルの場合何もしない(t *testing.T) {
	mock := newTestFS()
	p, _ := NewPane("/project", mock)

	// カーソルを"README.md"（ファイル）に合わせる（index 2）
	p.Cursor = 2
	if err := p.EnterDir(); err != nil {
		t.Fatalf("EnterDir() error = %v", err)
	}

	if p.Dir != "/project" {
		t.Errorf("Dir = %q, want %q", p.Dir, "/project")
	}
}

func TestParentDir_親ディレクトリに遷移(t *testing.T) {
	mock := newTestFS()
	p, _ := NewPane("/project/src", mock)

	if err := p.ParentDir(); err != nil {
		t.Fatalf("ParentDir() error = %v", err)
	}

	if p.Dir != "/project" {
		t.Errorf("Dir = %q, want %q", p.Dir, "/project")
	}

	// 直前にいた"src"にカーソルが復元されているか
	if p.Entries[p.Cursor].Name != "src" {
		t.Errorf("カーソル位置のエントリ = %q, want %q", p.Entries[p.Cursor].Name, "src")
	}
}

func TestParentDir_ルートディレクトリでは何もしない(t *testing.T) {
	mock := newTestFS()
	p, _ := NewPane("/", mock)

	if err := p.ParentDir(); err != nil {
		t.Fatalf("ParentDir() error = %v", err)
	}

	if p.Dir != "/" {
		t.Errorf("Dir = %q, want %q", p.Dir, "/")
	}
}

func TestToggleSelect_選択トグル(t *testing.T) {
	mock := newTestFS()
	p, _ := NewPane("/project", mock)

	p.ToggleSelect()
	if !p.Selected[0] {
		t.Error("ToggleSelect後 Selected[0] = false, want true")
	}

	p.ToggleSelect()
	if p.Selected[0] {
		t.Error("再ToggleSelect後 Selected[0] = true, want false")
	}
}

func TestSelectAll_全選択(t *testing.T) {
	mock := newTestFS()
	p, _ := NewPane("/project", mock)

	p.SelectAll()
	for i := range p.Entries {
		if !p.Selected[i] {
			t.Errorf("SelectAll後 Selected[%d] = false, want true", i)
		}
	}
}

func TestInvertSelection_選択反転(t *testing.T) {
	mock := newTestFS()
	p, _ := NewPane("/project", mock)

	p.Selected[0] = true
	p.Selected[2] = true

	p.InvertSelection()

	if p.Selected[0] {
		t.Error("InvertSelection後 Selected[0] = true, want false")
	}
	if !p.Selected[1] {
		t.Error("InvertSelection後 Selected[1] = false, want true")
	}
	if p.Selected[2] {
		t.Error("InvertSelection後 Selected[2] = true, want false")
	}
	if !p.Selected[3] {
		t.Error("InvertSelection後 Selected[3] = false, want true")
	}
}

func TestClearSelection_選択全解除(t *testing.T) {
	mock := newTestFS()
	p, _ := NewPane("/project", mock)

	p.Selected[0] = true
	p.Selected[1] = true

	p.ClearSelection()

	if len(p.Selected) != 0 {
		t.Errorf("ClearSelection後 Selected count = %d, want 0", len(p.Selected))
	}
}

func TestYank_選択ありの場合(t *testing.T) {
	mock := newTestFS()
	p, _ := NewPane("/project", mock)

	p.Selected[0] = true // docs
	p.Selected[2] = true // README.md

	yb := p.Yank()

	if yb.SrcDir != "/project" {
		t.Errorf("SrcDir = %q, want %q", yb.SrcDir, "/project")
	}
	if len(yb.Entries) != 2 {
		t.Fatalf("Entries count = %d, want 2", len(yb.Entries))
	}

	expected := []string{
		filepath.Join("/project", "docs"),
		filepath.Join("/project", "README.md"),
	}
	for i, e := range yb.Entries {
		if e != expected[i] {
			t.Errorf("Entries[%d] = %q, want %q", i, e, expected[i])
		}
	}
}

func TestYank_選択なしの場合カーソル行(t *testing.T) {
	mock := newTestFS()
	p, _ := NewPane("/project", mock)

	p.Cursor = 1 // src

	yb := p.Yank()

	if len(yb.Entries) != 1 {
		t.Fatalf("Entries count = %d, want 1", len(yb.Entries))
	}
	expected := filepath.Join("/project", "src")
	if yb.Entries[0] != expected {
		t.Errorf("Entries[0] = %q, want %q", yb.Entries[0], expected)
	}
}

func TestEnterDir_選択がクリアされる(t *testing.T) {
	mock := newTestFS()
	p, _ := NewPane("/project", mock)

	p.Selected[0] = true
	p.Selected[1] = true

	p.Cursor = 1 // src
	if err := p.EnterDir(); err != nil {
		t.Fatalf("EnterDir() error = %v", err)
	}

	if len(p.Selected) != 0 {
		t.Errorf("EnterDir後 Selected count = %d, want 0", len(p.Selected))
	}
}

func TestChangeDir_ディレクトリ変更(t *testing.T) {
	mock := newTestFS()
	p, _ := NewPane("/project", mock)

	if err := p.ChangeDir("/project/docs"); err != nil {
		t.Fatalf("ChangeDir() error = %v", err)
	}

	if p.Dir != "/project/docs" {
		t.Errorf("Dir = %q, want %q", p.Dir, "/project/docs")
	}
	if p.Cursor != 0 {
		t.Errorf("Cursor = %d, want 0", p.Cursor)
	}
	if len(p.Entries) != 1 {
		t.Errorf("Entries count = %d, want 1", len(p.Entries))
	}
}
