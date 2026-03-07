package ui

import (
	"testing"

	"github.com/myuron/graftx/internal/fs"
)

func TestIconForEntry_ディレクトリ(t *testing.T) {
	entry := fs.Entry{Name: "docs", IsDir: true}
	got := iconForEntry(entry)
	want := "\uf115" // フォルダアイコン
	if got != want {
		t.Errorf("iconForEntry(%v) = %q, want %q", entry, got, want)
	}
}

func TestIconForEntry_Go(t *testing.T) {
	entry := fs.Entry{Name: "main.go", IsDir: false}
	got := iconForEntry(entry)
	want := "\ue627" // Goアイコン
	if got != want {
		t.Errorf("iconForEntry(%v) = %q, want %q", entry, got, want)
	}
}

func TestIconForEntry_拡張子大文字(t *testing.T) {
	entry := fs.Entry{Name: "README.MD", IsDir: false}
	got := iconForEntry(entry)
	want := "\uf48a" // Markdownアイコン
	if got != want {
		t.Errorf("iconForEntry(%v) = %q, want %q", entry, got, want)
	}
}

func TestIconForEntry_ファイル名完全一致(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Makefile", "\ue779"},
		{"Dockerfile", "\uf308"},
		{".gitignore", "\ue702"},
		{"go.mod", "\ue627"},
		{"go.sum", "\ue627"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := fs.Entry{Name: tt.name, IsDir: false}
			got := iconForEntry(entry)
			if got != tt.want {
				t.Errorf("iconForEntry(%v) = %q, want %q", entry, got, tt.want)
			}
		})
	}
}

func TestIconForEntry_各拡張子(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"app.py", "\ue73c"},
		{"index.js", "\ue74e"},
		{"style.css", "\ue749"},
		{"index.html", "\ue736"},
		{"data.json", "\ue60b"},
		{"config.yaml", "\ue60b"},
		{"README.md", "\uf48a"},
		{"image.png", "\uf1c5"},
		{"archive.zip", "\uf410"},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			entry := fs.Entry{Name: tt.filename, IsDir: false}
			got := iconForEntry(entry)
			if got != tt.want {
				t.Errorf("iconForEntry(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestIconForEntry_未知の拡張子(t *testing.T) {
	entry := fs.Entry{Name: "unknown.xyz123", IsDir: false}
	got := iconForEntry(entry)
	want := "\uf15b" // デフォルトファイルアイコン
	if got != want {
		t.Errorf("iconForEntry(%v) = %q, want %q", entry, got, want)
	}
}

func TestIconForEntry_拡張子なし(t *testing.T) {
	entry := fs.Entry{Name: "LICENSE", IsDir: false}
	got := iconForEntry(entry)
	// LICENSEはファイル名完全一致にない場合、デフォルトアイコン
	want := "\uf15b"
	if got != want {
		t.Errorf("iconForEntry(%v) = %q, want %q", entry, got, want)
	}
}
