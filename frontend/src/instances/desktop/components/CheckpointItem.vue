<template>
    <div @click="enterCheckpoint($event, index)" class="checkpoint-item" v-esc="closeAllCheckpoints" :key="index"
        :class="{ 'checkpoint-item-recent': justViewed === index, 'checkpoint-active': checkpoint.hash === taskHash }"
        :style="{ animationDelay: index < 10 ? `${(index - 1) * 0.05}s` : '0s' }">

        <div class="checkpoint-item-content" :class="{ 'checkpoint-item-content-active': isExpanded === index }">
            <div class="checkpoint-item-profile">
                <div class="profile-picture" :style="{ backgroundColor: checkpoint.avatarColor }"
                    v-tooltip="checkpoint.author">
                    <img class="profile-img"
                        :src="checkpoint.author_profile ? checkpoint.author_profile : '/icons/default_profile_picture.svg'">
                </div>
            </div>
            <div class="checkpoint-item-data">
                <div class="checkpoint-item-meta" :class="{ 'checkpoint-item-meta-active': isExpanded === index }"
                    @mouseenter="trayStates.handleHover($event)" @mouseleave="trayStates.resetScroll($event)">
                    <div class="checkpoint-item-label-text checkpoint-item-message"
                        style="overflow: hidden; text-overflow: ellipsis;">
                        {{ checkpoint.comment }}
                    </div>

                </div>
                <div class="checkpoint-item-meta">
                    <p class="checkpoint-item-label-text">{{ utils.formatDate(checkpoint.created_at) }}</p>
                </div>
            </div>

            <div v-if="isExpanded === index" class="checkpoint-item-actions close-button">
                <ActionButton :plainBackground="true" :icon="getAppIcon('close')" v-tooltip="'Close'"
                    @click="leaveCheckpoint($event, index)" />
            </div>
        </div>

        <div v-if="checkpoint.preview && isExpanded === index" class="checkpoint-item-thumb"
            :class="{ 'checkpoint-item-thumb-active': isExpanded === index }">

            <div class="checkpoint-item-snapshot">
                <img v-if="checkpoint.preview" class="thumbs" :src='checkpoint.preview'>
                <img v-else class="thumbs" src='/page-states/no_image.png'>
            </div>

        </div>

        <div v-if="checkpoint.preview && isExpanded !== index" class="checkpoint-item-attachment">
            <ActionButton :icon="getAppIcon('paper-clip')" v-tooltip="'View attachment'"/>
        </div>

        <div v-if="isExpanded === index" class="menu-divider"></div>

        <div v-if="isExpanded !== index" class="checkpoint-item-actions">
            <ActionButton :icon="getAppIcon('revert')" v-tooltip="'Revert to this Checkpoint'"
                @click="revertToVersion(checkpoint.ownerId, checkpoint.checkpoint_id)" />
            <ActionButton v-if="!checkpoint.is_downloaded" :icon="getAppIcon('cloud-down')"
                v-tooltip="'Download Check Point'" @click="downloadCheckpoint(checkpoint.checkpoint_id)" />
            <ActionButton v-if="checkpoint.is_downloaded" :icon="getAppIcon('launch')" v-tooltip="'Open Check Point'"
                @click="viewVersion(checkpoint.ownerId, checkpoint.checkpoint_id)" />
        </div>

        <div v-else class="full-checkpoint-item-actions">
            <ActionButton :label="'Revert'" :icon="getAppIcon('revert')" v-tooltip="'Revert to this Checkpoint'"
                @click="revertToVersion(checkpoint.ownerId, checkpoint.checkpoint_id)" />
            <ActionButton :label="'Download'" v-if="!checkpoint.is_downloaded" :icon="getAppIcon('cloud-down')"
                v-tooltip="'Download Check Point'" @click="downloadCheckpoint(checkpoint.checkpoint_id)" />
            <ActionButton :label="'Open'" v-if="checkpoint.is_downloaded" :icon="getAppIcon('launch')"
                v-tooltip="'Open Check Point'" @click="viewVersion(checkpoint.ownerId, checkpoint.checkpoint_id)" />
            <ActionButton :label="'Delete'" :icon="getAppIcon('trash')" v-tooltip="'Delete Checkpoint'"
                @click="prepDeletePopUpModal(checkpoint.checkpoint_id)" />
        </div>

    </div>
</template>

<script setup>
import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
    const icon = iconStore.getAppIcon(iconName);
    return icon
};
// imports
import { ref, onMounted, onBeforeUnmount, computed, nextTick, watch } from 'vue';
import utils from '@/services/utils';

// services
import { SyncService } from "@/../bindings/clustta/services";
import { CheckpointService } from "@/../bindings/clustta/services";

