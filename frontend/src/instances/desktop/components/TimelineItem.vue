<template>
    
    <div  class="trash-item-container">
        <div class="trash-item">
            <div @click="toggleVersions(timelineItemIndex)" class="trash-item-meta">
                <div class="profile-picture" :style="{ backgroundColor: timelineItem.avatarColor }"
                    v-tooltip="timelineItem.author_name">
                    <img class="profile-img"
                        :src="timelineItem.author_profile ? timelineItem.author_profile : '/icons/default_profile_picture.svg'">
                </div>
                <div ref="trash_name" class="meta">
                    <div class="trash-item-name" @mouseenter="trayStates.handleHover($event)"
                        @mouseleave="trayStates.resetScroll($event)">
                        <div class="checkpoint-item-label-text" style="font-weight: 400;">
                            {{ timelineItem.comment.replace(/_/g, " ") }}
                        </div>
                    </div>
                    <div class="trash-item-name" @mouseenter="trayStates.handleHover($event)"
                        @mouseleave="trayStates.resetScroll($event)">
                        <div class="checkpoint-item-label-text" style="overflow: hidden; font-size: 12px; text-overflow: ellipsis;">{{
                            utils.formatDate(timelineItem.created_at) }}</div>
                    </div>
                </div>
            </div>

            <div v-if="isExpanded !== timelineItemIndex" class="checkpoint-item-actions">

                <span @click="revertProject(timelineItem.created_at)" class="single-action-button" v-tooltip="'Revert'">
                    <img class="small-icons" :src="getAppIcon('revert')">
                </span>

                <span @click="toggleVersions(timelineItemIndex)" class="single-action-button" v-tooltip="'Expand'">
                    <img class="small-icons" src="/icons/chevron_down_white_slim.svg"
                        :class="{ 'is-active': isExpanded === timelineItemIndex }">
                </span>
            </div>

            <div v-else class="checkpoint-item-actions">

                <span @click="revertProject(timelineItem.created_at)" class="single-action-button" v-tooltip="'Revert'">
                    <img class="small-icons" :src="getAppIcon('revert')">
                </span>

                <span @click="toggleVersions(timelineItemIndex)" class="single-action-button" v-tooltip="'Close'">
                    <img class="small-icons" :src="getAppIcon('close')"
                        :class="{ 'is-active': isExpanded === timelineItemIndex }">
                </span>
            </div>

        </div>

        <transition name="expand" appear>
            <div v-if="timelineItem.task_paths.length" v-show="isExpanded === timelineItemIndex"
                class="trash-checkpoints-root">
                <div class="trash-checkpoints">
                    <div class="checkpoint-item-children" v-for="(taskPath, index) in timelineItem.task_paths" :key="index">
                        <div class="trash-item-meta">
                            <span><img class="small-icons" :src="getAppIcon('generic')"></span>
                            <div ref="trash_name" class="trash-item-label" @mouseenter="handleHover($event)"
                                @mouseleave="resetScroll($event)">
                                <div @click="selectItem(taskPath)" class="trash-item-label-text">{{ taskPath }}</div>
                            </div>
                        </div>

                    </div>
                </div>
            </div>
        </transition>

    </div>
</template>

<script setup>
import { computed, onMounted, ref, nextTick } from 'vue';
import { CheckpointService, CollectionService, AssetService, TrashService } from "@/../bindings/clustta/services";
import utils from '@/services/utils';
import { useCollectionStore } from '@/stores/collections';

import { useTrayStates } from '@/stores/TrayStates';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useAssetStore } from '@/stores/assets';
import { useStageStore } from '@/stores/stages';
import { useCommonStore } from '@/stores/common';

const props = defineProps({

    timelineItem: Object,
    timelineItemIndex: Number,
    isExpanded: {
        type: Number,
        default: ''
    }

});

const emit = defineEmits(['update-expanded']);

const trayStates = useTrayStates();
const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const assetStore = useAssetStore();
const stage = useStageStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();

const getAppIcon = (iconName) => {
    const icon = iconStore.getAppIcon(iconName);
    return icon
};

