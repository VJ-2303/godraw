# godraw documentation

godraw is a local, native whiteboard for Linux. It is written in Go and drawn
with the GPU through the
[MyGo](https://github.com/egoist/mygo) framework — no web view, no JavaScript.
It reads and writes tldraw's `.tldr` files, so drawings made here open in
tldraw and the other way round.

This folder documents how the project is built and how it behaves. The
user-facing introduction is the repository `README.md`; nothing else here is
marketing.

## Contents

### Architecture

- [architecture/overview.md](architecture/overview.md) — layers, process lifecycle, data flow, module map.
- [architecture/document-model.md](architecture/document-model.md) — the Go records for a drawing: pages, shapes, bindings, assets, history snapshots.
- [architecture/editor-engine.md](architecture/editor-engine.md) — the canvas without a window: camera, selection, undo, gestures, arrows, selection, culling.
- [architecture/ui-layer.md](architecture/ui-layer.md) — the window: canvas painting, panels, gallery, clipboard, export.
- [architecture/geometry.md](architecture/geometry.md) — `internal/geo`: vectors, boxes, matrices, paths.

### Reference

- [reference/tools.md](reference/tools.md) — every drawing tool and what it does.
- [reference/keyboard-shortcuts.md](reference/keyboard-shortcuts.md) — every key binding.
- [reference/file-format.md](reference/file-format.md) — the `.tldr` file: schema versions, record encodings, read and write guarantees.

### Development

- [development/building.md](development/building.md) — toolchain, build, test, release, install script.
- [development/known-gaps.md](development/known-gaps.md) — places where the README overstates what the code does.

## Scope

This documentation covers the code as it is on `main`. Claims trace to source
files named in the text. Where the README and the code disagree, the code is
described and the disagreement is listed in
[development/known-gaps.md](development/known-gaps.md).
