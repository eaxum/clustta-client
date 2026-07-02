# Clustta 0.4.36 Release Copy

## GitHub Release

### Clustta 0.4.36

## Improvements

### Checkpoint Workflow

Creating checkpoints is now smoother when you are working with several changes at once. You can create checkpoints from selected modified or new items, review what will be included, remove items from the list, and jump straight to the affected asset from the checkpoint flow.

The single-asset checkpoint modal can also show existing checkpoints beside the new checkpoint form, making it easier to check recent history before saving another version.

### Safer File Drops

File drops now handle common replacement workflows better. You can choose whether files dropped from your operating system overwrite matching files or create a numbered copy.

When a dropped file overwrites a tracked asset, Clustta can automatically create a checkpoint so the replacement is captured in history.

### Cleaner Asset Browsing

Project browsing now gives you clearer control over what is visible. Tasks, resources, collections, assets, and untracked files are easier to show or hide, and "Only Assets" views now work better inside nested collections.

Untracked files are also handled more consistently across project and collection views, including a new purge action for cleaning untracked files while keeping tracked assets and collections.

### Studio Collaboration

Studio permissions are now clearer and more consistent. Create and modify actions better reflect both a user's role and their assigned collection scope, so disabled actions are easier to understand.

Private studios also get better profile handling, studio renaming support, usage information in settings, and smoother account switching.

### Integrations & Links

Kitsu sync now follows directory mappings more closely when deciding asset names, especially when the mapping defines the final filename.

Portfolio links are more flexible too, with support for YouTube, Vimeo, Google Drive, bit.ly, and TinyURL links. Linked assets also have a new Copy URL action.

### Updates & Packaging

Clustta now checks for updates from inside the app and routes users to the right destination for their install type, including website builds, Microsoft Store, Apple App Store, and Linux packages.

This release also includes macOS signing and packaging improvements, refreshed macOS icon and thumbnail handling, and release manifest automation for published builds.

### Smaller UX Polish

- Large lists scroll more smoothly.
- Collaborator and assignee menus now include search and clearer empty states.
- Checkpoint timeline items are easier to scan and open.
- Asset-template messaging is clearer when a project has no templates yet.
- The app now uses "Fetch" language where it is restoring missing working files.

## Bug Fixes

- **Drag and drop** - Fixed stale root-drop hints, stuck drag previews, and several drop-target edge cases.
- **Checkpoint navigation** - Fixed navigation from checkpoint timeline items back to the related asset.
- **Thumbnails** - Fixed preview refreshes after asset changes so thumbnails stay current.
- **Light mode** - Fixed icon contrast issues in affected menus and selectors.
- **Asset tags** - Fixed tag updates so asset changes sync more reliably.
- **Project refresh** - Fixed in-place refresh behavior and metadata loading for linked items.
- **Studio roles** - Fixed newly added roles not appearing immediately.
- **Kitsu setup** - Improved error display and cleanup while editing integration settings.

**Full Changelog**: `70b9920d16ac7da92f1fb11499020e908de2d719...HEAD`


Creating checkpoints is now smoother when you are working with several changes at once. This release also improves file drops, project browsing, studio collaboration, integrations, update checks, and a batch of workflow fixes.

**Checkpoint Workflow**
Creating checkpoints is now smoother when you are working with several changes at once.
You can create checkpoints from selected modified or new items, review what will be included, remove items from the list, and jump straight to the affected asset from the checkpoint flow.

The single-asset checkpoint modal can also show existing checkpoints beside the new checkpoint form, making it easier to check recent history before saving another version.

**Better File Drops(Experimental)**
File drops now handle common replacement workflows better. You can choose whether files dropped from your operating system overwrite matching files or create a numbered copy.

When a dropped file overwrites a tracked asset, Clustta can automatically create a checkpoint so the replacement is captured in history.

**Cleaner Asset Browsing**
Project browsing now gives you clearer control over what is visible. Tasks, resources, collections, assets, and untracked files are easier to show or hide, and "Only Assets" views
now work better inside nested collections.

