package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"godraw/internal/doc"
)

// DrawingItem represents a drawing file in the central library.
type DrawingItem struct {
	Path      string
	Name      string
	ModTime   time.Time
	Size      int64
	ThumbPath string
	Thumbnail *ui.Bitmap
}

// defaultDirectory returns ~/Documents/godraw, creating it if needed.
func defaultDirectory() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "drawings"
	}
	dir := filepath.Join(home, "Documents", "godraw")
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

// NewApp creates an App with the central library directory.
func NewApp(files []string) *App {
	a := New(doc.New(), nil)
	a.dir = defaultDirectory()
	a.loadDrawings()

	if len(files) > 0 {
		a.ImportFile(files[0])
	} else if len(a.drawings) > 0 {
		a.OpenDrawing(a.drawings[0].Path)
		a.galleryOpen = true
	} else {
		a.CreateAndOpen("")
		a.galleryOpen = true
	}
	return a
}

func (a *App) SetWindow(w *mygo.Window) { a.win = w }

func (a *App) OpenGallery() {
	a.SaveCurrent()
	a.loadDrawings()
	a.galleryOpen = true
}

func (a *App) CloseGallery() {
	a.galleryOpen = false
}

// loadDrawings scans the library directory for .tldr files.
func (a *App) loadDrawings() {
	if a.dir == "" {
		a.dir = defaultDirectory()
	}
	entries, err := os.ReadDir(a.dir)
	if err != nil {
		return
	}
	var list []*DrawingItem
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") || !strings.EqualFold(filepath.Ext(e.Name()), ".tldr") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		base := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		thumbPath := filepath.Join(a.dir, "."+base+".thumb.png")
		item := &DrawingItem{
			Path:      filepath.Join(a.dir, e.Name()),
			Name:      base,
			ModTime:   info.ModTime(),
			Size:      info.Size(),
			ThumbPath: thumbPath,
		}
		if bData, err := os.ReadFile(thumbPath); err == nil {
			if bmp, err := ui.DecodeBitmap(bData); err == nil {
				item.Thumbnail = bmp
			}
		}
		list = append(list, item)
	}
	slices.SortFunc(list, func(x, y *DrawingItem) int {
		if x.ModTime.After(y.ModTime) {
			return -1
		} else if x.ModTime.Before(y.ModTime) {
			return 1
		}
		return strings.Compare(x.Name, y.Name)
	})
	a.drawings = list
}

// filteredDrawings returns drawings matching the search query.
func (a *App) filteredDrawings() []*DrawingItem {
	q := strings.ToLower(strings.TrimSpace(a.searchQuery))
	if q == "" {
		return a.drawings
	}
	var out []*DrawingItem
	for _, d := range a.drawings {
		if strings.Contains(strings.ToLower(d.Name), q) {
			out = append(out, d)
		}
	}
	return out
}

func (a *App) nextDefaultName() string {
	names := map[string]bool{}
	for _, d := range a.drawings {
		names[strings.ToLower(d.Name)] = true
	}
	for i := 1; ; i++ {
		name := fmt.Sprintf("Drawing %d", i)
		if !names[strings.ToLower(name)] {
			return name
		}
	}
}

// OpenDrawing loads a drawing from disk into the canvas editor.
func (a *App) OpenDrawing(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	d, err := doc.Read(f)
	if err != nil {
		return
	}
	a.currentPath = path
	a.currentName = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	a.ed.Replace(d.Snapshot())
	a.dirty = false
	a.galleryOpen = false
}

// CreateAndOpen creates a new .tldr file and immediately opens it.
func (a *App) CreateAndOpen(name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = a.nextDefaultName()
	}
	// Sanitize name for filesystem
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	path := filepath.Join(a.dir, name+".tldr")

	// Avoid overwriting existing file
	if _, err := os.Stat(path); err == nil {
		for i := 2; ; i++ {
			candidate := filepath.Join(a.dir, fmt.Sprintf("%s (%d).tldr", name, i))
			if _, err := os.Stat(candidate); os.IsNotExist(err) {
				path = candidate
				name = fmt.Sprintf("%s (%d)", name, i)
				break
			}
		}
	}

	d := doc.New()
	f, err := os.Create(path)
	if err == nil {
		_ = doc.Write(f, d)
		f.Close()
	}

	a.currentPath = path
	a.currentName = name
	a.ed.Replace(d.Snapshot())
	a.dirty = false
	a.galleryOpen = false
}

// SaveCurrent writes the active drawing and updates its thumbnail.
func (a *App) SaveCurrent() {
	if a.currentPath == "" || a.ed == nil {
		return
	}
	f, err := os.Create(a.currentPath)
	if err != nil {
		return
	}
	_ = doc.Write(f, a.ed.Doc)
	f.Close()
	a.dirty = false

	// Generate and cache thumbnail
	base := strings.TrimSuffix(filepath.Base(a.currentPath), filepath.Ext(a.currentPath))
	thumbPath := filepath.Join(a.dir, "."+base+".thumb.png")
	if thumbBytes, ok := a.RenderThumbnail(320, 180); ok {
		_ = os.WriteFile(thumbPath, thumbBytes, 0o644)
	}
}

