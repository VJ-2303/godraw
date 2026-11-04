# Geometry (`internal/geo`)

Canvas mathematics. Pure Go, no dependencies. Everything pixel-facing is built
from these primitives; the window never does raw trig.

## Types

| Type | Meaning |
|------|---------|
| `Vec` | point or direction `(X, Y)` |
| `Box` | axis-aligned rectangle `(X, Y, W, H)` |
| `Mat` | affine transform, column-major `A B C D E F` where `x' = A·x + C·y + E` |
| `Path` | vector path of `Move`/`Line`/`Quad`/`Cube`/`Close` segments |
| `Rng` | xorshift32 generator, deterministic from a string seed |
| `Hash` | FNV-1a, used to seed the wobble of hand-drawn strokes |

## Vectors and boxes

Standard operations: add/sub/scale, dot/cross/length/distance/normalize,
rotate, `Per` (quarter turn, y-down screen convention), `Lerp`, `Angle`.

`Box` provides containment (`Contains`, `ContainsBox`), overlap tests
(`Overlaps`), `Union`, `Expand`, and `BoxOf`/`BoxFrom` constructors.

## Matrices

`Compose(x, y, rotation)` is the transform of a shape placed at `(x, y)`
turned by `rotation`. `Mul` composes parent-then-child order; `PageMat` in
the editor folds a shape's ancestor chain. `Invert` handles degenerate
determinants by returning `Identity`. `Rotation()` extracts the 2×2 part's
angle — the window uses it to rotate the text-edit overlay to match its shape.

## Paths

`Path` is renderer-agnostic: `internal/app` flattens it for the canvas, and
`internal/editor` flattens the same path for SVG export and for hit-testing
outlines. Notable constructors: `EllipsePath`, `RoundedPoly`, `Smooth`
(catmull-rom-ish smoothing for strokes), `Flatten(n)` (subdivide into points),
`SVG(m)`.

Freehand strokes are stored as polylines with pressure; the display layer
turns them into ribbon outlines (variable-width filled polygons) so that
pressure shows up in the ink, not just the width.

## Intersection and hit tests

- `SegmentsCross` — proper segment intersection, with endpoints included.
- `SegmentPolyline` — all crossings of a segment with a (optionally closed)
  polyline, in order.
- `SegmentHitsPolyline` — crossing within a margin; used for eraser and arrow
  endpoints.
- `BoxHitsPolyline` — polygon crosses a box or lies inside it.
- `PointInPolygon` — non-zero winding rule.
- `Circle` — circle through three points (used to smooth arc arrows);
  `ShortAngle`, `NormAngle` — angle bookkeeping.
- `RamerDouglasPeucker`-style `Simplify(tol)` — stroke decimation.
- `DistToPolyline` / `PolyLength` / `NearestOnSegment` — measurement.

## Deterministic randomness

`NewRng(seed)` derives a stable stream from any string. The hand-drawn
(`dash: "draw"`) rendering of a shape is seeded by the shape's id, so the
wobble is identical every frame and every restart: geometry is a pure
function of the record, not of the frame.

## Conventions

- Angles are radians; zero is +x, positive is clockwise (y points down).
- `Mat.A`/`D` hold the x/y basis columns; translation is `E`/`F`.
- `Box.Corners()` is clockwise from top-left; `PageMat` transforms shapes
  parent-first, which is how nested group rotation stays sane.
