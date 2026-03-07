// Package selector はリポジトリ選択のためのコマンド実行を行うパッケージ。
package selector

import (
	"os"
	"os/exec"
	"strings"
)

// CommandRunner はリポジトリ選択コマンドを実行するインターフェース。
type CommandRunner interface {
	// SelectRepository はリポジトリ選択UIを表示し、選択されたパスを返す。
	// キャンセル時は ("", nil) を返す。
	SelectRepository() (string, error)
	// ListRepositories はghqで管理されているリポジトリの一覧を返す。
	ListRepositories() ([]string, error)
}

// DefaultRunner は ghq + fzf を使ったCommandRunnerの実装。
type DefaultRunner struct{}

// SelectRepository は ghq list --full-path | fzf を実行し、選択されたリポジトリのパスを返す。
// fzfがキャンセルされた場合は ("", nil) を返す。
func (r *DefaultRunner) SelectRepository() (string, error) {
	cmd := exec.Command("bash", "-c", "ghq list --full-path | fzf")
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		// fzfキャンセル時（ExitError）は空文字列を返す
		if _, ok := err.(*exec.ExitError); ok {
			return "", nil
		}
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

// ListRepositories は ghq list --full-path を実行し、リポジトリの一覧を返す。
func (r *DefaultRunner) ListRepositories() ([]string, error) {
	cmd := exec.Command("ghq", "list", "--full-path")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	output := strings.TrimSpace(string(out))
	if output == "" {
		return []string{}, nil
	}
	return strings.Split(output, "\n"), nil
}
