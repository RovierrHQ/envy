#!/usr/bin/env sh
# Remove envy binary and PATH entries added by install.sh.
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/RovierrHQ/envy/main/scripts/uninstall.sh | sh
#   sh uninstall.sh --yes

set -eu

INSTALL_DIR="${ENVY_INSTALL_DIR:-$HOME/.local/bin}"
BIN="${INSTALL_DIR}/envy"
YES=0

for arg in "$@"; do
  case "$arg" in
    -y|--yes) YES=1 ;;
    -h|--help)
      echo "Usage: uninstall.sh [--yes]"
      echo "Env:   ENVY_INSTALL_DIR"
      exit 0
      ;;
    *)
      echo "unknown argument: $arg" >&2
      exit 1
      ;;
  esac
done

if [ "$YES" -eq 0 ]; then
  printf "Remove envy from %s and strip PATH markers? [y/N] " "$INSTALL_DIR"
  read -r ans || true
  case "$ans" in
    y|Y|yes|YES) ;;
    *)
      echo "Aborted."
      exit 0
      ;;
  esac
fi

if [ -f "$BIN" ] || [ -L "$BIN" ]; then
  rm -f "$BIN"
  echo "Removed ${BIN}"
else
  echo "No binary at ${BIN}"
fi

strip_path_file() {
  file=$1
  marker="# envy PATH"
  [ -f "$file" ] || return 0
  if ! grep -Fq "$marker" "$file" 2>/dev/null; then
    return 0
  fi
  tmp=$(mktemp)
  # Drop marker line and the following PATH line
  awk -v m="$marker" '
    $0 == m { skip=1; next }
    skip { skip=0; next }
    { print }
  ' "$file" >"$tmp"
  mv "$tmp" "$file"
  echo "Cleaned PATH in ${file}"
}

strip_path_file "$HOME/.zshrc"
strip_path_file "$HOME/.bashrc"
strip_path_file "$HOME/.bash_profile"
strip_path_file "$HOME/.profile"

fish_cfg="$HOME/.config/fish/config.fish"
if [ -f "$fish_cfg" ] && grep -Fq "# envy PATH" "$fish_cfg" 2>/dev/null; then
  tmp=$(mktemp)
  awk '
    $0 == "# envy PATH" { skip=1; next }
    skip { skip=0; next }
    { print }
  ' "$fish_cfg" >"$tmp"
  mv "$tmp" "$fish_cfg"
  echo "Cleaned PATH in ${fish_cfg}"
fi

# Remove empty default install dir only
if [ "$INSTALL_DIR" = "$HOME/.local/bin" ] && [ -d "$INSTALL_DIR" ]; then
  rmdir "$INSTALL_DIR" 2>/dev/null || true
fi

echo "Done. Open a new terminal so PATH updates apply."
