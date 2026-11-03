package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"

	"godraw/internal/doc"
)

func TestHomeGalleryFlow(t *testing.T) {
	tempDir := t.TempDir()

	a := NewApp(nil)
	a.dir = tempDir
	a.loadDrawings()

	if !a.galleryOpen {
		t.Fatalf("expected galleryOpen = true on launch")
	}

	// Create and open new drawing
	a.CreateAndOpen("Diagram Alpha")
	if a.galleryOpen {
		t.Fatalf("expected galleryOpen = false after creation")
	}
	if a.currentName != "Diagram Alpha" {
		t.Errorf("currentName = %q, want 'Diagram Alpha'", a.currentName)
	}

	// Add shape and save
	rect := doc.NewShape(doc.NewID("shape"), doc.TypeGeo)
	rect.X, rect.Y, rect.W, rect.H = 50, 50, 100, 80
	a.ed.Doc.Shapes[rect.ID] = rect
	a.SaveCurrent()

	if _, err := os.Stat(a.currentPath); os.IsNotExist(err) {
		t.Fatalf("file not saved to %q", a.currentPath)
	}

	// Open gallery overlay
	a.OpenGallery()
	if !a.galleryOpen {
		t.Fatalf("expected galleryOpen = true after OpenGallery")
	}
	if len(a.drawings) != 1 {
		t.Fatalf("expected 1 drawing, got %d", len(a.drawings))
	}
	if a.drawings[0].Name != "Diagram Alpha" {
		t.Errorf("drawing name = %q, want 'Diagram Alpha'", a.drawings[0].Name)
	}

	// Rename drawing
	a.renameTarget = a.drawings[0]
	a.renameName = "Diagram Beta"
	a.confirmRename()

	if len(a.drawings) != 1 || a.drawings[0].Name != "Diagram Beta" {
		t.Errorf("expected renamed to Diagram Beta, got %+v", a.drawings)
	}
	if _, err := os.Stat(filepath.Join(tempDir, "Diagram Beta.tldr")); os.IsNotExist(err) {
		t.Errorf("Diagram Beta.tldr does not exist")
	}

	// Render view with ui.Tester
	tt := ui.NewTester(a.View, 1280, 820)
	tt.Frame()

	// Delete drawing
	a.deleteTarget = a.drawings[0]
	a.confirmDelete()

	if len(a.drawings) != 0 {
		t.Errorf("expected 0 drawings after delete, got %d", len(a.drawings))
	}
}
