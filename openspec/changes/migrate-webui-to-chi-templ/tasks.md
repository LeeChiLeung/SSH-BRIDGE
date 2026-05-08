## 1. UI Migration Foundation

- [x] 1.1 Add `chi` and `templ` to the project and create the root `ui/` directory structure for routes, pages, components, and assets.
- [x] 1.2 Define the program entry strategy for legacy vs new WebUI, including how the application selects which UI implementation to serve.

## 2. Shared Integration Boundary

- [x] 2.1 Introduce the minimal adapter layer needed for the new UI to consume the existing tunnel controller, profile store, and runtime view data without duplicating business logic.
- [x] 2.2 Document and enforce the rule that new UI work must reuse shared core modules rather than reimplementing tunnel/profile behavior inside `ui/`.

## 3. First Migration Slice

- [x] 3.1 Port one low-risk page or panel, such as the list home page or shell layout, into the new `chi + templ` UI stack.
- [x] 3.2 Verify that legacy UI remains functional while the migrated page is served through the new stack.

## 4. Future-Ready Structure

- [x] 4.1 Establish component and layout conventions in `ui/` that are suitable for later i18n work and page-by-page migration.
- [x] 4.2 Add tests or validation steps that cover new UI routing, legacy fallback behavior, and shared-core integration boundaries.
