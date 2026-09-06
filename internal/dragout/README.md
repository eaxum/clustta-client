# Native file drag-out preview

Windows implementation of copy-only file export, using the shell data-object and OLE sequence from [drag-rs](https://github.com/spacedriveapp/drag-rs/blob/35739c3e5e5b05165ce08947d7148357e8dbe640/crates/drag/src/platform_impl/windows/mod.rs). Reference commit: `35739c3e5e5b05165ce08947d7148357e8dbe640`. Attribution is retained in `LICENSE.drag-rs`; include it when distributing the preview.

Windows uses Go and a C shim against system libraries. macOS uses Go and an Objective-C AppKit shim following the file-URL/session approach in the same drag-rs revision. It adds no third-party dependency or C++ runtime. Linux uses a Go/C GTK3 backend with the same system library dependency already required by Wails. Builds without cgo and other platforms return unsupported. Standalone module extraction is deferred until native application testing passes.

Copyright (c) 2026 Eaxum for the port and integration. See `NOTICE` for attribution and the repository `LICENSE` for Eaxum's contributions. The upstream `LICENSE.drag-rs` notice is preserved separately; include both notices when distributing the adapted code. A future standalone module's license must be selected explicitly when it is extracted.

## Architecture

`Asset.vue` renders `ExportDragHandle.vue` in place of the asset-type icon on hover in both grid and list views, using `grip-vertical` at 0.5 opacity. The asset component passes selection eligibility based on local file state (`normal`, `modified`, or `outdated`); mixed selections containing unavailable files cannot export. `useExportDrag.js` snapshots selected asset IDs and calls the generated `DragOutService` binding. The service reads the trusted calling window from `application.WindowKey`, authenticates the current user, and delegates path resolution to `internal/bridge/assets`.

The request contains `project_path`, `project_id`, and `asset_ids`. The resolver opens the existing local project database read-only, verifies its ID and membership, enforces the existing asset visibility rules without requiring download or structure-editing permissions, and constructs paths from repository metadata. It rejects links, missing files, directories, trashed assets, and canonical paths outside the project root. Pending local renames use the stored local path. It never downloads files or accepts frontend asset paths.

The `wails` subpackage dispatches `dragout.Begin` with `application.InvokeAsync`. Windows delivers completion after OLE returns; macOS returns to the AppKit event loop and completes through `NSDraggingSource`; Linux returns to GTK and completes after `drag-end`. Only the service goroutine waits. A `runtime/cgo.Handle` retains the Go callback until completion, while the native source remains retained for the session. The binding promise resolves to `dropped` or `cancelled`, or rejects with an actionable error. OLE initialization and cleanup run on the same UI thread. An active-session guard stays held until native completion, including after binding cancellation. `dropped` means target acceptance, not completion of the target's import/copy.

## Enable and test

The handle is enabled for Windows, macOS, and Linux desktop development builds when the native backend is available. To include it in a production preview build, set `VITE_NATIVE_DRAG_OUT=true` while building the frontend. Set it to `false` to disable the handle in development too. Production builds without that flag hide it. Unsupported platforms and web builds hide it.

Focused checks:

```powershell
go test ./internal/dragout ./internal/bridge/assets
go test ./services -run TestDragOut
node --test frontend/src/lib/exportDrag.test.js
```

Automated coverage includes a real Windows shell data object for files in different directories, Unicode/spaced paths, missing files, real project-schema resolution, permissions, canonical path containment, and frontend selection rules. It does not simulate an OS drop.

Verification on the implementation machine: focused Go/Node tests and desktop builds pass. The user confirmed drag-out works in Clustta. The file-symlink subtest requires a Windows privilege unavailable here; a directory-junction escape test covers the Windows reparse-point case separately. The web build currently fails on the existing `ExportService` import in `frontend/src/stores/exports.js`, which the web-adapter package does not export; that unrelated adapter remains unchanged.

## Manual release gate

Verify in Clustta:

- Single and multiple files into Explorer/Finder and a creative application; sources remain unchanged.
- Escape, rejected drops, quick mouse release, and repeated drags restore the UI.
- Modifier keys never turn export into a move.
- Grid/list selection, double-click, internal moves, inbound drops, and context menus still work.
- Missing files, mixed collection/asset selections, unavailable permissions, and pending renames behave as described.
- Closing or reloading the source window during a drag does not leave a second session active.

Record OS version and target application versions before enabling the feature by default in production. Custom previews, external-drag return-to-window handoff, file promises, and other platform backends are deferred.

## macOS verification

The AppKit implementation must be compiled and exercised on a Mac with Xcode command-line tools and cgo enabled. This Windows workspace cannot compile or run Cocoa. Run `go test ./internal/dragout`, build the desktop app with the existing Wails workflow, and perform the manual release gate above when changing the macOS backend. The user has confirmed the macOS implementation and manual checks work.

The backend exports existing file URLs with system file icons, allows only copy, ignores modifier-key operation changes, and cancels if the initiating mouse button has already been released. Paths are converted from UTF-8 with `NSURL fileURLWithPath`, including spaces and Unicode. It uses the Wails `NSWindow` content view and constructs the drag event at the current mouse position because the binding arrives asynchronously.

Specifically test Finder and the intended DCCs, multiple files from different directories, Unicode names, Escape, rejected drops, rapid release, repeated sessions, and closing the source window during a drag. Native tests cover input rejection and completion on an invalid window; they do not simulate an AppKit drag session.

## Linux verification

The GTK3 backend starts a manual copy-only drag on the Wails window without configuring automatic GTK drag sources. It publishes percent-encoded `text/uri-list` data, uses the default GTK drag icon, and retains the payload until `drag-end`. Failure marks the session cancelled; window destruction requests cancellation. Cleanup disconnects only the export session's signal handlers.

The asynchronous binding has no original GDK mouse event, so GTK starts with the current pointer position/time and button 1. It checks that the window is active and the left button is still held. This needs separate verification on X11 and native Wayland, where compositor input-serial requirements may affect initiation. No X11-specific APIs are used and XWayland alone does not verify native Wayland.

On a Linux machine with the existing Wails GTK3 development environment, run:

```sh
go test ./internal/dragout -count=1
```

Then build Clustta normally and test single/multiple files into the desktop file manager and intended DCCs, Unicode and URI-special filenames, Escape, rejected drops, quick release, repeated drags, modifier keys, and closing the source window. Sources must remain intact and internal/inbound dragging must continue to work. Test both `GDK_BACKEND=x11` and `GDK_BACKEND=wayland` in compatible desktop sessions; record the compositor, desktop, and target application versions.

Linux tests cover URI round-tripping, missing files, mixed invalid selections, and native invalid-window completion without needing a display. The Windows development machine cannot compile GTK: its Ubuntu WSL environment lacks the compiler/GTK development tools, and Docker is not running. Native compilation and the Linux desktop test matrix remain pending.
