import { Browser } from '@wailsio/runtime';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';
import { isValidWeblink } from '@/lib/pointer';
import { AssetService, CheckpointService, FSService } from '@/services';
import { useBrowserTreeStore } from '@/stores/browserTree';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';

export const useAssetLauncher = () => {
  const browserTreeStore = useBrowserTreeStore();
  const modals = useDesktopModalStore();
  const notificationStore = useNotificationStore();
  const projectStore = useProjectStore();
  const trayStates = useTrayStates();
  const userStore = useUserStore();
  const { t } = useI18n();

  const preLaunchTrustKey = (trust) => {
    return `clustta-prelaunch-trust:${projectStore.activeProject.id}:${trust.hook_id}`;
  };

  const launchProjectAsset = async (assetId) => {
    try {
      const projectPath = projectStore.activeProject.uri;
      const trust = await FSService.PreparePreLaunch(
        projectPath,
        projectStore.getActiveProjectUrl,
        assetId
      );
      if (trust?.required && localStorage.getItem(preLaunchTrustKey(trust)) !== trust.digest) {
        confirmPreLaunchTrust(assetId, trust);
        return;
      }
      await FSService.LaunchProjectAsset(projectPath, assetId);
    } catch (error) {
      notificationStore.errorNotification(t('notifications.errorLaunchingAsset'), error);
    }
  };

  const confirmPreLaunchTrust = (assetId, trust) => {
    trayStates.dangerousActionTitle = t('blocks.trustLaunchHookTitle', { name: trust.hook_name });
    trayStates.dangerousActionMessage = t('blocks.trustLaunchHookMessage', {
      scripts: trust.scripts.join(', '),
    });
    trayStates.dangerousActionIcon = 'alert';
    trayStates.dangerousActionConfirmLabel = t('blocks.trustAndLaunch');
    trayStates.dangerousActionConfirmText = '';
    trayStates.dangerousActionShowInput = false;
    trayStates.dangerousActionShowToggle = false;
    trayStates.dangerousActionFunction = async () => {
      localStorage.setItem(preLaunchTrustKey(trust), trust.digest);
      await launchProjectAsset(assetId);
    };
    modals.setModalVisibility('confirmDangerousActionModal', true);
  };

  const renameAssetPath = async (asset) => {
    await AssetService.ApplyPathUpdate(projectStore.activeProject.uri, asset.id);
    asset.local_path = '';
    asset.file_status = await AssetService.GetAssetState(projectStore.activeProject.uri, asset.id);
    browserTreeStore.patchItemsById(asset.id, {
      file_status: asset.file_status,
      local_path: '',
    });
    emitter.emit('refresh-browser');
  };

  const launchAssetFile = async (asset) => {
    const filePath = asset.pointer || asset.file_path;
    if (await FSService.Exists(filePath)) {
      if (asset.type === 'untracked_asset') {
        await FSService.LaunchFile(filePath);
        return;
      }
      await launchProjectAsset(asset.id);
      return;
    }

    if (asset.type === 'untracked_asset') return;

    try {
      const response = await CheckpointService.Revert(
        projectStore.activeProject.uri,
        projectStore.getActiveProjectUrl,
        [asset.id]
      );
      if (!response?.restored_asset_ids?.includes(asset.id)) return;
      browserTreeStore.markAssetsAvailable(response.restored_asset_ids);
      await launchProjectAsset(asset.id);
    } catch (error) {
      notificationStore.errorNotification(t('notifications.errorFetchingAsset'), error);
    }
  };

  const confirmRenameBeforeLaunch = (asset) => {
    trayStates.dangerousActionTitle = t('blocks.renameRequiredTitle');
    trayStates.dangerousActionMessage = t('blocks.renameRequiredMessage', {
      oldName: asset.local_path?.split(/[\\/]/).pop(),
      newName: `${asset.name}${asset.extension}`,
    });
    trayStates.dangerousActionIcon = 'alert';
    trayStates.dangerousActionConfirmLabel = t('common.proceed');
    trayStates.dangerousActionConfirmText = '';
    trayStates.dangerousActionShowInput = false;
    trayStates.dangerousActionShowToggle = false;
    trayStates.dangerousActionFunction = async () => {
      await renameAssetPath(asset);
      await launchAssetFile(asset);
    };
    modals.setModalVisibility('confirmDangerousActionModal', true);
  };

  const launchAsset = async (asset) => {
    if (!asset || !userStore.canDo('pull_chunk')) return;
    if (asset.is_link && isValidWeblink(asset.pointer)) {
      Browser.OpenURL(asset.pointer);
      return;
    }
    if (asset.local_path) {
      confirmRenameBeforeLaunch(asset);
      return;
    }
    await launchAssetFile(asset);
  };

  return { launchAsset, renameAssetPath };
};
