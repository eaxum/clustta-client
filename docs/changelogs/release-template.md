# Clustta Release Copy Template

Use this template when preparing release notes for GitHub Releases, Canny, Apple App Store, and Microsoft Store.

Replace every value in square brackets. Keep the language user-facing, concise, and focused on what changed for the user.

## Release Inputs

- Version: `[0.0.00]`
- Previous version/tag: `[v0.0.00]`
- New version/tag: `[v0.0.00]`
- Compare range: `[[previous-tag-or-commit]...[new-tag-or-commit]]`
- Release headline: `[short headline]`
- Canny types: `[new]`, `[improved]`, `[fixed]`

## GitHub Release

### Clustta `[0.0.00]`

## Improvements

### `[Feature Area]`

`[One short paragraph explaining the user-facing improvement. Say what users can do now, not how it was implemented.]`

`[Optional second paragraph with a supporting workflow detail.]`

`[Duplicate this feature-area block only when the release has another major user-facing theme.]`

### Smaller UX Polish

- `[Small improvement written as a user-facing sentence.]`
- `[Duplicate this bullet only when needed.]`

## Bug Fixes

- **`[Fix Area]`** - `[Plain-language description of what was fixed.]`
- `[Duplicate this bullet only when needed.]`

**Full Changelog**: `[[previous-tag-or-commit]...[new-tag-or-commit]]`

## Canny Changelog

### Title

Clustta `[0.0.00]` : `[short headline]`

### Types

`[new, improved, fixed]`

### Body

`[One short intro paragraph. Match the GitHub release's main theme, but keep it tighter for Canny.]`

**`[New or Improved]: [Feature Area 1]`**
`[Copy or lightly adapt the matching GitHub paragraph. Keep line breaks natural for Canny.]`

`[Optional second paragraph from the GitHub section.]`

`[Duplicate this Canny feature block only when the release has another major user-facing theme.]`

**Improved: Smaller UX Polish**

- `[Small improvement copied from GitHub if relevant to Canny.]`
- `[Duplicate this bullet only when needed.]`

**Fixed**

- **`[Fix Area]`** - `[Plain-language description copied or adapted from GitHub.]`
- `[Duplicate this bullet only when needed.]`

## Apple App Store

### What's New in This Version

Clustta `[0.0.00]` improves `[short release theme]`.

- `[Short user-facing change.]`
- `[Short combined polish/fixes line.]`
- `[Duplicate bullets only when needed.]`

## Microsoft Store

### Release Notes

Clustta `[0.0.00]` improves `[short release theme]`.

- `[Short user-facing change.]`
- `[Short combined polish/fixes line.]`
- `[Duplicate bullets only when needed.]`

## Short Store Summary

`[One sentence summary for store metadata or social copy.]`

## Writing Checklist

- Keep release notes user-facing and avoid internal implementation detail.
- Use GitHub for the fuller version: feature sections plus a bug-fix list.
- Use Canny for the announcement version: title, types, then bold feature labels.
- Keep Apple and Microsoft Store notes shorter than GitHub and Canny.
- Use active, direct wording: "Create", "View", "Choose", "Fixes", "Improves".
- Avoid mentioning experimental, internal, or backend details unless users need to know.
- Confirm platform-specific claims before publishing.
