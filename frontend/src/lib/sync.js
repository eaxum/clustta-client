import { SyncService } from "@/../bindings/clustta/services";
import { useNotificationStore } from "@/stores/notifications";
import { useProjectStore } from "@/stores/projects";
import { useTrayStates } from "@/stores/TrayStates";
import emitter from '@/lib/mitt';

export async function syncData() {
  const projectStore = useProjectStore();
  const notificationStore = useNotificationStore();
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