Untracked files are also handled more consistently across project and collection views, including a new purge action for cleaning untracked files while keeping tracked assets and collections.

**Improved: Studio Collaboration**
Studio permissions are now clearer and more consistent. Create and modify actions better reflect both a user's role and their assigned collection scope, so disabled actions are easier to understand.

Private studios also get better profile handling, studio renaming support, usage information in settings, and smoother account switching.

**Improved: Integrations & Links**
Kitsu sync now follows directory mappings more closely when deciding asset names, especially when the mapping defines the final filename.

Portfolio links are more flexible too, with support for YouTube, Vimeo, Google Drive, bit.ly, and TinyURL links. Linked assets also have a new Copy URL action.

**Improved: Updates & Packaging**
Clustta now checks for updates from inside the app and routes users to the right destination for their install type, including website builds, Microsoft Store, Apple App Store, and Linux packages.

This release also includes macOS signing and packaging improvements, refreshed macOS icon and
thumbnail handling, and release manifest automation for published builds.

**Improved: Smaller UX Polish**

- Large lists scroll more smoothly.
- Collaborator and assignee menus now include search and clearer empty states.
- Checkpoint timeline items are easier to scan and open.
- Asset-template messaging is clearer when a project has no templates yet.
- The app now uses "Fetch" language where it is restoring missing working files.

**Fixed**

- **Drag and drop** - Fixed stale root-drop hints, stuck drag previews, and several drop-target edge cases.
- **Checkpoint navigation** - Fixed navigation from checkpoint timeline items back to the related asset.
- **Thumbnails** - Fixed preview refreshes after asset changes so thumbnails stay current.
- **Light mode** - Fixed icon contrast issues in affected menus and selectors.
- **Asset tags** - Fixed tag updates so asset changes sync more reliably.
- **Project refresh** - Fixed in-place refresh behavior and metadata loading for linked items.
- **Studio roles** - Fixed newly added roles not appearing immediately.
- **Kitsu setup** - Improved error display and cleanup while editing integration settings.

## Apple App Store

### What's New in This Version

Clustta 0.4.36 improves checkpoints, file drops, asset browsing, studio collaboration, integrations,
and update checks.

- Create checkpoints from selected modified or new items.
- View existing checkpoints while creating a new checkpoint.
- Choose whether dropped files overwrite matching files or create numbered copies.
- Automatically checkpoint tracked assets when a dropped file overwrites them.
- Browse projects with clearer controls for tasks, resources, collections, assets, and untracked files.
- Use improved "Only Assets" views inside nested collections.
- Purge untracked files from projects or collections while keeping tracked work intact.
- Check for updates from inside the app.
- Use more portfolio link types, including Google Drive, bit.ly, and TinyURL.
- Improved Kitsu sync naming, studio permissions, collaborator search, thumbnails, scrolling, and drag-and-drop behavior.

## Microsoft Store

### Release Notes

Clustta 0.4.36 improves checkpoints, file drops, asset browsing, studio collaboration, integrations,
and update checks.

- Create checkpoints from selected modified or new items.
- View existing checkpoints while creating a new checkpoint.
- Choose whether dropped files overwrite matching files or create numbered copies.
- Automatically checkpoint tracked assets when a dropped file overwrites them.
- Browse projects with clearer controls for tasks, resources, collections, assets, and untracked files.
- Use improved "Only Assets" views inside nested collections.
- Purge untracked files from projects or collections while keeping tracked work intact.
- Check for updates from inside the app with Microsoft Store-aware routing.
- Use more portfolio link types, including Google Drive, bit.ly, and TinyURL.
- Improved Kitsu sync naming, studio permissions, collaborator search, thumbnails, scrolling, and drag-and-drop behavior.

## Short Store Summary

Better checkpoints, safer file drops, cleaner asset browsing, improved studio collaboration, expanded
portfolio links, and built-in update checks.
