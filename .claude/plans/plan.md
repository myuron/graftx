# Plan: プレビュー機能の追加

## 要件
- フォーカス中ペインの下に、カーソル行ファイルのプレビューを常時表示する。
  - Source フォーカス時は Source ペインの下、Dest フォーカス時は Dest ペインの下。
- テキストファイルは内容、ディレクトリは中の一覧を表示。バイナリ／読み込み失敗は不可メッセージ。

## タスク
- [x] 1. ブランチ作成（feat/preview-pane）
- [x] 2. FileSystem に ReadFile を追加（osfs 実装 + テスト用モック更新）
- [x] 3. 失敗するテストを書く（previewLines / isBinary / render 統合）
- [x] 4. previewLines・isBinary・renderPreview を実装しテストをパス
- [x] 5. render() にフォーカス側の縦分割を統合
- [x] 6. gofmt → golangci-lint → go test ./... → nix build
- [ ] 7. PR 作成
