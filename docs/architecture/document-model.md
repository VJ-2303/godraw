# Document model (`internal/doc`)

The document, as read from and written to a `.tldr` file. Records in the
maps are treated as immutable: an edit clones a record and replaces it in the
map, so a snapshot of a drawing is a shallow copy of the maps, cheap enough to
take on every commit, and the undo history keeps the records themselves rather
than diffs.

## Top-level record: `Drawing`

```go
type Drawing struct {
    Name     string
    GridSize float64            // default 10
    Meta     json.RawMessage
    Pages    []*Page
    Shapes   map[ID]*Shape
    Bindings map[ID]*Binding
    Assets   map[ID]*Asset
    Others   []json.RawMessage  // records this app does not model, written back as-is
}
```

- `doc.New()` returns a drawing with one page, `page:page`, `Index: "a1"`.
- `Drawing.Snapshot()` clones the maps and the page slice; records are shared.
- `Equal` compares two snapshots by record identity. Nothing in the app calls
  it today.

## IDs

`ID` is a plain string with the record kind first: `"shape:…"`, `"page:…"`,
`"binding:…"`, `"asset:…"`. `doc.NewID(kind)` fills 21 random characters
from `[0-9a-zA-Z]`. `IsPageID` reports `"page:"` prefixes.

## Shapes

`Shape` is a flat union of the props of all ten types. Each type reads only
its own. Types:

```
geo  draw  highlight  arrow  line  text  note  frame  image  group
```

Important fields:

- `W, H, GrowY`, `Geo` (rectangle … cloud, 20 kinds), `Dash` (`draw`,
  `dashed`, `dotted`, `solid`), `Fill` (`none`, `semi`, `solid`, `pattern`,
  `fill`), `Size` (`s`, `m`, `l`, `xl`), `Font` (`draw`, `sans`, `serif`,
  `mono`), `Align` / `VAlign` (labels), `TextAlign` (text shapes).
- `Segments []Segment` — freehand stroke, split at `Shift`. Each segment is a
  run of `PointZ` (x, y, pressure). Drawn strokes are reproduced exactly from
  the file.
- `Points []LinePoint` — vertices of a line shape, each with an id and a
  fractional index. Dragging a vertex creates or removes one.
- Arrows: `Kind` (`arc` | `elbow`), `Bend`, `Start`/`End`,
  `HeadStart`/`HeadEnd` (eight kinds), `LabelPosition`, `ElbowMidPoint`.
- `RichText` is the ProseMirror document of the file; `Text` is the plain
  text derived from it. `SetText`/`EditChanged` rebuild plain text from the
  editor; on write, `RichText` goes back unchanged if `Text` was not edited
  (`richOf` comparison in `tldr.go`), otherwise a fresh minimal document is
  built from the plain text.
- `Raw map[string]json.RawMessage` holds props this app does not model; they
  are written back unchanged.
- `Cache any` holds derived geometry (bounds, outlines, display state) built
  by the editor. It is cleared by `Clone`/`Moved` and belongs to the thread
  that edits the document.

### Cloning

`Clone()` copies the record and drops `Cache`. Slices and maps are shared, not
deep-copied: callers must replace them, never write into them. This is the
invariant the undo history relies on (a `change[T]` holds `{before, after}`
pointers). `Moved()` copies the record for repositioning; geometry derived
from the shape's own form — not its position — stays valid.

## Bindings

`Binding` ties one end of an arrow to a target shape: `From` (the arrow),
`To` (the target), `Terminal` (`start` | `end`), `Anchor` (normalized
coordinates on the target), `IsExact`, `IsPrecise`, `Snap` (`"none"` written
when empty — required by tldraw's schema). Arrows follow their targets because
the editor recomputes arrow geometry from bindings each time it needs it
(`arrowMemo`, invalidated when bindings change).

## Pages and fractional indices

Page order and child order within a page are independent of list position.
Each record carries an `Index` string that sorts between any two neighbors
(`internal/doc/index.go`):

- Alphabet `0-9A-Za-z`; `FirstIndex` is `a1`.
- The head letter encodes the integer part's length; `a`–`z` count upward,
  `A`–`Z` downward.
- `IndexBetween(a, b)` returns a key between them; `IndicesBetween` calls it
  repeatedly, halving the remaining room so bulk inserts stay short.
- `incInt`/`decInt` carry or borrow across the integer part when its head
  letter saturates.

Reordering never renumbers anything, which is why undo can hold records by
identity. Pages sort by `Index` on read; children sort by `Index` (ties broken
by id) under each parent.

## Read/write asymmetry

The next document in memory and the next file are not identical views:

- Reading: session records (`camera`, `instance`, `instance_page_state`,
  `instance_presence`, `pointer`) are dropped. Everything else either maps to
  a modeled record or lands in `Others` and is written back verbatim.
- Writing: fields this app models are always rebuilt from the record, so a
  modeled-but-stale raw prop does not leak back. Unknown shape types are
  written with their raw props untouched, except `w`/`h` if a resize changed
  them.

Empty file (e.g. `touch x.tldr`) reads as a fresh drawing. A file whose
`tldrawFileFormatVersion` is absent reads as "not a tldraw file"; a value > 1
is rejected as newer than this app understands. See
[reference/file-format.md](../reference/file-format.md) for the byte-level
details.
