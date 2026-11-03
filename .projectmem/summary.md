# projectmem - godraw

_Last updated: 2026-11-03_

## Project purpose
A local, native whiteboard in the manner of [tldraw](https://github.com/tldraw/tldraw), written in Go and drawn with [MyGo](https://github.com/egoist/mygo)'s native UI: no web view, no JavaScript. It reads and writes tldraw's `.tldr` files.

## Recent issues
- [DONE] #0001 panic: mygo: NewWindow called before the application is ready; create windows in App.WhenReady [main.go:36] -> Instantiate window inside mygo.App.WhenReady callback [main.go:37] (fixed)

## Decisions
- Removed examples directory and showcase demo code
- Removed unused struct fields identified by staticcheck in internal/app and internal/editor
- Enabled Frameless: true for seamless tiling in Hyprland and edge-to-edge canvas
- Removed in-window GTK menubar on Linux/Windows and added New/Open/Save to in-app More menu and shortcuts
- Removed all macOS/Windows-specific menu structures, DMG references, and Windows syso artifacts for Linux-only target
- Centralize drawings in ~/Documents/godraw with native home gallery screen and remove canvas file-handling menus

## Notes
- init
- add showcase
- [ImgBot] Optimize images (#1)
- gotcha: go.mod and mygo v0.7.4 require go >= 1.27.3; host mise Go is 1.27.1 with GOTOOLCHAIN=local, requires GOTOOLCHAIN=auto or Go upgrade in mise

## Key files
- `.agents/skills/mygo-maintenance/SKILL.md`
- `.agents/skills/mygo-maintenance/agents/openai.yaml`
- `.gitignore`
- `README.md`
- `e2e.go`
- `go.mod`
- `go.sum`
- `internal/app/app.go`
- `internal/app/app_test.go`
- `internal/app/canvas.go`
- `internal/app/export.go`
- `internal/app/keys.go`
- `internal/app/panels.go`
- `internal/app/text.go`
- `internal/assets/assets.go`
- `internal/assets/fonts/OFL.txt`
- `internal/assets/fonts/ShantellSans.ttf`
- `internal/assets/icons/LICENSE`
- `internal/assets/icons/align-center-horizontal.svg`
- `internal/assets/icons/align-center-vertical.svg`

## Open questions
- None logged yet.
