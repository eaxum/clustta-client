# Clustta 0.4.37 Release Copy

## Release Inputs

- Version: `0.4.37`
- Previous version/tag: `v0.4.36`
- New version/tag: `v0.4.37`
- Compare range: `d82d1533048c05183440ea0d42f938b9a5d351ba...HEAD`
- Release headline: `Faster collaboration with safer, smarter workflows`
- Canny types: `new`, `improved`, `fixed`

## GitHub Release

### Clustta 0.4.37

## Improvements

### Faster Studio Collaboration

Changes to statuses, assignees, collection sharing, and task or resource state now reach Studio immediately when you are online. If the server cannot be reached, Clustta safely saves the change locally and tells you when a manual sync is needed.

Bulk assignment, sharing, and task-state changes now update the current view in place, avoiding unnecessary browser refreshes while you work.

### Clearer Role-Based Access

Project roles and collection assignments now control actions more consistently throughout Clustta. Creating, editing, moving, assigning, tagging, changing status, managing dependencies, deleting, and creating checkpoints are available only when your role and assigned collection scope allow them.

Asset moves are limited to collections you can modify, and Project Settings now opens only for eligible collaborators and shows only the tabs their permissions allow. Updated roles are also reflected when project data refreshes, so access changes take effect without reopening the project.

### Smarter Kitsu Workflows

Multiple Kitsu task types can now map to one Clustta asset and act as production steps instead of creating duplicate assets. Directory mappings include per-task output names and a new `<OutputName>` placeholder, giving each step a predictable output path.

New assets created during Kitsu sync now inherit their mapped Kitsu status, keeping production state aligned from the start.

### Leaner Project Storage

A new Metadata Only option in Advanced Settings reduces disk usage by not retaining downloaded file data inside `.clst` archives. Transferred chunks are cleaned up automatically and project archives are compacted, while files remain available to fetch again when needed.

### Better Browsing and Filtering

Browser filters now understand both asset and collection types, combine more reliably with search, and remain consistent when switching workspace tabs. Breadcrumbs show active filter context and accurate result or task counts, including in Kanban and My Assets views.

Search also accepts asset and collection IDs, and the Details pane lets you view and copy those IDs. Asset and collection state loading has been streamlined so browsing and searching larger projects stays responsive.

### Smaller UX Polish

- Revert is now available directly from the header when the current view has changes.
- Collection parent names are shown correctly in the Details pane.
- Assignees, collection states, collapsed collections, and other affected controls have more consistent spacing and alignment.
- The dependency graph no longer shows a redundant sidebar filter when the same workflows are available from the Details pane.
- Workspace tabs, asset-type filters, and other affected actions use clearer, current icons.
- Update status text remains readable across themes and changes dynamically with the current update state.

## Bug Fixes

- **Offline Studio changes** - Status, assignment, sharing, and task-state changes now fall back to a local update only when the server cannot be reached, with a clear manual-sync warning.
- **Role permissions** - Fixed asset and collection actions appearing or running outside a collaborator's role or assigned collection scope.
- **Asset moves** - Fixed scoped collaborators being able to choose destinations outside their assigned collections.
- **Project Settings** - Fixed settings access and tabs not consistently following project-management permissions.
- **Role refreshes** - Fixed updated roles not taking effect when project data refreshes.
- **Workspace filters** - Fixed filters and search carrying incorrect results or behavior between workspace tabs.
- **Search performance** - Fixed unnecessary asset, collection, and child-state loading during searches.
- **Collection details** - Fixed incorrect parent names and inconsistent assignee or state alignment in expanded and collapsed layouts.
- **Checkpoint permissions** - Fixed checkpoint actions appearing for assignment-scoped users who cannot modify the selected items.
- **Kitsu sync** - Fixed newly created assets starting without their mapped Kitsu status.
- **Linux updates** - Improved update metadata for DEB and RPM packages and corrected Flatpak release publishing paths.

**Full Changelog**: `d82d1533048c05183440ea0d42f938b9a5d351ba...HEAD`

## Canny Changelog

### Title

Clustta 0.4.37 : Faster collaboration with safer, smarter workflows

### Types

`new, improved, fixed`

### Body

Clustta 0.4.37 makes Studio collaboration faster and safer, adds more flexible Kitsu production mappings, reduces optional project storage, and improves browsing across large projects.

