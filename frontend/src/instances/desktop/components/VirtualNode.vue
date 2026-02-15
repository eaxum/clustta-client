<template>
  <div class="custom-node-root no-drag" :style="nodeStyle" @dblclick="selectItem" :class="{ 'full-width': forList }">
    <Handle v-if="data.parentId" :style="nodeStyle" class="handle" type="target" :position="Position.Left" />
    <Handle v-if="data.nodeId" :style="nodeStyle" type="source" :position="Position.Right" />
    <div class="virtual-node-root" >
      <div class="virtual-node-container">
        <div v-if="itemTypeIcon" class="virtual-node-icon-container">
          <img class="large-icons" :src="itemTypeIcon">
        </div>
        <!-- <div v-if="commonStore.showThumbs" class="virtual-node-preview-container">
          <div class="virtual-node-preview-image">
            <img v-if="data.preview" class="screenshot-thumb" :src="data.preview">
          </div>
        </div> -->
        <div class="virtual-node-icon-container">
          <img v-if="data.icon" class="large-icons no-filter" :src="data.icon">
        </div>
        <div class="virtual-node-content" v-tooltip="data.task_path">
          <div class="virtual-node-details">
            {{ utils.capitalizeStr(data.name) }}
          </div>
        </div>
        
        <!-- <div v-if="forList && data.type === 'task'" class="virtual-node-count">
          <div class="virtual-node-details">
            {{ dependenciesCount }}
          </div>
        </div> -->
        <div class="virtual-node-actions" v-if="userStore.canDo('manage_dependencies')" >
          <ActionButton v-if="(data.depth === 1 || isDependency) && showRemove" :icon="getAppIcon('minus-circle')" v-tooltip="$t('components.virtualNode.remove')"
            @click="removeDependency" />
          <ActionButton v-if="forList && showAdd" :icon="getAppIcon('plus-circle')" v-tooltip="$t('components.virtualNode.addDependency')"
            @click="addDependency" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>


// imports
import { computed, ref, onMounted } from 'vue';
import { Handle } from '@vue-flow/core';
import { Position } from '@vue-flow/core';
import utils from '@/services/utils';
import emitter from '@/lib/mitt';

// states/store imports
import { useCommonStore } from '@/stores/common';
import { useUserStore } from '@/stores/users';
import { useIconStore } from '@/stores/icons';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// states/stores
const iconStore = useIconStore();
const userStore = useUserStore();
const commonStore = useCommonStore();

// props
const props = defineProps({
  data: Object,
  showAdd: { type: Boolean, default: false },
  showRemove: { type: Boolean, default: false },
  forList: { type: Boolean, default: false },
  isDependency: { type: Boolean, default: false },
});

// computed props
const nodeStyle = computed(() => {
  const item = props.data;
  let itemType;
  if (item.entity_type_id) {
    itemType = 'entity';
    return {
      background: 'var(--entity-item-color)',
      outline: '1px solid var(--entity-item-color)',
      outlineOffset: '-1px'
    }
  } else if (item.task_type_id) {
    if (item.is_resource) {
      itemType = 'resource';
      return {
        background: 'var(--resource-item-color)',
        outline: '1px solid var(--resource-item-color)',
        outlineOffset: '-1px'
      }
    } else {
      itemType = 'task';
      return {
        background: 'var(--task-item-color)',
        outline: '1px solid var(--task-item-color)',
        outlineOffset: '-1px'
      }
    }
  }
});

const itemTypeIcon = computed(() => {
  const item = props.data;
  return item.entity_type_icon ? getAppIcon(item.entity_type_icon) : getAppIcon(item.task_type_icon)
});

const dependenciesCount = computed(() => {
  const item = props.data;
  if(!item.dependencies) return
  if(item?.type !== 'task' ) return 0
  return item.dependencies.length + item.entity_dependencies.length
});

// methods
const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const removeDependency = () => {
  let itemType = props.data.type
  emitter.emit('removeDependency', { id: props.data.id, itemType: itemType });
};

const addDependency = () => {
  let itemType = props.data.type
  emitter.emit('addDependency', { id: props.data.id, itemType: itemType });
};

const selectItem = () => {
  if (!props.data.task_type_id) {
    return
  }
  emitter.emit('selectItem', { message: props.data.id });
};

</script>

<style scoped>
@import "@/assets/desktop.css";

:deep(.vue-flow__handle) {
  width: 6px;
  border-radius: 3px;
  cursor: pointer;
  height: 50%;
  background: #ffffff;
  border: 1px solid #fff;
}

