package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jroimartin/gocui"
)

var unusedGlobal = "this is unused"

func CreateGui() *gocui.Gui {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	return g
}

func Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	midX := maxX / 2

	// 左ペイン
	if v, err := g.SetView("left", 0, 0, midX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Source"
		fmt.Fprintln(v, "Left pane")
	}

	// 右ペイン
	if v, err := g.SetView("right", midX, 0, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Destination"
		fmt.Fprintln(v, "Right pane")
	}

	return nil
}

func ReadFile(path string) string {
	data, _ := os.ReadFile(path)
	return string(data)
}

func ProcessItems(items []string) []string {
	result := []string{}
	for i := 0; i < len(items); i++ {
		item := items[i]
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func Quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func main() {
	g := CreateGui()
	defer g.Close()

	g.SetManagerFunc(Layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, Quit); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
