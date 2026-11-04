# Known gaps

Where the README and `plan.md` state things the code does not do. The code is
the source of truth; this list exists so nobody builds on the false claims.

## README claims not backed by the code

| README says | Reality |
|---|---|
| "autosave and crash recovery are built in" | Nothing saves without an explicit action (`Ctrl+S`, switching drawings, `Ctrl+N`, opening the gallery). No timer, no write-on-close, no recovery file. An empty `.tldr` reads as a fresh drawing, but nothing ever writes one automatically. |
| "godraw picks up changes other apps make to a file you have open" | No file watching. An externally modified drawing is only re-read on explicit open. |
| "Each drawing opens in its own window" | One window total. The library is a modal overlay in it; opening another drawing swaps the editor's document (`OpenDrawing` → `Replace`). |
| "recent files are one click away" | No recents list or menu. The gallery lists everything in `~/Documents/godraw`, newest first, filtered by search. |
| plan.md: "Auto-generate/cache thumbnails on drawing save" marked shipped | No thumbnail is generated or cached anywhere. The gallery shows a compact text list with an icon, not card previews. |

## Code that does nothing today

- `App.dirty` is set on every change and cleared on save; nothing reads it.
  There is no close prompt or unsaved-changes indicator (the window is
  frameless and there is no quit-time save either).
- `CopyPNG()` (`internal/app/export.go`) renders the selection/page as a
  2× PNG onto the clipboard but is wired to no button, menu item or shortcut.
- `Editor.ReadOnly` and `Editor.Debug` are enforced throughout the engine
  but set nowhere in the app; they are not reachable features.
- `showSizeBadge` is read by the canvas, but no UI toggles `sizeBadge`; it is
  always on.

## Verified behaviors worth knowing

- `Shift+/` is the only binding for the help dialog (no bare `?` binding;
  the help dialog lists it as `Shift+/`).
- `[` / `]` reorder; `Ctrl+[` / `Ctrl+]` jump to back/front.
- Grid snap only engages when shape-snapping did not (see `translateGesture`);
  holding `Ctrl` with "Always Snap" on disables snapping rather than falling
  back to the grid.
