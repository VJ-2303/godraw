# projectmem - godraw

_Last updated: 2026-11-03_

## Project purpose
A local, native whiteboard in the manner of [tldraw](https://github.com/tldraw/tldraw), written in Go and drawn with [MyGo](https://github.com/egoist/mygo)'s native UI: no web view, no JavaScript. It reads and writes tldraw's `.tldr` files.

## Recent issues
- [DONE] #0002 Search box has double border/padding artifact; thumbnail rendering looks like AI mockup and pollutes ~/Documents/godraw [internal/app/home.go:300] -> Removed double-bordered search wrapper and replaced AI thumbnail gallery with compact native document list [internal/app/home.go:300] (fixed)
- [DONE] #0001 panic: mygo: NewWindow called before the application is ready; create windows in App.WhenReady [main.go:36] -> Instantiate window inside mygo.App.WhenReady callback [main.go:37] (fixed)

## Decisions
- Removed examples directory and showcase demo code
- Removed unused struct fields identified by staticcheck in internal/app and internal/editor
- Enabled Frameless: true for seamless tiling in Hyprland and edge-to-edge canvas
- Removed in-window GTK menubar on Linux/Windows and added New/Open/Save to in-app More menu and shortcuts
- Removed all macOS/Windows-specific menu structures, DMG references, and Windows syso artifacts for Linux-only target
- Centralize drawings in ~/Documents/godraw with native home gallery screen and remove canvas file-handling menus
- Use centered dialog overlay for drawings gallery instead of full-screen page
- Cache shape pageBounds and precompute viewport culling boxes per frame
- Default to light theme instead of OS dark theme; dark theme only when explicitly set [internal/app/app.go:211]
- Apply Option B warm paper light theme: #f8f7f4 canvas, #212124 warm ink text, 10% alpha grid dots, warm panel shadow [internal/editor/theme.go:76]
- Refactor: refactor: remove system theme option, default to light [internal/app/app.go]
- Refined light theme to match MyGo zinc chrome: canvas Text #1d1d1d, Selection #2563eb (unified with chrome accent), Grid alpha-based #1d1d1d @ 14%; canvas stays cool #f9fafb [internal/editor/theme.go:77]

## Notes
- init
- add showcase
- [ImgBot] Optimize images (#1)
- gotcha: go.mod and mygo v0.7.4 require go >= 1.27.3; host mise Go is 1.27.1 with GOTOOLCHAIN=local, requires GOTOOLCHAIN=auto or Go upgrade in mise
- New feature: feat: linux release pipeline — tag-triggered build and user-local install script [install.sh]
- New feature: feat: clean flat app icon — white stroke on accent blue [resources/icon.png]

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
