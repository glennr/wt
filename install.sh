#!/usr/bin/env sh
set -eu

# wt installer
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/glennr/wt/main/install.sh | sh
#   INSTALL_DIR=/usr/local/bin sh install.sh

REPO="glennr/wt"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

main() {
  platform="$(detect_platform)"
  arch="$(detect_arch)"

  echo "Installing wt to $INSTALL_DIR ..."
  mkdir -p "$INSTALL_DIR"

  if try_github_release "$platform" "$arch"; then
    :
  elif has go && source_dir="$(source_tree_dir)"; then
    echo "No prebuilt binary found — building from local source ..."
    go build -C "$source_dir" -o "$INSTALL_DIR/wt" .
  elif has go; then
    echo "No prebuilt binary found — building from source with go install ..."
    GOBIN="$INSTALL_DIR" go install "github.com/$REPO@latest"
  else
    echo "Error: no prebuilt binary for $platform/$arch and Go is not installed."
    echo "Install Go from https://go.dev/dl/ and retry."
    exit 1
  fi

  verify_install
}

try_github_release() {
  platform="$1"
  arch="$2"

  # Need curl or wget
  if ! has curl && ! has wget; then
    return 1
  fi

  # Get latest release tag
  tag="$(get_latest_tag)" || return 1
  [ -z "$tag" ] && return 1

  url="https://github.com/$REPO/releases/download/${tag}/wt_${tag#v}_${platform}_${arch}.tar.gz"

  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' EXIT

  echo "Downloading $url ..."
  if has curl; then
    curl -fsSL "$url" -o "$tmpdir/wt.tar.gz" 2>/dev/null || return 1
  else
    wget -q "$url" -O "$tmpdir/wt.tar.gz" 2>/dev/null || return 1
  fi

  tar -xzf "$tmpdir/wt.tar.gz" -C "$tmpdir"
  install -m 755 "$tmpdir/wt" "$INSTALL_DIR/wt"
  echo "Installed wt $tag"
}

get_latest_tag() {
  if has curl; then
    curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" 2>/dev/null \
      | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"//;s/".*//'
  elif has wget; then
    wget -qO- "https://api.github.com/repos/$REPO/releases/latest" 2>/dev/null \
      | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"//;s/".*//'
  else
    return 1
  fi
}

detect_platform() {
  case "$(uname -s)" in
    Linux*)  echo "linux" ;;
    Darwin*) echo "darwin" ;;
    *)       echo "$(uname -s | tr '[:upper:]' '[:lower:]')" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)  echo "amd64" ;;
    aarch64|arm64)  echo "arm64" ;;
    armv7l)         echo "armv7" ;;
    *)              echo "$(uname -m)" ;;
  esac
}

verify_install() {
  if "$INSTALL_DIR/wt" --version >/dev/null 2>&1; then
    echo "OK: $("$INSTALL_DIR/wt" --version)"
  else
    echo "OK: wt installed to $INSTALL_DIR/wt"
  fi

  # PATH hint
  case ":${PATH}:" in
    *":$INSTALL_DIR:"*) ;;
    *)
      echo ""
      echo "Add $INSTALL_DIR to your PATH:"
      echo "  echo 'export PATH=\"$INSTALL_DIR:\$PATH\"' >> ~/.bashrc"
      ;;
  esac
}

has() {
  command -v "$1" >/dev/null 2>&1
}

source_tree_dir() {
  # Return the source tree path if install.sh is run from within it
  dir="$(cd "$(dirname "$0")" 2>/dev/null && pwd)" || return 1
  if [ -f "$dir/go.mod" ] && grep -q "module github.com/$REPO" "$dir/go.mod" 2>/dev/null; then
    echo "$dir"
    return 0
  fi
  return 1
}

main
