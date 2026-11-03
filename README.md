# godraw

A local, native whiteboard for Linux. Pure Go, drawn straight to your GPU with
[MyGo](https://github.com/egoist/mygo)'s native UI: no web view, no JavaScript.

![godraw, drawing a hero board](docs/screenshots/demo.png)

## Why godraw

Most whiteboard apps are a browser engine with a whiteboard inside. godraw is a single
static binary that draws with the GPU directly. It starts instantly, and your drawings
are plain files on your own disk — no account, no cloud, works fully offline.

## Install

One line, no root needed:

```sh
curl -fsSL https://raw.githubusercontent.com/VJ-2303/godraw/main/install.sh | sh
```

This puts `godraw` in `~/.local/bin` and adds it to your applications menu. Grab a
tarball or `.deb` from [Releases](https://github.com/VJ-2303/godraw/releases) instead,
or uninstall any time with `install.sh --uninstall`.

## What it does

**Tools.** Select, hand, draw, eraser, arrow, text, note, image, line, highlighter, laser,
frame, and 20 geo shapes. Keep a tool selected with `Q`. Press `?` for every shortcut.

**Editing.** Multiple selection and brush selection, move, resize, rotate, flip, duplicate,
alt-drag to copy, snapping (hold `Ctrl`), grid, group and ungroup, lock, reorder, align and
distribute, copy and paste, undo and redo, and multiple pages.

**Shapes.**
- Hand-drawn, dashed, dotted and solid strokes
- None, semi, solid, pattern and full fills
- Four sizes and four fonts, including a handwriting font
- Opacity
- Labels on shapes, notes and arrows
- Arrows that bind to shapes and follow them, bend, or route as elbows, with eight arrowheads
- Lines with draggable vertices and curves
- Frames that clip and hold their contents

**Your drawings.** Every drawing is a `.tldr` file in `~/Documents/godraw` — plain JSON you
own, readable by other apps that use the format. The home screen lists your drawings:
search, rename, delete, or start a new one. Each drawing opens in its own window, autosave
and crash recovery are built in, recent files are one click away, and godraw picks up
changes other apps make to a file you have open.

**Images in and out.** Paste images from the clipboard, drop files onto the window, and
export the selection or the whole page as PNG (at 2× resolution) or SVG.

**View.** Pan, zoom and pinch, zoom to fit, minimap, dark mode, focus mode.

## Performance

godraw stays fast by doing less work, not by caching harder:

- **No browser engine.** The UI is drawn by native GPU rendering. There is no DOM, no
  JavaScript, and no web view process — startup is immediate and memory use stays low.
- **Culling.** Each frame computes a culling box per shape from cached page bounds, and
  only shapes that intersect the viewport are drawn. Zoomed into a corner of a huge board,
  everything off screen costs nothing.
- **Cached display lists.** Each shape's display list is built once and reused across
  moves and zooms.
- **Cheap ordering.** Shapes and pages are ordered by fractional indices, so reordering
  never renumbers anything, and undo holds the records themselves.
- **Tiled export.** PNG export renders large pages in tiles and stitches them, so pages
  far larger than the screen or a single GPU texture still export, capped at 8192 px per
  side.

## Credits

- [Shantell Sans](https://github.com/arrowtype/shantell-sans), the handwriting font (SIL OFL)
- [Lucide](https://lucide.dev), the icons (ISC)

Licenses are in `internal/assets`.
