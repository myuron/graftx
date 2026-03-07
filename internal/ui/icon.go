package ui

import (
	"path/filepath"
	"strings"

	"github.com/myuron/graftx/internal/fs"
)

// デフォルトアイコン
const (
	iconDir     = "\uf115" // フォルダアイコン
	iconDefault = "\uf15b" // デフォルトファイルアイコン
)

// ファイル名完全一致のアイコンマッピング
var iconByName = map[string]string{
	"Makefile":       "\ue779",
	"CMakeLists.txt": "\ue779",
	"Dockerfile":     "\uf308",
	".dockerignore":  "\uf308",
	".gitignore":     "\ue702",
	".gitmodules":    "\ue702",
	".gitattributes": "\ue702",
	"go.mod":         "\ue627",
	"go.sum":         "\ue627",
	".env":           "\uf462",
	".env.local":     "\uf462",
	"package.json":   "\ue71e",
	"tsconfig.json":  "\ue628",
	"Gemfile":        "\ue739",
	"Rakefile":       "\ue739",
	"Cargo.toml":     "\ue7a8",
	"Cargo.lock":     "\ue7a8",
}

// 拡張子→アイコンのマッピング
var iconByExt = map[string]string{
	// Go
	".go": "\ue627",
	// Python
	".py":  "\ue73c",
	".pyx": "\ue73c",
	".pyi": "\ue73c",
	// JavaScript / TypeScript
	".js":  "\ue74e",
	".mjs": "\ue74e",
	".cjs": "\ue74e",
	".jsx": "\ue7ba",
	".ts":  "\ue628",
	".tsx": "\ue7ba",
	// Web
	".html": "\ue736",
	".htm":  "\ue736",
	".css":  "\ue749",
	".scss": "\ue749",
	".sass": "\ue749",
	".less": "\ue749",
	// Data / Config
	".json":    "\ue60b",
	".yaml":    "\ue60b",
	".yml":     "\ue60b",
	".toml":    "\ue60b",
	".xml":     "\ue619",
	".csv":     "\uf1c3",
	".sql":     "\uf1c0",
	".graphql": "\ue662",
	// Documentation
	".md":       "\uf48a",
	".markdown": "\uf48a",
	".rst":      "\uf48a",
	".txt":      "\uf15c",
	".tex":      "\uf15c",
	// Shell
	".sh":   "\uf489",
	".bash": "\uf489",
	".zsh":  "\uf489",
	".fish": "\uf489",
	// Ruby
	".rb":  "\ue739",
	".erb": "\ue739",
	// Rust
	".rs": "\ue7a8",
	// Java / Kotlin
	".java":   "\ue738",
	".kt":     "\ue634",
	".kts":    "\ue634",
	".gradle": "\ue738",
	// C / C++
	".c":   "\ue61e",
	".h":   "\ue61e",
	".cpp": "\ue61d",
	".hpp": "\ue61d",
	".cc":  "\ue61d",
	// C#
	".cs": "\uf81a",
	// Swift
	".swift": "\ue755",
	// PHP
	".php": "\ue73d",
	// Lua
	".lua": "\ue620",
	// Elixir / Erlang
	".ex":  "\ue62d",
	".exs": "\ue62d",
	".erl": "\ue7b1",
	// Haskell
	".hs": "\ue777",
	// Scala
	".scala": "\ue737",
	// R
	".r":   "\uf25d",
	".rmd": "\uf25d",
	// Image
	".png":  "\uf1c5",
	".jpg":  "\uf1c5",
	".jpeg": "\uf1c5",
	".gif":  "\uf1c5",
	".svg":  "\uf1c5",
	".ico":  "\uf1c5",
	".webp": "\uf1c5",
	// Archive
	".zip": "\uf410",
	".tar": "\uf410",
	".gz":  "\uf410",
	".bz2": "\uf410",
	".xz":  "\uf410",
	".7z":  "\uf410",
	".rar": "\uf410",
	// Binary / Executable
	".exe":  "\uf489",
	".bin":  "\uf489",
	".wasm": "\ue6a1",
	// Lock
	".lock": "\uf023",
	// PDF
	".pdf": "\uf1c1",
	// Docker
	".dockerfile": "\uf308",
	// Diff / Patch
	".diff":  "\uf440",
	".patch": "\uf440",
}

// iconForEntry はfs.Entryに対応するNerd Fontアイコン文字を返す。
func iconForEntry(entry fs.Entry) string {
	if entry.IsDir {
		return iconDir
	}

	// ファイル名完全一致
	if icon, ok := iconByName[entry.Name]; ok {
		return icon
	}

	// 拡張子マッチ（小文字化）
	ext := strings.ToLower(filepath.Ext(entry.Name))
	if icon, ok := iconByExt[ext]; ok {
		return icon
	}

	return iconDefault
}
