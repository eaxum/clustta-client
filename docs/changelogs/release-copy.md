# Clustta 0.4.38 Release Copy

## Release Inputs

- Version: `0.4.38`
- Previous version/tag: `v0.4.37`
- New version/tag: `v0.4.38`
- Compare range: `3359b6b89f4b7c68f24047bba9c6a3a7a46c96dd...HEAD`
- Release headline: `Faster browsing, flexible storage, and connected DCC workflows`
- Canny types: `new`, `improved`, `fixed`

## GitHub Release

### Clustta 0.4.38

## Improvements

### Connected DCC Workflows

Clustta now provides shared asset workflows for DCC integrations, with studio-aware actions and faster workspace loading. This foundation supports consistent checkpoint, fetch, revert, and build experiences across connected creative applications.

### Flexible Studio Storage

Private Studios can choose Compact or Deflated storage per project and convert between modes with administrator confirmation and live progress. A new project management view makes storage settings and linked Studio usage easier to understand.

### Faster, Steadier Browsing

Browser views now preserve position during silent updates and refresh only the data that changed. Nested collections, fetched file states, filtered results, and breadcrumb counts stay more accurate without interrupting navigation.

### Safer Batch Actions

Agent batch actions now group scoped local changes into a single plan with selective approval previews, making recursive type and suffix changes easier to review before applying them.

### Smaller UX Polish

- Keyboard shortcuts and context menus now show permission-aware action hints.
- Modals, cloud controls, badges, icons, and border styling have been refreshed.

## Bug Fixes

- **Sync and offline changes** - Fixed restricted items appearing during updates and added safe local fallback for metadata changes that cannot sync immediately.
- **Browser state** - Fixed nested asset fetch states, filtered descendant visibility, and file states after fetching.
- **Checkpoints** - Fixed checkpoints being backdated and corrected timeline gradients across date groups.
- **Kitsu integration** - Fixed sync previews incorrectly marking projects as unsynced.
- **Task updates** - Fixed batch task status changes not updating the current view correctly.
- **Project storage** - Fixed metadata-only chunk retention and dedicated Studio storage selection or confirmation issues.
- **Accounts and themes** - Fixed an entitlement loading race and icons that did not follow the active theme.

**Full Changelog**: `3359b6b89f4b7c68f24047bba9c6a3a7a46c96dd...HEAD`

## Canny Changelog

### Title

Clustta 0.4.38 : Faster browsing, flexible storage, and connected DCC workflows

### Types

`new, improved, fixed`

### Body

Clustta 0.4.38 adds shared DCC asset workflows, flexible private Studio storage, faster browsing, and safer batch actions.

**New: Connected DCC Workflows**

Connected creative applications can use consistent, studio-aware asset workflows with faster workspace loading.

**New: Flexible Studio Storage**

Private Studios can choose Compact or Deflated storage per project and convert between modes with administrator confirmation and live progress.

**Improved: Faster Browsing**

Silent updates preserve your position and refresh only changed data, while nested asset states, filters, and counts remain accurate.

**Improved: Safer Batch Actions**

Review scoped batch changes in one plan and selectively approve them before they are applied.

**Fixed**

- Fixed restricted item visibility and offline metadata fallback during sync.
- Fixed nested fetch states, filtered results, and browser file states.
- Fixed checkpoint dates, timeline gradients, and Kitsu sync previews.
- Fixed batch task updates, project storage edge cases, entitlement loading, and themed icons.

## Apple App Store

### What's New in This Version

Clustta 0.4.38 improves connected workflows, project storage, browsing, and reliability.

- Use faster, studio-aware asset workflows from connected creative applications.
- Choose and convert private Studio project storage modes with clear progress.
- Keep your browser position while project data updates quietly in place.
- Review safer batch actions and see clearer permission-aware shortcuts.
- Includes fixes for sync, fetched assets, checkpoints, task updates, and project storage.

## Microsoft Store

### Release Notes

Clustta 0.4.38 improves connected workflows, project storage, browsing, and reliability.

- Use faster, studio-aware asset workflows from connected creative applications.
- Choose and convert private Studio project storage modes with clear progress.
- Keep your browser position while project data updates quietly in place.
- Review safer batch actions and see clearer permission-aware shortcuts.
- Includes fixes for sync, fetched assets, checkpoints, task updates, and project storage.

## Short Store Summary

Connected DCC workflows, flexible Studio storage, faster browsing, and reliability fixes.
