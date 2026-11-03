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

// Mode is whether the app is on the Home page or in the Canvas editor.
type Mode int

const (
	ModeHome Mode = iota
	ModeCanvas
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
	} else {
		a.mode = ModeHome
	}
	return a
}

func (a *App) SetWindow(w *mygo.Window) { a.win = w }

func (a *App) GoHome() {
	a.mode = ModeHome
	a.loadDrawings()
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

// OpenDrawing loads a drawing from disk and switches to canvas mode.
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
	a.mode = ModeCanvas
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
	a.mode = ModeCanvas
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
	if thumbBytes, ok := a.RenderThumbnail(220, 135); ok {
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

// homeView renders the Home Page file gallery.
func (a *App) homeView(c *ui.Context) {
	t := c.Theme()
	items := a.filteredDrawings()

	ui.Column(c.Key("home-root")).Fill().Background(t.Background).Children(func() {
		// Top Header Bar
		ui.Row(c.Key("home-header")).FillWidth().Height(56).Padding(0, 24, 0, 24).
			AlignItems(ui.Center).Border(1, t.Border.Alpha(0.6)).Background(t.Surface).Children(func() {

			// Branding
			ui.Row(c.Key("brand")).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Icon(c, icon("pencil")).Size(18, 18).TextColor(t.Accent)
				ui.Text(c, "godraw").FontSize(17).Bold()
			})

			ui.Box(c).Width(32)

			// Search Box
			ui.Row(c.Key("search-box")).Width(280).Height(34).Radius(7).
				Background(t.Background).Border(1, t.Border).PaddingX(8).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Icon(c, icon("search")).Size(14, 14).TextColor(t.TextMuted)
				ui.TextInput(c.Key("search-input"), &a.searchQuery).Grow(1).Placeholder("Search drawings…")
			})

			ui.Box(c).Grow(1)

			// Header Right Actions
			ui.Row(c.Key("header-right")).Gap(10).AlignItems(ui.Center).Children(func() {
				themeIcon := "moon"
				if t.Dark {
					themeIcon = "sun"
				}
				iconButton(c, "theme-toggle", themeIcon, "Toggle theme", false).OnClick(func() {
					a.ToggleDark()
				})

				btn := ui.ButtonBase(c.Key("new-drawing-btn")).Height(34).Padding(0, 14).Radius(7).
					Background(t.Accent).TextColor(t.AccentText).Center().Gap(6)
				btn.Children(func() {
					ui.Icon(c, icon("plus")).Size(15, 15)
					ui.Text(c, "New Drawing").FontSize(13).Bold()
				})
				btn.OnClick(func() {
					a.newName = a.nextDefaultName()
					a.newModal = true
				})
			})
		})

		// Main Content Container
		ui.Column(c.Key("home-body")).FillWidth().Grow(1).Padding(24, 24, 24, 24).Children(func() {
			// Section Header
			ui.Row(c.Key("section-header")).FillWidth().Height(32).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "Drawings").FontSize(14).Bold()
				ui.Box(c).Width(6)
				ui.Text(c, fmt.Sprintf("(%d)", len(items))).FontSize(13).TextColor(t.TextMuted)
			})

			ui.Box(c).Height(16)

			// Empty State
			if len(items) == 0 {
				ui.Column(c.Key("empty-state")).Fill().Center().Gap(12).Children(func() {
					ui.Icon(c, icon("frame")).Size(48, 48).TextColor(t.TextMuted.Alpha(0.35))
					if a.searchQuery != "" {
						ui.Text(c, "No drawings match your search").FontSize(14).TextColor(t.TextMuted)
					} else {
						ui.Text(c, "No drawings yet").FontSize(15).Bold()
						ui.Text(c, "Create your first drawing to start sketching").FontSize(13).TextColor(t.TextMuted)
						ui.Box(c).Height(4)
						createBtn := ui.ButtonBase(c.Key("empty-create")).Height(34).Padding(0, 16).Radius(7).
							Background(t.Accent).TextColor(t.AccentText).Center()
						createBtn.Children(func() {
							ui.Text(c, "+ Create Drawing").FontSize(13).Bold()
						})
						createBtn.OnClick(func() {
							a.newName = a.nextDefaultName()
							a.newModal = true
						})
					}
				})
				return
			}

			// Grid of Drawing Cards
			ui.GridView(c, &a.gridState, len(items), 220, 185, func(i int) {
				item := items[i]
				card := ui.Column(c.Key(item.Path)).Size(220, 185).Radius(9).
					Border(1, t.Border).Background(t.Surface).Clip()
				if card.Hovered() {
					card.Background(t.SurfaceHover)
				}

				card.Children(func() {
					// Thumbnail Area
					thumb := ui.Box(c.Key("thumb")).Width(220).Height(125).Background(t.Background).Center().Clip()
					if item.Thumbnail != nil {
						thumb.Children(func() {
							ui.Image(c, item.Thumbnail).Fit(ui.Contain).Size(210, 115)
						})
					} else {
						thumb.Children(func() {
							ui.Icon(c, icon("pencil")).Size(28, 28).TextColor(t.TextMuted.Alpha(0.25))
						})
					}
					thumb.OnClick(func() {
						a.OpenDrawing(item.Path)
					})

					// Divider
					ui.Box(c).Height(1).FillWidth().Background(t.Border.Alpha(0.7))

					// Card Info Row
					ui.Row(c.Key("info")).FillWidth().Height(58).Padding(0, 10, 0, 10).AlignItems(ui.Center).Children(func() {
						infoCol := ui.Column(c.Key("meta")).Grow(1).Gap(2).Justify(ui.Center)
						infoCol.Children(func() {
							ui.Text(c, item.Name).FontSize(13).Bold().SingleLine()
							ui.Text(c, formatRelativeTime(item.ModTime)).FontSize(11).TextColor(t.TextMuted).SingleLine()
						})
						infoCol.OnClick(func() {
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
			}).Grow(1)
		})
	})

	// Modals
	a.newDrawingModal(c)
	a.renameDrawingModal(c)
	a.deleteDrawingModal(c)
}

func (a *App) newDrawingModal(c *ui.Context) {
	ui.Modal(c, &a.newModal, func() {
		ui.Column(c).Width(340).Gap(14).Children(func() {
			ui.Text(c, "New Drawing").FontSize(16).Bold()
			in := ui.TextInput(c.Key("new-name-input"), &a.newName).AutoFocus().Label("Name")
			ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
				cancelBtn := ui.ButtonBase(c.Key("cancel-new")).Padding(0, 12).Height(32).Radius(6)
				cancelBtn.Children(func() { ui.Text(c, "Cancel").FontSize(13) })
				cancelBtn.OnClick(func() { a.newModal = false })

				createBtn := ui.ButtonBase(c.Key("confirm-new")).Padding(0, 14).Height(32).Radius(6).
					Background(c.Theme().Accent).TextColor(c.Theme().AccentText)
				createBtn.Children(func() { ui.Text(c, "Create").FontSize(13).Bold() })
				createBtn.OnClick(func() {
					a.CreateAndOpen(a.newName)
					a.newModal = false
				})
			})
			if in.Submitted() {
				a.CreateAndOpen(a.newName)
				a.newModal = false
			}
		})
	})
}

