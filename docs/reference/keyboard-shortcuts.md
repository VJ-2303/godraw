# Keyboard shortcuts

All bindings live in `internal/app/keys.go` (`shortcuts`) and the in-app help
dialog (`?`). `ui.Cmd` is `Ctrl` on Linux. Bindings are inert while a text
editor is open, a page is being renamed, or the help dialog is shown. In a
text editor, only `Escape` / `Ctrl+Enter` (stop) and, for one-line frame
headings, `Enter` (commit) do anything.

## Tools

| Key | Action |
|-----|--------|
| `V` | Select |
| `H` | Hand (or hold `Space`) |
| `D`, `B` | Draw |
| `Shift+D` | Highlighter |
| `E` | Eraser |
| `A` | Arrow |
| `T` | Text |
| `N` | Note |
| `L` | Line |
| `F` | Frame |
| `K` | Laser |
| `R` | Geo: rectangle |
| `O` | Geo: ellipse |
| `Q` | Toggle tool lock |

## File and view

| Key | Action |
|-----|--------|
| `Ctrl+S` | Save |
| `Ctrl+O` | Drawings gallery |
| `Ctrl+N` | Save and create new drawing |
| `Ctrl+U` | Insert image |
| `Ctrl+/` | Toggle dark mode |
| `Ctrl+.` | Focus mode |
| `Ctrl+'` | Toggle grid |
| `Shift+/` | Keyboard shortcuts |
| `Ctrl+=` | Zoom in |
| `Ctrl+-` | Zoom out |
| `Shift+0` | Zoom to 100% |
| `Shift+1` | Zoom to fit |
| `Shift+2` | Zoom to selection |
| `Ctrl+Shift+Z` / `Ctrl+Y` | Redo |

## Editing

| Key | Action |
|-----|--------|
| `Ctrl+Z` | Undo |
| `Ctrl+A` | Select all (unlocked; focused group, else page) |
| `Ctrl+D` | Duplicate |
| `Ctrl+C` / `Ctrl+X` / `Ctrl+V` | Copy / cut / paste |
| `Ctrl+G` / `Ctrl+Shift+G` | Group / ungroup |
| `Shift+L` | Lock / unlock |
| `Delete`, `Backspace` | Delete |
| `Escape` | Cancel (clears gesture, selection, editing) |
| `Enter` | Edit the single selected text-like shape |
| `[` / `]` | Send backward / bring forward |
| `Ctrl+[` / `Ctrl+]` | Send to back / bring to front |
| `Arrows` | Nudge 1 DIP (screen); `Shift+Arrows` 10 DIP |

## Align and arrange

| Key | Action |
|-----|--------|
| `Alt+A` / `Alt+H` / `Alt+D` | Align left / center-h / right |
| `Alt+W` / `Alt+V` / `Alt+S` | Align top / center-v / bottom |
| `Shift+H` | Flip horizontal |
| `Shift+V` | Flip vertical |

Distribute horizontally/vertically is context-menu only
(`Distribute Horizontally` / `Vertically`, minimum 3 selected).

## Drag modifiers

| Key | While dragging |
|-----|----------------|
| `Shift` | Constrain: axis (move), proportions (resize), 15° (rotate) |
| `Alt` | Copy (move); resize from center |
| `Ctrl` | Snap to shapes (or, with "Always Snap" on, do not snap); free an arrow end |
| `Ctrl` + brush | Select only fully-contained shapes |
| `Space` | Temporary pan |

## Mouse

- Double-click canvas: create text.
- Double-click text-like shape / frame heading / group: edit, rename, or
  enter the group.
- Double-click zoom percentage: fit to page; single click resets to 100%.
- Right-click: context menu (arrange, align, flip, lock, move-to-page, export).
- Scroll: pan. `Ctrl`+scroll: zoom at pointer. Pinch: zoom.
- Click empty canvas: clear selection (right-click also on empty canvas).
