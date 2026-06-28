# Plan: UI を gocui から Bubble Tea へ移行

## タスク
- [x] 1. ブランチ作成（feat/migrate-bubbletea）
- [x] 2. go get（bubbletea / lipgloss / bubbles）
- [x] 3. gomarkdoc で docs/lib/{bubbletea,lipgloss,textinput}.md 生成
- [x] 4. App を tea.Model 化（Init/Update/View）— 通常モードのペイン描画・移動
- [x] 5. 入力モード（textinput）・確認モード・検索/フィルタ・作成/リネーム移植
- [x] 6. セレクタポップアップ移植（中央モーダル）
- [x] 7. gocui 依存コード削除（SetGui/Layout/render*/keybinding登録/inputEditor/selector View群）
- [x] 8. main.go 置換、flake.nix vendorHash 更新、docs/lib/gocui.md 削除、design.md/CLAUDE.md 更新
- [x] 9. gofmt → golangci-lint → go test ./... → nix build
- [ ] 10. PR 作成
