package main

import (
	"fmt"
	"os"

	"github.com/jroimartin/gocui"
	"github.com/myuron/graftx/internal/fs"
	"github.com/myuron/graftx/internal/selector"
	"github.com/myuron/graftx/internal/ui"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "カレントディレクトリの取得に失敗: %v\n", err)
		os.Exit(1)
	}

	app, err := ui.NewApp(cwd, &fs.OSFS{}, &selector.DefaultRunner{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "アプリケーションの初期化に失敗: %v\n", err)
		os.Exit(1)
	}

	for {
		g, err := gocui.NewGui(gocui.OutputNormal)
		if err != nil {
			fmt.Fprintf(os.Stderr, "GUI初期化に失敗: %v\n", err)
			os.Exit(1)
		}

		app.ExitReason = ui.ExitReasonNone
		if err := app.SetGui(g); err != nil {
			g.Close()
			fmt.Fprintf(os.Stderr, "GUI設定に失敗: %v\n", err)
			os.Exit(1)
		}

		err = g.MainLoop()
		g.Close()

		if err != nil && err != gocui.ErrQuit {
			fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
			os.Exit(1)
		}

		switch app.ExitReason {
		case ui.ExitReasonQuit:
			return
		case ui.ExitReasonSelectRepo:
			// フェーズ3で実装: fzf起動 → SourcePane初期化
			result, err := app.Selector.SelectRepository()
			if err != nil {
				app.Status = fmt.Sprintf("リポジトリ選択エラー: %v", err)
				continue
			}
			if result == "" {
				// キャンセル
				continue
			}
			// SourcePaneの初期化/ChangeDir
			if err := app.SetSourceDir(result); err != nil {
				app.Status = fmt.Sprintf("ソースディレクトリ設定エラー: %v", err)
			}
		default:
			return
		}
	}
}
