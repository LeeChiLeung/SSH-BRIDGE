#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_NAME="${APP_NAME:-timber-ssh-go}"
VERSION="${VERSION:-0.0.0-dev}"
DIST_DIR="${DIST_DIR:-$ROOT_DIR/dist}"
WORK_DIR="$DIST_DIR/.work"
DEFAULT_TARGETS="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64"
TARGETS="${TARGETS:-$DEFAULT_TARGETS}"
GOFLAGS_RELEASE="${GOFLAGS_RELEASE:--trimpath -buildvcs=false}"
APPIMAGETOOL_BIN="${APPIMAGETOOL_BIN:-appimagetool}"
APPIMAGE_RUNTIME_FILE="${APPIMAGE_RUNTIME_FILE:-}"

usage() {
  cat <<'EOF'
Usage:
  scripts/release.sh build
  scripts/release.sh deb [arch]
  scripts/release.sh appimage [arch]
  scripts/release.sh all

Environment:
  VERSION            Artifact version string. Default: 0.0.0-dev
  DIST_DIR           Output directory. Default: ./dist
  TARGETS            Space-separated build matrix. Default:
                     linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64
  APPIMAGETOOL_BIN   AppImage tool path override. Default: appimagetool
  APPIMAGE_RUNTIME_FILE
                     Optional local AppImage runtime file passed to appimagetool
  KEEP_WORK          Preserve intermediate packaging work directories when set to 1
EOF
}

log() {
  printf '[release] %s\n' "$*"
}

fail() {
  printf '[release] error: %s\n' "$*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required tool: $1"
}

sanitize_version() {
  printf '%s' "$1" | tr '/ ' '--'
}

versioned_name() {
  local os="$1"
  local arch="$2"
  local ext="${3:-}"
  printf '%s_%s_%s_%s%s' "$APP_NAME" "$(sanitize_version "$VERSION")" "$os" "$arch" "$ext"
}

binary_path() {
  local os="$1"
  local arch="$2"
  local ext=""
  if [[ "$os" == "windows" ]]; then
    ext=".exe"
  fi
  printf '%s/%s' "$DIST_DIR" "$(versioned_name "$os" "$arch" "$ext")"
}

map_deb_arch() {
  case "$1" in
    amd64|arm64) printf '%s' "$1" ;;
    *) fail "unsupported Debian architecture: $1" ;;
  esac
}

map_appimage_arch() {
  case "$1" in
    amd64) printf 'x86_64' ;;
    arm64) printf 'aarch64' ;;
    *) fail "unsupported AppImage architecture: $1" ;;
  esac
}

ensure_dir() {
  mkdir -p "$1"
}

cleanup_workdir() {
  local dir="$1"
  if [[ "${KEEP_WORK:-0}" == "1" ]]; then
    log "keeping work directory $dir"
    return
  fi
  rm -rf "$dir"
}

build_target() {
  local os="$1"
  local arch="$2"
  local out
  out="$(binary_path "$os" "$arch")"
  ensure_dir "$DIST_DIR"
  log "building $os/$arch -> $out"
  GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build $GOFLAGS_RELEASE -o "$out" .
}

build_all() {
  require_cmd go
  ensure_dir "$DIST_DIR"
  for target in $TARGETS; do
    local os="${target%/*}"
    local arch="${target#*/}"
    [[ -n "$os" && -n "$arch" ]] || fail "invalid target entry: $target"
    build_target "$os" "$arch"
  done
}

render_deb_control() {
  local arch="$1"
  sed \
    -e "s/__VERSION__/$(sanitize_version "$VERSION")/g" \
    -e "s/__ARCH__/$arch/g" \
    "$ROOT_DIR/packaging/deb/control.tpl"
}

