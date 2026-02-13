import { createI18n } from 'vue-i18n';
import en from './locales/en.json';
import es from './locales/es.json';
import fr from './locales/fr.json';

// Get the user's language preference from settings
// This will be loaded from the backend settings service
let defaultLocale = 'en';

// Try to get language from localStorage for initial load
// The actual value will be loaded from backend settings after app initialization
try {
  const savedLocale = localStorage.getItem('clustta_language');
  if (savedLocale) {
    defaultLocale = savedLocale;
  }
} catch (e) {
  console.warn('Failed to load language preference from localStorage:', e);
}

const i18n = createI18n({
  legacy: false, // Use Composition API mode
  locale: defaultLocale,
  fallbackLocale: 'en',
  messages: {
    en,
    es,
    fr
  },
  // Global injection for $t, $tc, etc.
  globalInjection: true
});

export default i18n;

// Helper function to change locale dynamically
export function setLocale(locale) {
  if (i18n.global.availableLocales.includes(locale)) {
    i18n.global.locale.value = locale;
    // Save to localStorage for persistence across sessions
    try {
      localStorage.setItem('clustta_language', locale);
    } catch (e) {
      console.warn('Failed to save language preference to localStorage:', e);
    }
  } else {
    console.warn(`Locale '${locale}' is not available. Available locales:`, i18n.global.availableLocales);
  }
}

// Helper function to get current locale
export function getLocale() {
  return i18n.global.locale.value;
}

// Helper function to get available locales
export function getAvailableLocales() {
  return i18n.global.availableLocales;
}
