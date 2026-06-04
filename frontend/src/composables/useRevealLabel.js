import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { usePlatformStore } from '@/stores/platform';

/**
 * Composable for platform-aware "reveal in file manager" labels.
 * Returns the correct wording for the host OS: Finder on macOS,
 * Explorer on Windows, and File Manager on Linux.
 */
export function useRevealLabel() {
  const { t } = useI18n();
  const platformStore = usePlatformStore();

  // Suffix key for the host platform's file manager
  const platformKey = computed(() => {
    if (platformStore.isMac) return 'Finder';
    if (platformStore.isLinux) return 'FileManager';
    return 'Explorer';
  });

  // "Reveal in Finder/Explorer/File Manager"
  const revealLabel = computed(() => t(`common.revealIn${platformKey.value}`));

  // "Show in Finder/Explorer/File Manager"
  const showLabel = computed(() => t(`common.showIn${platformKey.value}`));

  return { revealLabel, showLabel };
}
