// Package fs はファイルシステム操作を抽象化するパッケージ。
// Entry構造体とFileSystemインターフェースを提供する。
package fs

// Entry はファイル/ディレクトリ1つを表す構造体。
type Entry struct {
	Name  string // ファイル名（ディレクトリ名）
	IsDir bool   // ディレクトリならtrue
}

// FileSystem はファイルシステム操作を抽象化するインターフェース。
// テスト時にモックを注入できる。
type FileSystem interface {
	// ReadDir はディレクトリのエントリ一覧を返す。
	// ディレクトリを先、ファイルを後にソートし、それぞれアルファベット順。
	ReadDir(path string) ([]Entry, error)
	// Copy はsrcをdstにコピーする。ディレクトリの場合は再帰的にコピーする。
	Copy(src, dst string) error
	// Remove はファイル/ディレクトリを完全削除する。
	Remove(path string) error
	// Trash はファイル/ディレクトリをOS標準のゴミ箱に移動する。
	Trash(path string) error
	// Rename はファイル/ディレクトリをリネームする。
	Rename(oldPath, newPath string) error
	// Create はファイルまたはディレクトリを作成する。isDirがtrueならディレクトリを作成する。
	Create(path string, isDir bool) error
}
