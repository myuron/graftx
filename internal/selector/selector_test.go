package selector

import (
	"fmt"
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

func TestListRepositories_正常パス(t *testing.T) {
	repos := []string{"/home/user/repos/project1", "/home/user/repos/project2"}
	var runner CommandRunner = &mockRunner{repoList: repos}

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
}

func TestListRepositories_エラー(t *testing.T) {
	var runner CommandRunner = &mockRunner{listErr: fmt.Errorf("command not found")}

	_, err := runner.ListRepositories()
	if err == nil {
		t.Fatal("ListRepositories() error = nil, want error")
	}
}

func TestListRepositories_空リスト(t *testing.T) {
	var runner CommandRunner = &mockRunner{repoList: []string{}}

	result, err := runner.ListRepositories()
	if err != nil {
		t.Fatalf("ListRepositories() error = %v", err)
	}
	if len(result) != 0 {
		t.Errorf("ListRepositories() len = %d, want 0", len(result))
	}
}
