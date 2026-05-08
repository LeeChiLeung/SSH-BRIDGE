## Context

The application is a Go binary with a lightweight local Web console and SSH runtime logic. That makes distribution relatively straightforward at the executable level, but there is currently no repository-owned release pipeline that defines which platforms to build, how to lay out artifacts, or how Linux installable packages should be assembled.

Packaging introduces a second layer of concerns beyond compilation: archive naming, package metadata, filesystem layout, desktop integration for AppImage, and packaging tool availability. Those concerns should be isolated into scripts and metadata files so application code remains focused on tunnel behavior.

## Goals / Non-Goals

**Goals:**
- Provide one or more repository scripts that build release artifacts into a predictable `dist/` directory.
- Produce cross-platform binaries for the primary targets without requiring manual command construction each release.
- Produce a `.deb` package that installs the application binary and related packaging assets in a consistent Linux filesystem layout.
- Produce an `AppImage` artifact for portable Linux distribution.
- Make the packaging workflow reproducible enough for local release use and future CI automation.

**Non-Goals:**
- Implementing platform-native installers for Windows or macOS in this change.
- Shipping automatic update infrastructure.
- Signing packages or notarizing binaries in the first version.
- Embedding user-specific runtime configuration such as private keys or saved profiles into package outputs.

## Decisions

Use repository-managed shell scripts as the primary build entrypoint.
Rationale: The project is already Go-first and lightweight. Shell scripts keep the first release workflow transparent and easy to run locally without introducing a large external release framework.
Alternatives considered:
- A GoReleaser migration was rejected for the first step because it adds more abstraction than needed before the artifact shapes are settled.

Standardize all outputs under `dist/` with per-target file naming.
Rationale: Contributors need one obvious place to inspect release outputs, and packaging steps need stable input/output paths.
Alternatives considered:
- Writing artifacts into ad hoc temporary directories was rejected because it makes verification and cleanup harder.

Use a declarative Debian packaging tool configuration plus a script wrapper.
Rationale: Debian packaging needs metadata such as package name, version, architecture, install paths, and optional service or desktop assets. A declarative config keeps those values reviewable and easier to evolve.
Alternatives considered:
- Hand-writing `dpkg-deb` directory trees in shell only was rejected because metadata becomes harder to maintain cleanly.

Build AppImage from an explicit `AppDir` assembly step.
Rationale: AppImage packaging expects a filesystem tree with binary and desktop metadata. Building that tree explicitly keeps the process understandable and easier to debug.
Alternatives considered:
- Using opaque one-command wrappers without repository metadata was rejected because troubleshooting would be harder.

Keep versioning script-driven, with a default development version that can be overridden by environment variable or flag.
Rationale: Packaging output names need a version string even before a full release tagging workflow exists.
Alternatives considered:
- Hardcoding version numbers in packaging files was rejected because it creates unnecessary edit churn.

## Risks / Trade-offs

- [Packaging scripts may depend on tools not installed on every machine] → Fail early with clear checks and document required tools in the script output.
- [Cross-compilation targets may drift from actual user needs] → Start with a small explicit matrix and keep scripts easy to extend.
- [AppImage desktop metadata may be incomplete for some desktop environments] → Ship a pragmatic first version focused on portable execution rather than deep desktop integration.
- [Package outputs can accidentally include local files] → Build from explicit inputs only and keep runtime data directories outside packaging manifests.

## Migration Plan

1. Add release script(s) and metadata files that define binary targets and package inputs.
2. Add Debian packaging configuration and AppImage assembly metadata.
3. Build artifacts into `dist/` and verify expected file names and layouts.
4. Document or encode tool checks so future local or CI runs fail predictably when prerequisites are missing.

Rollback is straightforward: remove the packaging scripts and metadata without affecting the application runtime.

## Open Questions

- Should the `.deb` package install a desktop entry for launching the local Web console, or should the first version stay CLI-first and only package the binary?
