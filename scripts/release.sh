#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_NAME="${APP_NAME:-timber-ssh-go}"
DIST_DIR="${DIST_DIR:-$ROOT_DIR/dist}"
WORK_DIR="$DIST_DIR/.work"
DEFAULT_TARGETS="linux/amd64 linux/arm64 windows/amd64"
TARGETS="${TARGETS:-$DEFAULT_TARGETS}"
GOFLAGS_RELEASE="${GOFLAGS_RELEASE:--trimpath -buildvcs=false}"
APPIMAGETOOL_BIN="${APPIMAGETOOL_BIN:-appimagetool}"
APPIMAGE_RUNTIME_FILE="${APPIMAGE_RUNTIME_FILE:-$ROOT_DIR/public/obsolete-runtime-x86_64}"
APPIMAGE_COMP="${APPIMAGE_COMP:-xz}"

usage() {
  cat <<'EOF'
Usage:
  scripts/release.sh build
  scripts/release.sh deb [arch]
  scripts/release.sh appimage [arch]
  scripts/release.sh all

Environment:
  VERSION            Artifact version string. If omitted, derive from git tag/describe
  DIST_DIR           Output directory. Default: ./dist
  TARGETS            Space-separated build matrix. Default:
                     linux/amd64 linux/arm64 windows/amd64
  APPIMAGETOOL_BIN   AppImage tool path override. Default: appimagetool
  APPIMAGE_RUNTIME_FILE
                     Local AppImage runtime file passed to appimagetool.
                     Default: ./public/obsolete-runtime-x86_64
  APPIMAGE_COMP      Squashfs compression for AppImage. Default: xz
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

resolve_version() {
  if [[ -n "${VERSION:-}" ]]; then
    printf '%s' "$(printf '%s' "$VERSION" | sed 's/^v//')"
    return
  fi
  if command -v git >/dev/null 2>&1 && git -C "$ROOT_DIR" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    local described
    described="$(git -C "$ROOT_DIR" describe --tags --dirty --always 2>/dev/null || true)"
    if [[ -n "$described" ]]; then
      printf '%s' "$(printf '%s' "$described" | sed 's/^v//')"
      return
    fi
  fi
  printf '0.0.0-dev'
}

resolve_commit() {
  if command -v git >/dev/null 2>&1 && git -C "$ROOT_DIR" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    git -C "$ROOT_DIR" rev-parse --short=12 HEAD 2>/dev/null || printf 'unknown'
    return
  fi
  printf 'unknown'
}

resolve_build_date() {
  date -u +"%Y-%m-%dT%H:%M:%SZ"
}

VERSION="$(resolve_version)"
BUILD_COMMIT="${BUILD_COMMIT:-$(resolve_commit)}"
BUILD_DATE="${BUILD_DATE:-$(resolve_build_date)}"

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
  local ldflags
  out="$(binary_path "$os" "$arch")"
  ldflags="-X main.buildVersion=$VERSION -X main.buildCommit=$BUILD_COMMIT -X main.buildDate=$BUILD_DATE"
  if [[ "$os" == "windows" ]]; then
    ldflags="$ldflags -H=windowsgui"
  fi
  ensure_dir "$DIST_DIR"
  log "building $os/$arch -> $out"
  GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build $GOFLAGS_RELEASE -ldflags "$ldflags" -o "$out" .
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
  build_target linux "$arch"
  binary="$(binary_path linux "$arch")"

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
  build_target linux "$arch"
  binary="$(binary_path linux "$arch")"

  local work="$WORK_DIR/appimage-${arch}"
  local appdir="$work/AppDir"
  local out="$DIST_DIR/$(versioned_name linux "$arch" ".AppImage")"

  rm -rf "$work"
  ensure_dir "$appdir/usr/bin"
  ensure_dir "$appdir/usr/share/applications"
  ensure_dir "$appdir/usr/share/icons/hicolor/32x32/apps"
  ensure_dir "$appdir/usr/share/icons/hicolor/scalable/apps"
  ensure_dir "$DIST_DIR"

  cp "$binary" "$appdir/usr/bin/$APP_NAME"
  chmod 0755 "$appdir/usr/bin/$APP_NAME"
  cp "$ROOT_DIR/packaging/appimage/$APP_NAME.desktop" "$appdir/usr/share/applications/$APP_NAME.desktop"
  cp "$ROOT_DIR/packaging/appimage/$APP_NAME.desktop" "$appdir/$APP_NAME.desktop"
  if [[ -f "$ROOT_DIR/icon/32x32.png" ]]; then
    cp "$ROOT_DIR/icon/32x32.png" "$appdir/usr/share/icons/hicolor/32x32/apps/$APP_NAME.png"
  fi
  cp "$ROOT_DIR/packaging/appimage/$APP_NAME.svg" "$appdir/usr/share/icons/hicolor/scalable/apps/$APP_NAME.svg"
  cp "$ROOT_DIR/packaging/appimage/$APP_NAME.svg" "$appdir/$APP_NAME.svg"
  ln -sf "$APP_NAME.svg" "$appdir/.DirIcon"
cat >"$appdir/AppRun" <<EOF
#!/usr/bin/env bash
set -euo pipefail
HERE="\$(cd "\$(dirname "\${BASH_SOURCE[0]}")" && pwd)"
if [[ "\$#" -eq 0 ]]; then
  exec "\$HERE/usr/bin/$APP_NAME" --tray --web-addr 127.0.0.1:8848
fi
exec "\$HERE/usr/bin/$APP_NAME" "\$@"
EOF
  chmod 0755 "$appdir/AppRun"

  local tmp_out="$out.tmp"
  rm -f "$tmp_out"

  local -a appimagetool_args
  appimagetool_args=(--comp "$APPIMAGE_COMP" "$appdir" "$tmp_out")
  if [[ -n "$APPIMAGE_RUNTIME_FILE" ]]; then
    [[ -f "$APPIMAGE_RUNTIME_FILE" ]] || fail "missing AppImage runtime file: $APPIMAGE_RUNTIME_FILE"
    appimagetool_args=(--runtime-file "$APPIMAGE_RUNTIME_FILE" "${appimagetool_args[@]}")
  fi

  ARCH="$appimage_arch" "$APPIMAGETOOL_BIN" "${appimagetool_args[@]}"
  mv -f "$tmp_out" "$out"
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
