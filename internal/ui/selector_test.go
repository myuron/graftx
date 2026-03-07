package ui

import (
	"testing"
)

func TestFilterRepoList_完全一致(t *testing.T) {
	app := &App{}
	app.repoList = []string{
		"/home/user/repos/project-alpha",
		"/home/user/repos/project-beta",
		"/home/user/repos/other-gamma",
	}

	app.repoFilterQuery = "project"
	app.filterRepoList()

	if len(app.filteredRepoList) != 2 {
		t.Fatalf("filteredRepoList len = %d, want 2", len(app.filteredRepoList))
	}
	if app.filteredRepoList[0] != "/home/user/repos/project-alpha" {
		t.Errorf("filteredRepoList[0] = %q, want %q", app.filteredRepoList[0], "/home/user/repos/project-alpha")
	}
	if app.filteredRepoList[1] != "/home/user/repos/project-beta" {
		t.Errorf("filteredRepoList[1] = %q, want %q", app.filteredRepoList[1], "/home/user/repos/project-beta")
	}
}

func TestFilterRepoList_大文字小文字無視(t *testing.T) {
	app := &App{}
	app.repoList = []string{
		"/home/user/repos/MyProject",
		"/home/user/repos/other",
	}

	app.repoFilterQuery = "myproject"
	app.filterRepoList()

	if len(app.filteredRepoList) != 1 {
		t.Fatalf("filteredRepoList len = %d, want 1", len(app.filteredRepoList))
	}
	if app.filteredRepoList[0] != "/home/user/repos/MyProject" {
		t.Errorf("filteredRepoList[0] = %q, want %q", app.filteredRepoList[0], "/home/user/repos/MyProject")
	}
}

func TestFilterRepoList_空クエリ(t *testing.T) {
	app := &App{}
	app.repoList = []string{
		"/home/user/repos/project1",
		"/home/user/repos/project2",
	}

	app.repoFilterQuery = ""
	app.filterRepoList()

	if len(app.filteredRepoList) != 2 {
		t.Fatalf("filteredRepoList len = %d, want 2", len(app.filteredRepoList))
	}
}

func TestFilterRepoList_一致なし(t *testing.T) {
	app := &App{}
	app.repoList = []string{
		"/home/user/repos/project1",
	}

	app.repoFilterQuery = "nonexistent"
	app.filterRepoList()

	if len(app.filteredRepoList) != 0 {
		t.Fatalf("filteredRepoList len = %d, want 0", len(app.filteredRepoList))
	}
}

func TestFilterRepoList_カーソルリセット(t *testing.T) {
	app := &App{}
	app.repoList = []string{
		"/home/user/repos/project1",
		"/home/user/repos/project2",
	}
	app.repoSelectorCursor = 5

	app.repoFilterQuery = ""
	app.filterRepoList()

	if app.repoSelectorCursor != 0 {
		t.Errorf("repoSelectorCursor = %d, want 0", app.repoSelectorCursor)
	}
}

func TestSelectorCursorDown_通常移動(t *testing.T) {
	app := &App{}
	app.filteredRepoList = []string{"a", "b", "c"}
	app.repoSelectorCursor = 0

	app.selectorMoveCursorDown()

	if app.repoSelectorCursor != 1 {
		t.Errorf("repoSelectorCursor = %d, want 1", app.repoSelectorCursor)
	}
}

func TestSelectorCursorDown_末尾で停止(t *testing.T) {
	app := &App{}
	app.filteredRepoList = []string{"a", "b", "c"}
	app.repoSelectorCursor = 2

	app.selectorMoveCursorDown()

	if app.repoSelectorCursor != 2 {
		t.Errorf("repoSelectorCursor = %d, want 2", app.repoSelectorCursor)
	}
}

func TestSelectorCursorUp_通常移動(t *testing.T) {
	app := &App{}
	app.filteredRepoList = []string{"a", "b", "c"}
	app.repoSelectorCursor = 2

	app.selectorMoveCursorUp()

	if app.repoSelectorCursor != 1 {
		t.Errorf("repoSelectorCursor = %d, want 1", app.repoSelectorCursor)
	}
}

func TestSelectorCursorUp_先頭で停止(t *testing.T) {
	app := &App{}
	app.filteredRepoList = []string{"a", "b", "c"}
	app.repoSelectorCursor = 0

	app.selectorMoveCursorUp()

	if app.repoSelectorCursor != 0 {
		t.Errorf("repoSelectorCursor = %d, want 0", app.repoSelectorCursor)
	}
}

func TestSelectorCursorDown_空リスト(t *testing.T) {
	app := &App{}
	app.filteredRepoList = []string{}
	app.repoSelectorCursor = 0

	app.selectorMoveCursorDown()

	if app.repoSelectorCursor != 0 {
		t.Errorf("repoSelectorCursor = %d, want 0", app.repoSelectorCursor)
	}
}