**Improved: Faster Studio Collaboration**

Changes to statuses, assignees, collection sharing, and task or resource state now reach Studio immediately when you are online. If the server cannot be reached, Clustta safely saves the change locally and tells you when a manual sync is needed.

Bulk changes also update the current view in place without an unnecessary browser refresh.

**Improved: Clearer Role-Based Access**

Project roles and collection assignments now consistently control which actions are available. Collaborators see only the creation, editing, moving, assignment, status, tagging, dependency, deletion, and checkpoint actions allowed within their assigned scope.

Asset destinations are limited to collections collaborators can modify, while Project Settings shows only the permitted tabs. Role changes are reflected when project data refreshes.

**New: Smarter Kitsu Workflows**

Multiple Kitsu task types can now map to one Clustta asset as production steps. Configure an output name for each task type and use the new `<OutputName>` placeholder to create predictable paths without duplicating assets.

New assets created during sync also inherit their mapped Kitsu status.

**New: Metadata-Only Storage**

Enable Metadata Only in Advanced Settings to reduce disk usage by leaving downloaded file data out of `.clst` archives. Clustta cleans up transferred chunks and compacts the archive automatically, while keeping files available to fetch again.

**Improved: Browsing and Filtering**

Filter assets and collections by type, combine filters with search more reliably, and keep the same behavior across workspace tabs. Breadcrumbs now show filter context and accurate result counts in browser, Kanban, and My Assets views.

You can also search by asset or collection ID, then view and copy the ID from the Details pane.

**Improved: Smaller UX Polish**

- Revert is available directly from the header when changes are present.
- Large-project browsing and search avoid unnecessary state loading.
- Collection layouts and affected controls have clearer alignment and icons.
- Update status text is clearer and responds to the current update state.

**Fixed**

- **Role permissions** - Fixed actions appearing or running outside a collaborator's role or assigned collection scope.
- **Asset moves** - Fixed destination choices outside a scoped collaborator's assigned collections.
- **Project Settings** - Fixed settings access and tabs not consistently following project-management permissions.
- **Workspace filters** - Fixed inconsistent filter and search behavior when moving between workspace tabs.
- **Collection details** - Fixed incorrect parent names and inconsistent assignee or state layouts.
- **Checkpoint permissions** - Fixed checkpoint actions for assignment-scoped collaborators.
- **Kitsu sync** - Fixed new assets not receiving their mapped Kitsu status.
- **Linux updates** - Improved DEB, RPM, and Flatpak release metadata handling.

## Apple App Store

### What's New in This Version

Clustta 0.4.37 improves Studio collaboration, role-based access, Kitsu workflows, project storage, and browsing.

- Apply status, assignment, sharing, and task-state changes to Studio immediately when online, with safe local fallback when offline.
- Work with clearer role-based actions, collection-scoped asset moves, and permission-aware Project Settings.
- See bulk assignment, sharing, and task-state changes update without refreshing the browser.
- Map multiple Kitsu task types to one asset, configure per-step output names, and carry mapped statuses into new assets.
- Reduce disk usage with the new Metadata Only storage option.
- Filter and search assets and collections more consistently across workspaces.
- Search by asset or collection ID, then view and copy IDs from the Details pane.
- Access Revert from the header and enjoy clearer counts, layouts, icons, and update messages.

## Microsoft Store

### Release Notes

Clustta 0.4.37 improves Studio collaboration, role-based access, Kitsu workflows, project storage, and browsing.

- Apply status, assignment, sharing, and task-state changes to Studio immediately when online, with safe local fallback when offline.
- Work with clearer role-based actions, collection-scoped asset moves, and permission-aware Project Settings.
- See bulk assignment, sharing, and task-state changes update without refreshing the browser.
- Map multiple Kitsu task types to one asset, configure per-step output names, and carry mapped statuses into new assets.
- Reduce disk usage with the new Metadata Only storage option.
- Filter and search assets and collections more consistently across workspaces.
- Search by asset or collection ID, then view and copy IDs from the Details pane.
- Access Revert from the header and enjoy clearer counts, layouts, icons, and update messages.

## Short Store Summary

Faster Studio updates, safer role-based access, multi-step Kitsu mappings, and metadata-only storage.
