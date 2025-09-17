<template>
	<div ref="navigatorRoot" class="navigator-root" @scroll="disableMenu()">
        <div  class="navigator-item-container"
        :style="gridStyles">

        <GridItem v-for="(child, index) in rootItems" :child="child" :key="child.index" :index="index" 
         @mousedown="onMouseDown($event, child, index)"
        @mouseup="onMouseUp($event, child, index)" :ref="el => handleRef(child.id, el?.$el || el)" />

	</div>
</div>
</template>

<script setup>
// imports
import { computed, ref, nextTick, onMounted, onUnmounted } from 'vue';

import GridItem from '@/instances/common/components/GridItem.vue';

// state imports
import { useMenu } from '@/stores/menu';
import { useStageStore } from '@/stores/stages';
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useAssetStore } from '@/stores/assets';
import { useProjectStore } from '@/stores/projects';
import { useDndStore } from '@/stores/dnd';

// states/stores
const menu = useMenu();
const stage = useStageStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const assetStore = useAssetStore();
const projectStore = useProjectStore();
const dndStore = useDndStore();

const props = defineProps({
  rootItems: { type: Array, default: [] }
});


const gridStyles = computed(() => ({
  display: 'grid',
  boxSizing: 'border-box',
  gridTemplateColumns: `repeat(auto-fill, minmax(${commonStore.gridSize}px, 1fr))`,
  gap: '10px',
  width: '100%',
//   backgroundColor: 'forestGreen'
}));

const navigatorRoot = ref(null);

const handleRef = async (id, el) => {
    if (!el) {
    dndStore.removeRef(id);
    dndStore.removeVisibleItemsRef(id);
    return
    }

    await nextTick()

    const domElement = el instanceof HTMLElement ? el : el.$el

    if (domElement) {
    dndStore.addRef(id, domElement);
    }
};

const dragTimer = ref(null);

const onMouseDown = (event, item, index) => {
    const id = item.id;

    const allItems = props.rootItems;

    let itemType;

    if (item.entity_type_id) {
    itemType = 'entity';
    } else if (item.task_type_id) {
    if (item.is_resource) {
        itemType = 'resource';
    } else {
        itemType = 'task';
    }
    } else if (item.item_type) {
    itemType = item.item_type;
    }

    if(!stage.markedItems.includes(id) || stage.cmdOrCtrlKey(event)){
    stage.handleClick(event, item, itemType, allItems);
    }

    menu.disableAllMenus();
    event.stopPropagation();
    dragItem(event, id);
};

const dragItem = (event, id) => {
    if(stage.operationActive) return
    dragTimer.value = setTimeout(() => {
    onDragStart(event, id);
    }, dndStore.dragDelay);
}

const onMouseUp = (event, item) => {

    if(dndStore.draggedItemId) return 

    const id = item.id;

    const allItems = props.rootItems;

    let itemType;

    if (item.entity_type_id) {
    itemType = 'entity';
    } else if (item.task_type_id) {
    if (item.is_resource) {
        itemType = 'resource';
    } else {
        itemType = 'task';
    }
    } else if (item.item_type) {
    itemType = item.item_type;
    }

    if (stage.markedItems.includes(id) && !stage.cmdOrCtrlKey(event)) {
    stage.handleClick(event, item, itemType, allItems);
    }

    console.log('mouse-up')
    clearTimeout(dragTimer.value);
};

const onDragStart = (e, id) => {
    stage.firstSelectedItemId = '';

    if (!id) {
    return
    }
    dndStore.onDragStart(e, id);
};

// methods
const disableMenu = () => {
	menu.disableAllMenus();
};

onMounted(() => {
});

onUnmounted(() => {
});

</script>

<style scoped>
@import "@/assets/desktop.css";

.navigator-root {
	/* padding: .4rem; */
	padding-right: .4rem;
	position: relative;
	height: 100%;
    height: 100%;
	width: 100%;
	display: flex;
	box-sizing: border-box;
	align-items: flex-start;
	justify-content: center;
	overflow: hidden;
    overflow-y: scroll;
	/* min-width: 550px; */
	/* background-color: forestgreen; */
    border-radius: var(--small-radius);
	/* background-color: var(--steel); */
}



.navigator-root::-webkit-scrollbar {
  width: 6px;
}

.virtua-scroll-tray::-webkit-scrollbar {
  width: 4px;
}

.navigator-root::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: rgba(255, 255, 255, 0.295);

}

.navigator-root::-webkit-scrollbar-track {
  border-radius: 10px;
}

.navigator-item-container {
  position: relative;
  height: 100%;
  height: min-content;
  width: 100%;
  box-sizing: border-box;
  align-items: flex-start;
  justify-content: flex-start;
  overflow: hidden;
  /* background-color: firebrick; */
  flex-direction: column;
  gap: .5rem;
  padding: 1rem;
}

.navigator-item-container-grid {
	display: grid;
	box-sizing: border-box;
	grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
	gap: 10px;
	width: 100%;
}
</style>







