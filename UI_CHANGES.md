# UI Changes - Language Selector

## Settings Page - Appearance Section

### Before (Original)
```
┌─────────────────────────────────────────────────────────────┐
│ Settings                                                     │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ Appearance                                                   │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ 🎨 Icon scheme                          [Dropdown ▼]  │   │
│ │    Toggle between different icon styles...            │   │
│ ├──────────────────────────────────────────────────────┤   │
│ │ 🌙 Theme                                [Dropdown ▼]  │   │
│ │    Light or Dark mode.                               │   │
│ ├──────────────────────────────────────────────────────┤   │
│ │ 🔲 Default View: Grid                   [Toggle On]  │   │
│ │    Choose between grid or list view as default.      │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### After (With Language Selector) - English
```
┌─────────────────────────────────────────────────────────────┐
│ Settings                                                     │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ Appearance                                                   │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ 🎨 Icon scheme                          [Dropdown ▼]  │   │
│ │    Toggle between different icon styles...            │   │
│ ├──────────────────────────────────────────────────────┤   │
│ │ 🌙 Theme                                [Dropdown ▼]  │   │
│ │    Light or Dark mode.                               │   │
│ ├──────────────────────────────────────────────────────┤   │
│ │ 🌐 Language                             [English ▼]  │   │  ← NEW!
│ │    Language                                           │   │
│ ├──────────────────────────────────────────────────────┤   │
│ │ 🔲 Default View: Grid                   [Toggle On]  │   │
│ │    Choose between grid or list view as default.      │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### After (With Language Selector) - Spanish
```
┌─────────────────────────────────────────────────────────────┐
│ Configuración                                                │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ Apariencia                                                   │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ 🎨 Esquema de iconos                    [Dropdown ▼]  │   │
│ │    Alterna entre diferentes estilos de iconos...     │   │
│ ├──────────────────────────────────────────────────────┤   │
│ │ 🌙 Tema                                 [Dropdown ▼]  │   │
│ │    Modo claro u oscuro.                              │   │
│ ├──────────────────────────────────────────────────────┤   │
│ │ 🌐 Idioma                               [Spanish ▼]  │   │
│ │    Idioma                                             │   │
│ ├──────────────────────────────────────────────────────┤   │
│ │ 🔲 Vista Predeterminada: Cuadrícula     [Toggle On]  │   │
│ │    Elige entre vista de cuadrícula o lista...        │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### After (With Language Selector) - French
```
┌─────────────────────────────────────────────────────────────┐
│ Paramètres                                                   │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ Apparence                                                    │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ 🎨 Schéma d'icônes                      [Dropdown ▼]  │   │
│ │    Basculez entre différents styles d'icônes...      │   │
│ ├──────────────────────────────────────────────────────┤   │
│ │ 🌙 Thème                                [Dropdown ▼]  │   │
│ │    Mode clair ou sombre.                             │   │
│ ├──────────────────────────────────────────────────────┤   │
│ │ 🌐 Langue                               [French ▼]   │   │
│ │    Langue                                             │   │
│ ├──────────────────────────────────────────────────────┤   │
│ │ 🔲 Vue Par Défaut: Grille               [Toggle On]  │   │
│ │    Choisissez entre la vue grille ou liste...        │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Language Dropdown Menu

When clicked, the Language dropdown shows:
```
┌─────────────────┐
│ English      ✓  │  ← Currently selected
│ Spanish         │
│ French          │
└─────────────────┘
```

## Notification Toast - Language Changed

### English
```
┌─────────────────────────────────────────┐
│ Language Updated                        │
│ Language changed to Spanish             │
└─────────────────────────────────────────┘
```

### Spanish
```
┌─────────────────────────────────────────┐
│ Idioma Actualizado                      │
│ Idioma cambiado a Spanish              │
└─────────────────────────────────────────┘
```

### French
```
┌─────────────────────────────────────────┐
│ Langue Mise à Jour                      │
│ Langue changée en French                │
└─────────────────────────────────────────┘
```

## EULA Modal

### Before (Hardcoded English)
```
┌──────────────────────────────────────────────────────────┐
│ Clustta End User License Agreement                       │
├──────────────────────────────────────────────────────────┤
│                                                           │
│ [EULA Text Content...]                                   │
│                                                           │
│                                                           │
├──────────────────────────────────────────────────────────┤
│         [Decline]              [Accept]                  │
└──────────────────────────────────────────────────────────┘
```

### After - Spanish
```
┌──────────────────────────────────────────────────────────┐
│ Acuerdo de Licencia de Usuario Final de Clustta         │
├──────────────────────────────────────────────────────────┤
│                                                           │
│ [EULA Text Content...]                                   │
│                                                           │
│                                                           │
├──────────────────────────────────────────────────────────┤
│        [Rechazar]             [Aceptar]                  │
└──────────────────────────────────────────────────────────┘
```

### After - French
```
┌──────────────────────────────────────────────────────────┐
│ Accord de Licence Utilisateur Final Clustta             │
├──────────────────────────────────────────────────────────┤
│                                                           │
│ [EULA Text Content...]                                   │
│                                                           │
│                                                           │
├──────────────────────────────────────────────────────────┤
│        [Refuser]              [Accepter]                 │
└──────────────────────────────────────────────────────────┘
```

## Other Translated Sections

### Data Management Section - English
```
┌──────────────────────────────────────────────────────┐
│ Data Management                                       │
├──────────────────────────────────────────────────────┤
│ 🧹 Clear recents                              →       │
│    Clear recent projects from the side pane.         │
└──────────────────────────────────────────────────────┘
```

### Data Management Section - Spanish
```
┌──────────────────────────────────────────────────────┐
│ Gestión de Datos                                     │
├──────────────────────────────────────────────────────┤
│ 🧹 Borrar recientes                           →       │
│    Borrar proyectos recientes del panel lateral.    │
└──────────────────────────────────────────────────────┘
```

### Resources & Support Section - French
```
┌──────────────────────────────────────────────────────┐
│ Ressources et Support                                │
├──────────────────────────────────────────────────────┤
│ 📖 Documentation                              ↗       │
│    Documentation complète et référence API.          │
├──────────────────────────────────────────────────────┤
│ 🎥 Guides Vidéo                               ↗       │
│    Regarder des tutoriels et des procédures...       │
├──────────────────────────────────────────────────────┤
│ ❓ Communauté et Support                      ↗       │
│    Rejoignez notre communauté Discord...             │
└──────────────────────────────────────────────────────┘
```

## User Flow

1. **User opens Settings**
   ```
   Menu → Settings → General
   ```

2. **User sees Language option** (new)
   ```
   🌐 Language                    [English ▼]
      Language
   ```

3. **User clicks dropdown**
   ```
   ┌─────────────────┐
   │ English      ✓  │
   │ Spanish         │
   │ French          │
   └─────────────────┘
   ```

4. **User selects Spanish**
   ```
   → UI immediately updates to Spanish
   → Notification appears: "Idioma Actualizado"
   → Setting saves to backend automatically
   ```

5. **User closes and reopens app**
   ```
   → App loads in Spanish automatically
   → Settings shows "Spanish" selected
   ```

## Implementation Notes

- The globe icon (🌐) is used to represent language settings
- Language changes apply immediately without page refresh
- All translated UI elements update in real-time
- The dropdown shows language names in English for consistency
- Success notification appears in the newly selected language
- Language preference persists across sessions
