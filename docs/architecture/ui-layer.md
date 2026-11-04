# Window layer (`internal/app`)

The MyGo UI. One `App` per window. `main.go` is the only caller in
production; tests exercise it headless through `ui.Tester` with a fake
`Host`.

## `App` and the `Host` interface

```go
type Host interface {
    Changed(d *doc.Drawing)   // a commit happened; called with a snapshot
    Update(fn func())          // run fn on the main thread, then paint
    Window() *mygo.Window      // nil in tests
}
```

Production wiring passes `host = nil` to `app.New`; `App.SetWindow` stores
the real `*mygo.Window`, and `runUpdate`/`getWindow` fall back to it.
`OnChange` sets `a.dirty = true` (nothing acts on it; see
[known-gaps](../development/known-gaps.md)).

## View tree

View paints the canvas and its panels first; `View` then appends the modal
dialogs (gallery, rename, delete) in that order, so the dialogs paint above
the canvas. The help dialog is built inside `canvasView`, before the
dialogs. Unfocused mode (`FocusMode`) hides the panels, not the canvas.

Focus mode (`Ctrl+.`) hides everything except the canvas.

## Canvas paint

`paintCanvas` fills the background (or transparent when exporting), draws the
grid (dots at `GridSize` when `Grid` is on), then walks the page tree in
paint order, rendering each visible shape from the editor's cached display
list. Frames paint their heading and a mask for children. Overlays painted on
top: snap guides, hover/selection outlines, resize/rotate handles, arrow
labels' boxes, brush rectangle, size badge (width × height next to a resize),
eraser/laser trails (fading polylines).

Only shapes intersecting the viewport are drawn; the rest cost a bounds check.

## Text editing

When `Editing` is set, a `ui.TextAreaBase` is positioned, sized, font-set and
rotated to sit exactly over the shape's text box (`textEditor` in `app.go`).
Auto-size text shapes get extra room on their growth side so typing does not
wrap early. `Escape` or `Ctrl+Enter` commits; in a frame heading, `Enter`
commits too. While an editor is active, canvas shortcuts are suspended.

## Drawing library (gallery)

Drawings live in `~/Documents/godraw` as `<name>.tldr` (`defaultDirectory`).
There is no in-canvas file picker; everything goes through the gallery, which
is a modal opened with the Drawings button (top-left, or `Ctrl+O`):

- Grid of rows sorted by mtime (newest first); each row shows name,
  modification time (relative) and file size.
- Search filters by name, case-insensitive.
- `New` creates a `Drawing N` file and opens it.
- Clicking a row saves the current drawing and opens the chosen one in place;
  the current file is saved first.
- Rename and delete act on the file on disk; delete asks in a modal and is
  irreversible.
- External `.tldr` files arrive by `godraw file.tldr` (imported into the
  library with a ` (2)` suffix on name collision), or by dropping them on the
  window, or through the OS file association (see `mygo.json`).
- The drawing open at startup with no CLI args is the most recent one.

Saving happens only on explicit actions: `Ctrl+S`, opening another drawing,
creating a new one, or opening the gallery. There is no autosave, no crash
recovery, and no file watching (see `known-gaps.md`).

## Panels

- **Toolbar** (bottom center): Select, Hand, Draw, Eraser, Arrow, Text,
  Note, insert-image, geo picker (20 shapes), Line, Highlighter, Laser,
  Frame, then the tool-lock toggle (`Q`). Tooltips carry the key binding.
- **Style panel** (top right, appears when the selection or the active tool
  has style): color grid (13 tldraw colors), opacity slider (0.1–1), fill
  (none/semi/solid/pattern/fill), stroke (draw/dashed/dotted/solid), size
  (S/M/L/XL), font (draw/sans/serif/mono), alignment, arrow kind
  (arc/elbow), arrowheads, and line spline (line/curve). Values reflect the
  shared style of the selection; setting applies where the tool/selection
  accepts it.
- **Top-left**: gallery button with current drawing name, page picker
  (list, add, rename by double-click, delete), undo/redo, delete, duplicate,
  More menu (export, grid, always-snap, minimap, focus mode, theme, shortcuts).
- **Bottom-left**: zoom out / percentage (click = 100%, double-click = fit) /
  zoom in / fit, plus minimap toggle and shortcuts button. The minimap is a
  200×130 overview of content bounds; click or drag to re-center, selected
  shapes outlined in the accent color.
- **Help** (`Shift+/`): the full shortcut list, two columns.

## Keyboard

All bindings are in `internal/app/keys.go`; see
[reference/keyboard-shortcuts.md](../reference/keyboard-shortcuts.md).
`ui.Cmd` maps to `Ctrl` on Linux. Shortcuts are inert while a text editor, a
page rename field, or the help dialog is open.

The right-click context menu (`contextMenu`) offers cut/copy/paste/duplicate,
group/ungroup, reorder and align submenus, flip, lock, move-to-page, export,
delete.

## Clipboard

Copying writes a custom format `application/vnd.godraw.shapes+json` (the
shapes as a `.tldr` document) plus, when a plain text projection exists, a
text representation. Paste order: godraw shapes → PNG image → plain text.
Images pasted or dropped are inserted as data-URL assets at the pointer (or
drop point). Supported container types by extension: png, jpg/jpeg, gif, webp.

## Export

`RenderPNG(scale)` renders the export bounds (selection, else page) into a
raster at `scale` (2× for the PNG action), in 1024-pixel tiles with per-tile
clip masks, stitched into one image; pages wider/taller than 8192 device
pixels are clamped proportionally. `editor.SVG(dark)` emits a vector version
from the same display lists with `<path>`/`<text>` elements. Both are written
through a save dialog; no file is produced until a path is chosen.

## Fonts and theme

Shantell Sans (the `draw` font) is embedded and registered once at window
creation (`newMeasurer`). Font mapping: `draw` → Shantell Sans, `serif` →
Georgia, `sans` → system-ui, `mono` → monospace. Themes are the light and
dark palettes from `internal/editor/theme.go` (tldraw's swatch names); light
is the default, dark is toggled with `Ctrl+/`.
