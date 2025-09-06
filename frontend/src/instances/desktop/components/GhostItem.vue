<template>
    <div :style="{
        left: `${dragPosition.x}px`,
        top: `${dragPosition.y}px`, transform: 'translate(-50%, -50%)'
    }" class="file-drop">

        <div class="ghost-item-icon-container">
            <img class="large-icons" :src="canDrop ? '/entity-icons/generic.svg' : '/icons/home.svg'">
        </div>

        <div class="file-drop-indicator">
            Drop in {{ dropLocation }}
        </div>
    </div>
</template>

<script setup>

// imports
import { computed, onMounted } from 'vue';
import { useDndStore } from '@/stores/dnd';
import { useEntityStore } from '@/stores/entity';


const dndStore = useDndStore();
const entityStore = useEntityStore();

const dropLocation = computed(() => {
    if (!dndStore.targetItemId) {
        return 'Root'
    } else {
        const entityName = entityStore.findEntity(dndStore.targetItemId).name;
        return entityName
    }
});

const props = defineProps({
    isDragging: Boolean,
    canDrop: Boolean,
    dragPosition: { type: Object, default: { x: 0, y: 0 } },
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.file-drop {
    position: absolute;
    z-index: 99999;
    width: 260px;
    height: 70px;
    background-color: firebrick;
    border-radius: var(--normal-radius);
    border: var(--transparent-line);
    background-color: var(--white);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: .3rem;
    padding-left: .5rem;
    box-sizing: border-box;
    overflow: hidden;

    outline: var(--transparent-line);
    outline-offset: -1px;
    background-color: var(--light-steel);

}

.ghost-item-icon-container {
    display: flex;
    box-sizing: border-box;
    align-items: center;
    justify-content: center;
    min-width: min-content;
    padding: .1rem;
    overflow: hidden;
    height: 100%;
    /* background-color: firebrick; */
    padding: .3rem;
    box-sizing: border-box;
}

.file-drop-text {
    font-weight: 600;
    font-size: xx-large;
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
}

.file-drop-indicator {
    width: 100%;
    display: flex;
    /* background-color: cadetblue; */
    font-weight: 300;
    height: 100%;
    align-items: center;
    justify-content: flex-start;
    padding: .3rem;
    box-sizing: border-box;
    color: white;
}
</style>