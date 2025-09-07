<template>
  <div class="custom-node-root no-drag" :style="nodeStyle" @dblclick="selectItem" :class="{ 'not-node': !showAdd }">
    <Handle v-if="data.parentId" :style="nodeStyle" class="handle" type="target" :position="Position.Left" />
    <Handle v-if="data.nodeId" :style="nodeStyle" type="source" :position="Position.Right" />
    <div class="task-item-root" v-tooltip="data.task_path">
      <div class="task-item-container">
        <div v-if="itemTypeIcon" class="task-item-icon-container">
          <img class="large-icons" :src="itemTypeIcon">
        </div>
        <div v-if="commonStore.showThumbs" class="task-item-preview-container">
          <div class="task-item-preview-image">
            <img v-if="data.preview" class="screenshot-thumb" :src="data.preview">
          </div>
        </div>
        <div class="task-item-icon-container">
          <img v-if="data.icon" class="large-icons no-filter" :src="data.icon">
        </div>
        <div class="task-item-content">
          <div class="task-item-details">
            {{ utils.capitalizeStr(data.name) }}
          </div>
        </div>
        
        <div v-if="forList && data.type === 'task'" class="task-item-count">
          <div class="task-item-details">
            {{ dependenciesCount }}
          </div>
        </div>
        <div class="task-item-actions" v-if="userStore.canDo('update_task')" >
          <ActionButton v-if="data.depth === 1 || isDependency" :icon="getAppIcon('minus-circle')" v-tooltip="'Remove'"
            @click="removeDependency" />
          <ActionButton v-if="forList" :icon="getAppIcon('plus-circle')" v-tooltip="'Add dependency'"
            @click="addDependency" />
        </div>
      </div>
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
import { computed, ref, onMounted } from 'vue';
import { Handle } from '@vue-flow/core';
import { Position } from '@vue-flow/core';
import utils from '@/services/utils';
import emitter from '@/lib/mitt';

// states/store imports
import { useCommonStore } from '@/stores/common';
import { useUserStore } from '@/stores/users';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'

// states/stores
const userStore = useUserStore();
const commonStore = useCommonStore();

// props
const props = defineProps({
  data: Object,
  showAdd: { type: Boolean, default: false },
  forList: { type: Boolean, default: false },
  isDependency: { type: Boolean, default: false },
});


const nodeStyle = computed(() => {
  const item = props.data;
  let itemType;
  if (item.entity_type_id) {
    itemType = 'entity';
    return {
      background: 'var(--entity-item-color)',
      border: '1px solid var(--entity-item-color)'
    }
  } else if (item.task_type_id) {
    if (item.is_resource) {
      itemType = 'resource';
      return {
        background: 'var(--resource-item-color)',
        border: '1px solid var(--resource-item-color)'
      }
    } else {
      itemType = 'task';
      return {
        background: 'var(--task-item-color)',
        border: '1px solid var(--task-item-color)'
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
  if(item?.type !== 'task' ) return 0
  return item.dependencies.length + item.entity_dependencies.length
});

// methods
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

.task-item-root {

  display: flex;
  flex-direction: column;
  gap: .2rem;
  color: var(--white);
  align-items: center;
  padding: .3rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: flex-end;
  background-color: darkblue;
  background-color: var(--dark-steel);
  border-radius: 10px;
  overflow: hidden;
  background-color: var(--steel);

}

.task-item-root:hover {
  outline: var(--transparent-line);
  outline-offset: -1.5px;
}

.task-item-container {
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

.task-item-container-selected {
  outline: 1.5px solid rgb(255, 255, 255);
  outline-offset: -1.5px;
}

.task-item-container-selected:hover {
  outline: 1.5px solid rgb(255, 255, 255);
  outline-offset: -1.5px;
}

.subtask-item-root {
  display: flex;
  flex-direction: column;
  gap: .2rem;
  color: var(--white);
  align-items: center;
  box-sizing: border-box;
  width: 100%;
  overflow: hidden;
  padding-left: 50px;
}


.subtask-item-root-collapsed {
  height: 0px;
}

.subtask-item-container {
  display: flex;
  gap: .5rem;
  color: var(--white);
  align-items: center;
  padding: .4rem;
  box-sizing: border-box;
  width: 100%;
  height: 50px;
  background-color: var(--dark-steel);
  border-bottom: var(--transparent-line);
}

.subtask-item-container:last-child {
  border-bottom: 0px;
}

.subtask-item-container:hover {
  border-radius: 10px;
  background-color: var(--light-steel);
  outline: var(--transparent-line);
  outline-offset: -1px;
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

.task-item-preview-container {

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

.task-item-preview-image {
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

.task-item-icon-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  min-width: min-content;
  padding: .1rem;
  overflow: hidden;
  height: 100%;
}

.task-item-content {
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

.task-item-count {
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

.task-item-details {
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

.task-item-meta {
  display: flex;
  padding: .2rem;
  box-sizing: border-box;
  align-items: center;
  gap: .4rem;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.task-item-tag {
  display: flex;
  box-sizing: border-box;
  overflow: hidden;
  padding: .1rem .4rem;
  font-size: 12px;
  background-color: black;
  border-radius: 20px;
}


.task-item-status-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  padding: .4rem;
  height: 100%;
}

.task-item-status {
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

.task-item-actions {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: min-content;
  min-width: max-content;
  gap: .7rem;
  height: 100%;
}

.task-item-assignee {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: min-content;
  min-width: max-content;
  gap: .7rem;
  height: 100%;
}

.not-node {
  width: 100%;
}
</style>