// ImportFile copies an external .tldr file into ~/Documents/godraw/ and opens it.
func (a *App) ImportFile(src string) {
	base := filepath.Base(src)
	dest := filepath.Join(a.dir, base)
	if src != dest {
		// If exists, make unique
		if _, err := os.Stat(dest); err == nil {
			stem := strings.TrimSuffix(base, filepath.Ext(base))
			for i := 2; ; i++ {
				candidate := filepath.Join(a.dir, fmt.Sprintf("%s (%d).tldr", stem, i))
				if _, err := os.Stat(candidate); os.IsNotExist(err) {
					dest = candidate
					break
				}
			}
		}
		in, err := os.Open(src)
		if err == nil {
			out, err := os.Create(dest)
			if err == nil {
				_, _ = io.Copy(out, in)
				out.Close()
			}
			in.Close()
		}
	}
	a.OpenDrawing(dest)
}

func (a *App) confirmRename() {
	if a.renameTarget == nil {
		a.renameModal = false
		return
	}
	newName := strings.TrimSpace(a.renameName)
	if newName != "" && newName != a.renameTarget.Name {
		newPath := filepath.Join(a.dir, newName+".tldr")
		if err := os.Rename(a.renameTarget.Path, newPath); err == nil {
			newThumb := filepath.Join(a.dir, "."+newName+".thumb.png")
			_ = os.Rename(a.renameTarget.ThumbPath, newThumb)
		}
	}
	a.renameModal = false
	a.renameTarget = nil
	a.loadDrawings()
}

func (a *App) confirmDelete() {
	if a.deleteTarget == nil {
		a.deleteModal = false
		return
	}
	_ = os.Remove(a.deleteTarget.Path)
	_ = os.Remove(a.deleteTarget.ThumbPath)
	a.deleteModal = false
	a.deleteTarget = nil
	a.loadDrawings()
}

func formatRelativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "Just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m <= 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h <= 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", h)
	case d < 48*time.Hour:
		return "Yesterday"
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("Jan 02")
	}
}

// galleryDialog renders the centered drawings overlay.
func (a *App) galleryDialog(c *ui.Context) {
	if !a.galleryOpen {
		return
	}
	t := c.Theme()
	items := a.filteredDrawings()
	isSearching := strings.TrimSpace(a.searchQuery) != ""

	ui.Modal(c, &a.galleryOpen, func() {
		ui.Column(c.Key("gallery-dialog")).Width(680).Height(460).Gap(12).Children(func() {
			// Header
			ui.Row(c.Key("gallery-header")).FillWidth().Height(34).AlignItems(ui.Center).Children(func() {
				ui.Row(c.Key("gallery-title")).Gap(8).AlignItems(ui.Baseline).Children(func() {
					ui.Text(c, "Drawings").FontSize(17).Bold()
					if len(a.drawings) > 0 {
						ui.Text(c, fmt.Sprintf("%d", len(a.drawings))).FontSize(12).TextColor(t.TextMuted)
					}
				})

				ui.Box(c).Grow(1)

				ui.Row(c.Key("header-actions")).Gap(8).AlignItems(ui.Center).Children(func() {
					ui.Row(c.Key("search-box")).Width(180).Height(30).Radius(6).
						Background(t.Surface.Alpha(0.5)).Border(1, t.Border.Alpha(0.7)).Padding(0, 8).Gap(6).AlignItems(ui.Center).Children(func() {
						ui.Icon(c, icon("search")).Size(13, 13).TextColor(t.TextMuted)
						ui.TextInput(c.Key("search-input"), &a.searchQuery).Grow(1).Placeholder("Search…")
					})

					newBtn := ui.ButtonBase(c.Key("new-drawing-btn")).Height(30).Padding(0, 12).Radius(6).
						Background(t.Accent).TextColor(t.AccentText).Center().Gap(6)
					newBtn.Children(func() {
						ui.Icon(c, icon("plus")).Size(13, 13)
						ui.Text(c, "New").FontSize(12).Bold()
					})
					newBtn.OnClick(func() {
						a.SaveCurrent()
						a.CreateAndOpen("")
					})

					closeBtn := iconButton(c, "close-gallery", "x", "Close — Esc", false)
					closeBtn.OnClick(func() {
						a.galleryOpen = false
					})
				})
			})

			// Hairline divider
			ui.Box(c).Height(1).FillWidth().Background(t.Border.Alpha(0.4))

			// Grid Content
			ui.Column(c.Key("gallery-body")).FillWidth().Grow(1).Children(func() {
				if isSearching && len(items) == 0 {
					ui.Column(c.Key("empty-search")).Fill().Center().Gap(6).Children(func() {
						ui.Text(c, "No drawings found").FontSize(14).TextColor(t.TextMuted)
					})
					return
				}

				total := len(items)
				if !isSearching {
					total = 1 + len(items)
				}

				ui.GridView(c, &a.gridState, total, 210, 165, func(i int) {
					if !isSearching && i == 0 {
						a.renderNewCard(c, t)
						return
					}
					idx := i
					if !isSearching {
						idx = i - 1
					}
					a.renderDrawingCard(c, t, items[idx])
				}).Grow(1)
			})
		})
	})
}

