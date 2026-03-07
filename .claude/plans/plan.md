# コピー完了後にコピー元の*マークを消す

## 問題
- `paste`/`pasteOverwrite`後、コピー元ペインの`Selected`がクリアされず`*`が残る

## 修正箇所
- [ ] `keybinding.go`: `clearYankSourceSelection`ヘルパーメソッド追加
- [ ] `keybinding.go`: `paste`関数でコピー成功後にヘルパーを呼ぶ
- [ ] `keybinding.go`: `pasteOverwrite`関数でコピー成功後にヘルパーを呼ぶ
