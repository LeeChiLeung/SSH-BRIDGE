## Why

The project can already run locally, but there is no repeatable release workflow for distributing it to other machines. A first-class packaging flow is needed now so operators can reliably produce multi-platform binaries and Linux installable artifacts without hand-assembling files on every release.

## What Changes

- Add scripted multi-platform builds for the Go binary with a predictable `dist/` output layout.
- Add Debian package generation for Linux distribution and installation workflows.
- Add AppImage packaging for portable Linux desktop-style distribution.
- Define the release artifact structure, naming, and build entrypoints so contributors can produce the same outputs consistently.
- Keep packaging logic outside the application runtime, using build scripts and packaging metadata instead of changing tunnel behavior.

## Capabilities

### New Capabilities
- `release-artifact-builds`: Repeatable multi-platform binary builds with a standardized artifact directory and naming scheme
- `linux-package-distribution`: Debian and AppImage packaging flows for distributing the compiled application on Linux

### Modified Capabilities
- None.

## Impact

- Adds build and packaging scripts, packaging metadata, and release output conventions.
- May add packaging tool configuration such as `nfpm` and AppImage metadata files.
- Affects contributor and release workflows, but not the SSH tunnel runtime behavior itself.
