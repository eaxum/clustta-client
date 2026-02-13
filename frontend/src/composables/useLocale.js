import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { setLocale as setI18nLocale, getLocale as getI18nLocale, getAvailableLocales } from '@/i18n';
import { SettingsService } from '@/services';

/**
 * Composable for managing locale/language settings
 * Provides convenient methods for getting/setting current locale and syncing with user settings
 */
export function useLocale() {
  const { t, locale } = useI18n();

  // Language code to name mapping
  const languageNames = {
    'en': 'English',
    'es': 'Spanish',
    'fr': 'French'
  };

  // Get current locale
  const currentLocale = computed(() => locale.value);

  // Get current language name
  const currentLanguage = computed(() => languageNames[locale.value] || 'English');

  // Get available locales
  const availableLocales = computed(() => getAvailableLocales());

  /**
   * Changes the current locale and persists it to backend settings
   * @param {string} newLocale - The locale code to switch to (e.g., 'en', 'es', 'fr')
   * @returns {Promise<boolean>} - Returns true if successful, false otherwise
   */
  async function setLocale(newLocale) {
    try {
      // Update i18n locale
      setI18nLocale(newLocale);
      
      // Persist to backend settings
      await SettingsService.SetLanguage(newLocale);
      
      return true;
    } catch (error) {
      console.error('Failed to set locale:', error);
      return false;
    }
  }

  /**
   * Loads the user's language preference from backend settings
   * @returns {Promise<string>} - Returns the loaded locale code
   */
  async function loadUserLocale() {
    try {
      const savedLocale = await SettingsService.GetLanguage();
      if (savedLocale && availableLocales.value.includes(savedLocale)) {
        setI18nLocale(savedLocale);
        return savedLocale;
      }
      return 'en'; // Default to English
    } catch (error) {
      console.error('Failed to load user locale:', error);
      return 'en';
    }
  }

  /**
   * Gets the display name for a locale code
   * @param {string} localeCode - The locale code (e.g., 'en', 'es', 'fr')
   * @returns {string} - The display name (e.g., 'English', 'Spanish', 'French')
   */
  function getLocaleName(localeCode) {
    return languageNames[localeCode] || localeCode;
  }

  /**
   * Gets the locale code from a display name
   * @param {string} languageName - The display name (e.g., 'English', 'Spanish')
   * @returns {string} - The locale code (e.g., 'en', 'es')
   */
  function getLocaleCode(languageName) {
    const entry = Object.entries(languageNames).find(([_, name]) => name === languageName);
    return entry ? entry[0] : 'en';
  }

  return {
    t,
    currentLocale,
    currentLanguage,
    availableLocales,
    languageNames,
    setLocale,
    loadUserLocale,
    getLocaleName,
    getLocaleCode,
  };
}
