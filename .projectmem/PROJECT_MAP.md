# Project Map - godraw

## Project purpose
A local, native whiteboard in the manner of [tldraw](https://github.com/tldraw/tldraw), written in Go and drawn with [MyGo](https://github.com/egoist/mygo)'s native UI: no web view, no JavaScript. It reads and writes tldraw's `.tldr` files.

## Stack
- Go 1.27.3
- Native UI toolkit: [MyGo](https://github.com/egoist/mygo) (v0.7.4)
- Key packages: purego, typesetting, golang.org/x/image

## Architecture & Structure
- `main.go`: Application entry point, multi-document lifecycle, native desktop menu and drag-and-drop.
- `internal/app/`: Native UI window layer built with MyGo. Canvas view, toolbars, style panels, keyboard shortcuts, export actions.
- `internal/editor/`: Core canvas state engine independent of window UI. Handles tools, gestures, camera/viewport, selection, undo/redo, shapes rendering.
- `internal/doc/`: Data model matching tldraw schema. Reads and writes `.tldr` JSON documents.
- `internal/geo/`: 2D geometry primitives (points, vectors, boxes, affine transforms, paths, collision/intersection).
- `internal/assets/`: Embedded font (Shantell Sans) and SVG icons.

## Key Files to Read First
- `main.go`: App lifecycle and window orchestration.
- `internal/app/app.go`: Window UI structure and event binding.
- `internal/editor/editor.go`: Core canvas state and tool orchestration.
- `internal/doc/doc.go`: Document schema and serialization.
