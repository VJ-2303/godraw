# Editor engine (`internal/editor`)

Everything the canvas does that is not pixels. No MyGo imports; the package is
testable headless. `PackageName`: `package editor`.

```go
type Editor struct {
    Doc     *doc.Drawing
    Page    doc.ID
    Measure Measure            // text layout callback, supplied by app
    Tool    Tool
    ToolLocked, Grid, Snap, DarkMode, FocusMode, ReadOnly, Debug bool
    Next    Style              // style of the next shape; the style panel sets it
    // ...
}
```

`New(d, measure)` starts on the drawing's first page. `Replace(d)` swaps the
document (file reload), discards history, and resets transient state.

## Camera

`screen = (page + (X, Y)) * Z` — `Camera{X, Y, Z}`, one per page (`cams` map).
`ToPage`/`ToScreen` convert; `ViewBox()` is the visible region in page
coordinates. `SetCam` clamps `Z` to `[0.05, 8]`. `ZoomIn`/`ZoomOut` step
through `zoomStops` (0.05, 0.1, 0.25, 0.5, 1, 2, 4, 8); `ZoomToFit` fills the
viewport with the page's content bounds; `FitNewPage` frames a page on first
view; `ZoomToSelection` frames the selection.

## History (undo/redo)

A transaction opens with `begin` and closes with `commit`. Edits inside it are
recorded as `change[T]{before, after}` maps (shapes, bindings, assets, pages,
name) in one `entry`. Escape-while-dragging calls `rollback`. `Undo`/`Redo`
replay the entry forward or backward, restoring selection and page as they
were at each end of it.

Details:

- Committing drops entries whose `before == after` pointer (the gesture
  returned to what it started with produces no undo step).
- Slider style edits (opacity) leave the transaction open; `Undo` calls
  `commit()` first to close it, so one slider session is one step.
- `OnChange` fires once per commit/undo/redo, not per drag frame.
- Undo also covers page operations and document rename; selection and active
  page are restored per entry.

## Gestures

A press starts a gesture; it gets `move`,`up`,`cancel`:

- `selectDown` — hit-test; resize/rotate/arrow/line-point handles win over
  shapes; otherwise start `pressShape` (drag to move, with `Alt` = copy,
  `Shift` = axis-locked, `Ctrl` = snap/free) or a brush (`pressCanvas`).
  Double-click enters groups / starts text editing.
- `translateGesture` — moves `topSelected()` (selection minus anything whose
  ancestor is also selected). `Alt` clones live. On drop, shapes are
  reparented into or out of the frame under the pointer (`reparentMoved`).
- `resizeGesture` / `rotateGesture` — `Shift` constrains ratio / 15° steps;
  `Alt` resizes from the center.
- `createBox` / `clickCreate` / `lineCreate` / `freehand` / `arrowCreate` /
  `arrowEndDrag` / `arrowBendDrag` / `linePointDrag` / `eraserGesture` /
  `laserGesture` / `panGesture`.
- Escape cancels any gesture and restores the pre-gesture records.

The eraser marks shapes while dragging and deletes them on release; it skips
locked shapes, descends into groups, and matches strokes by segment/distance
plus filled interiors.

## Selection rules

- Clicking a child selects the **outermost group**, not the child.
- `Enter` or double-click on a group focuses it; clicks inside a focused group
  select within it; clicking outside drops the focus.
- `SelectAll` selects the unlocked children of the focused group, or of the
  page.
- `topSelected()` drops selected shapes whose ancestor is also selected — the
  move/align/duplicate units.
- Locked shapes cannot be selected, moved, resized, or erased (handles
  disappear), but they still render and can be unlocked.

## Hit-testing and culling

`hitShape` is type-aware (filled closed shapes hit anywhere; strokes hit
within a margin of `8/Z`; frames hit in the body or on the heading).

Each shape's display geometry is computed once and cached (`Shape.Cache`), and
is invalidated by `Clone`. `PageBounds` boxes are derived from it. Each frame,
`scene.culled` skips shapes whose bounds miss the viewport; culling is the
main performance lever on huge boards.

Display lists are built by `Editor.Display(s)` into `[]Prim`; `Prim`s are
consumed by both the canvas (`internal/app/canvas.go`) and SVG export, so
both see identical geometry.

## Arrows

Arrows are records bound, on one or both ends, to target shapes. The editor
recomputes arrow geometry from the bindings on demand and memoizes it
(`arrowMemo`); the memo is dropped when the document or bindings change.

- Drag from empty canvas with the arrow tool (or drag an existing arrow's
  end): near a shape, a binding is created (`bindTerminal` with the normalized
  anchor); `Ctrl` while dragging leaves the end free.
- Binding endpoints re-snap to the shape's outline on move; the outline is
  the shape's own geometry (`geoOutline`), not its bounding box.
- `Kind: "arc"` renders as a quadratic curve through `Bend` (draggable at the
  midpoint); `Kind: "elbow"` routes orthogonally.
- Labels sit at `LabelPosition` along the path; label font size follows
  `Size` (s 18, m 20, l 24, xl 28).

## Snapping

Two independent mechanisms, deliberately XOR-gated:

- **Shape/grid alignment while translating**: snaps when `Ctrl` is held and
  "Always Snap" (the `Snap` flag) is off, or when `Ctrl` is not held and
  "Always Snap" is on (`if p.Mods.Cmd != e.Snap` in `translateGesture`).
  Targets are edges and centers of other shapes; threshold is `8/Z` page
  units; guide lines are drawn while snapped (`e.Snaps`).
- **Grid snap**: if `Grid` is on and shape-snapping did not engage, positions
  round to the drawing's `GridSize` (default 10).

## Frames

A frame clips its children (the window paints a clip mask from the frame's
path). A shape dropped fully inside a frame is reparented into it; a shape
dragged out is reparented back out. `DropOn` marks the frame under the pointer
during a drag (highlighted). New frames adopt the shapes they wrap
(`adoptIntoFrame`). `frameHeading` is the draggable name strip; double-click
in the body writes a text shape there.

## Export bounds

`ExportBounds` returns the union of visual bounds of the selection, or of all
shapes on the page when nothing is selected, expanded by stroke width. This is
what "export the selection, else the page" means everywhere PNG/SVG export or
clipboard-copy is invoked.

## Read-only and debug

`ReadOnly` gates every mutating action and all creation tools except laser/
select, but no UI path sets it — it exists for embedding and tests. `Debug`
gates `DebugJSON`, which dumps the document; nothing in the UI exposes it.
