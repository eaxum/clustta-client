# Testing and Validation Guide

## Status

✅ **Frontend Build**: Successfully builds with no errors
✅ **Go Backend**: Language methods added to Settings struct and service layer
✅ **Frontend Services**: Web mode adapter ready with Language methods
⏳ **Wails Bindings**: Need to be generated via `wails3 dev` or `wails3 package`

## How Wails Bindings Work

When you run `wails3 dev` or build the app, Wails automatically:

1. Scans all Go services registered in `main.go`
2. Generates TypeScript bindings for each exported method
3. Places them in `frontend/bindings/clustta/services/`

The new `GetLanguage()` and `SetLanguage()` methods in `services/settings_service.go` will automatically generate:

```typescript
// In frontend/bindings/clustta/services/settingsservice.js
export function GetLanguage() {
    return $Call.ByID(2840987543, "GetLanguage");
}

export function SetLanguage(language) {
    return $Call.ByID(2840987543, "SetLanguage", language);
}
```

## Testing Steps for Desktop Mode

### 1. Build/Run the App
```bash
# Development mode
make client  # or: wails3 dev

# Production build
make build
```

### 2. Test Language Switching
1. Open the app
2. Navigate to Settings → General
3. Find the "Language" dropdown (below Theme setting)
4. Select a different language (Spanish or French)
5. Verify:
   - UI elements change immediately
   - Notification appears in the new language
   - Setting persists after restart

### 3. Test Persistence
1. Change language to Spanish
2. Close the app completely
3. Reopen the app
4. Verify the UI is still in Spanish
5. Check Settings → General shows "Spanish" selected

### 4. Test Fallback Behavior
The system uses English as fallback:
- If a translation key is missing in the selected language, English text appears
- If a translation key doesn't exist at all, the key path is shown

### 5. Test Component Translations

**EulaModal** (`frontend/src/instances/desktop/modals/EulaModal.vue`):
- Title: Should change based on language
- "Accept" and "Decline" buttons: Translated
- Trigger: Start app fresh (no EULA accepted)

**General Settings** (`frontend/src/instances/desktop/settings/General.vue`):
- All section headers (Appearance, Data Management, etc.)
- All setting labels and descriptions
- Notification messages when changing settings
- Trigger: Open Settings from menu

## Web Mode Testing

Since web mode uses localStorage-based settings, it can be tested without Wails:

```bash
cd frontend
npm run dev:web
```

Then:
1. Open http://localhost:1420 in browser
2. Navigate to Settings
3. Change language
4. Verify changes in localStorage:
   ```javascript
   // In browser console
   JSON.parse(localStorage.getItem('clustta_settings'))
   // Should show: { "language": "es", ... }
   ```

## Expected Behavior

### Language Dropdown Options
- English (default)
- Spanish (Español)
- French (Français)

### Translated Components
| Component | Elements Translated |
|-----------|-------------------|
| EulaModal | Title, Accept/Decline buttons |
| General Settings | All section headers, labels, descriptions, notifications |

### Not Translated (By Design)
- PopUpModal: Uses dynamic button labels passed as props
- GeneralButton: Labels passed as props
- ActionButton: Labels passed as props
- User-generated content (project names, comments, etc.)

## Verifying Implementation

### Check Go Backend
```bash
# Verify Language field in Settings struct
grep -A 5 "type Settings struct" internal/settings/user.go

# Verify GetLanguage and SetLanguage functions exist
grep -A 10 "func GetLanguage\|func SetLanguage" internal/settings/user.go

# Verify service exposure
grep -A 10 "GetLanguage\|SetLanguage" services/settings_service.go
```

### Check Frontend Files
```bash
# Verify i18n is registered
grep "i18n" frontend/src/main.js

# Verify locale files exist
ls frontend/src/i18n/locales/

# Verify composable exists
ls frontend/src/composables/useLocale.js

# Verify General.vue uses translations
grep "\$t(" frontend/src/instances/desktop/settings/General.vue
```

## Known Limitations

1. **Bindings Generation**: Desktop mode requires running `wails3 dev` to generate bindings
2. **Partial Translation**: Only key components are translated (EulaModal, General Settings)
3. **Language Persistence**: In web mode, uses localStorage; in desktop mode, uses Go settings file

## Adding More Translations

To translate additional components:

1. Add translation keys to all locale files (`en.json`, `es.json`, `fr.json`)
2. Replace hardcoded strings with `$t('key.path')`
3. Import `useI18n` in script section if needed
4. Test in both languages

Example:
```vue
<!-- Before -->
<button>Save Changes</button>

<!-- After -->
<button>{{ $t('common.save') }}</button>
```

## Troubleshooting

### Issue: Language not changing
- Check browser console for errors
- Verify SettingsService.SetLanguage() is being called
- Check localStorage/settings file for language value

### Issue: Missing translations
- Check if the key exists in the locale file
- Verify the key path is correct (case-sensitive)
- Check console for missing translation warnings

### Issue: Wails bindings not found (desktop mode)
- Run `wails3 dev` to generate bindings
- Check `frontend/bindings/clustta/services/settingsservice.js`
- Verify service is registered in `main.go`

## Screenshots Needed

For PR review, capture:
1. Settings page with Language dropdown (English selected)
2. Settings page after switching to Spanish
3. Settings page after switching to French
4. EULA modal in different languages (if accessible)
5. Notification toast showing language change message
