package selector

import (
	"fmt"
	"os/exec"
	"testing"
)

// mockRunner はテスト用のCommandRunnerモック。
type mockRunner struct {
	result   string
	err      error
	repoList []string
	listErr  error
}

func (m *mockRunner) SelectRepository() (string, error) {
	return m.result, m.err
}

func (m *mockRunner) ListRepositories() ([]string, error) {
	return m.repoList, m.listErr
}

func TestSelectRepository_正常パス(t *testing.T) {
	var runner CommandRunner = &mockRunner{result: "/home/user/repos/myproject", err: nil}

	result, err := runner.SelectRepository()
	if err != nil {
		t.Fatalf("SelectRepository() error = %v", err)
	}
	if result != "/home/user/repos/myproject" {
		t.Errorf("SelectRepository() = %q, want %q", result, "/home/user/repos/myproject")
	}
}

func TestSelectRepository_キャンセル(t *testing.T) {
	var runner CommandRunner = &mockRunner{result: "", err: nil}

	result, err := runner.SelectRepository()
	if err != nil {
		t.Fatalf("SelectRepository() error = %v", err)
	}
	if result != "" {
		t.Errorf("SelectRepository() = %q, want %q", result, "")
	}
}

func TestSelectRepository_エラー(t *testing.T) {
	var runner CommandRunner = &mockRunner{result: "", err: fmt.Errorf("command not found")}

	_, err := runner.SelectRepository()
	if err == nil {
		t.Fatal("SelectRepository() error = nil, want error")
	}
}

// fakeExecCommand はテスト用のexecCommandスタブを作成する。
// 指定した出力を返すヘルパープロセスとしてテスト自身を再実行する。
func fakeExecCommand(output string) func(string, ...string) *exec.Cmd {
	return func(name string, args ...string) *exec.Cmd {
		cmd := exec.Command("echo", "-n", output)
		return cmd
	}
}

func TestDefaultRunner_ListRepositories_複数リポジトリ(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	execCommand = fakeExecCommand("/home/user/repos/project1\n/home/user/repos/project2")

	runner := &DefaultRunner{}
	result, err := runner.ListRepositories()
	if err != nil {
		t.Fatalf("ListRepositories() error = %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("ListRepositories() len = %d, want 2", len(result))
	}
	if result[0] != "/home/user/repos/project1" {
		t.Errorf("ListRepositories()[0] = %q, want %q", result[0], "/home/user/repos/project1")
	}
	if result[1] != "/home/user/repos/project2" {
		t.Errorf("ListRepositories()[1] = %q, want %q", result[1], "/home/user/repos/project2")
	}
}

func TestDefaultRunner_ListRepositories_末尾改行のトリム(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	execCommand = fakeExecCommand("/home/user/repos/project1\n")

	runner := &DefaultRunner{}
	result, err := runner.ListRepositories()
	if err != nil {
		t.Fatalf("ListRepositories() error = %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("ListRepositories() len = %d, want 1", len(result))
	}
	if result[0] != "/home/user/repos/project1" {
		t.Errorf("ListRepositories()[0] = %q, want %q", result[0], "/home/user/repos/project1")
	}
}

func TestDefaultRunner_ListRepositories_空出力(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	execCommand = fakeExecCommand("")

	runner := &DefaultRunner{}
	result, err := runner.ListRepositories()
	if err != nil {
		t.Fatalf("ListRepositories() error = %v", err)
	}
	if len(result) != 0 {
		t.Errorf("ListRepositories() len = %d, want 0", len(result))
	}
}

func TestDefaultRunner_ListRepositories_空白のみ(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	execCommand = fakeExecCommand("  \n  \n  ")

	runner := &DefaultRunner{}
	result, err := runner.ListRepositories()
	if err != nil {
		t.Fatalf("ListRepositories() error = %v", err)
	}
	if len(result) != 0 {
		t.Errorf("ListRepositories() len = %d, want 0", len(result))
	}
}
