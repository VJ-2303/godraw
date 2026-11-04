# Drawing tools

Selected from the bottom toolbar, or with a key. `Q` toggles tool lock: when
locked, the tool stays selected after each use instead of reverting to
Select.

| Tool | Key | What a press does |
|------|-----|-------------------|
| Select | `V` | Hit-test; drag to move; brush from empty canvas; handles resize/rotate/bend; arrows select and re-bind |
| Hand | `H` | Drag to pan (also: hold `Space` from any tool) |
| Draw | `D` / `B` | Freehand stroke with pressure; `Shift` draws straight segments |
| Eraser | `E` | Marks shapes en route, deletes on release |
| Arrow | `A` | Drag to create; ends bind to shapes on drop; drag existing ends to re-bind; `Ctrl` leaves ends free |
| Text | `T` | Click or drag to create a text shape and start editing |
| Note | `N` | Click to drop a sticky note; double-click enters its text |
| Line | `L` | Two-click polyline; drag the midpoint dot to add a vertex; `Spline: curve` smooths it |
| Highlighter | `Shift+D` | Translucent freehand |
| Laser | `K` | Fades from the trail after 1 second; never edits the document |
| Frame | `F` | Drag to draw a frame; it adopts what it wraps; rename via the heading |
| Geo | `R`, `O`, … | Drag to draw the active geo shape; `R` = rectangle, `O` = ellipse |

There is no separate image tool; the toolbar's image button and `Ctrl+U` open
a file picker and insert as an image shape.

## Geo shapes

20 kinds, shown in the toolbar popover (`editor.GeoKinds`):

```
rectangle  ellipse  triangle  diamond  pentagon
hexagon    octagon  star      rhombus  rhombus-2
oval       trapezoid arrow-right  arrow-left  arrow-up  arrow-down
x-box      check-box cloud        heart
```

`X-box` and `check-box` keep their inner mark even when the outline is
solid/filled.

## Gestures common to tools

- `Shift` while drawing/resizing: constrain to axis / keep proportions. In
  rotate: 15° steps. In translate: axis-locked movement.
- `Alt` while moving: drag a copy. `Alt` while resizing: resize from the
  center.
- `Ctrl` while moving: snap to other shapes (inverted when "Always Snap" is
  on; grid-snap when `Ctrl` matches that setting and `Grid` is on).
- `Ctrl` while brushing: only shapes fully contained in the brush rectangle
  are selected.

## Handles

`Select` shows handles on the selection: corners/edges resize, the knob above
rotates, arrow midpoints bend, line vertices drag, frame heading drags the
frame. Hovering a handle sets the cursor (`CursorEW`, `CursorNWSE`, `CursorRotate`, …).

## Tool styles

New shapes take `e.Next`, the `Style` shown in the style panel: color, fill,
dash, size, font, align, valign, geo kind, arrow kind/heads/spline, opacity.
Tools apply only the props their shape type understands
(`toolHasStyle`/`SetStyle` in `actions.go`), and setting a style on a mixed
selection applies to the members that accept it.

## Text and notes

Text and notes both own a rich-text document in the file; the editor shows and
edits the plain projection and regenerates the rich document on save. Empty
text gets the shape deleted on stop-editing (no empty shapes left behind, no
history entry). A note's default size is 200×200; `GrowY` lets it grow
downward as text is added.

## What the eraser erases

Frames, draw/highlight/line strokes, geo shapes, text, notes, arrows, images —
anything on the page except groups and locked shapes. Groups are traversed;
a whole group is erased if any part is hit. Erasing never touches the other
pages.
