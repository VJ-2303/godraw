#!/bin/sh
# Installs godraw from GitHub Releases into ~/.local — no root needed.
#
#   curl -fsSL https://raw.githubusercontent.com/VJ-2303/godraw/main/install.sh | sh
#   install.sh v0.1.0     # a specific version
#   install.sh --uninstall
#
# Environment: GODRAW_REPO (default VJ-2303/godraw) overrides the repository.
set -eu

REPO="${GODRAW_REPO:-VJ-2303/godraw}"
PREFIX="${HOME}/.local"
BIN_DIR="${PREFIX}/bin"
SHARE_DIR="${PREFIX}/share"

case "${1:-}" in
--uninstall)
  rm -rf "${BIN_DIR}/godraw" \
    "${SHARE_DIR}/applications/godraw.desktop" \
    "${SHARE_DIR}/godraw" \
    "${SHARE_DIR}/mime/packages/godraw.xml"
  echo "godraw uninstalled."
  exit 0
  ;;
"") ;;
*) VERSION="$1" ;;
esac

# -- Pick a downloader and the platform. --------------------------------------

if command -v curl >/dev/null 2>&1; then
  fetch() { curl -fsSL -o "$2" "$1"; }
elif command -v wget >/dev/null 2>&1; then
  fetch() { wget -qO "$2" "$1"; }
else
  echo "error: need curl or wget" >&2
  exit 1
fi

case "$(uname -m)" in
x86_64) ARCH="amd64" ;;
aarch64 | arm64) ARCH="arm64" ;;
*)
  echo "error: unsupported architecture: $(uname -m)" >&2
  exit 1
  ;;
esac

# -- Work out which version to fetch. -----------------------------------------

if [ -z "${VERSION:-}" ]; then
  VERSION=$(fetch "https://api.github.com/repos/${REPO}/releases/latest" - | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n 1)
  if [ -z "$VERSION" ]; then
    echo "error: no release found at https://github.com/${REPO}/releases" >&2
    exit 1
  fi
fi
ASSET="godraw-${VERSION#v}-linux-${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
echo "Downloading ${REPO} ${VERSION} (${ARCH})..."
if ! fetch "$URL" "${TMP}/${ASSET}"; then
  echo "error: ${ASSET} not found — check https://github.com/${REPO}/releases" >&2
  exit 1
fi

# The checksums file names assets under their build directory; strip the
# directory so it matches the file we downloaded.
fetch "https://github.com/${REPO}/releases/download/${VERSION}/sha256sums.txt" "${TMP}/sha256sums.txt"
if ! grep "${ASSET}" "${TMP}/sha256sums.txt" | sed 's|linux-[a-z0-9]*/||' | (cd "$TMP" && sha256sum -c -) >/dev/null 2>&1; then
  echo "error: checksum verification failed for ${ASSET}" >&2
  exit 1
fi

# -- Install. -----------------------------------------------------------------

APP_DIR="${SHARE_DIR}/godraw"
mkdir -p "${BIN_DIR}" "${SHARE_DIR}/applications" "${APP_DIR}" "${SHARE_DIR}/mime/packages"
tar -xzf "${TMP}/${ASSET}" -C "$TMP"

install -m 755 "${TMP}/godraw" "${BIN_DIR}/godraw"
install -m 644 "${TMP}/godraw.desktop" "${SHARE_DIR}/applications/godraw.desktop"
install -m 644 "${TMP}/godraw.png" "${APP_DIR}/godraw.png"
[ -f "${TMP}/godraw.xml" ] && install -m 644 "${TMP}/godraw.xml" "${SHARE_DIR}/mime/packages/godraw.xml"

# The packaged desktop entry runs bare `godraw` and names the icon, which
# both fail outside a system install; point them at the installed paths.
sed -i -e "s|^Exec=.*|Exec=${BIN_DIR}/godraw %U|" \
  -e "s|^Icon=.*|Icon=${APP_DIR}/godraw.png|" \
  "${SHARE_DIR}/applications/godraw.desktop"

command -v update-desktop-database >/dev/null 2>&1 && update-desktop-database "${SHARE_DIR}/applications"
command -v update-mime-database >/dev/null 2>&1 && update-mime-database "${SHARE_DIR}/mime"

case ":${PATH}:" in
*":${BIN_DIR}:"*) ;;
*)
  echo
  echo "Add ${BIN_DIR} to your PATH, e.g.:"
  echo "  echo 'export PATH=\"${BIN_DIR}:\$PATH\"' >> ~/.profile"
  ;;
esac

echo "Installed godraw ${VERSION#v} to ${BIN_DIR}/godraw"
echo "Uninstall any time with: install.sh --uninstall"
