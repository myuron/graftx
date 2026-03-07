package fs

import (
	"os"
	"path/filepath"
	"testing"
)

// テスト用のディレクトリ構造を作成するヘルパー。
func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// ファイルを作成
	for _, name := range []string{"banana.txt", "apple.txt", "cherry.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// ディレクトリを作成
	for _, name := range []string{"docs", "src", "assets"} {
		if err := os.Mkdir(filepath.Join(dir, name), 0755); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

func TestReadDir_ソート順_ディレクトリ先_アルファベット順(t *testing.T) {
	dir := setupTestDir(t)
	osFS := &OSFS{}

	entries, err := osFS.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	// 期待する順序: ディレクトリ（アルファベット順）→ ファイル（アルファベット順）
	expected := []Entry{
		{Name: "assets", IsDir: true},
		{Name: "docs", IsDir: true},
		{Name: "src", IsDir: true},
		{Name: "apple.txt", IsDir: false},
		{Name: "banana.txt", IsDir: false},
		{Name: "cherry.txt", IsDir: false},
	}

	if len(entries) != len(expected) {
		t.Fatalf("ReadDir() returned %d entries, want %d", len(entries), len(expected))
	}

	for i, e := range entries {
		if e.Name != expected[i].Name || e.IsDir != expected[i].IsDir {
			t.Errorf("entries[%d] = %+v, want %+v", i, e, expected[i])
		}
	}
}

func TestReadDir_空ディレクトリ(t *testing.T) {
	dir := t.TempDir()
	osFS := &OSFS{}

	entries, err := osFS.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("ReadDir() returned %d entries, want 0", len(entries))
	}
}

func TestReadDir_存在しないディレクトリ(t *testing.T) {
	osFS := &OSFS{}

	_, err := osFS.ReadDir("/nonexistent/path")
	if err == nil {
		t.Error("ReadDir() error = nil, want error")
	}
}

func TestCopy_ファイルコピー(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	content := []byte("hello world")

	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatal(err)
	}

	osFS := &OSFS{}
	if err := osFS.Copy(src, dst); err != nil {
		t.Fatalf("Copy() error = %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("コピー先ファイル読み込みエラー: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("コピー先の内容 = %q, want %q", string(got), string(content))
	}
}

func TestCopy_ディレクトリ再帰コピー(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "srcdir")
	dstDir := filepath.Join(dir, "dstdir")

	// ネストしたディレクトリ構造を作成
	if err := os.MkdirAll(filepath.Join(srcDir, "sub"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("a"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "sub", "b.txt"), []byte("b"), 0644); err != nil {
		t.Fatal(err)
	}

	osFS := &OSFS{}
	if err := osFS.Copy(srcDir, dstDir); err != nil {
		t.Fatalf("Copy() error = %v", err)
	}

	// コピー先のファイルを検証
	got, err := os.ReadFile(filepath.Join(dstDir, "a.txt"))
	if err != nil {
		t.Fatalf("a.txt読み込みエラー: %v", err)
	}
	if string(got) != "a" {
		t.Errorf("a.txt = %q, want %q", string(got), "a")
	}

	got, err = os.ReadFile(filepath.Join(dstDir, "sub", "b.txt"))
	if err != nil {
		t.Fatalf("sub/b.txt読み込みエラー: %v", err)
	}
	if string(got) != "b" {
		t.Errorf("sub/b.txt = %q, want %q", string(got), "b")
	}
}

func TestRemove_ファイル削除(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	osFS := &OSFS{}
	if err := osFS.Remove(path); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("ファイルが削除されていない")
	}
}

func TestRemove_ディレクトリ削除(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "subdir")
	if err := os.MkdirAll(filepath.Join(target, "nested"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "nested", "f.txt"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	osFS := &OSFS{}
	if err := osFS.Remove(target); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Error("ディレクトリが削除されていない")
	}
}

func TestTrash_ゴミ箱移動(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "trash_me.txt")
	if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	osFS := &OSFS{}
	if err := osFS.Trash(path); err != nil {
		t.Fatalf("Trash() error = %v", err)
	}

	// 元の場所からは消えていることを検証
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("ファイルが元の場所から消えていない")
	}

	// ゴミ箱にファイルが存在することを検証
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	trashPath := filepath.Join(home, ".Trash", "trash_me.txt")
	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		t.Error("ファイルがゴミ箱に移動されていない")
	}

	// テスト後のクリーンアップ: ゴミ箱からファイルを削除
	os.Remove(trashPath)
}

func TestRename_リネーム(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "old.txt")
	newPath := filepath.Join(dir, "new.txt")
	if err := os.WriteFile(oldPath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	osFS := &OSFS{}
	if err := osFS.Rename(oldPath, newPath); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}

	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Error("リネーム前のファイルが残っている")
	}

	got, err := os.ReadFile(newPath)
	if err != nil {
		t.Fatalf("リネーム先の読み込みエラー: %v", err)
	}
	if string(got) != "test" {
		t.Errorf("リネーム先の内容 = %q, want %q", string(got), "test")
	}
}

func TestCreate_ファイル作成(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "newfile.txt")

	osFS := &OSFS{}
	if err := osFS.Create(path, false); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("作成したファイルのStat失敗: %v", err)
	}
	if info.IsDir() {
		t.Error("ファイルとして作成したのにディレクトリになっている")
	}
}

func TestCreate_ディレクトリ作成(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "newdir")

	osFS := &OSFS{}
	if err := osFS.Create(path, true); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("作成したディレクトリのStat失敗: %v", err)
	}
	if !info.IsDir() {
		t.Error("ディレクトリとして作成したのにファイルになっている")
	}
}
