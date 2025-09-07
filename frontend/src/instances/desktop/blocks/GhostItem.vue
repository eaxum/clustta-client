<template>
    <div class="ghost-item-main">
      <div v-if="dndStore.draggedItemId  && !commonStore.useGrid" class="drop-propmt-message">
        <div class="drop-propmt-message-content" :class="{ 'drop-propmt-message-error' : isErrMsg }" >
          {{ dropMessage }}
        </div>
      </div>
      <div v-if="stage.markedItems.length === 1" class="ghost-item-wrapper">
        <Collection v-if="data.type === 'entity'" :isGhost="true" :entity="data" :index="index" />
        <Asset v-if="data.type === 'task'" :isGhost="true" :task="data" :index="index" />
        <Collection v-if="data.type === 'untracked_entity'" :isGhost="true" :isUntracked="true" :entity="data" :index="index" />
        <Asset v-if="data.type === 'untracked_task'" :isGhost="true" :isUntracked="true" :task="data" :index="index" />
      </div>
      <div v-else class="single-ghost-item">
        <div class="box depth-1"> {{ stage.markedItems.length + ' items -' }}  {{ dropMessage}}</div>
        <div class="box depth-2"></div>
        <div class="box depth-3"></div>
        <div class="box depth-4"></div>
      </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref, onMounted, onBeforeUnmount, nextTick, watch } from 'vue';

import { useDndStore } from '@/stores/dnd';
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useStageStore } from '@/stores/stages';
import { useCommonStore } from '@/stores/common';
import { useMenu } from '@/stores/menu';


const dndStore = useDndStore();
const collectionStore = useCollectionStore();
const assetStore = useAssetStore();
const stage = useStageStore();
const menu = useMenu();
const commonStore = useCommonStore();

import Asset from '@/instances/desktop/blocks/Asset.vue'
import Collection from '@/instances/desktop/blocks/Collection.vue'

// props
const props = defineProps({
  data: Object,
  index: Number,
});

const isErrMsg = ref('false');

// computed
const dropMessage = computed(() => {
  if (dndStore.isOverlapping) {
    const targetItem = dndStore.targetItem;
    const targetItemType = targetItem?.type;
    const draggedItem = dndStore.draggedItem;
    const draggedItemType = draggedItem?.type;

    // When ALT key is active, always move to project root
    if (dndStore.altKeyActive) {
      isErrMsg.value = false;
      return 'Release to move to project root. ESC key to cancel.';
    }

    // Rule 1: You can't drag anything over an 'untracked_task'
    if (targetItemType === 'untracked_task') {
      isErrMsg.value = true;
      return 'Cannot drop here - ' + targetItem.name + ' is untracked';
    }

    // Rule 2: You can only drag 'untracked_task' and 'untracked_entity' over an 'untracked_entity'
    if (targetItemType === 'untracked_entity') {
      if (draggedItemType === 'untracked_task' || draggedItemType === 'untracked_entity') {
        isErrMsg.value = false;
        return 'Release to move into this folder';
      } else {
        isErrMsg.value = true;
        return 'Cannot drop here - You cant move a tracked item into an untracked one';
      }
    }

    // Rule 3: You can drag anything over an 'entity'
    if (targetItemType === 'entity') {
      isErrMsg.value = false;
      return 'Release to move this item into ' + targetItem.name;
    }

    // Rule 4: Only an 'entity' can be dragged over a 'task'
    if (targetItemType === 'task') {
      if (draggedItemType === 'entity' || draggedItemType === 'task') {
        isErrMsg.value = false;
        return 'Release to make this a dependency of ' + targetItem.name;
      } else {
        isErrMsg.value = true;
        return 'Cannot drop here - ' + draggedItem.name + ' is untracked' ;
      }
    }

    // Default fallback
    return 'Cannot drop here';
  }

  if (!dndStore.targetItem) {
    if (!dndStore.altKeyActive) {
      isErrMsg.value = false;
      return 'ALT key to move to project root. Release or ESC key to cancel.';
    } else {
      return 'Release to move to project root.';
    }
  }
  
  return 'Cannot drop here';
});

</script>

<style scoped>
@import "@/assets/desktop.css";

.root-content {
  display: flex;
  gap: .2rem;
  color: white;
  align-items: center;
  padding-left: .5rem;
  box-sizing: border-box;
  width: 100%;
  height: 60px;
  /* justify-content: flex-end; */
  align-items: center;
  background-color: var(--dark-steel);
  border-radius: 10px;
  overflow: hidden;
  padding-left: 10px;

  outline: var(--transparent-line);
  outline-offset: -1px;

}

.ghost-item-main {
  width: 100%;
  height: min-content;
  position: relative;
  border-radius: 10px;
  box-sizing: border-box;
  overflow: hidden;
  padding: 0px .1rem;
}

.ghost-outline {
  /* border-radius: var(--large-radius); */
}

.is-overlapping {
  background-color: #228b2266;
}

.drop-propmt-message {
  z-index: 100;
  display: flex;
  width: 100%;
  /* width: max-content; */
  height: 100%;
  /* background-color: firebrick; */
  position: absolute;
  right: 50%;
  top: 0;
  transform: translateX(50%);
  justify-content: center;
  align-items: center;
  /* border-radius: var(--large-radius); */
}

.drop-propmt-message-content {
  display: flex;
  width: max-content;
  height: max-content;
  font-size: 14px;
  padding: .2rem .5rem;
  /* background-color: goldenrod; */
  background-color: var(--black-steel);
  font-weight: 100;
  border-radius: 5px;
  text-wrap: nowrap;
}
.drop-propmt-message-error{
  background-color: crimson;
}

.ghost-item-wrapper {
    /* opacity: 0.2; */
  }

  .single-ghost-item{
  z-index: 10000000000;
  gap: .2rem;
  box-sizing: border-box;
  width: 100%;
  height: 80px;
  align-items: center;


  position: relative;

  }


.box {
  /* font-weight: bold; */
  position: absolute;
  width: 98%;
  height: 60px;
  border-radius: 10px;
  background-color: var(--dark-steel);
  outline: var(--transparent-line);
  outline-offset: -1px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.depth-1 {
  top: 0;
  left: 0;
  z-index: 3;
  outline: var(--solid-line);
  outline-offset: -1px;
}

.depth-2 {
  top: 5px;
  left: 5px;
  z-index: 2;
  opacity: .75;
}

.depth-3 {
  top: 10px;
  left: 10px;
  z-index: 1;
  opacity: .5;
}

.depth-4 {
  top: 15px;
  left: 15px;
  z-index: 1;
  opacity: .25;
}
</style>