const toggleVersions = (index) => {
    if (props.isExpanded !== index) {
        emit('update-expanded', index)

        nextTick(() => {
            const element = document.querySelectorAll('.trash-item-container')[index];
            // element.scrollIntoView({ behavior: 'smooth', block: 'center' });
        });
    } else {
        emit('update-expanded', -1)
    }
};

const formatName = (name, type) => {
    // return name
    return utils.formatDate(name.slice(-20));

    if (type.includes('checkpoint')) {
        // return name.slice(0, -20);
        return utils.formatDate(name.slice(-20));
    } else return name;

}

const handleHover = (event) => {
    let element = event.target;
    const elementChild = event.target.children[0];

    const isOverflowing = element.scrollWidth > element.offsetWidth;
    const scrollDist = element.scrollWidth - element.offsetWidth;
    if (isOverflowing) {

        elementChild.style.transform = 'translateX(' + -(scrollDist) + 'px)';
        elementChild.style.transition = (scrollDist / 12) + 's linear';
        // //console.log(elementChild.style.transition);

    }
};
const resetScroll = (event) => {
    let element = event.target;
    const elementChild = event.target.children[0];

    elementChild.style.transform = 'translateX(0px)';
    elementChild.style.transition = 0 + 's linear';

}

const revertProject = async (createdAt) => {
    CheckpointService.RevertProject(projectStore.activeProject.uri, projectStore.getActiveProjectUrl, createdAt)
        .then(async (response) => {
            notificationStore.addNotification(
                "Project Successfully Reverted",
                "",
                "success",
                false
            )
            assetStore.refreshAllFilesStatus()
        })
        .catch((error) => {
            console.log(error)
            notificationStore.errorNotification("Error Reverting Project", error)
        });
};

const selectedTaskId = ref('');

const selectItem = async (taskPath) => {
    await findItem(taskPath);
    // scrollIntoView();

}

// collectionStore.navigateToCollection(entity);
//   commonStore.navigatorMode = true;

const findItem = async (taskPath) => {
    const allEntities = await CollectionService.GetCollections(projectStore.activeProject.uri)
    const allTasks = await AssetService.GetAssets(projectStore.activeProject.uri)
    const task = allTasks.find((item) => item.task_path === taskPath);
    const taskId = task?.id;
    const taskParent = allEntities.find((item) => item.id === task.entity_id );
    
    if(taskParent){
        collectionStore.navigateToCollection(taskParent);
        commonStore.navigatorMode = true;
    } 
    
    stage.deselectAllItems();
    assetStore.selectAsset(taskId)
    stage.firstSelectedItemId = taskId;
    stage.markedItems = [taskId];
    selectedTaskId.value = taskId;

    return
    const taskParents = assetStore.getAssetEntity(taskId, true);
    const taskParentIds = taskParents.map((item) => item.id)

    for(const parentId of taskParentIds){
        if(parentId in stage.expandedEntities){
            
        } else {
            stage.expandEntity(parentId)
        }
    }
    
    stage.deselectAllItems();
    assetStore.selectAsset(taskId)
    stage.firstSelectedItemId = taskId;
    stage.markedItems = [taskId];
    selectedTaskId.value = taskId;
    
};

const scrollIntoView = () => {
    
    const taskId = selectedTaskId.value;

    const item_class = 'virtua-item-header';
    const allElements = Array.from(document.querySelectorAll(`.${item_class}`));
    const allItemsIds = allElements.map((item) => item.id);
    const currentIndex = allItemsIds.indexOf(taskId);
    const selectedElement = allElements[currentIndex];

    nextTick(() => {
        selectedElement.scrollIntoView({
            behavior: 'auto',
            block: 'center'
        });
    });

};


onMounted(() => {
});

</script>

<style scoped>
@import "@/assets/desktop.css";

.is-active {
    transform: rotate(180deg);

}

.trash-item-container {
    position: relative;
    cursor: auto;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    width: 100%;
    gap: .5rem;
    display: flex;
    gap: .2rem;
    color: var(--white);
    align-items: center;
    box-sizing: border-box;
    width: 100%;
    border-radius: 10px;
    overflow: hidden;
    height: 60px;
    min-height: max-content;
    border-radius: 12px;
    outline: var(--transparent-line);
    outline-offset: -1px;
    background-color: var(--dark-steel);
}


