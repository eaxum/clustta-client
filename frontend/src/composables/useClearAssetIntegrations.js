import { useI18n } from 'vue-i18n';

import { IntegrationService } from '@/services';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';

export function useClearAssetIntegrations() {
  const modals = useDesktopModalStore();
  const notificationStore = useNotificationStore();
  const projectStore = useProjectStore();
  const trayStates = useTrayStates();
  const { t } = useI18n();

  const confirmClearAssetIntegrations = (assetIDs) => {
    const projectURI = projectStore.activeProject?.uri;
    if (!assetIDs.length || !projectURI) return;

    trayStates.popUpModalTitle = t('kitsu.clearSelectedAssetsTitle', { count: assetIDs.length });
    trayStates.popUpModalMessage = t('kitsu.clearSelectedAssetsMessage', { count: assetIDs.length });
    trayStates.popUpModalIcon = 'plug-cancel';
    trayStates.popUpModalButtons = [t('common.cancel'), t('kitsu.clearSelectedAssets')];
    trayStates.popUpModalFunction = async () => {
      try {
        await IntegrationService.UnlinkAssets(projectURI, assetIDs);
        modals.disableAllModals();
        notificationStore.addNotification(
          t('kitsu.selectedAssetsCleared', { count: assetIDs.length }),
          '',
          'success',
        );
      } catch (error) {
        notificationStore.errorNotification(t('kitsu.selectedAssetsClearFailed'), error);
      }
    };
    modals.setModalVisibility('popUpModal', true);
  };

  return { confirmClearAssetIntegrations };
}
