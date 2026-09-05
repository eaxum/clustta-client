# Native file drag-out preview

Windows implementation of copy-only file export, using the shell data-object and OLE sequence from [drag-rs](https://github.com/spacedriveapp/drag-rs/blob/35739c3e5e5b05165ce08947d7148357e8dbe640/crates/drag/src/platform_impl/windows/mod.rs). Reference commit: `35739c3e5e5b05165ce08947d7148357e8dbe640`. Attribution is retained in `LICENSE.drag-rs`; include it when distributing the preview.

The implementation uses Go and a C shim against existing Windows system libraries. It adds no third-party dependency or C++ runtime. macOS, Linux, and builds without cgo return unsupported. Standalone module extraction is deferred until native application testing passes.

Copyright (c) 2026 Eaxum for the port and integration. See `NOTICE` for attribution and the repository `LICENSE` for Eaxum's contributions. The upstream `LICENSE.drag-rs` notice is preserved separately; include both notices when distributing the adapted code. A future standalone module's license must be selected explicitly when it is extracted.

## Architecture

`Asset.vue` renders `NativeDragHandle.vue` in both grid and list actions. `useNativeDrag.js` snapshots selected asset IDs and calls the generated `DragOutService` binding. The service reads the trusted calling window from `application.WindowKey`, authenticates the current user, and delegates path resolution to `internal/bridge/assets`.

The request contains `project_path`, `project_id`, and `asset_ids`. The resolver opens the existing local project database read-only, verifies its ID and membership, enforces `pull_chunk` and the existing asset visibility rules, and constructs paths from repository metadata. It rejects links, missing files, directories, trashed assets, and canonical paths outside the project root. Pending local renames use the stored local path. It never downloads files or accepts frontend asset paths.

The `wails` subpackage dispatches native work with `application.InvokeAsync`. The binding promise resolves to `dropped` or `cancelled`, or rejects with an actionable error. OLE initialization and cleanup run on the same UI thread. An active-session guard stays held until the native call returns, including after binding cancellation. `dropped` means target acceptance, not completion of the target's import/copy.

## Enable and test

The handle is enabled for Windows desktop development builds. To include it in a production preview build, set `VITE_NATIVE_DRAG_OUT=true` while building the frontend. Set it to `false` to disable the handle in development too. Production builds without that flag hide it. Other platforms and web builds hide it.

Focused checks:

```powershell
go test ./internal/dragout ./internal/bridge/assets
go test ./services -run TestDragOut
node --test frontend/src/lib/nativeDrag.test.js
```

Automated coverage includes a real Windows shell data object for files in different directories, Unicode/spaced paths, missing files, real project-schema resolution, permissions, canonical path containment, and frontend selection rules. It does not simulate an OS drop.

Verification on the implementation machine: focused Go/Node tests and desktop builds pass. The user confirmed drag-out works in Clustta. The file-symlink subtest requires a Windows privilege unavailable here; a directory-junction escape test covers the Windows reparse-point case separately. The web build currently fails on the existing `ExportService` import in `frontend/src/stores/exports.js`, which the web-adapter package does not export; that unrelated adapter remains unchanged.

## Manual release gate

Verify in Clustta:

- Single and multiple files into Explorer and a creative application; sources remain unchanged.
- Escape, rejected drops, quick mouse release, and repeated drags restore the UI.
- Modifier keys never turn export into a move.
- Grid/list selection, double-click, internal moves, inbound drops, and context menus still work.
- Missing files, mixed collection/asset selections, unavailable permissions, and pending renames behave as described.
- Closing or reloading the source window during a drag does not leave a second session active.

Record Windows version and target application versions before enabling the feature by default in production. Custom previews, external-drag return-to-window handoff, file promises, and other platform backends are deferred.
