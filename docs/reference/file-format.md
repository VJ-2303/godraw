# The `.tldr` file

On-disk format read and written by `internal/doc/tldr.go`. Compatible with
tldraw; godraw reads tldraw files and writes files tldraw reads. All text is
UTF-8 JSON with UTF-8 escapes only where Go's encoder emits them; key order of
known objects is fixed so files are stable across saves and diff well.

## Envelope

```json
{
  "tldrawFileFormatVersion": 1,
  "schema": { "schemaVersion": 2, "sequences": { "com.tldraw.shape.geo": 12, … } },
  "records": [ … ]
}
```

- `tldrawFileFormatVersion` must be `1`. Absent → `ErrNotTldraw` ("not a
  tldraw file"). `> 1` → rejected: "this file is from a newer version of
  tldraw".
- `schema.sequences` names the record-type versions of the
  `@tldraw/tlschema` 5.6.0 release (see `fileSchema` in `tldr.go`). tldraw
  migrates older files forward, so files godraw writes open there.
- Keys of `records` objects are written in a fixed order (`obj` in `tldr.go`),
  and shape/binding/asset records are written in sorted-id order. Saving the
  same drawing twice with no edits produces byte-identical output modulo the
  trailing newline Go's encoder adds.

## Records godraw writes

| typeName | id | notes |
|----------|----|-------|
| `document` | fixed `document:document` | `name`, `gridSize` (10 if unset), `meta` |
| `page` | `page:…` | `name`, fractional `index`, `meta` |
| `shape` | `shape:…` | see per-type props below |
| `binding` | `binding:…` | `fromId`, `toId`, `props.terminal/start\|end`, `normalizedAnchor {x,y}`, `isExact`, `isPrecise`, `snap` (`"none"` if unset) |
| `asset` | `asset:…` | images as data URLs in `props.src` |

Session records — `camera`, `instance`, `instance_page_state`,
`instance_presence`, `pointer` — are never written.

Records of any other type (e.g. `com.tldraw.shape.bookmark`, `user`) are
preserved verbatim: on read they land unexamined in `Others` and on write they
go back out as they came in.

## Shape records

```json
{ "typeName":"shape", "id":"shape:…", "type":"geo", "parentId":"page:page",
  "index":"a1", "x":…, "y":…, "rotation":0, "opacity":1, "isLocked":false,
  "meta":{}, "props":{ … } }
```

`rotation` is radians, clockwise. `index` is a fractional-index string; see
`architecture/document-model.md`.

### Per-type `props`

**geo** — `w h geo dash growY url scale flipX flipY color labelColor fill
size font align verticalAlign richText`

**draw / highlight** — `segments[] color dash size isComplete isClosed isPen
scale scaleX scaleY` (highlight omits `dash`, `fill`, `isClosed`)

**arrow** — `kind labelColor color fill dash size arrowheadStart arrowheadEnd
font start{x,y} end{x,y} bend richText labelPosition scale elbowMidPoint`

**line** — `color dash size spline points{ key:{id,index,x,y} } scale`; point
keys are object keys, and are sorted by `index` on read.

**text** — `color size w font textAlign autoSize scale richText`

**note** — `color labelColor size font fontSizeAdjustment align verticalAlign
growY url scale richText textLastEditedBy`

**frame** — `w h name color`

**image** — `w h playing url assetId crop{ topLeft, bottomRight, isCircle? }
flipX flipY altText`

**group** — `{}` (children carry transforms)

**Unknown types** — the raw `props` object is written back unchanged, except
`w`/`h` if a resize changed them.

### Stroke segments

`segments` entries are `{type:"free"|"straight", path?, points?, dim}`.
`path` is the compact encoding godraw and tldraw both emit:

1. Concatenate little-endian IEEE-754 values: first point as `float32 x`,
   `y`, (`float32 z` when `dim == 3`, i.e. pen pressure).
2. Each following point as **deltas** from the decoder's position, as IEEE
   binary16 (`float16To`/`float16From` in `tldr.go`), two bytes each per axis.
3. Base64-standard encode the whole buffer.

Writing is delta-from-decoder (each written delta is quantized, then the
running position advances by the *quantized* delta), so rounding error does
not accumulate along a long stroke. `dim == 2` points carry an implicit `z =
0.5`. Files with the pre-compaction `points: [{x,y,z}]` form are read but
always written in `path` form.

### Rich text

`richText` is a ProseMirror JSON document. On read it is converted to plain
text (`RichToPlain`: one line per `paragraph`/`heading`, `bulletList` → `• `,
`orderedList` → `n. `, `hardBreak` → newline). On write:

- if the shape's text was never edited since read, the original `richText` is
  written back byte-for-byte;
- otherwise a minimal document is built from the plain text,
  `{type:"doc",content:[{type:"paragraph",content:[{type:"text",text:…}]}]}`,
  one paragraph per line.

Alignment values written by older tldraw releases carry a `-legacy` suffix
(e.g. `middle-legacy`); it is stripped on read.

## Pages and order

Pages are independent records sorted by `index` on read. Child order inside a
page or group is the `index` string of the child records; `index` sorts
lexicographically over a custom alphabet (see `FirstIndex` and
`IndexBetween`). Deleting and re-adding a shape gives it the current top
index; nothing is renumbered.

## Guarantees and limits

- Reading never partially fails: any shape/binding that errors aborts the
  read with the record id in the error (`shape <id>: …`).
- An empty file reads as a fresh drawing (used by crash-touched files; it is
  not a crash-recovery feature — nothing autosaves).
- A file with no page records gets a default `page:page`.
- Images stay inside the `.tldr` as base64 data URLs; there are no external
  asset references in practice. Image shapes referencing missing assets render
  as a placeholder until one arrives.
- The `.tldr` written by godraw is not pretty-printed; it is one line plus a
  trailing newline. It is intended to be diffed semantically, not eyeballed.