.handle {
  background: red;
}

.custom-node-root {
  text-align: center;
  position: relative;
  height: min-content;
  width: min-content;
  border-radius: 10px;
  color: black;
  cursor: pointer;
  text-align: center;
  background-color: transparent;
  border-radius: var(--large-radius);
}

.custom-node {
  display: flex;
  align-items: center;
  justify-content: center;
  /* background-color: tomato; */
  font-size: 12px;
  overflow: hidden;
  text-align: center;
  position: relative;
  width: 150px;
  height: 70px;
  padding: 10px;
  color: black;

}

.single-action-button-disabled {
  pointer-events: none;
}

.task-collapsed {
  transform: rotate(-90deg);
}

.task-expanded {
  transform: rotate(0deg);
}

.chevron-inactive {
  opacity: .2;
}

.virtual-node-root {
  display: flex;
  flex-direction: column;
  gap: .2rem;
  color: var(--white);
  align-items: center;
  /* padding: .3rem; */
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: flex-end;
  background-color: darkblue;
  background-color: var(--dark-steel);
  border-radius: 14px;
  overflow: hidden;
  background-color: var(--steel);
  /* border-radius: var(--very-large-radius); */

}

.virtual-node-root:hover {
  outline: var(--transparent-line);
  outline-offset: -1.5px;
}

.virtual-node-container {
  display: flex;
  color: var(--white);
  align-items: center;
  padding: .4rem;
  box-sizing: border-box;
  width: 100%;
  height: 50px;
  justify-content: space-between;
  transition: all .3s ease-out;
}

.virtual-node-container-selected {
  outline: 1.5px solid rgb(255, 255, 255);
  outline-offset: -1.5px;
}

.virtual-node-container-selected:hover {
  outline: 1.5px solid rgb(255, 255, 255);
  outline-offset: -1.5px;
}

.task-spacer {
  display: flex;
  align-items: center;
  justify-content: center;
  max-width: 40px;
  min-width: 40px;
  aspect-ratio: 1/1;
  height: 100%;
  overflow: hidden;
}

.task-spacer-empty {
  background-color: moccasin;
}

.checkboxes {
  width: 18px;
  height: 18px;
  border-radius: 4px;
  border: 2px solid yellow;
  background: #FFF;
  padding: 10px;
}

.action-column {
  text-align: center;
  padding: 2px;
  box-sizing: border-box;
  transition: all .3s ease-in;
}

.checkbox-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  padding: .1rem;
  overflow: hidden;
  min-width: min-content;
  height: 100%;
}

.virtual-node-preview-container {

  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  padding: .1rem;
  overflow: hidden;
  min-width: 60px;
  height: 100%;
  aspect-ratio: 16 / 9;
}

.virtual-node-preview-image {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  height: 100%;
  aspect-ratio: 16 / 9;
  background-color: var(--light-steel);
  border-radius: 5px;
}

.virtual-node-icon-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  min-width: min-content;
  padding: .1rem;
  overflow: hidden;
  height: 100%;
}

.virtual-node-content {
  gap: .4rem;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  /* background-color: indigo; */
  width: 100%;
  overflow: hidden;
}

.virtual-node-count {
  gap: .4rem;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  /* background-color: crimson; */
  width: 100%;
  width: min-content;
  min-width: min-content;
  overflow: hidden;
}

.virtual-node-details {
  padding: .2rem;
  flex-wrap: nowrap;
  overflow: hidden;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  height: min-content;
  white-space: nowrap;
  text-overflow: ellipsis;
  font-size: 14px;
}

.virtual-node-meta {
  display: flex;
  padding: .2rem;
  box-sizing: border-box;
  align-items: center;
  gap: .4rem;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.virtual-node-tag {
  display: flex;
  box-sizing: border-box;
  overflow: hidden;
  padding: .1rem .4rem;
  font-size: 12px;
  background-color: black;
  border-radius: 20px;
}


.virtual-node-status-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  padding: .4rem;
  height: 100%;
}

.virtual-node-status {
  display: flex;
  border-radius: var(--normal-radius);
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  width: 50px;
  padding: .4rem .4rem;
  height: max-content;
  background-color: firebrick;
  font-size: 12px;
  text-transform: uppercase;
  font-weight: 700;
  color: black;
}

.virtual-node-actions {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: min-content;
  min-width: max-content;
  gap: .7rem;
  height: 100%;
}

.virtual-node-assignee {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: min-content;
  min-width: max-content;
  gap: .7rem;
  height: 100%;
}

.full-width {
  width: 100%;
}
</style>

