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
		fmt.Fprintf(os.Stderr, "failed to get current directory: %v\n", err)
		os.Exit(1)
	}

	app, err := ui.NewApp(cwd, &fs.OSFS{}, &selector.DefaultRunner{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize GUI: %v\n", err)
		os.Exit(1)
	}
	defer g.Close()

	if err := app.SetGui(g); err != nil {
		fmt.Fprintf(os.Stderr, "failed to configure GUI: %v\n", err)
		os.Exit(1)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
