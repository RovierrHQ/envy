#!/usr/bin/env sh
# Install envy from GitHub Releases into ~/.local/bin and optionally update PATH.
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/RovierrHQ/envy/main/scripts/install.sh | sh
#   ENVY_VERSION=v0.2.0 sh install.sh
#   sh install.sh --no-path

set -eu

REPO="RovierrHQ/envy"
INSTALL_DIR="${ENVY_INSTALL_DIR:-$HOME/.local/bin}"
NO_PATH=0
VERSION="${ENVY_VERSION:-latest}"

for arg in "$@"; do
  case "$arg" in
    --no-path) NO_PATH=1 ;;
    --version=*) VERSION="${arg#--version=}" ;;
    -h|--help)
      echo "Usage: install.sh [--no-path] [--version=vX.Y.Z]"
      echo "Env:   ENVY_VERSION, ENVY_INSTALL_DIR"
      exit 0
      ;;
    *)
      echo "unknown argument: $arg" >&2
      exit 1
      ;;
  esac
done

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing required command: $1" >&2
    exit 1
  }
}

need curl
need tar
need mktemp

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  linux|darwin) ;;
  *)
    echo "unsupported OS: $os (use install.ps1 on Windows)" >&2
    exit 1
    ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *)
    echo "unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

if [ "$VERSION" = "latest" ]; then
  tag=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/${REPO}/releases/latest")
  tag=${tag%/}
  tag=${tag##*/}
else
  case "$VERSION" in
    v*) tag="$VERSION" ;;
    *) tag="v${VERSION}" ;;
  esac
fi

if [ -z "$tag" ] || [ "$tag" = "latest" ]; then
  echo "failed to resolve release tag" >&2
  exit 1
fi

ver=${tag#v}
asset="envy_${ver}_${os}_${arch}.tar.gz"
base="https://github.com/${REPO}/releases/download/${tag}"
tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

old_ver=""
if [ -x "${INSTALL_DIR}/envy" ]; then
  old_ver=$("${INSTALL_DIR}/envy" version 2>/dev/null || true)
fi

if [ -n "$old_ver" ]; then
  echo "Upgrading envy ${old_ver} -> ${tag} (${os}/${arch}) -> ${INSTALL_DIR}"
else
  echo "Installing envy ${tag} (${os}/${arch}) -> ${INSTALL_DIR}"
fi

curl -fsSL "${base}/${asset}" -o "${tmpdir}/${asset}"
curl -fsSL "${base}/checksums.txt" -o "${tmpdir}/checksums.txt"

expected=$(awk -v f="$asset" '$2 == f { print $1; exit }' "${tmpdir}/checksums.txt")
if [ -z "$expected" ]; then
  echo "checksum not found for ${asset}" >&2
  exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "${tmpdir}/${asset}" | awk '{ print $1 }')
elif command -v shasum >/dev/null 2>&1; then
  actual=$(shasum -a 256 "${tmpdir}/${asset}" | awk '{ print $1 }')
else
  echo "need sha256sum or shasum to verify download" >&2
  exit 1
fi

if [ "$actual" != "$expected" ]; then
  echo "checksum mismatch for ${asset}" >&2
  echo "  expected: ${expected}" >&2
  echo "  actual:   ${actual}" >&2
  exit 1
fi

tar -xzf "${tmpdir}/${asset}" -C "$tmpdir"
if [ ! -f "${tmpdir}/envy" ]; then
  echo "archive missing envy binary" >&2
  exit 1
fi

mkdir -p "$INSTALL_DIR"
install -m 755 "${tmpdir}/envy" "${INSTALL_DIR}/envy"

path_line='export PATH="$HOME/.local/bin:$PATH"'
if [ "$INSTALL_DIR" != "$HOME/.local/bin" ]; then
  path_line="export PATH=\"${INSTALL_DIR}:\$PATH\""
fi

append_path_file() {
  file=$1
  marker="# envy PATH"
  if [ -f "$file" ] && grep -Fq "$marker" "$file" 2>/dev/null; then
    return 0
  fi
  mkdir -p "$(dirname "$file")"
  {
    echo ""
    echo "$marker"
    echo "$path_line"
  } >>"$file"
  echo "Updated PATH in ${file}"
}

if [ "$NO_PATH" -eq 0 ]; then
  case "$(basename "${SHELL:-}")" in
    zsh) append_path_file "$HOME/.zshrc" ;;
    bash)
      if [ -f "$HOME/.bashrc" ] || [ ! -f "$HOME/.bash_profile" ]; then
        append_path_file "$HOME/.bashrc"
      else
        append_path_file "$HOME/.bash_profile"
      fi
      ;;
    fish)
      fish_cfg="$HOME/.config/fish/config.fish"
      marker="# envy PATH"
      if [ ! -f "$fish_cfg" ] || ! grep -Fq "$marker" "$fish_cfg" 2>/dev/null; then
        mkdir -p "$(dirname "$fish_cfg")"
        {
          echo ""
          echo "$marker"
          echo "fish_add_path ${INSTALL_DIR}"
        } >>"$fish_cfg"
        echo "Updated PATH in ${fish_cfg}"
      fi
      ;;
    *)
      append_path_file "$HOME/.profile"
      ;;
  esac
  case ":${PATH}:" in
    *":${INSTALL_DIR}:"*) ;;
    *) export PATH="${INSTALL_DIR}:${PATH}" ;;
  esac
fi

new_ver=$("${INSTALL_DIR}/envy" version 2>/dev/null || echo "$ver")
if [ -n "$old_ver" ]; then
  echo "Upgraded ${INSTALL_DIR}/envy (${old_ver} -> ${new_ver})"
else
  echo "Installed ${INSTALL_DIR}/envy (${new_ver})"
fi
echo "Done. Open a new terminal (or re-source your shell profile) if envy is not found."
echo "Upgrade anytime by re-running this script. Uninstall: scripts/uninstall.sh"
