import { SyncService } from "@/services";
import { useNotificationStore } from "@/stores/notifications";
import { useProjectStore } from "@/stores/projects";
import { useTrayStates } from "@/stores/TrayStates";
import { useAccountStore } from "@/stores/accounts";
import emitter from '@/lib/mitt';

// Guard function to check if remote features are available
function checkRemoteAccess() {
  const accountStore = useAccountStore();
  const notificationStore = useNotificationStore();
  
  if (accountStore.isOfflineMode) {
    notificationStore.addNotification(
      "Offline Mode",
      "Sync features are not available in offline mode. Sign in to enable sync.",
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

  
  notificationStore.cancleFunction = SyncService.CancelSync;
  notificationStore.canCancel = true;
  
  let syncOptions = {
    only_latest_checkpoints: false,
    task_dependencies: false,
    tasks: false,
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
  
  let syncOptions = {
      only_latest_checkpoints: false,
      task_dependencies: false,
      tasks: false,
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
      emitter.emit('refresh-browser')
    })
    .catch((error) => {
      console.error(error);
      notificationStore.errorNotification("Error Syncing Data", error);
    });
}

export async function syncFullData() {
  if (!checkRemoteAccess()) return;
  
  const projectStore = useProjectStore();
  const notificationStore = useNotificationStore();
  let syncOptions = {
    only_latest_checkpoints: false,
    task_dependencies: true,
    tasks: true,
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
      emitter.emit('refresh-browser')
    })
    .catch((error) => {
      console.error(error);
      notificationStore.errorNotification("Error Syncing Data", error);
    });
}