.trash-item {
    position: relative;
    cursor: auto;
    box-sizing: border-box;
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: center;
    width: 100%;
    padding-right: .5rem;
    overflow: hidden;
}

.checkpoint-item-children {
    position: relative;
    cursor: auto;
    box-sizing: border-box;
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: center;
    width: 100%;
    gap: 2rem;
    overflow: hidden;
}

.trash-item-container:hover {
    outline: var(--solid-line);
    outline-offset: -1.5px;

}

.trash-checkpoints-root {
    box-sizing: border-box;
    position: relative;
    height: max-content;
    width: 100%;
    overflow: hidden;
    padding: 0 .3rem;
    padding-bottom: .3rem;
}

.trash-checkpoints{
    border-radius: var(--small-radius);
    background-color: var(--light-steel);
    padding: .5rem 0;
    padding-right: 1rem;
    display: flex;
    flex-direction: column;
    gap: .5rem;
}

.trash-checkpoints-root-closed {
    position: relative;
    background-color: rgb(51, 51, 51);
    border-bottom-left-radius: 10px;
    border-bottom-right-radius: 10px;
    transition: height 0.2s ease-in-out;
    height: 100px;
    width: 100%;
    overflow: hidden;
}

.trash-item-meta {
    padding-left: .5rem;
    box-sizing: border-box;
    overflow: hidden;
    display: flex;
    flex-direction: row;
    align-items: center;
    gap: .5rem;
    width: 100%;
}

.tree {
    position: absolute;
    top: 0px;
    left: 19.6px;
    width: 1px;
    height: 100%;
    background-color: var(--white);
}

.status-pill {
    box-sizing: border-box;
    background-color: rgb(81, 224, 100);
    width: 15px;
    height: 15px;
    border-radius: 20px;
    margin: 0px 5px;
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
}

.status-line {
    box-sizing: border-box;
    background-color: rgb(81, 224, 100);
    width: .3rem;
    height: 50%;
    border-radius: 20px;
    margin: 0px 5px;
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
}

.status-pill-text {
    font-family: 'Inter', sans-serif;
    font-size: 12px;
    font-weight: 200;
    color: rgb(15, 15, 15);
}

.trash-item-label {
    align-items: center;
    overflow: hidden;
    width: 100%;
    display: flex;
    display: none;
    white-space: nowrap;
    overflow: hidden;
    display: flex;
}

.meta {
    flex-direction: column;
    align-items: flex-start;
    min-height: min-content;
    padding: .5rem 0;
    justify-content: center;
    gap: .2rem;
    overflow: hidden;
    width: 100%;
    display: flex;
    display: none;
    white-space: nowrap;
    overflow: hidden;
    display: flex;
}

.trash-item-name {
    color: rgb(224, 224, 224);
    height: min-content;
    width: 100%;
}

.trash-item-entity {
    color: rgb(219, 219, 219);
    background-color: rgba(0, 0, 0, 0.216);
    padding: .3rem;
    border-radius: 5px;
    font-size: 12px;
    transition: all 0.2s ease-in-out;
}

.trash-item-label-text {
    font-family: 'Inter', sans-serif;
    font-size: 14px;
    font-weight: 200;
    color: var(--white);
    text-overflow: ellipsis;
}

.trash-item-label-text:hover {
    text-decoration: underline;
}

.checkpoint-item-label-text {
    color: var(--white);
    font-size: 14px;
    height: min-content;
    overflow: hidden;
    width: 100%;
    text-overflow: ellipsis;
}

.checkpoint-item-actions{
    display: none;
}

.trash-item-container:hover .checkpoint-item-actions {
    display: flex;
}

.profile-picture {
    background-color: red;
    height: 24px;
    min-width: 24px;
    overflow: hidden;
    display: flex;
    align-items: center;
    border-radius: 24px;
}

.profile-img {
    width: 100%;
    height: 100%;
}
</style>