---
name: fix-coderabbit-review
description: |
  CodeRabbitのPRレビュー結果を取得し、修正計画を立て、ユーザー承認後にコード修正・コミット・プッシュまで行うスキル。
  ユーザーが「CodeRabbitのレビューを修正して」「PRのレビュー指摘を直して」「coderabbitの指摘を対応して」
  「PR #123のレビューを修正」「レビュー結果を反映して」などと言った場合にトリガーする。
  CodeRabbit、レビュー修正、PR指摘対応、レビュー反映に関する依頼には必ずこのスキルを使用すること。
---

# CodeRabbitレビュー修正スキル

CodeRabbitがPRに投稿したレビューコメントを取得し、修正計画を立て、ユーザーの承認を得てからコード修正を実施する。

## ワークフロー

### Step 1: PR特定とリポジトリ情報の取得

ユーザーがPR番号を指定した場合はそれを使う。指定がなければ現在のブランチから自動検出する。

```bash
# リポジトリのowner/repoを取得
gh repo view --json owner,name --jq '"\(.owner.login)/\(.name)"'

# 現在のブランチに紐づくPRを検出
gh pr view --json number,title,url,headRefName
```

PRが見つからない場合はユーザーにPR番号を確認する。

### Step 2: CodeRabbitレビューの取得

CodeRabbitのユーザー名は `coderabbitai[bot]` である。レビューは3箇所に投稿される。全て取得すること。

#### 2a: issueコメント（サマリーとウォークスルー）

CodeRabbitはPRのissueコメントにウォークスルー（変更概要）とpre-merge checksを投稿する。
このコメントから全体像を把握する。ただしHTML内部状態コメント（`<!-- internal state -->`）は無視すること。

```bash
gh api repos/{owner}/{repo}/issues/{pr_number}/comments \
  --jq '.[] | select(.user.login == "coderabbitai[bot]") | {id: .id, body: .body}'
```

#### 2b: レビュー本文（コメント一覧サマリー）

レビュー本文には「Actionable comments posted: N」という形で指摘数と、全コメントの要約が含まれる。
また「Prompt for all review comments with AI agents」セクションに全指摘の修正指示がまとまっている。

```bash
gh api repos/{owner}/{repo}/pulls/{pr_number}/reviews \
  --jq '.[] | select(.user.login == "coderabbitai[bot]") | {id: .id, state: .state, body: .body}'
```

#### 2c: インラインコメント（ファイル固有の指摘）

各ファイルの特定行に対する具体的な指摘。`path`と`original_line`で対象箇所を特定できる。
各コメントには重要度（例: `⚠️ Potential issue | 🟠 Major`）と「Prompt for AI Agents」セクションが含まれる。

```bash
gh api repos/{owner}/{repo}/pulls/{pr_number}/comments \
  --jq '.[] | select(.user.login == "coderabbitai[bot]") | {path: .path, line: .original_line, body: .body}'
```

### Step 3: レビュー指摘の解析と整理

CodeRabbitのコメントから以下の情報を抽出する：

1. **重要度**: コメント冒頭の絵文字タグから判定
   - `⚠️ Potential issue | 🟠 Major` → 重要
   - `🧹 Nitpick` → 軽微
   - `💡 Suggestion` → 提案

2. **修正指示**: 各コメントの「Prompt for AI Agents」セクションに具体的な修正手順が記載されている。
   このセクションの指示を修正方針の基盤として活用する。

3. **修正提案コード**: `📝 Committable suggestion` や ```diff ブロックに具体的なコード変更案が含まれる場合がある。

各指摘を以下の形式で整理する：
- 対象ファイルと行番号
- 重要度（Major / Nitpick / Suggestion）
- 指摘内容の要約（日本語）
- 修正方針の案

### Step 4: 修正計画の提示と承認

全体の修正計画をユーザーに提示する。重要度順（Major → Nitpick → Suggestion）に並べる。

```
## CodeRabbitレビュー修正計画

### PR #XXX: [PRタイトル]

| # | 重要度 | ファイル | 指摘内容 | 修正方針 |
|---|--------|---------|---------|---------|
| 1 | 🟠 Major | main.go:43 | 端末サイズが小さい場合のエラー | サイズチェックを追加 |
| 2 | 🟠 Major | main.go:49 | ReadFile未使用・エラー無視 | 関数を削除 |
| 3 | Nitpick | main.go:13 | CreateGuiがpanicする | error返却に変更 |
```

ユーザーに対して以下の選択肢を提供する：
- **全て修正**: 全指摘を修正する
- **個別選択**: 指摘ごとに修正する/しないを選ぶ
- **修正方針の変更**: 特定の指摘について修正方針を変えたい場合

ユーザーが個別選択した場合、各指摘について修正するかどうかを確認する。

### Step 5: コード修正

承認された修正計画に基づいてコードを修正する。

修正の進め方：
1. CodeRabbitの「Prompt for AI Agents」の指示を参考にしつつ、プロジェクトの規約に従って修正する
2. CodeRabbitが提供した`📝 Committable suggestion`のコードは参考にするが、そのまま適用せず文脈を確認する
3. CLAUDE.mdの規約に従う（テスト駆動開発、日本語コメント等）
4. 修正が他の部分に影響しないか確認する
5. 各指摘を修正するたびに、何を変更したか簡潔に報告する

### Step 6: テスト・フォーマット・リントの実行

CLAUDE.mdのワークフローに従い、修正後に以下を実施する：

```bash
# フォーマッター
gofmt -w .

# リンター（利用可能な場合）
golangci-lint run

# テスト
go test ./...
```

テストやリンターが失敗した場合は修正を見直し、ユーザーに報告する。

### Step 7: コミットとプッシュ

全ての修正が完了し、テストが通ったらコミットしてプッシュする。

コミットメッセージのフォーマット：
```
fix: CodeRabbitレビュー指摘の修正

- [修正内容1の要約]
- [修正内容2の要約]
...
```

プッシュ前にユーザーに確認を取ること。

### Step 8: 完了報告

修正内容のサマリーを表示する：

```
## 修正完了

### 修正した指摘
| # | ファイル | 修正内容 |
|---|---------|---------|
| 1 | main.go:43 | Layout関数にサイズチェックを追加 |

### スキップした指摘
| # | ファイル | 理由 |
|---|---------|------|
| 2 | main.go:49 | ユーザー判断でスキップ |

コミット: abc1234
プッシュ先: origin/branch-name
```
