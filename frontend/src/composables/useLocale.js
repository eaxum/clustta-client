import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { 
  setLocale as setI18nLocale, 
  getLocale as getI18nLocale, 
  getAvailableLocales,
  SUPPORTED_LANGUAGES,
  VALID_LOCALES
} from '@/i18n';
import { SettingsService } from '@/services';

/**
 * Composable for managing locale/language settings
 * Provides convenient methods for getting/setting current locale and syncing with user settings
 */
export function useLocale() {
  const { t, locale } = useI18n();

  // Get current locale
  const currentLocale = computed(() => locale.value);

  // Get current language name
  const currentLanguage = computed(() => SUPPORTED_LANGUAGES[locale.value] || 'English');

  // Get available locales
  const availableLocales = computed(() => VALID_LOCALES);

  /**
   * Changes the current locale and persists it to backend settings
   * @param {string} newLocale - The locale code to switch to (e.g., 'en', 'es', 'fr')
   * @returns {Promise<boolean>} - Returns true if successful, false otherwise
   */
  async function setLocale(newLocale) {
    try {
      // Validate locale before setting
      if (!VALID_LOCALES.includes(newLocale)) {
        console.error(`Invalid locale: ${newLocale}`);
        return false;
      }
      
      // Update i18n locale
      const success = setI18nLocale(newLocale);
      if (!success) {
        return false;
      }
      
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
      if (savedLocale && VALID_LOCALES.includes(savedLocale)) {
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
    return SUPPORTED_LANGUAGES[localeCode] || localeCode;
  }

  /**
   * Gets the locale code from a display name
   * @param {string} languageName - The display name (e.g., 'English', 'Spanish')
   * @returns {string} - The locale code (e.g., 'en', 'es')
   */
  function getLocaleCode(languageName) {
    const entry = Object.entries(SUPPORTED_LANGUAGES).find(([_, name]) => name === languageName);
    return entry ? entry[0] : 'en';
  }

  return {
    t,
    currentLocale,
    currentLanguage,
    availableLocales,
    languageNames: SUPPORTED_LANGUAGES,
    setLocale,
    loadUserLocale,
    getLocaleName,
    getLocaleCode,
  };
}
