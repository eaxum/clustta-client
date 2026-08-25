# Clustta 0.4.39 Release Copy

## Release Inputs

- Version: `0.4.39`
- Previous version/tag: `v0.4.38`
- New version/tag: `v0.4.39`
- Compare range: `9b6f5241ad6ddc60caea51edf617f96f39e0cb53...HEAD`
- Release headline: `Smarter project launches, stronger batch tools, and flexible exports`
- Canny types: `new`, `improved`, `fixed`

## GitHub Release

### Clustta 0.4.39

## Improvements

### Smarter Project Launches

Configure project pre-launch hooks for supported creative applications. Clustta can discover installed DCC versions, prepare project environments, and fetch required hook and environment dependencies before launch.

### More Capable Agent Workflows

The Agent composer now supports `/` quick commands, `~` project script references, and `@` asset or collection references. The Agent can also plan scoped batch moves, renames, dependency changes, copies, and edits with clearer approval previews.

### Pending Renames Across Sync

Local asset and collection renames now remain available while awaiting sync. Clustta preserves pending paths and applies remote path changes when items are fetched, keeping checkpoints, reverts, and local files aligned.

### Flexible Exports and Project Organization

Preview scoped asset exports, choose the columns to include, and save results in multiple file formats. Project settings now also include tag management, role duplication, and a no-tags filter.

### Improved Kitsu Workflows

Open the linked Kitsu task for a selected asset type, filter task outputs, and review selectable full-tree sync previews with clearer import and connection controls.

### Smaller UX Polish

- Checkpoints without messages now receive automatic version comments.
- Dependency paths, workspace labels, title bar controls, navigation, and approval previews are clearer and more consistent.

## Bug Fixes

- **Authentication** - Restored the login prompt when a sync token expires.
- **Project settings** - Fixed project configuration syncing and made environment variable updates safer.
- **Collaborators** - Fixed Studio sync token updates, stale user records, and collaborator limit enforcement.
- **Kitsu integration** - Fixed category metadata parsing and sync preview ordering.
- **Agent tools** - Fixed script resolution, shortcut display, translations, and scoped dependency operations.
- **Cross-platform UI** - Fixed macOS scrollbar and title bar spacing issues, plus Linux build command cleanup.

**Full Changelog**: `9b6f5241ad6ddc60caea51edf617f96f39e0cb53...HEAD`

## Canny Changelog

### Title

Clustta 0.4.39 : Smarter project launches, stronger batch tools, and flexible exports

### Types

`new, improved, fixed`

### Body

Clustta 0.4.39 adds configurable project launch workflows, more capable Agent batch tools, flexible exports, and improved project organization.

**New: Smarter Project Launches**

Configure pre-launch hooks for supported creative applications while Clustta discovers installed versions, prepares project environments, and fetches required dependencies.

**Improved: More Capable Agent Workflows**

Use `/` for quick commands, `~` to reference project scripts, and `@` to reference assets or collections. Scoped batch moves, renames, dependency changes, copies, and edits now have clearer approval previews.

**New: Pending Renames Across Sync**

Keep working with local asset and collection renames while they await sync, with pending and remote paths applied consistently when items are fetched.

**New: Flexible Exports and Organization**

Preview scoped exports, choose columns and file formats, manage project tags, duplicate roles, and filter items without tags.

**Improved: Kitsu Workflows**

Open linked tasks, filter outputs, and review selectable full-tree sync previews with clearer controls.

**Fixed**

- Fixed expired sync token login prompts and safer project setting updates.
- Fixed collaborator records and limits, Kitsu metadata, and preview ordering.
- Fixed Agent script resolution, shortcut display, translations, and cross-platform UI issues.

## Apple App Store

### What's New in This Version

Clustta 0.4.39 improves project launches, batch workflows, exports, and reliability.

- Configure pre-launch hooks and prepare DCC environments and dependencies before launch.
- Use `/` quick commands, `~` script references, and `@` asset or collection references in the Agent.
- Keep working with local renames while they await sync across fetched items, checkpoints, and reverts.
- Preview exports, choose columns and formats, and manage project tags and roles.
- Work more smoothly with Kitsu tasks, output filters, and sync previews.
- Includes fixes for authentication, sync paths, checkpoints, collaborators, and cross-platform UI.

## Microsoft Store

### Release Notes

Clustta 0.4.39 improves project launches, batch workflows, exports, and reliability.

- Configure pre-launch hooks and prepare DCC environments and dependencies before launch.
- Use `/` quick commands, `~` script references, and `@` asset or collection references in the Agent.
- Keep working with local renames while they await sync across fetched items, checkpoints, and reverts.
- Preview exports, choose columns and formats, and manage project tags and roles.
- Work more smoothly with Kitsu tasks, output filters, and sync previews.
- Includes fixes for authentication, sync paths, checkpoints, collaborators, and cross-platform UI.

## Short Store Summary

Smarter project launches, stronger Agent batch tools, flexible exports, and reliability fixes.

## Flathub Release

### Clustta 0.4.39

Clustta 0.4.39 adds smarter project launches, more capable Agent commands, flexible exports, pending renames across sync, and reliability improvements.

- Configure pre-launch hooks and prepare DCC environments and dependencies before launch.
- Use `/` quick commands, `~` project script references, and `@` asset or collection references in the Agent.
- Keep working with local renames while they await sync and apply remote paths consistently when items are fetched.
- Preview exports, choose columns and formats, manage project tags, and duplicate roles.
- Improve Kitsu workflows with linked tasks, output filters, and selectable sync previews.
- Includes fixes for authentication, project settings, collaborators, Agent tools, and cross-platform UI.
