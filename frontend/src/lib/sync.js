import { SyncService } from "@/services";
import { useNotificationStore } from "@/stores/notifications";
import { useProjectStore } from "@/stores/projects";
import { useTrayStates } from "@/stores/TrayStates";
import { useAccountStore } from "@/stores/accounts";
import { useEntitlementStore } from "@/stores/entitlements";
import { useUserStore } from "@/stores/users";
import emitter from '@/lib/mitt';

// Refreshes entitlements based on the current studio context.
export function refreshEntitlements() {
  const projectStore = useProjectStore();
  const entitlementStore = useEntitlementStore();
  const studio = projectStore.selectedStudio;
  if (projectStore.isCloudHosted && studio?.id) {
    entitlementStore.fetchStudioEntitlements(studio.id);
  } else {
    entitlementStore.fetchEntitlements();
  }
}

// Guard function to check if remote features are available
function checkRemoteAccess() {
  const accountStore = useAccountStore();
  const entitlementStore = useEntitlementStore();
  const notificationStore = useNotificationStore();
  
  if (accountStore.isOfflineMode) {
    notificationStore.addNotification(
      "Offline Mode",
      "Sync features are not available in offline mode. Sign in to enable sync.",
      "warning"
    );
    return false;
  }

  if (!entitlementStore.canSync) {
    notificationStore.addNotification(
      "Sync Unavailable",
      "Sync is not included in your current plan. Please upgrade to enable sync.",
      "warning"
    );
    return false;
  }

  return true;
}

export async function syncData() {
  if (!checkRemoteAccess()) return;
  
  const projectStore = useProjectStore();
  const notificationStore = useNotificationStore();
  const userStore = useUserStore();

  
  notificationStore.cancleFunction = SyncService.CancelSync;
  notificationStore.canCancel = true;
  
  let syncOptions = {
    only_latest_checkpoints: false,
    asset_dependencies: false,
    assets: false,
    templates: false,
  };
  await SyncService.SyncData(
    projectStore.activeProject.uri,
    projectStore.getActiveProjectUrl,
    false,
    syncOptions
  )
    .then(async () => {
      projectStore.activeProject.is_unsynced = false;
      await projectStore.reloadActiveProject();
      await userStore.reloadCurrentUser();
      refreshEntitlements();
      emitter.emit('refresh-browser')
    })
    .catch((error) => {
      console.error(error);
      notificationStore.errorNotification("Error Syncing Data", error);
    });
}

export async function pullData() {
  if (!checkRemoteAccess()) return;
  
  const projectStore = useProjectStore();
  const notificationStore = useNotificationStore();
  const userStore = useUserStore();
  
  let syncOptions = {
      only_latest_checkpoints: false,
      asset_dependencies: false,
      assets: false,
      templates: false,
  };

  await SyncService.PullData(
    projectStore.activeProject.uri,
    projectStore.getActiveProjectUrl,
    false,
    syncOptions
  )
    .then(async () => {
      projectStore.activeProject.is_unsynced = false;
      await projectStore.reloadActiveProject();
      await userStore.reloadCurrentUser();
      refreshEntitlements();
      emitter.emit('refresh-browser')
    })
    .catch((error) => {
      console.error(error);
      notificationStore.errorNotification("Error Syncing Data", error);
    });
}

// Non-destructive background merge used by the polling loop when the
// experimental UseUpdateSync flag is enabled. Safe to run while the
// project has unsynced local edits.
export async function updateProject() {
  if (!checkRemoteAccess()) return;

  const projectStore = useProjectStore();
  const userStore = useUserStore();

  await SyncService.UpdateProject(
    projectStore.activeProject.uri,
    projectStore.getActiveProjectUrl
  )
    .then(async () => {
      await projectStore.reloadActiveProject();
      await userStore.reloadCurrentUser();
      refreshEntitlements();
      emitter.emit('refresh-browser');
    })
    .catch((error) => {
      console.log(error);
    });
}

export async function syncFullData() {
  if (!checkRemoteAccess()) return;
  
  const projectStore = useProjectStore();
  const notificationStore = useNotificationStore();
  const userStore = useUserStore();
  let syncOptions = {
    only_latest_checkpoints: false,
    asset_dependencies: true,
    assets: true,
    templates: true,
  };
  await SyncService.SyncData(
    projectStore.activeProject.uri,
    projectStore.getActiveProjectUrl,
    true,
    syncOptions
  )
    .then(async () => {
      projectStore.activeProject.is_unsynced = false;
      await projectStore.reloadActiveProject();
      await userStore.reloadCurrentUser();
      refreshEntitlements();
      emitter.emit('refresh-browser')
    })
    .catch((error) => {
      console.error(error);
      notificationStore.errorNotification("Error Syncing Data", error);
    });
}
