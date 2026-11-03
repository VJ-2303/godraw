// godraw is a local whiteboard in the manner of tldraw, drawn with MyGo's
// native UI. It reads and writes tldraw's .tldr files.
package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"godraw/internal/app"
)

// fileArgs returns the drawings named on the command line.
func fileArgs() []string {
	var files []string
	for _, a := range os.Args[1:] {
		if !strings.EqualFold(filepath.Ext(a), ".tldr") {
			continue
		}
		if info, err := os.Stat(a); err != nil || !info.Mode().IsRegular() {
			log.Printf("godraw: can't open %s", a)
			continue
		}
		files = append(files, a)
	}
	return files
}

func main() {
	files := fileArgs()
	a := app.NewApp(files)

	mygo.App.WhenReady(func() {
		mygo.App.SetMenu(nil)
		win := mygo.NewWindow(mygo.WindowOptions{
			Width: 1280, Height: 820, MinWidth: 640, MinHeight: 420,
			StateKey:  "drawing",
			Frameless: true,
			Content:   ui.View(a.View),
		})
		win.SetMenu(mygo.NewMenu(nil))
		a.SetWindow(win)

		win.OnFileDrop(func(e *mygo.FileDropEvent) {
			var images []string
			for _, p := range e.Paths {
				if strings.EqualFold(filepath.Ext(p), ".tldr") {
					win.Update(func() { a.ImportFile(p) })
				} else {
					images = append(images, p)
				}
			}
			if len(images) > 0 {
				win.Update(func() { a.DropFiles(images, float64(e.X), float64(e.Y)) })
			}
		})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
