<template>
    <div class="app-root">
        <router-view />
    </div>
</template>

<script setup>

import { ref, onMounted, computed } from 'vue';
import { useNotificationStore } from './stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useSyncConflictStore } from '@/stores/syncConflict';
import { Events } from "@wailsio/runtime";
import emitter from '@/lib/mitt';

import { useAssetStore } from '@/stores/assets';
import { useProjectStore } from './stores/projects';
import { useStudioStore } from './stores/studio';
import { useUserStore } from './stores/users';
import { SyncService, ProjectService } from "@/services";
import { System } from "@wailsio/runtime";
import { LogService } from '@/services';
import { useStageStore } from './stores/stages';
import { useMenu } from '@/stores/menu';
import { useAccountStore } from '@/stores/accounts';
import { useEntitlementStore } from '@/stores/entitlements';
import { useSettingsStore } from '@/stores/settings';
import { useThemeStore } from '@/stores/theme';
import { usePlatformStore } from '@/stores/platform';

// Platform detection
const menu = useMenu();
const platformStore = usePlatformStore();
const projectStore = useProjectStore();
const assetStore = useAssetStore();
const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();
const syncConflictStore = useSyncConflictStore();
const themeStore = useThemeStore();
const settingsStore = useSettingsStore();
const stageStore = useStageStore();
const studioStore = useStudioStore();
const userStore = useUserStore();
const accountStore = useAccountStore();
const entitlementStore = useEntitlementStore();



const disableMenu = () => {
    // Only disable context menu on desktop
    if (platformStore.isWeb || System.IsDebug) {
        return
    }

    // prevent context menu
    document.addEventListener('contextmenu', e => {
        e.preventDefault();
        return false;
    }, { capture: true })

    // prevent drag click selection
    document.addEventListener('selectstart', e => {
        e.preventDefault();
        return false;
    }, { capture: true })
}

const handleProgressUpdate = (progressData) => {
    notificationStore.updateProgress(progressData);
};

const handleSyncConflict = (conflictData) => {
    console.log('Sync conflict detected:', conflictData);
    syncConflictStore.setConflicts(
        conflictData.projectPath,
        conflictData.remoteURL,
        conflictData.conflicts
    );
    modals.setModalVisibility('syncConflictModal', true);
};

const handleopenClusttaFile = async (filePath) => {
    if (!filePath) return;

    try {
        const fileInfo = await ProjectService.InspectClusttaFile(filePath);
        if (!fileInfo || !fileInfo.valid) {
            notificationStore.errorNotification("Invalid Project", "The selected file is not a valid Clustta project.");
            return;
        }

        await projectStore.openClusttaFile(fileInfo);
    } catch (error) {
        console.error('Failed to open project file:', error);
        notificationStore.errorNotification("Failed to Open Project", error);
    }
};

if (platformStore.isWeb) {
    emitter.on('progress-update', handleProgressUpdate);
    emitter.on('sync-conflict', handleSyncConflict);
} else {
    Events.On('progress-update', async (message) => {
        handleProgressUpdate(message.data);
    });
    Events.On('sync-conflict', async (message) => {
        handleSyncConflict(message.data);
    });
    Events.On('open-project-file', async (message) => {
        const filePath = message.data;
        console.log('Opening project file:', filePath);
        await handleopenClusttaFile(filePath);
    });
}

disableMenu();

async function updateFileStates() {
    if (!projectStore.projects.length) {
        return
    }
    try {
        await assetStore.refreshDisplayedFilesStatus()
    } catch (err) {
        console.log(err)
        LogService.LogError("error procesing file status: " + err.message)
    }
}
async function pullData() {
    let syncOptions = {
        only_latest_checkpoints: false,
        asset_dependencies: false,
        assets: false,
        templates: false,
    };
    await SyncService.PullData(
        projectStore.activeProject.uri, projectStore.getActiveProjectUrl, false, syncOptions
    )
        .then(async () => {
            await projectStore.reloadActiveProject()
            await userStore.reloadUsers()
            entitlementStore.fetchEntitlements();
            emitter.emit('refresh-browser');
        }).catch((error) => {
            console.log("Error Syncing Data", error)
            notificationStore.errorNotification("Error Syncing Data", error)
        })
}

const operationsActive = computed(() => {
    return stageStore.operationActive 
    || !!modals.activeModal || !!notificationStore.getProgress.running
    || !!menu.activeMenu || !assetStore.assetsLoaded || stageStore.activeStage !== 'browser'
});

// Periodically checks studio reachability when offline.
// Retries every 5 minutes until the server is reachable again.
function startConnectivityCheckInterval() {
    const RETRY_INTERVAL = 5 * 1000 * 60

    function run() {
        if (studioStore.appOnline) {
            setTimeout(run, RETRY_INTERVAL);
            return;
        }
        studioStore.checkStudioReachability().finally(() => {
            setTimeout(run, RETRY_INTERVAL);
        });
    }

    run();
}

// Polls for sync token changes when online.
// Only runs when a non-Personal studio is selected and the app is online.
function startCheckSycnTokenInterval() {
    function run() {
        if (!studioStore.appOnline) {
            setTimeout(run, 5000);
            return
        }
        if (operationsActive.value) {
            setTimeout(run, 1000);
            return
        }
        if (!projectStore.selectedStudio || (projectStore.selectedStudio.name == "Personal" && !projectStore.isR2Remote)) {
            setTimeout(run, 1000);
            return
        }
        if (!projectStore.activeProject) {
            setTimeout(run, 1000);
            return
        }
        if (!projectStore.getActiveProject.is_downloaded) {
            setTimeout(run, 1000);
            return
        }
        if (projectStore.getActiveProject.is_unsynced) {
            setTimeout(run, 1000);
            return
        }
        ProjectService.GetSyncToken(projectStore.getActiveProjectUrl)
            .then(async (token) => {
                studioStore.appOnline = true;
                if (token) {
                    let syncToken = projectStore.activeProject.sync_token
                    if (syncToken != token) {
                        stageStore.operationActive = true;
                        await projectStore.refreshActiveProject()
                        if (!projectStore.getActiveProject.is_downloaded) {
                            stageStore.operationActive = false;
                            return
                        }
                        if (projectStore.getActiveProject.is_unsynced) {
                            stageStore.operationActive = false;
                            return
                        }

                        await pullData().catch((error) => {
                            console.log(error)
                        })
                        stageStore.operationActive = false;
                    }
                }
            }).catch((error) => {
                if (projectStore.selectedStudio?.name !== 'Personal') {
                    studioStore.appOnline = false;
                }
            }).finally(() => {
                setTimeout(run, 5000);
            });

    }

    run();
}



onMounted(async () => {
    await settingsStore.initializeShowTypeIcons();
    if (!platformStore.isWeb) {
        startCheckSycnTokenInterval();
        startConnectivityCheckInterval();
    }
});
</script>


<style scoped>
@import "@/assets/tray.css";

.app-root {
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  position: fixed;
  top: 0;
  left: 0;
}
</style>


