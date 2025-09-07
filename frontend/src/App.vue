<template>
    <div>
        <ClusttaDesktop v-if="windowName === 'main'" />
    </div>
</template>

<script setup>

import { ref, onMounted, computed } from 'vue';
import { useNotificationStore } from './stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { Events } from "@wailsio/runtime";
import emitter from '@/lib/mitt';

// components
import ClusttaDesktop from '@/instances/desktop/ClusttaDesktop.vue';
import { useAssetStore } from '@/stores/assets';
import { useProjectStore } from './stores/projects';
import { SyncService, ProjectService } from "@/../bindings/clustta/services";
import { System, Window } from "@wailsio/runtime";
import { LogService } from '@/../bindings/clustta/services/index';
import { useStageStore } from './stores/stages';
import { useMenu } from '@/stores/menu';
import { useAccountStore } from '@/stores/accounts';


const windowNameTop = ref();
const projectStore = useProjectStore();
const assetStore = useAssetStore();
const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();
const menu = useMenu();

const stageStore = useStageStore();
const accountStore = useAccountStore();



const disableMenu = () => {
    if (System.IsDebug) {
        return
    }


    document.addEventListener('contextmenu', e => {
        e.preventDefault();
        return false;
    }, { capture: true })

    document.addEventListener('selectstart', e => {
        e.preventDefault();
        return false;
    }, { capture: true })
}

Events.On('progress-update', async (message) => {
    // let progressData = JSON.parse(message.payload);
    let progressData = message.data[0];
    notificationStore.updateProgress(progressData);
});


const windowName = ref(windowNameTop);

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
    return stageStore.operationActive || !!modals.activeModal || !!menu.activeMenu || !assetStore.tasksLoaded || stageStore.activeStage !== 'browser'
});

function startCheckSycnTokenInterval() {
    function run() {
        if (operationsActive.value) {
            setTimeout(run, 1000); // Schedule the next run only after the current one completes
            return
        }
        if (!projectStore.selectedStudio || projectStore.selectedStudio.name == "Personal") {
            setTimeout(run, 1000); // Schedule the next run only after the current one completes
            return
        }
        if (!projectStore.activeProject) {
            setTimeout(run, 1000); // Schedule the next run only after the current one completes
            return
        }
        if (!projectStore.getActiveProject.is_downloaded) {
            setTimeout(run, 1000); // Schedule the next run only after the current one completes
            return
        }
        if (projectStore.getActiveProject.is_unsynced) {
            setTimeout(run, 1000); // Schedule the next run only after the current one completes
            return
        }
        ProjectService.GetSyncToken(projectStore.getActiveProjectUrl)
            .then(async (token) => {
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
                console.log(error)
            }).finally(() => {
                setTimeout(run, 5000); // Schedule the next run only after the current one completes
            });

    }

    run(); // Start the loop
}

function startUpdateFileStatesInterval() {
    function run() {
        updateFileStates().finally(() => {
            setTimeout(run, 1000); // Schedule the next run only after the current one completes
        });
    }

    run(); // Start the loop
}


onMounted(async () => {
    windowNameTop.value = await Window.Name()
    
    // Initialize account store early in app lifecycle
    await accountStore.initialize();
    
    startCheckSycnTokenInterval()
});
</script>


<style scoped>
@import "@/assets/tray.css";
</style>


