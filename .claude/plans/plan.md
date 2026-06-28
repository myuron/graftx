# Plan: キーバインディングヘルパーの追加 (issue #9)

## 要件
- `?` でキーバインド一覧（ヘルプ）を中央モーダルで表示する。
- ヘルプ表示中は任意キー/Esc/q で閉じる。
- 既存の `?`（後方検索）は廃止する。/ の前方検索と n/N の双方向巡回は維持。

## タスク
- [x] 1. ブランチ作成（feat/keybinding-help）
- [x] 2. 失敗するテストを書く（? でヘルプ表示 / 任意キーで閉じる / render にヘルプ内容）
- [x] 3. 後方検索の削除（InputModeSearchBackward, searchBackward, inputPrefix/isTextInputMode）
- [x] 4. ヘルプ機能の実装（help.go: helpEntries, renderHelpModal, showHelp 状態, ルーティング）
- [x] 5. ステータスバーに [?]Help を追記 + README 更新
- [x] 6. gofmt → golangci-lint → go test ./... → nix build
- [ ] 7. PR 作成
