# Architecture overview

## What godraw is

A single static binary for Linux that draws a whiteboard natively. The UI is
gpu-rendered by MyGo (`github.com/egoist/mygo` v0.7.4), not by a browser
engine. Layout is `geo` ← `doc` ← `editor` ← `app` ← `main`: lower layers
never import upper ones, and only `main.go` and `internal/app` talk to MyGo.

## Process and window lifecycle

One process, one window.

1. `main.go` collects `.tldr` paths from the command line (`fileArgs`), logs
   and skips anything that is not a regular `.tldr` file.
2. `app.NewApp(files)` builds the window's state: an empty `doc.Drawing`, a
   scanner of the drawing library directory (`~/Documents/godraw`), and the
   first drawing to show. With CLI files, the first is imported; otherwise the
   most recently modified library drawing is opened with the gallery open;
   otherwise a new `Drawing N` is created and the gallery is open.
3. `mygo.App.WhenReady` creates the window: 1280×820 DIP, minimum 640×420,
   **frameless** (`Frameless: true`, so Hyprland can tile it edge-to-edge),
   `StateKey: "drawing"`. The system menu bar is removed; commands live in an
   in-app More menu.
4. The window registers one `OnFileDrop` handler: dropped `.tldr` files are
   imported into the library; dropped images are inserted at the drop point.
5. `mygo.App.Run()` starts the loop.

There is one window for all drawings. The library gallery is a modal overlay
in that window; switching drawings swaps the editor's document rather than
opening another window.

## Layers

| Layer | Path | Responsibility | Imports MyGo |
|-------|------|----------------|--------------|
| Entry | `main.go`, `e2e.go` | CLI args, window creation, file-drop | yes |
| Window | `internal/app/` | Paint, panels, gallery, keyboard, export, clipboard | yes |
| Engine | `internal/editor/` | Document state, camera, selection, history, gestures, hit-testing, display lists | no |
| Model | `internal/doc/` | `.tldr` records, (de)serialization, fractional indices | no |
| Geometry | `internal/geo/` | Vectors, boxes, affine matrices, paths, intersection tests | no |
| Assets | `internal/assets/` | Embedded Shantell Sans font (SIL OFL), Lucide icons (ISC) | no |

`internal/editor` and `internal/doc` are pure Go: they run headless, which is
what makes the engine testable without a display.

## Data flow

**Input path.** MyGo delivers pointer/key/scroll events → `App.input`
(`internal/app/app.go`) and `App.shortcuts` (`internal/app/keys.go`) →
`editor.Editor` (`PointerDown`/`PointerMove`/`PointerUp`, or a command).
The editor starts a *gesture* object for a press; `move`/`up`/`cancel` follow
until release or Escape. Mutations go through `put`/`update`/`transact`, which
bump the document and fire `OnChange` with a snapshot. The window sets
`dirty = true` on every change (it has no other effect; nothing autosaves).

**Render path.** Each frame runs the `ui.Painter` callback in `canvas.go`
(`paintCanvas` → `scene`). The scene walks the shape tree depth-first, computes
a culling box per shape from cached geometry, and skips shapes outside the
viewport. Visible shapes are emitted as primitives (`Prim`) built once and
cached in `doc.Shape.Cache`. Frames fill a clip mask for their children.

## Testing

`internal/editor`, `internal/doc`, `internal/geo` are tested as pure Go
(65 test functions in total). `internal/app` is tested with `ui.Tester`, a
headless harness with a fake `Host`; several tests render snapshots to golden
PNGs. `e2e.go` (build tag `e2e`) exposes the app to an automation socket when
`MYGO_AUTOMATION` names one.

## Packaging

`mygo.json` describes the app for `mygo build`: identifier
`io.github.vj-2303.godraw`, `.tldr` file association (MIME
`application/vnd.tldraw+json`), Linux target with no web view. Release is a
tag-triggered workflow; see
[development/building.md](../development/building.md).
