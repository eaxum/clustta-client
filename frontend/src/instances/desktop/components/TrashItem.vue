<template>
    <div class="trash-item-container">
        <div class="trash-item">
            <div class="trash-item-meta">
                <span><img class="small-icons" :src="trashItemIcon(trashItem.type)"></span>
                <div ref="trash_name" class="trash-item-label">
                    <!-- <p class="trash-item-label-text">{{ trashItem.name }} - {{ trashItem.entity_name }}</p> -->
                    <div class="trash-item-name" @mouseenter="trayStates.handleHover($event)"
                        @mouseleave="trayStates.resetScroll($event)">
                        <p class="trash-item-label-text" style="overflow: hidden; text-overflow: ellipsis;">{{
                            trashItem.name.replace(/_/g, " ") }}</p>
                    </div>
                    <div v-if="trashItem.type === 'task'" class="trash-item-entity">{{
                        trashItem.entity_name }}</div>
                    <div v-if="trashItem.type === 'resource'" class="trash-item-entity">{{
                        trashItem.entity_name }}</div>
                </div>
            </div>

            <span v-if="!trashItem.checkpoints.length" @click="restoreItem(trashItem.id, trashItem.type)"
                class="single-action-button" v-tooltip="'Restore'">
                <img class="small-icons" :src="getAppIcon('undo')">
            </span>

            <span v-else @click="toggleVersions(trashItemIndex)" class="single-action-button" v-tooltip="'Expand'">
                <img class="small-icons" src="/icons/chevron_down_white_slim.svg"
                    :class="{ 'is-active': isExpanded === trashItemIndex }">
            </span>

        </div>

        <transition name="expand" appear>
            <div v-if="trashItem.checkpoints.length" v-show="isExpanded === trashItemIndex" class="trash-checkpoints">
                <div class="tree"></div>
                <div class="trash-item" v-for="(checkpoint, index) in trashItem.checkpoints" :key="index">
                    <div class="trash-item-meta">
                        <span><img class="small-icons" src="/icons/checkpoint_tree.svg"></span>
                        <div ref="trash_name" class="trash-item-label" @mouseenter="handleHover($event)"
                            @mouseleave="resetScroll($event)">
                            <p class="trash-item-label-text">{{ formatName(checkpoint.name, checkpoint.type) }}</p>
                        </div>
                    </div>

                    <span @click="restoreItem(checkpoint.id, checkpoint.type)" class="single-action-button"
                        v-tooltip="'Restore'">
                        <img class="small-icons" :src="getAppIcon('undo')">
                    </span>

                </div>
            </div>
        </transition>

    </div>
</template>

<script setup>
import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
	const icon = iconStore.getAppIcon(iconName);
	return icon
};

import { computed, onMounted, ref, nextTick } from 'vue';
import { useTrayStates } from '@/stores/TrayStates';
import { TrashService } from "@/../bindings/clustta/services";
import utils from '@/services/utils';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';

const props = defineProps({

    trashItem: Object,
    trashItemIndex: Number,
    allTrash: Array

});

const trayStates = useTrayStates();

const notificationStore = useNotificationStore();
const projectStore = useProjectStore();

const isExpanded = ref(-1);

const toggleVersions = (index) => {
    // isExpanded.value = index
    if (isExpanded.value !== index) {
        isExpanded.value = index;
    }
    else isExpanded.value = -1;

    nextTick(() => {
        const element = document.querySelectorAll('.trash-item-container')[index];
        element.scrollIntoView({ behavior: 'smooth', block: 'start' });
    });

    console.log(isExpanded.value)
};

const formatName = (name, type) => {
    // return name
    return utils.formatDate(name.slice(-20));

    if (type.includes('checkpoint')) {
        // return name.slice(0, -20);
        return utils.formatDate(name.slice(-20));
    } else return name;

}
const trashItemIcon = (type) => {

    if (type === 'entity') {
        return '/icons/entity.svg';
    } else if (type === 'task') {
        return '/icons/task.svg';
    } else if (type === 'resource') {
        return '/icons/resources.svg';
    } else if (type === 'template') {
        return '/icons/categories.svg';
    } else {
        return '/icons/all.svg';
    }

};
const restoreItem = async (id, type) => {
    TrashService.Restore(projectStore.activeProject.uri, id, type)
        .then(async (response) => {
            trayStates.trashables = await TrashService.GetTrashs(projectStore.activeProject.uri)
            notificationStore.addNotification(
                "Item Restored",
                "",
                "success",
                false
            )
            trayStates.refreshData()
        })
        .catch((error) => {
            notificationStore.addNotification(
                "Error Restoring Item",
                error.message,
                "error",
                false
            )
        });
};

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

onMounted(() => {
});

</script>

<style scoped>
@import "@/assets/desktop.css";

.is-active {
    transform: rotate(180deg);
    /* background-color: chartreuse; */

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
    padding: .3rem;
    box-sizing: border-box;
    width: 100%;
    background-color: var(--dark-steel);
    border-radius: 10px;
    overflow: hidden;
    height: 60px;
    height: max-content;
}

/* .trash-item-container-active {
    background-color: burlywood;
} */

.trash-item {
    position: relative;
    cursor: auto;
    box-sizing: border-box;
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: center;
    /* height: 16%; */
    width: 100%;
    gap: .5rem;
    padding-right: .5rem;
    overflow: hidden;

}

.trash-item-container:hover {
    outline: var(--transparent-line);
    outline-offset: -1.5px;

}

.trash-checkpoints {
    position: relative;
    background-color: rgb(51, 51, 51);
    border-bottom-left-radius: 10px;
    border-bottom-right-radius: 10px;
    /* transition: height 0.2s ease-in-out; */
    height: max-content;
    width: 100%;
    overflow: hidden;

}

.trash-checkpoints-closed {
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
    height: 80%;
    overflow: hidden;
    display: flex;
    flex-direction: row;
    align-items: center;
    gap: .5rem;
    border-radius: 10px;
    width: 100%;
    flex: 1;
}

.tree {
    position: absolute;
    top: 0px;
    left: 19.6px;
    width: 1px;
    height: 100%;
    background-color: var(--white);

}

/* .trash-item-meta:hover {

    background-color: #ffffff15;
}

.trash-item-meta:active {

    background-color: #00000013;
} */

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
    /* font-weight: 200; */
    color: rgb(15, 15, 15);
    /* background-color: red; */


}

.trash-item-label {
    align-items: center;
    overflow: hidden;
    height: 100%;
    height: 50px;
    width: 100%;
    display: flex;
    white-space: nowrap;
    overflow: hidden;
    display: flex;
    gap: .5rem;
}

.trash-item-name {
    color: var(--white);
    white-space: nowrap;
    overflow: hidden;

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
    font-size: 16px;
    color: var(--white);
}

</style>