func (a *App) renderNewCard(c *ui.Context, t *ui.Theme) {
	card := ui.Column(c.Key("create-new-card")).Size(210, 165).Radius(8).
		Background(t.Surface.Alpha(0.2)).Border(1, t.Border.Alpha(0.6)).Center().Gap(8).Clip()
	if card.Hovered() {
		card.Border(1, t.Accent).Background(t.Surface.Alpha(0.4))
	}
	card.Children(func() {
		ui.Icon(c, icon("plus")).Size(22, 22).TextColor(t.TextMuted)
		ui.Text(c, "New drawing").FontSize(12).Bold()
	})
	card.OnClick(func() {
		a.SaveCurrent()
		a.CreateAndOpen("")
	})
}

func (a *App) renderDrawingCard(c *ui.Context, t *ui.Theme, item *DrawingItem) {
	card := ui.Column(c.Key(item.Path)).Size(210, 165).Radius(8).
		Background(t.Surface.Alpha(0.25)).Border(1, t.Border.Alpha(0.6)).Padding(6).Gap(6).Clip()
	if card.Hovered() {
		card.Border(1, t.Accent).Background(t.Surface.Alpha(0.45))
	}

	card.Children(func() {
		// Framed thumbnail sheet
		thumb := ui.Box(c.Key("thumb")).Width(198).Height(112).Radius(5).
			Background(t.Background).Border(1, t.Border.Alpha(0.4)).Center().Clip()
		if item.Thumbnail != nil {
			thumb.Children(func() {
				ui.Image(c, item.Thumbnail).Fit(ui.Contain).Size(190, 104)
			})
		} else {
			thumb.Children(func() {
				ui.Icon(c, icon("frame")).Size(18, 18).TextColor(t.TextMuted.Alpha(0.2))
			})
		}
		thumb.OnClick(func() {
			a.SaveCurrent()
			a.OpenDrawing(item.Path)
		})

		// Meta info row
		ui.Row(c.Key("info")).Width(198).Height(30).AlignItems(ui.Center).Children(func() {
			metaCol := ui.Column(c.Key("meta")).Grow(1).Gap(1).Justify(ui.Center)
			metaCol.Children(func() {
				ui.Text(c, item.Name).FontSize(12).Bold().SingleLine()
				ui.Text(c, formatRelativeTime(item.ModTime)).FontSize(10).TextColor(t.TextMuted).SingleLine()
			})
			metaCol.OnClick(func() {
				a.SaveCurrent()
				a.OpenDrawing(item.Path)
			})

			moreBtn := iconButton(c, "card-more", "ellipsis", "Options", false)
			moreBtn.PopoverMenuButton(func(m *ui.Menu) {
				if m.Item("Rename…").Chosen() {
					a.renameTarget = item
					a.renameName = item.Name
					a.renameModal = true
				}
				if m.Item("Delete").Chosen() {
					a.deleteTarget = item
					a.deleteModal = true
				}
			})
		})
	})
}

func (a *App) renameDrawingModal(c *ui.Context) {
	ui.Modal(c, &a.renameModal, func() {
		ui.Column(c).Width(340).Gap(14).Children(func() {
			ui.Text(c, "Rename Drawing").FontSize(16).Bold()
			in := ui.TextInput(c.Key("rename-name-input"), &a.renameName).AutoFocus().Label("Name")
			ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
				if ui.Button(c, "Cancel").Clicked() {
					a.renameModal = false
				}
				if ui.PrimaryButton(c, "Rename").Clicked() || in.Submitted() {
					a.confirmRename()
				}
			})
		})
	})
}

func (a *App) deleteDrawingModal(c *ui.Context) {
	if a.deleteTarget == nil {
		return
	}
	t := c.Theme()
	ui.Modal(c, &a.deleteModal, func() {
		ui.Column(c).Width(360).Gap(14).Children(func() {
			ui.Text(c, "Delete Drawing").FontSize(16).Bold()
			ui.Text(c, fmt.Sprintf("Delete %q? This cannot be undone.", a.deleteTarget.Name)).
				FontSize(13).TextColor(t.TextMuted)
			ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
				if ui.Button(c, "Cancel").Clicked() {
					a.deleteModal = false
				}
				delBtn := ui.ButtonBase(c.Key("confirm-delete")).Padding(0, 14).Height(32).Radius(6).
					Background(t.Danger).TextColor(ui.RGB(255, 255, 255))
				delBtn.Children(func() { ui.Text(c, "Delete").FontSize(13).Bold() })
				delBtn.OnClick(func() {
					a.confirmDelete()
				})
			})
		})
	})
}
