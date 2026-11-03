// godraw is a local whiteboard in the manner of tldraw, drawn with MyGo's
// native UI. It reads and writes tldraw's .tldr files.
package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"godraw/internal/app"
	"godraw/internal/doc"
)

// window is what a document's window holds.
type window struct {
	doc *mygo.Document[*doc.Drawing]
	win *mygo.Window
	app *app.App
}

func (w *window) Changed(d *doc.Drawing) { w.doc.Set(d) }

func (w *window) Update(fn func()) {
	if w.win != nil {
		w.win.Update(fn)
	}
}

func (w *window) Window() *mygo.Window { return w.win }

func (w *window) New() {
	if docs != nil {
		go docs.New()
	}
}

func (w *window) Open() {
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Parent:  w.win,
			Filters: []mygo.FileFilter{{Name: "tldraw Drawing", Extensions: []string{"tldr"}}},
		})
		if err == nil && len(paths) > 0 {
			for _, p := range paths {
				go openDrawing(p)
			}
		}
	}()
}

func (w *window) Save() {
	go func() {
		if err := w.doc.Save(); err != nil {
			log.Printf("godraw: save: %v", err)
		}
	}()
}

func (w *window) SaveAs() {
	go func() {
		if err := w.doc.SaveAs(""); err != nil {
			log.Printf("godraw: save as: %v", err)
		}
	}()
}

// docs holds all open drawings.
var docs *mygo.Documents[*doc.Drawing]

// openDrawing opens a drawing in a window of its own, off the main thread.
var openDrawing func(path string)

// fileArgs returns the drawings named on the command line. MyGo opens them
// itself in apps the mygo CLI built; this covers go run and go build too, and
// a file asked for twice just focuses its window.
func fileArgs() []string {
	var files []string
	for _, a := range os.Args[1:] {
		if !isDrawing(a) {
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

// isDrawing reports whether a path names a drawing.
func isDrawing(path string) bool { return strings.EqualFold(filepath.Ext(path), ".tldr") }

func main() {
	files := fileArgs()
	openDrawing = func(path string) {
		if _, err := docs.Open(path); err != nil {
			log.Printf("godraw: %s: %v", path, err)
		}
	}
	docs = mygo.NewDocuments(mygo.DocumentsOptions[*doc.Drawing]{
		Types: []mygo.DocumentType{{Name: "tldraw Drawing", Extensions: []string{"tldr"}}},
		New:   doc.New,
		// With files to open, no untitled document comes up beside them.
		NoUntitled: len(files) > 0,
		Read:       func(r io.Reader) (*doc.Drawing, error) { return doc.Read(r) },
		Write:      func(w io.Writer, d *doc.Drawing) error { return doc.Write(w, d) },
		// Records don't change once made, so a drawing compares by them:
		// undoing back to what the file has leaves it unedited.
		Equal: doc.Equal,
		Window: func(d *mygo.Document[*doc.Drawing]) *mygo.Window {
			w := &window{doc: d}
			w.app = app.New(d.Content().Snapshot(), w)
			w.win = mygo.NewWindow(mygo.WindowOptions{
				Width: 1280, Height: 820, MinWidth: 640, MinHeight: 420,
				StateKey:  "drawing",
				Frameless: true,
				Content:   ui.View(w.app.View),
			})
			w.win.SetMenu(mygo.NewMenu(nil))
			// Dropped drawings open; dropped images land where they fall.
			w.win.OnFileDrop(func(e *mygo.FileDropEvent) {
				var images []string
				for _, p := range e.Paths {
					if isDrawing(p) {
						go openDrawing(p)
					} else {
						images = append(images, p)
					}
				}
				w.win.Update(func() { w.app.DropFiles(images, float64(e.X), float64(e.Y)) })
			})
			// Reverted, or changed by another app: show what the file has.
			d.OnReload(func() {
				w.win.Update(func() { w.app.Editor().Replace(d.Content().Snapshot()) })
			})
			return w.win
		},
	})
	mygo.App.WhenReady(func() {
		mygo.App.SetMenu(nil)
		if len(files) == 0 {
			return
		}
		go func() {
			opened := 0
			for _, f := range files {
				if _, err := docs.Open(f); err != nil {
					log.Printf("godraw: %s: %v", f, err)
					continue
				}
				opened++
			}
			if opened == 0 {
				docs.New() // nothing else would show a window
			}
		}()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
