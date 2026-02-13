# Internationalization (i18n) Guide

This document explains how the internationalization system works in the Clustta client and how to add new translations.

## Overview

The Clustta client uses `vue-i18n` v11 for internationalization support. The system supports multiple languages with English as the fallback language.

## Architecture

### Configuration

- **Main config**: `frontend/src/i18n/index.js`
- **Locale files**: `frontend/src/i18n/locales/` directory
  - `en.json` - English (base/fallback language)
  - `es.json` - Spanish
  - `fr.json` - French

### Backend Integration

Language preferences are stored in the user settings:
- **Go Backend**: `internal/settings/user.go` - `Settings.Language` field
- **Service Layer**: `services/settings_service.go` - `GetLanguage()` and `SetLanguage()` methods
- **Frontend Service**: `frontend/src/services/adapters/settingsservice.js` - Language persistence

## Usage in Components

### Basic Translation

Use the `$t()` function in templates:

```vue
<template>
  <h1>{{ $t('settings.title') }}</h1>
  <p>{{ $t('common.accept') }}</p>
</template>
```

### Translation with Parameters

For dynamic values, use interpolation:

```vue
<template>
  <p>{{ $t('notifications.languageChanged', { language: 'English' }) }}</p>
</template>
```

In the locale file:
```json
{
  "notifications": {
    "languageChanged": "Language changed to {language}"
  }
}
```

### Using i18n in Script

```vue
<script setup>
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

const message = t('common.success');
</script>
```

## Adding New Translations

### 1. Add Keys to Locale Files

Update all locale files in `frontend/src/i18n/locales/`:

**en.json**:
```json
{
  "myFeature": {
    "title": "My Feature",
    "description": "This is my feature"
  }
}
```

**es.json**:
```json
{
  "myFeature": {
    "title": "Mi Característica",
    "description": "Esta es mi característica"
  }
}
```

**fr.json**:
```json
{
  "myFeature": {
    "title": "Ma Fonctionnalité",
    "description": "C'est ma fonctionnalité"
  }
}
```

### 2. Use in Components

Replace hardcoded strings with translation keys:

```vue
<!-- Before -->
<h2>My Feature</h2>

<!-- After -->
<h2>{{ $t('myFeature.title') }}</h2>
```

## Key Naming Convention

Use descriptive, namespaced keys:
- ✅ `settings.theme.label`
- ✅ `notifications.saveSuccess`
- ❌ `theme` (too generic)
- ❌ `msg1` (not descriptive)

### Common Namespaces

- `common` - Shared UI elements (buttons, labels, etc.)
- `settings` - Settings page content
- `notifications` - Notification messages
- `eula` - EULA modal content
- `languages` - Language names

## Changing Language

The language preference is persisted in:
1. Backend user settings (Go)
2. Frontend localStorage (for quick access)

To change language programmatically:

```javascript
import { setLocale } from '@/i18n';

setLocale('es'); // Switch to Spanish
```

## Available Languages

Current languages:
- `en` - English
- `es` - Spanish (Español)
- `fr` - French (Français)

### Adding a New Language

1. Create a new locale file: `frontend/src/i18n/locales/de.json`
2. Copy structure from `en.json` and translate all keys
3. Import and register in `frontend/src/i18n/index.js`:

```javascript
import de from './locales/de.json';

const i18n = createI18n({
  // ...
  messages: {
    en,
    es,
    fr,
    de  // Add new language
  }
});
```

4. Add language to dropdown in `frontend/src/instances/desktop/settings/General.vue`
5. Update `en.json` to include the language name:

```json
{
  "languages": {
    "en": "English",
    "es": "Spanish",
    "fr": "French",
    "de": "German"
  }
}
```

## Fallback Behavior

If a translation key is missing in the selected language, the system will:
1. Try to find the key in the fallback language (English)
2. Display the key path if not found in any language

Example:
- Selected: Spanish (`es`)
- Key: `settings.newFeature`
- If missing in Spanish → Shows English version
- If missing in English → Shows `settings.newFeature`

## Testing Translations

1. Change language in Settings → General → Language dropdown
2. Navigate through the app to verify translations
3. Check for missing keys (they'll appear as key paths)

## Components Using i18n

Currently translated components:
- `frontend/src/instances/desktop/modals/EulaModal.vue`
- `frontend/src/instances/desktop/settings/General.vue`

## Future Enhancements

- [ ] Create a composable `useLocale()` for common i18n operations
- [ ] Add RTL (Right-to-Left) support for Arabic/Hebrew
- [ ] Implement lazy-loading for locale files
- [ ] Add locale-specific date/time formatting
- [ ] Create a translation coverage report tool

## Wails Desktop Mode

For desktop mode, the Wails build process auto-generates TypeScript bindings for Go services in `frontend/bindings/clustta/services/settingsservice.js`. These bindings will include `GetLanguage()` and `SetLanguage()` methods after running:

```bash
make client  # or: wails3 dev
```

The frontend service adapter (`frontend/src/services/adapters/wails.js`) automatically uses these bindings when running in desktop mode.
