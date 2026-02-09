<template>
  <div v-if="data" class="ghost-card" :class="ghostCardClasses" :style="ghostCardStyles">
    <div class="ghost-item-main">
      <div v-if="stage.markedItems.length === 1" class="ghost-item-wrapper">
        <Collection v-if="data.type === 'entity'" :loadingChildren="false" :isGhost="true" :entity="data" :index="index" />
        <Asset v-if="data.type === 'task'" :loadingAssetState="false" :isGhost="true" :task="data" :index="index" />
        <Collection v-if="data.type === 'untracked_entity'" :isGhost="true" :isUntracked="true" :entity="data" :index="index" />
        <Asset v-if="data.type === 'untracked_task'" :isGhost="true" :isUntracked="true" :task="data" :index="index" />
      </div>
      <div v-else-if="stage.markedItems.length" class="single-ghost-item">
        <div class="box depth-1">{{ stage.markedItems.length + ' items -' }} {{ dropMessage }}</div>
        <div class="box depth-2"></div>
        <div class="box depth-3"></div>
        <div class="box depth-4"></div>
      </div>
      <div v-else class="ghost-item-backdrop">
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onBeforeUnmount, ref, watch } from 'vue';

// components
import Asset from '@/instances/desktop/blocks/Asset.vue';
import Collection from '@/instances/desktop/blocks/Collection.vue';

// stores
import { useCommonStore } from '@/stores/common';
import { useDndStore } from '@/stores/dnd';
import { usePromptStore } from '@/stores/prompts';
import { useStageStore } from '@/stores/stages';

const commonStore = useCommonStore();
const dndStore = useDndStore();
const promptStore = usePromptStore();
const stage = useStageStore();

// props
const props = defineProps({
  data: Object,
  index: Number,
});

// refs
const currentPromptId = ref(null);
const isErrMsg = ref(false);

// computed properties

// Returns the drop message based on current drag state.
const dropMessage = computed(() => {
  if (dndStore.isOverlapping) {
    const targetItem = dndStore.targetItem;
    const targetItemType = targetItem?.type;
    const draggedItem = dndStore.draggedItem;
    const draggedItemType = draggedItem?.type;

    if (dndStore.altKeyActive) {
      isErrMsg.value = false;
      return 'Release to move to project root. ESC key to cancel.';
    }

    if (targetItemType === 'untracked_task') {
      isErrMsg.value = true;
      return 'Cannot drop here - ' + targetItem.name + ' is untracked';
    }

    if (targetItemType === 'untracked_entity') {
      if (draggedItemType === 'untracked_task' || draggedItemType === 'untracked_entity') {
        isErrMsg.value = false;
        return 'Release to move into this folder';
      } else {
        isErrMsg.value = true;
        return 'Cannot drop here - You cant move a tracked item into an untracked one';
      }
    }

    if (targetItemType === 'entity') {
      isErrMsg.value = false;
      return 'Release to move this item into ' + targetItem.name;
    }

    if (targetItemType === 'task') {
      if (draggedItemType === 'entity' || draggedItemType === 'task') {
        isErrMsg.value = false;
        return 'Release to make this a dependency of ' + targetItem.name;
      } else {
        isErrMsg.value = true;
        return 'Cannot drop here - ' + draggedItem.name + ' is untracked';
      }
    }

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

// Returns dynamic classes for the ghost card wrapper.
const ghostCardClasses = computed(() => ({
  'active': dndStore.draggedItemId !== null || dndStore.ghostCardStyle.leaving,
  'single-ghost': !commonStore.useGrid && stage.markedItems.length === 1,
  'leaving': dndStore.ghostCardStyle.leaving,
  'no-target': dndStore.ghostCardStyle.leaving && !dndStore.targetItem,
}));

// Returns inline styles for positioning the ghost card.
// Centers the ghost card on the mouse cursor.
const ghostCardStyles = computed(() => {
  const width = 300;
  const height = stage.markedItems.length === 1 ? 60 : 80;
  return {
    width: `${width}px`,
    left: `${dndStore.ghostCardStyle.pos.x - width / 2}px`,
    top: `${dndStore.ghostCardStyle.pos.y - height / 2}px`,
    transform: dndStore.ghostCardStyle.transform,
  };
});

// watchers
watch(dropMessage, (newMessage, oldMessage) => {
  if (newMessage && newMessage !== oldMessage) {
    if (currentPromptId.value) {
      promptStore.clearPrompt(currentPromptId.value);
    }
    const promptType = isErrMsg.value ? 'error' : 'info';
    currentPromptId.value = promptStore.addPrompt(newMessage, promptType);
  }
}, { immediate: true });

watch(() => dndStore.draggedItemId, (newValue) => {
  if (!newValue && currentPromptId.value) {
    promptStore.clearPrompt(currentPromptId.value);
    currentPromptId.value = null;
  }
});

// lifecycle hooks
onBeforeUnmount(() => {
  if (currentPromptId.value) {
    promptStore.clearPrompt(currentPromptId.value);
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.ghost-card {
  position: fixed;
  z-index: 100;
  user-select: none;
  pointer-events: none;
  opacity: 0;
  transform-origin: center;
  transform: scale(1) rotate(0);
  transition: transform 0.04s ease-in-out;
  border-radius: var(--large-radius);
}

.ghost-card.active {
  opacity: 1;
}

.ghost-card.single-ghost {
  border-radius: var(--large-radius);
}

.ghost-card.leaving {
  transition: opacity 0.15s ease, transform 0.15s ease;
  opacity: 0 !important;
  transform: scale(0.8) !important;
}

.ghost-card.leaving.no-target {
  transform: scale(1.2) !important;
}

.ghost-item-main {
  width: 100%;
  height: min-content;
  position: relative;
  box-sizing: border-box;
  overflow: hidden;
  padding: 0px .1rem;
}

.ghost-item-backdrop{
  width: 300px;
  height: 60px;
  background-color: var(--dark-steel);
  border-radius: var(--large-radius);
}

.single-ghost-item {
  z-index: 10000000000;
  gap: .2rem;
  box-sizing: border-box;
  width: 100%;
  height: 80px;
  align-items: center;
  position: relative;
}

.box {
  position: absolute;
  width: 98%;
  height: 60px;
  border-radius: var(--large-radius);
  background-color: var(--dark-steel);
  outline: var(--transparent-line);
  outline-offset: -1px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--white);
}

.depth-1 {
  font-size: 14px;
  top: 0;
  left: 0;
  z-index: 3;
  outline-offset: -1px;
  padding: 1rem;
  box-sizing: border-box;
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



