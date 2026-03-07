package selector

import (
	"fmt"
	"testing"
)

// mockRunner はテスト用のCommandRunnerモック。
type mockRunner struct {
	result string
	err    error
}

func (m *mockRunner) SelectRepository() (string, error) {
	return m.result, m.err
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
