# Clustta 0.4.42 Release Copy

## Release Inputs

- Version: `0.4.42`
- Previous commit: `973cff4abbdcb745fcf4fa54316292b10f3b62a3`
- New version/tag: `v0.4.42`
- Compare range: `973cff4abbdcb745fcf4fa54316292b10f3b62a3...HEAD`
- Release headline: `AYON integration and stronger DCC workflows`
- Canny types: `new`, `improved`, `fixed`

## GitHub Release

### Clustta 0.4.42

Clustta 0.4.42 introduces AYON integration and expands the Clustta Bridge with safer, more complete workflows for DCC applications.

## New Features

### AYON Integration

Connect Clustta to AYON with a service user API key, link projects, and import AYON folders and tasks into Clustta. Task types, statuses, assignments, and hierarchy are preserved so teams can keep production tracking and workfiles aligned.

AYON workfiles can use Clustta checkpoints from Blender, while previews and task thumbnails can be sent back to AYON. Existing workfiles can be prepared and imported without requiring a preconfigured Clustta template.

### Expanded DCC Workflows

The Clustta Bridge now gives DCC integrations stable workflows for tracking existing project files, creating checkpoints, changing status, syncing work, getting the latest version, building dependencies, and reverting checkpoints.

Project and asset context is resolved directly from the current file, so integrations no longer depend on the project currently open in Clustta Desktop. Direct file references can also be imported as linked dependencies.

## Improvements

### Safer Checkpoints

Bridge-driven checkpoints now capture a stable copy before processing a file and report an actionable error if the source changes during capture.

### Integration and Workspace Polish

- Integration screens now use provider-specific names and icons, including new AYON branding.
- AYON task thumbnails are uploaded with previews, and checkpoint creation gives clearer feedback on smaller displays.
- JetBrains Mono is now available as an interface font.
- FreeCAD files now display the correct file icon.
- Newly created collections are selected consistently and created in the intended location in column and navigator views.

## Bug Fixes

- **Details pane** - The active pane now stays synchronized with the current selection and no longer renders overlapping content.
- **Project navigation** - Returning to a project preserves the correct details selection.
- **Asset grid** - Asset action icons no longer clip at smaller grid sizes.
- **Project sync** - Syncing an asset now includes its ancestor collections when required.
- **Dependencies** - New asset and collection dependencies consistently use the linked dependency type.
- **Open-source notices** - Third-party licence notices now load reliably from packaged application assets.
- **Authentication** - Studio requests no longer send obsolete identity headers.

**Full Changelog**: `973cff4abbdcb745fcf4fa54316292b10f3b62a3...HEAD`

## Canny Changelog

### Title

Clustta 0.4.42 : AYON integration and stronger DCC workflows

### Types

`new, improved, fixed`

### Body

Clustta 0.4.42 connects production tracking and workfiles through AYON, while making DCC workflows safer and more complete.

**New: AYON Integration**

Connect with a service user API key, link an AYON project, and import its folders and tasks into Clustta with task types, statuses, assignments, and hierarchy intact.

Use Clustta checkpoints for AYON workfiles in Blender, and send previews and task thumbnails back to AYON. Workfiles can also be prepared and imported without a preconfigured Clustta template.

**New: Expanded DCC Workflows**

DCC integrations can now track existing project files, create checkpoints, update status, sync work, get the latest version, build dependencies, and revert checkpoints through the Clustta Bridge.

**Improved: Safer Checkpoints**

Bridge-driven checkpoints capture a stable source copy and warn when a file changes during capture.

**Improved: Smaller UX Polish**

- Choose JetBrains Mono as your interface font.
- See provider-specific names and icons across integration screens.
- Create collections in the intended column or navigator location.
- View the correct icon for FreeCAD files.

**Fixed**

- **Details pane** - Selection changes no longer cause stale or overlapping panes.
- **Project navigation** - Returning to a project preserves the correct details selection.
- **Asset grid** - Action icons remain visible at smaller grid sizes.
- **Project sync** - Required ancestor collections are included when syncing assets.
- **Dependencies** - New dependencies consistently use the linked type.
- **Open-source notices** - Third-party licence information loads reliably in packaged builds.

## Apple App Store

### What's New in This Version

Clustta 0.4.42 adds AYON integration and improves DCC workflows.

- Connect AYON, link projects, and import production folders and tasks.
- Use Clustta checkpoints with AYON workfiles and send previews back to AYON.
- Track files, create checkpoints, sync work, get the latest version, and manage dependencies through supported DCC integrations.
- Create safer checkpoints from stable source copies.
- Choose JetBrains Mono and benefit from details pane, grid, navigation, sync, and reliability fixes.

## Microsoft Store

### Release Notes

Clustta 0.4.42 adds AYON integration and improves DCC workflows.

- Connect AYON, link projects, and import production folders and tasks.
- Use Clustta checkpoints with AYON workfiles and send previews back to AYON.
- Track files, create checkpoints, sync work, get the latest version, and manage dependencies through supported DCC integrations.
- Create safer checkpoints from stable source copies.
- Choose JetBrains Mono and benefit from details pane, grid, navigation, sync, and reliability fixes.

## Short Store Summary

Connect AYON and manage safer, more complete DCC workfile workflows with Clustta.

## Flathub Release

### Clustta 0.4.42

Clustta 0.4.42 adds AYON integration and improves DCC workflows.

- Connect AYON, link projects, and import production folders and tasks.
- Use Clustta checkpoints with AYON workfiles and send previews back to AYON.
- Track files, create checkpoints, sync work, get the latest version, and manage dependencies through supported DCC integrations.
- Create safer checkpoints from stable source copies.
- Choose JetBrains Mono and benefit from details pane, grid, navigation, sync, and reliability fixes.
