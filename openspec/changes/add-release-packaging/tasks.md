## 1. Add release build scripts and output layout

- [x] 1.1 Create a repository build entrypoint that compiles the application for the chosen multi-platform target matrix.
- [x] 1.2 Standardize release outputs under `dist/` with stable versioned artifact names.
- [x] 1.3 Add prerequisite checks and usage guidance for build and packaging scripts.

## 2. Add Linux packaging workflows

- [x] 2.1 Add repository-managed metadata and script support for generating a `.deb` package.
- [x] 2.2 Add AppDir assembly assets and script support for generating an AppImage artifact.
- [x] 2.3 Ensure packaging workflows consume explicit repository inputs and do not include local runtime data such as saved profiles.

## 3. Validate release artifacts

- [x] 3.1 Verify that the multi-platform build entrypoint produces the expected binaries under `dist/`.
- [x] 3.2 Verify that Debian packaging produces a `.deb` artifact with the expected naming and output location.
- [x] 3.3 Verify that AppImage packaging produces an AppImage artifact with the expected naming and output location.
