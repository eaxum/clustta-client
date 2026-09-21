# Clustta 0.4.40 Release Copy

## Release Inputs

- Version: `0.4.40`
- Previous version/tag: `v0.4.39`
- New version/tag: `v0.4.40`
- Compare range: `77923fb287640e029654fffd59bdba1aef85f565...HEAD`
- Release headline: `Precise dependency builds, visible activity, and drag-and-drop exports`
- Canny types: `new`, `improved`, `fixed`

## GitHub Release

### Clustta 0.4.40

## New Features

### Precise Dependency Builds

Pin dependencies to specific checkpoints or follow tagged versions, then preview the exact files and versions a build will download. Conflicts resolve more efficiently, and dependency rows can take you straight to the related asset or collection.

### Checkpoint Tags

Tag checkpoints for use in dependency builds, making it easy to follow a named version instead of selecting a fixed checkpoint.

### Checkpoint Provenance

Record which asset checkpoint a new checkpoint was derived from. Choose a source asset, then use its latest checkpoint or select an exact version. Clustta saves that link with the checkpoint to preserve its lineage across the project.

### Editable Checkpoints

Update an existing checkpoint's comment, tags, and source checkpoint.

### Activity Panel

Track project downloads and other long-running work from the new Activity panel. Review progress and details, minimize the panel while continuing to work, and cancel supported downloads if needed.

### Drag-and-Drop File Exports

Drag assets from Clustta into the system file manager or another supported application. Native drag-out is available on Windows, macOS, and Linux, with platform-appropriate behavior and clearer availability hints.

### Jack of All Trades Template

Start mixed-media projects with a new template containing open-source starter files for Blender, Krita, GIMP, and Audacity, organized with useful tags.

## Improvements

### Safer Project Compatibility and Sync

Clustta now checks project schema and protocol compatibility before operations that could modify a project. Confirmed replicas can migrate safely, local changes are tracked more completely.

### Clearer Downloads and Onboarding

Download progress now uses clearer localized phases and supports cancellation. Storage setup suggests sensible default paths and provides better macOS permission guidance.

### Refined Browsing and Project Controls

- Search supported dropdown lists more quickly and use direct checkboxes in filter menus.
- Choose whether file type icons appear in the browser.
- Use a streamlined list view, animated tab indicators, and clearer shortcut labels.
- Open project settings from the collaborators pane and unassign users from asset details.

## Bug Fixes

- **Authentication** - Preserved project and user state when signing in again after a session expires.
- **Dependencies** - Fixed checkpoint build permissions, dependency graph views, extension visibility, and conflict resolution.
- **Sync** - Preserved custom asset types and icons when making a personal project remote and improved discard errors for local changes.
- **Projects** - Fixed folder reveal actions so they open the selected project's folder.
- **Assignments** - Handled deleted assignees correctly and improved assignee controls in asset details.
- **Interface** - Improved default sizing, dark theme defaults, filters, skeleton spacing, responsive empty states, and other visual details.

**Full Changelog**: `77923fb287640e029654fffd59bdba1aef85f565...HEAD`

## Canny Changelog

### Title

Clustta 0.4.40 : Precise dependency builds, visible activity, and drag-and-drop exports

### Types

`new, improved, fixed`

### Body

Clustta 0.4.40 adds precise dependency version controls, checkpoint provenance, visible background activity, and cross-platform drag-and-drop exports.

**New: Precise Dependency Builds**

Pin dependencies to exact checkpoints or follow tagged versions, preview what a build will download, and move directly from dependency rows to related assets and collections.

**New: Checkpoint Tags**

Tag checkpoints for dependency builds so dependencies can follow a named version.

**New: Checkpoint Provenance**

Record where a checkpoint came from by choosing a source asset and linking either its latest checkpoint or an exact version. Clustta stores the link as part of the checkpoint's lineage.

**New: Editable Checkpoints**

Update an existing checkpoint's comment, tags, and source checkpoint without recreating it.

**New: Activity Panel**

Follow project downloads and other long-running work, inspect details, minimize the panel, and cancel supported downloads.

**New: Drag-and-Drop File Exports**

Drag eligible assets and collections into the system file manager or another supported application on Windows, macOS, and Linux.

**Improved: Compatibility, Sync, and Onboarding**

Benefit from safer project compatibility checks and migrations, more complete local change tracking, synced dependency selections, clearer download progress, and better storage setup guidance.

**Fixed**

- Fixed expired-session reauthentication without losing project context.
- Fixed dependency permissions, graph views, conflict resolution, and extension visibility.
- Fixed custom type and icon preservation when moving personal projects online.
- Fixed selected project folder reveal actions, deleted assignees, and several interface details.

## Apple App Store

### What's New in This Version

Clustta 0.4.40 brings more precise dependency builds, visible activity, drag-and-drop exports, and safer project updates.

- Pin dependencies to exact checkpoints or tagged versions and preview build downloads.
- Tag checkpoints, record their source lineage, and edit checkpoint details after creation.
- Track downloads and other long-running work in the new Activity panel.
- Drag eligible assets and collections into supported apps and file locations.
- Start mixed-media work with the new Jack of All Trades project template.
- Includes improvements to project compatibility, sync, onboarding, authentication, and the interface.

## Microsoft Store

### Release Notes

Clustta 0.4.40 brings more precise dependency builds, visible activity, drag-and-drop exports, and safer project updates.

- Pin dependencies to exact checkpoints or tagged versions and preview build downloads.
- Tag checkpoints, record their source lineage, and edit checkpoint details after creation.
- Track downloads and other long-running work in the new Activity panel.
- Drag eligible assets and collections into supported apps and file locations.
- Start mixed-media work with the new Jack of All Trades project template.
- Includes improvements to project compatibility, sync, onboarding, authentication, and the interface.

## Short Store Summary

Precise dependency builds, checkpoint provenance, visible activity, and drag-and-drop exports.

## Flathub Release

### Clustta 0.4.40

Clustta 0.4.40 adds precise dependency version controls, checkpoint provenance, visible background activity, cross-platform drag-and-drop exports, and safer project updates.

- Pin dependencies to exact checkpoints or tagged versions and preview build downloads.
- Tag checkpoints, record their source lineage, and edit checkpoint details after creation.
- Track downloads and other long-running work in the new Activity panel.
- Drag eligible assets and collections into supported apps and file locations on Windows, macOS, and Linux.
- Start mixed-media work with the new Jack of All Trades project template.
- Includes improvements to project compatibility, sync, onboarding, authentication, dependency navigation, and interface polish.