// stores/state imports
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';
import { useAssetStore } from '@/stores/assets';
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';

// components
import CheckpointListSkeleton from '@/instances/common/components/CheckpointListSkeleton.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import { useProjectStore } from '@/stores/projects';
import { FSService, TaskService } from '@/../bindings/clustta/services/index';

// props
const props = defineProps({
    checkpoint: {
        type: Object,
        required: true
    },
    index: {
        type: Number,
        required: true
    },
    taskHash: {
        type: String,
        default: ''
    },
    isExpanded: {
        type: Number,
        default: ''
    }
});

// emits
const emit = defineEmits(['refreshCheckpoints', 'update-task-hash', 'update-expanded']);

// stores/states
const modals = useDesktopModalStore();
const userStore = useUserStore();
const trayStates = useTrayStates();
const assetStore = useAssetStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();

// refs
const itemVersionId = ref(null);
const elements = ref([]);
// const isExpanded = ref(-1);
const justViewed = ref(-1);

// methods
const downloadCheckpoint = (checkpointId) => {
    const callback = (progress) => {
        notificationStore.updateProgress(progress);
    }
    notificationStore.cancleFunction = SyncService.CancelSync
    notificationStore.canCancel = true
    SyncService.DownloadCheckpoint(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, checkpointId)
        .then((response) => {
            emit('refreshCheckpoints');
        })
        .catch((error) => {
            notificationStore.resetProgress()
            notificationStore.addNotification(
                "Error Downloading Checkpoint",
                error.message,
                "error",
                false
            )
        });
};

const revertToVersion = (id, checkpointId) => {
    CheckpointService.RevertToCheckpoint(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, id, checkpointId)
        .then((response) => {
            emit('update-task-hash');
            emit('refreshCheckpoints');
            assetStore.refreshDisplayedFilesStatus()
        })
        .catch((error) => {
            console.log(error)
            notificationStore.addNotification(
                "Error Reverting to Version",
                error.message,
                "error",
                false
            )
        });
    trayStates.navigateBack();
};

const viewVersion = (id, checkpointId) => {
    CheckpointService.ViewCheckpoint(projectStore.activeProject.uri, checkpointId, assetStore.selectedTask.name, assetStore.selectedTask.extension)
        .then((response) => {
            //console.log(response)
        })
        .catch((error) => {
            notificationStore.addNotification(
                "Error Reverting to Version",
                error.message,
                "error",
                false
            )
        });
};

let timeoutId;

const enterCheckpoint = (event, index) => {
    // console.log(index)
    // justViewed.value = index;
    if (props.isExpanded !== index) {
        emit('update-expanded', index)

        nextTick(() => {
            const element = document.querySelectorAll('.checkpoint-item')[index];
            element.scrollIntoView({ behavior: 'smooth', block: 'start' });
        });
    };
};

const leaveCheckpoint = (event, index) => {
    emit('update-expanded', -1)
    nextTick(() => {
        const element = document.querySelectorAll('.checkpoint-item')[index];
        element.scrollIntoView({ behavior: 'smooth', block: 'center' });
    });
};

const closeAllCheckpoints = () => {
    emit('update-expanded', -1)
}

const deleteVersion = async () => {
    let checkpoint_id = itemVersionId.value;
    CheckpointService.DeleteCheckpoint(projectStore.activeProject.uri, checkpoint_id)
        .then((response) => {
            emit('refreshCheckpoints');
        })
        .catch((error) => {
            notificationStore.addNotification(
                "Error Deleting Checkpoint",
                error.message,
                "error",
                false
            )
        });

    modals.setModalVisibility('popUpModal', false);
    modals.setModalVisibility('projectMenu', false);
};

const prepDeletePopUpModal = (checkpointId) => {
    trayStates.popUpModalTitle = "Delete Checkpoint";
    trayStates.popUpModalMessage = "Are you sure you want to delete this Checkpoint?";
    trayStates.popUpModalIcon = 'trash';
    itemVersionId.value = checkpointId;
    trayStates.popUpModalFunction = deleteVersion;
    modals.setModalVisibility('popUpModal', true);
};

onMounted(() => {
    // Access the list elements
    elements.value = document.querySelectorAll('.checkpoint-item-meta');
});

onBeforeUnmount(() => {
});

</script>

<style scoped>
@import "@/assets/desktop.css";

.checkpoint-item {
    z-index: 1;
    box-sizing: border-box;
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: center;
    width: 100%;
    padding: .2rem .5rem ;
    flex-wrap: wrap;
    opacity: 0;
    animation: fadeIn .4s ease-in-out forwards;
    border-radius: 12px;
    outline: var(--transparent-line);
    outline-offset: -1px;
    background-color: var(--dark-steel);
}

