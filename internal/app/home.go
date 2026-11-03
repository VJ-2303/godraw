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
	Path    string
	Name    string
	ModTime time.Time
	Size    int64
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
		list = append(list, &DrawingItem{
			Path:    filepath.Join(a.dir, e.Name()),
			Name:    base,
			ModTime: info.ModTime(),
			Size:    info.Size(),
		})
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
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	path := filepath.Join(a.dir, name+".tldr")

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

// SaveCurrent writes the active drawing to its file.
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
}

// ImportFile copies an external .tldr file into ~/Documents/godraw/ and opens it.
func (a *App) ImportFile(src string) {
	base := filepath.Base(src)
	dest := filepath.Join(a.dir, base)
	if src != dest {
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
		_ = os.Rename(a.renameTarget.Path, newPath)
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

func formatBytes(b int64) string {
	switch {
	case b < 1024:
		return fmt.Sprintf("%d B", b)
	case b < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	default:
		return fmt.Sprintf("%.1f MB", float64(b)/(1024*1024))
	}
}

// galleryDialog renders the centered drawings list overlay.
func (a *App) galleryDialog(c *ui.Context) {
	if !a.galleryOpen {
		return
	}
	t := c.Theme()
	items := a.filteredDrawings()

	ui.Modal(c, &a.galleryOpen, func() {
		ui.Column(c.Key("gallery-dialog")).Width(480).Gap(12).Children(func() {
			// Header
			ui.Row(c.Key("gallery-header")).FillWidth().Height(34).AlignItems(ui.Center).Children(func() {
				ui.Row(c.Key("gallery-title")).Gap(6).AlignItems(ui.Baseline).Children(func() {
					ui.Text(c, "Drawings").FontSize(16).Bold()
					if len(a.drawings) > 0 {
						ui.Text(c, fmt.Sprintf("(%d)", len(a.drawings))).FontSize(12).TextColor(t.TextMuted)
					}
				})

				ui.Box(c).Grow(1)

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

			// Search Bar (clean native input without double-border wrapper)
			ui.TextInput(c.Key("search-input"), &a.searchQuery).
				Placeholder("Search drawings…").FillWidth().Height(34)

			// Divider
			ui.Box(c).Height(1).FillWidth().Background(t.Border.Alpha(0.4))

			// List Content
			if len(items) == 0 {
				ui.Column(c.Key("empty-search")).FillWidth().Height(90).Center().Gap(6).Children(func() {
					msg := "No drawings yet"
					if a.searchQuery != "" {
						msg = fmt.Sprintf("No drawings match %q", a.searchQuery)
					}
					ui.Text(c, msg).FontSize(13).TextColor(t.TextMuted)
				})
				return
			}

			ui.Column(c.Key("drawing-list")).FillWidth().MaxHeight(320).Gap(2).Children(func() {
				for _, item := range items {
					it := item
					isActive := it.Path == a.currentPath

					row := ui.Row(c.Key(it.Path)).FillWidth().Height(46).Radius(6).
						Padding(0, 10).Gap(10).AlignItems(ui.Center)
					if isActive {
						row.Background(t.SurfacePressed)
					} else if row.Hovered() {
						row.Background(t.SurfaceHover)
					}

					row.Children(func() {
						iconColor := t.TextMuted
						if isActive {
							iconColor = t.Accent
						}
						ui.Icon(c, icon("pencil")).Size(16, 16).TextColor(iconColor)

						col := ui.Column(c.Key("meta")).Grow(1).Gap(2).Justify(ui.Center)
						col.Children(func() {
							title := ui.Text(c, it.Name).FontSize(13).SingleLine()
							if isActive {
								title.Bold().TextColor(t.Accent)
							}
							metaText := fmt.Sprintf("%s · %s", formatRelativeTime(it.ModTime), formatBytes(it.Size))
							ui.Text(c, metaText).FontSize(11).TextColor(t.TextMuted).SingleLine()
						})
						col.OnClick(func() {
							if !isActive {
								a.SaveCurrent()
								a.OpenDrawing(it.Path)
							} else {
								a.galleryOpen = false
							}
						})

						renameBtn := iconButton(c, "rename-"+it.Name, "pencil", "Rename…", false)
						renameBtn.OnClick(func() {
							a.renameTarget = it
							a.renameName = it.Name
							a.renameModal = true
						})

						delBtn := iconButton(c, "del-"+it.Name, "trash-2", "Delete", false)
						delBtn.OnClick(func() {
							a.deleteTarget = it
							a.deleteModal = true
						})
					})
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