package_deb() {
  require_cmd ar
  require_cmd tar
  require_cmd gzip

  local arch="${1:-amd64}"
  local deb_arch
  deb_arch="$(map_deb_arch "$arch")"
  local binary
  binary="$(binary_path linux "$arch")"
  [[ -f "$binary" ]] || build_target linux "$arch"

  local pkg_name
  pkg_name="$(versioned_name linux "$arch" ".deb")"
  local out="$DIST_DIR/$pkg_name"
  local work="$WORK_DIR/deb-${arch}"
  local data_dir="$work/data"
  local control_dir="$work/control"

  rm -rf "$work"
  ensure_dir "$data_dir/usr/bin"
  ensure_dir "$control_dir"
  ensure_dir "$DIST_DIR"

  cp "$binary" "$data_dir/usr/bin/$APP_NAME"
  chmod 0755 "$data_dir/usr/bin/$APP_NAME"
  render_deb_control "$deb_arch" >"$control_dir/control"
  printf '2.0\n' >"$work/debian-binary"

  (
    cd "$control_dir"
    tar -czf "$work/control.tar.gz" .
  )
  (
    cd "$data_dir"
    tar -czf "$work/data.tar.gz" .
  )
  rm -f "$out"
  (
    cd "$work"
    ar r "$out" debian-binary control.tar.gz data.tar.gz >/dev/null
  )
  cleanup_workdir "$work"
  log "created $out"
}

package_appimage() {
  require_cmd "$APPIMAGETOOL_BIN"

  local arch="${1:-amd64}"
  local appimage_arch
  appimage_arch="$(map_appimage_arch "$arch")"
  local binary
  binary="$(binary_path linux "$arch")"
  [[ -f "$binary" ]] || build_target linux "$arch"

  local work="$WORK_DIR/appimage-${arch}"
  local appdir="$work/AppDir"
  local out="$DIST_DIR/$(versioned_name linux "$arch" ".AppImage")"

  rm -rf "$work"
  ensure_dir "$appdir/usr/bin"
  ensure_dir "$appdir/usr/share/applications"
  ensure_dir "$appdir/usr/share/icons/hicolor/scalable/apps"
  ensure_dir "$DIST_DIR"

  cp "$binary" "$appdir/usr/bin/$APP_NAME"
  chmod 0755 "$appdir/usr/bin/$APP_NAME"
  cp "$ROOT_DIR/packaging/appimage/$APP_NAME.desktop" "$appdir/usr/share/applications/$APP_NAME.desktop"
  cp "$ROOT_DIR/packaging/appimage/$APP_NAME.desktop" "$appdir/$APP_NAME.desktop"
  cp "$ROOT_DIR/packaging/appimage/$APP_NAME.svg" "$appdir/usr/share/icons/hicolor/scalable/apps/$APP_NAME.svg"
  cp "$ROOT_DIR/packaging/appimage/$APP_NAME.svg" "$appdir/$APP_NAME.svg"
  ln -sf "$APP_NAME.svg" "$appdir/.DirIcon"
cat >"$appdir/AppRun" <<EOF
#!/usr/bin/env bash
set -euo pipefail
HERE="\$(cd "\$(dirname "\${BASH_SOURCE[0]}")" && pwd)"
if [[ "\$#" -eq 0 ]]; then
  exec "\$HERE/usr/bin/$APP_NAME" --web --web-addr 127.0.0.1:8848
fi
exec "\$HERE/usr/bin/$APP_NAME" "\$@"
EOF
  chmod 0755 "$appdir/AppRun"

  local -a appimagetool_args
  appimagetool_args=("$appdir" "$out")
  if [[ -n "$APPIMAGE_RUNTIME_FILE" ]]; then
    [[ -f "$APPIMAGE_RUNTIME_FILE" ]] || fail "missing AppImage runtime file: $APPIMAGE_RUNTIME_FILE"
    appimagetool_args=(--runtime-file "$APPIMAGE_RUNTIME_FILE" "${appimagetool_args[@]}")
  fi

  ARCH="$appimage_arch" "$APPIMAGETOOL_BIN" "${appimagetool_args[@]}"
  cleanup_workdir "$work"
  log "created $out"
}

main() {
  local command="${1:-}"
  case "$command" in
    build)
      build_all
      ;;
    deb)
      package_deb "${2:-amd64}"
      ;;
    appimage)
      package_appimage "${2:-amd64}"
      ;;
    all)
      build_all
      package_deb amd64
      package_appimage amd64
      ;;
    ""|-h|--help|help)
      usage
      ;;
    *)
      fail "unknown command: $command"
      ;;
  esac
}

main "$@"
