package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
)

// OSFS はFileSystemインターフェースのOS標準実装。
type OSFS struct{}

// ReadDir はディレクトリのエントリ一覧を返す。
// ディレクトリを先、ファイルを後にソートし、それぞれアルファベット順。
func (o *OSFS) ReadDir(path string) ([]Entry, error) {
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, 0, len(dirEntries))
	for _, de := range dirEntries {
		entries = append(entries, Entry{
			Name:  de.Name(),
			IsDir: de.IsDir(),
		})
	}

	// ディレクトリ先・アルファベット順ソート
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return entries[i].Name < entries[j].Name
	})

	return entries, nil
}

// Copy はsrcをdstにコピーする。ディレクトリの場合は再帰的にコピーする。
func (o *OSFS) Copy(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return o.copyDir(src, dst)
	}
	return o.copyFile(src, dst)
}

// copyFile はファイルをコピーする。
func (o *OSFS) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(dstFile, srcFile)
	closeErr := dstFile.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

// copyDir はディレクトリを再帰的にコピーする。
func (o *OSFS) copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	dirEntries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range dirEntries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := o.copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := o.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// Remove はファイル/ディレクトリを完全削除する。
func (o *OSFS) Remove(path string) error {
	return os.RemoveAll(path)
}

// Trash はファイル/ディレクトリをmacOSのゴミ箱（~/.Trash）に移動する。
// macOS専用。他のプラットフォームではエラーを返す。
func (o *OSFS) Trash(path string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("Trash is only supported on macOS")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	name := filepath.Base(path)
	trashPath := filepath.Join(home, ".Trash", name)

	// 同名ファイルが既にゴミ箱にある場合はサフィックスを付与
	if _, err := os.Stat(trashPath); err == nil {
		const maxAttempts = 1000
		found := false
		for i := 1; i <= maxAttempts; i++ {
			candidate := fmt.Sprintf("%s %d", trashPath, i)
			if _, err := os.Stat(candidate); os.IsNotExist(err) {
				trashPath = candidate
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("ゴミ箱内で一意のパスが見つかりません: %s", name)
		}
	}

	return os.Rename(path, trashPath)
}

// Rename はファイル/ディレクトリをリネームする。
func (o *OSFS) Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

// Create はファイルまたはディレクトリを作成する。isDirがtrueならディレクトリを作成する。
func (o *OSFS) Create(path string, isDir bool) error {
	if isDir {
		return os.MkdirAll(path, 0755)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	return f.Close()
}