@keyframes fadeIn {
    from {
        opacity: 0;
    }

    to {
        opacity: 1;
    }
}

.checkpoint-item-attachment{
    display: flex;
    /* background-color: crimson; */
}

.checkpoint-item:hover .checkpoint-item-attachment {
    display: none;
}

.menu-divider{
	height: 5px;
	margin-top: 10px;
}

.checkpoint-item-recent {
    background-color: rgba(0, 0, 0, 0.185);
    border-radius: 10px;
    outline: var(--solid-line);
}

.checkpoint-item-thumb {
    box-sizing: border-box;
    position: relative;
    width: 96px;
    aspect-ratio: 16/9;
    overflow: hidden;
    padding-left: .4rem;
}

.checkpoint-item-snapshot {
    position: relative;
    overflow: hidden;
    box-sizing: border-box;
    /* background-color: rgb(92, 92, 92); */
    height: 100%;
    width: 30%;
    width: 100%;
    aspect-ratio: 16/9;
    border-radius: 8px;
}


.thumbs {
    width: 100%;
    box-sizing: border-box;
    height: 100%;
    overflow: hidden;
    display: flex;
    align-items: center;
    object-fit: cover;
}

.checkpoint-item-thumb-active {
    width: 100%;
    aspect-ratio: 16/9;
    padding-left: .0;
}

.checkpoint-item:focus-within .checkpoint-item-thumb:hover {
    background-color: darkorange;
    width: 100%;
    height: 200px;
}

.checkpoint-item-static {
    position: absolute;
    background-color: darkkhaki;
    width: 333px;
    height: 200px;
    display: flex;
    justify-content: end;
    align-items: center;
}

.checkpoint-actions {
    z-index: 1;
    box-sizing: border-box;
    background-color: coral;
    overflow: hidden;
    width: 333px;
    width: max-content;
    height: 80%;
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: 1rem;
    align-items: center;
    justify-content: space-evenly;
}

.checkpoint-item-content {
    width: 100%;
    box-sizing: border-box;
    height: 100%;
    overflow: hidden;
    display: flex;
    flex-direction: row;
    overflow: hidden;
    align-items: center;
    padding: .2rem .2rem;
    padding-left: .4rem;
    flex: 1;
    height: max-content;
    /* background-color: indigo; */
    gap: .5rem;
    position: relative;
}

.checkpoint-item-content-active {
    padding-top: .5rem;
    flex: auto;
    align-items: flex-start;
}

.checkpoint-item-profile {
    box-sizing: border-box;
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: center;
    min-width: min-content;
    height: 100%;
    /* background-color: goldenrod; */
}

.checkpoint-item-data {
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
    /* background-color: royalblue; */
}

.checkpoint-item-meta {
    box-sizing: border-box;
    overflow: hidden;
    display: flex;
    height: 100%;
    width: 100%;
    white-space: nowrap;
    align-items: center;
    height: 20px;
    text-overflow: ellipsis;
}

.checkpoint-item-meta-active {
    width: 100%;
    display: flex;
    white-space: pre-wrap;
    height: min-content;
    /* margin-bottom: 10px; */
    /* border-bottom: var(--transparent-line); */
    transition: all 0.2s linear;

}

.checkpoint-item-label-text {
    /* background-color: indianred; */
    font-family: 'Inter', sans-serif;
    font-size: 12px;
    color: var(--silver);
    width: 100%;
    max-width: 100%;
    /* display: flex; */
}

.checkpoint-item-message {
    font-weight: 400;
    text-overflow: ellipsis;
    font-size: 14px;

}

.large {
    font-size: 14px;
}

.checkpoint-item-actions {
    box-sizing: border-box;
    display: none;
    /* display: flex; */
    flex-direction: column;
    flex-direction: row;
    max-width: 100%;
    align-self: stretch;
    align-items: center;
    justify-content: space-evenly;
    gap: 2%;
}


.checkpoint-item:hover {
    outline: var(--solid-line);
}

.checkpoint-item:hover .checkpoint-item-actions {
    display: flex;
}

.close-button{
    position: absolute;
    top: 3;
    right: 0;
}

.full-checkpoint-item-actions {
    padding: .3rem .2rem;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    flex-direction: row;
    max-width: 100%;
    width: 100%;
    align-self: stretch;
    align-items: center;
    justify-content: space-evenly;
}

.checkpoint-active {
    background-color: var(--solid-blue-steel);
}

.checkpoint-active:hover {
    background-color: var(--solid-blue-steel);
}

.profile-picture {
    background-color: red;
    height: 24px;
    min-width: 24px;
    overflow: hidden;
    display: flex;
    align-items: center;
    border-radius: 24px;
    /* padding: 5px; */
}

.profile-img {
    width: 100%;
    height: 100%;
}
</style>



