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
import { SyncService, ProjectService } from "@/services";
import { System } from "@wailsio/runtime";
import { LogService } from '@/services';
import { useStageStore } from './stores/stages';
import { useMenu } from '@/stores/menu';
import { useAccountStore } from '@/stores/accounts';
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
const stageStore = useStageStore();
const accountStore = useAccountStore();



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

const handleOpenProjectFile = async (filePath) => {
//  TODO implement reading clustta files
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
        await handleOpenProjectFile(filePath);
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
        task_dependencies: false,
        tasks: false,
        templates: false,
    };
    await SyncService.PullData(
        projectStore.activeProject.uri, projectStore.getActiveProjectUrl, false, syncOptions
    )
        .then(async () => {
            await projectStore.reloadActiveProject()
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

function startCheckSycnTokenInterval() {
    function run() {
        if (operationsActive.value) {
            setTimeout(run, 1000);
            return
        }
        if (!projectStore.selectedStudio || projectStore.selectedStudio.name == "Personal") {
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
        // When offline, still check periodically but at a longer interval to detect reconnection
        const checkInterval = projectStore.serverActive ? 5000 : 60000;
        ProjectService.GetSyncToken(projectStore.getActiveProjectUrl)
            .then(async (token) => {
                projectStore.serverActive = true;
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
                    } else {
                        // console.log("sync token is the same")
                    }
                }
            }).catch((error) => {
                projectStore.serverActive = false;
            }).finally(() => {
                setTimeout(run, checkInterval);
            });

    }

    run(); // Start the loop
}

function startUpdateFileStatesInterval() {
    function run() {
        updateFileStates().finally(() => {
            setTimeout(run, 1000);
        });
    }

    run(); // Start the loop
}


onMounted(async () => {
    if (!platformStore.isWeb) {
        startCheckSycnTokenInterval();
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


