# Building, testing and releasing

## Toolchain

- Go **1.27.3** (`go.mod`). Local `mise` toolchains older than that need
  `GOTOOLCHAIN=auto` (or an upgrade), or the module fails to load.
- MyGo **v0.7.4**, pinned, plus the `mygo` CLI via the `tool` directive in
  `go.mod`: `go tool mygo …`.
- Linux only. No macOS/Windows build tags, no cgo web view (the `mygo.json`
  Linux target sets `"webview": false`).

## Daily commands

```sh
go build ./...          # compile
go test ./...           # run the suite: 65 test functions
go tool mygo build      # produces build/ with binary + packages
go tool mygo build -platform linux/amd64,linux/arm64   # release build
```

`go vet`/`gofmt` are not wired into a makefile or CI beyond `go test`; the
release workflow is the only automation, and it gates on tests.

## Tests

```
internal/editor    engine: tools, gestures, undo, arrows, export, SVG, round-trip
internal/app       window, panels, opacity, clipboard/menu, export PNG, gallery,
                   theme; snapshot tests render to golden PNGs
internal/doc       .tldr read/write round-trip, fractional indices
internal/geo       vectors, matrices, paths
```

`internal/app` tests run headless through MyGo's `ui.Tester`; `testHost`
captures `Changed` snapshots. `e2e.go` (build tag `e2e`) adds
`MYGO_AUTOMATION` socket automation for an outer driver; it is not built by
default.

Benchmark: `BenchmarkPaint` in `internal/app/app_test.go`.

## Release

Tag `v*` on `main`:

1. Workflow `.github/workflows/release.yml` runs `go test ./...`.
2. `mygo.json`'s `version` is rewritten from the tag via `sed`.
3. `go tool mygo build -platform linux/amd64,linux/arm64` produces per-arch
   directories under `build/` with `godraw-<version>-linux-<arch>.tar.gz`, a
   `.deb`, and a `PKGBUILD`.
4. `sha256sum`s are collected into `sha256sums.txt` (paths stripped of the
   arch directory by the installer).
5. `softprops/action-gh-release` attaches tarballs, `.deb`s, `PKGBUILD`s and
   the checksums file to the GitHub release, with generated notes.

Assets: Shantell Sans (OFL, `internal/assets/fonts/OFL.txt`) and Lucide icons
(ISC, `internal/assets/icons/LICENSE`) ship embedded; the binary is static.

## Install script

`install.sh` (repo root) installs from GitHub Releases, no root required:

- `curl -fsSL …/install.sh | sh`, or `sh install.sh v0.1.0` for a specific
  tag; `sh install.sh --uninstall` removes all of it.
- Downloader: `curl` (5 retries) or `wget`; requires one of them.
- Architecture: `x86_64` → amd64, `aarch64|arm64` → arm64.
- Resolves the latest release tag through the GitHub API when none is given;
  `GODRAW_REPO` overrides the repository (`VJ-2303/godraw` default).
- Verifies the tarball against `sha256sums.txt` from the same release before
  installing.
- Installs to `~/.local/bin/godraw`, `~/.local/share/applications/godraw.desktop`,
  `~/.local/share/godraw/godraw.png`, and an optional MIME entry
  (`mime/packages/godraw.xml`); rewrites `Exec=`/`Icon=` in the desktop file to
  absolute installed paths; runs `update-desktop-database` /
  `update-mime-database` when present.

## Repository layout

```
main.go                 entry point and CLI file handling
e2e.go                  automation plugin (build tag e2e)
mygo.json               packaging metadata (identifier, .tldr association, Linux)
install.sh              user-local release installer
internal/assets/        embedded font and icons
internal/doc/           .tldr model and serialization
internal/editor/        canvas engine (no UI)
internal/geo/           2D geometry
internal/app/           window: paint, panels, gallery, export
docs/                   this documentation
```
