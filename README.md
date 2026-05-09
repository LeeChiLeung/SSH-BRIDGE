# Timber SSH Go

A local SSH tunneling tool written in Go, featuring local port forwarding, multi-instance profiles, dual Web consoles (Legacy & Modern), system tray integration, and distribution artifacts for Linux and Windows.

## Use Cases

* Forwarding local ports to remote intranet services via an SSH jump server.
* Saving multiple tunnel profiles and managing them via a list or dashboard.
* Monitoring connection status, reconnection logs, health checks, up/down speeds, and cumulative traffic locally.
* Importing/Exporting profile configurations via JSON.
* Packaging into standalone binaries, `.deb` packages, or `AppImage`.

## Core Features

* **Local Port Forwarding**: Listen on local TCP ports and forward traffic to target addresses behind SSH.
* **Connection Persistence**: Exponential backoff reconnection, SSH keep-alive, and half-open connection detection.
* **Multi-instance Management**: Manage multiple tunnel profiles within a single process.
* **Local WebUI**: Create, edit, clone, delete, start, stop, import, and export profiles.
* **System Tray**: Quick access to the console and application exit.
* **Runtime Metrics**: Real-time up/down speeds, total traffic, active connection count, and runtime logs.
* **Version Injection**: `--version` flag outputs version, commit hash, and build timestamp.

## Requirements

* Go 1.20+
* Linux or Windows (for the default build matrix)
* `appimagetool` (required for AppImage packaging)
* `ar`, `tar`, `gzip` (required for `.deb` packaging)

## Quick Start

### Run a single tunnel directly:

```bash
go run . \
  --ssh-host ssh.example.com \
  --ssh-port 22 \
  --ssh-user user \
  --ssh-key ~/.ssh/example_id_rsa \
  --insecure-ignore-host-key \
  --local-addr 127.0.0.1 \
  --local-port 15432 \
  --remote-host 127.0.0.1 \
  --remote-port 5432

```

This is equivalent to the following SSH command:

```bash
ssh -p 22 -N -C -L 127.0.0.1:15432:127.0.0.1:5432 user@ssh.example.com

```

### Check version:

```bash
go run . --version

```

### Start the Legacy Web Console:

```bash
go run . --web

```

### Start the Modern Web Console:

```bash
go run . --web-new

```

Default access URL: `[http://127.0.0.1:8848](http://127.0.0.1:8848)`

### Start in Tray Mode:

```bash
go run . --tray

```

Tray mode concurrently starts:

* Legacy Console: `[http://127.0.0.1:8848](http://127.0.0.1:8848)`
* Modern Console: `[http://127.0.0.1:8849](http://127.0.0.1:8849)`

*Note: The "Open Console" option in the tray menu currently opens the Modern Console.*

## Common Run Modes

### 1. Direct Tunnel Mode

Best for single CLI-based forwarding:

```bash
go run . \
  --ssh-host ssh.example.com \
  --ssh-port 22 \
  --ssh-user user \
  --ssh-key ~/.ssh/example_id_rsa \
  --remote-host 127.0.0.1 \
  --remote-port 5432 \
  --local-addr 127.0.0.1 \
  --local-port 15432

```

### 2. `--env` Port Injection

If using `0` for the local port (auto-assignment), use `--env` to output the shell export command:

```bash
eval "$(go run . \
  --ssh-host ssh.example.com \
  --ssh-user user \
  --ssh-key ~/.ssh/example_id_rsa \
  --local-port 0 \
  --env)"

```

The program will output something like: `export TUNNEL_LOCAL_PORT=43017`.

### 3. Web Console Modes

* **Legacy**: `go run . --web --web-addr 127.0.0.1:8848`
* **Modern**: `go run . --web-new --web-addr 127.0.0.1:8848`

**Modern WebUI Features:**

* Dashboard with tunnel profile lists.
* Control panel for Start/Stop/Edit/Clone/Delete.
* JSON Import/Export.
* Real-time status: `disconnected`, `connecting`, `connected`, `retrying`.
* Status indicators: Gray (Inactive), Blue (Connecting), Green (Success), Red (Failed).
* Health Checks: Listener status and SSH link status.
* Log panel, traffic stats, and active connection monitoring.

### 4. Tray Mode

```bash
go run . --tray --web-addr 127.0.0.1:8848

```

* In Windows, double-clicking the `.exe` defaults to Tray Mode using the GUI subsystem (no console window).

## CLI Arguments

Key arguments include:

* SSH: `--ssh-host`, `--ssh-port`, `--ssh-user`, `--ssh-key`, `--ssh-password`
* Remote: `--remote-host`, `--remote-port`
* Local: `--local-addr`, `--local-port`
* Security: `--insecure-ignore-host-key`, `--known-hosts`
* Modes: `--env`, `--web`, `--web-new`, `--tray`, `--web-addr`
* Info: `--version`, `--help`

## Configuration Storage

The WebUI persists tunnel configurations in a profile file.

**Default Location:**

* **Linux**: `$XDG_CONFIG_HOME/timber-ssh-go/profiles.json`
* **Fallback**: `./profiles.json` in the current directory.

*Note: Profiles save SSH passwords and private key paths, but not the actual contents of the private key files.*

### JSON Import/Export

The Modern Console allows exporting all profiles to a single JSON file. During import:

* A pre-check is performed.
* Conflicts are identified by tunnel name (title).
* Users can choose to "Skip" or "Merge" on conflict.
* A summary (added/merged/skipped/invalid) is displayed upon completion.

## Build and Release

Default build matrix:

```bash
./build.sh

```

Explicit build commands:

```bash
./build.sh build            # Build standalone binary
./build.sh deb amd64        # Build .deb package for amd64
./build.sh appimage amd64   # Build AppImage for amd64

```

Supported platforms: `linux/amd64`, `linux/arm64`, `windows/amd64`. Artifacts are output to the `dist/` directory.

### Versioning

The build script handles versioning automatically:

* Uses the `VERSION` environment variable if provided.
* Otherwise, defaults to `git describe --tags --dirty --always`.

Example:

```bash
VERSION=0.0.2 ./scripts/release.sh all

```

## Known Limitations

* **macOS Support**: Not included in the default build matrix due to tray icon dependencies that hinder cross-compilation from Linux.
* **AppImage**: Packaging depends on local `appimagetool` and specific runtime files.
* **Documentation**: CLI/WebUI defaults are subject to change; check `--help` for the most up-to-date info.

## Project Status

The core functionality is complete, including SSH forwarding, multi-profile management, dual Web consoles, tray integration, traffic metrics, and release scripts for Linux/Windows. For further feature extensions, please refer to the change logs in `openspec/changes/`.