func (a *App) renameDrawingModal(c *ui.Context) {
	ui.Modal(c, &a.renameModal, func() {
		ui.Column(c).Width(340).Gap(14).Children(func() {
			ui.Text(c, "Rename Drawing").FontSize(16).Bold()
			in := ui.TextInput(c.Key("rename-name-input"), &a.renameName).AutoFocus().Label("Name")
			ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
				cancelBtn := ui.ButtonBase(c.Key("cancel-rename")).Padding(0, 12).Height(32).Radius(6)
				cancelBtn.Children(func() { ui.Text(c, "Cancel").FontSize(13) })
				cancelBtn.OnClick(func() { a.renameModal = false })

				confirmBtn := ui.ButtonBase(c.Key("confirm-rename")).Padding(0, 14).Height(32).Radius(6).
					Background(c.Theme().Accent).TextColor(c.Theme().AccentText)
				confirmBtn.Children(func() { ui.Text(c, "Rename").FontSize(13).Bold() })
				confirmBtn.OnClick(func() {
					a.confirmRename()
				})
			})
			if in.Submitted() {
				a.confirmRename()
			}
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
				cancelBtn := ui.ButtonBase(c.Key("cancel-delete")).Padding(0, 12).Height(32).Radius(6)
				cancelBtn.Children(func() { ui.Text(c, "Cancel").FontSize(13) })
				cancelBtn.OnClick(func() { a.deleteModal = false })

				delBtn := ui.ButtonBase(c.Key("confirm-delete")).Padding(0, 14).Height(32).Radius(6).
					Background(ui.RGBA(220, 50, 50, 1)).TextColor(ui.RGB(255, 255, 255))
				delBtn.Children(func() { ui.Text(c, "Delete").FontSize(13).Bold() })
				delBtn.OnClick(func() {
					a.confirmDelete()
				})
			})
		})
	})
}
