<template>
  <div class="entity-item-main" v-stop-propagation>

    <div class="entity-spacer">
      <span v-if="!entity.task_type_id" @click="expandItem" class="single-action-button">
        <img class="small-icons entity-collapsed" :class="{ 'entity-expanded': isExpanded }"
          :src="getAppIcon('chevron-down')">
      </span>

      <span v-else class="single-action-button">
        <img class="large-icons" :src="templateIcon">
      </span>
    </div>

    <div class="entity-item-root">

      <div class="entity-item-container drop-zone">

        <div class="entity-item-icon-container">
          <img v-if="!entity.task_type_id" class="large-icons" :src="getAppIcon(workflowItemIcon)">
          <img v-else class="large-icons" :src="getAppIcon(workflowTaskIcon)">
        </div>

        <div class="entity-item-content selection-area">
          <div class="entity-item-details">
            {{ entity.name }}
          </div>
        </div>

        <ActionButton v-if="isParent && !selectable" @click="editWorkflowItem" :icon="getAppIcon('edit')"
          v-tooltip="'Edit Workflow'" />
        <ActionButton v-if="isParent && !selectable" @click="deleteWorkflowItem" :icon="getAppIcon('trash')"
          v-tooltip="'Delete Workflow'" />
        <ActionButton v-if="selectable" :label="'Add workflow'" @click="selectWorkflowItem"
          :icon="getAppIcon('plus-circle')" />

      </div>


      <transition name="expand-task" @enter="utils.startTransition" @after-enter="utils.endTransition"
        @before-leave="utils.startTransition" @after-leave="utils.endTransition">
        <div v-if="isExpanded" class="entity-child-root">

          <WorkflowItem v-for="workflowItem in entity.links" :entity="workflowItem" />
          <WorkflowItem v-for="workflowItem in entity.entities" :entity="workflowItem" />
          <WorkflowItem v-for="workflowItem in entity.tasks" :entity="workflowItem" />

        </div>
      </transition>

    </div>
  </div>
</template>

<script setup>


// imports
import { computed, ref, onMounted, onBeforeUnmount, nextTick, watch } from 'vue';
import utils from '@/services/utils';

// states/store imports
import { useIconStore } from '@/stores/icons';
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useTemplateStore } from '@/stores/template';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// emits

const iconStore = useIconStore();
const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const templateStore = useTemplateStore();

// props
const props = defineProps({
  entity: Object,
  index: Number,
  isExpanded: { type: Boolean, default: false },
  isParent: { type: Boolean, default: false },
  selectable: { type: Boolean, default: false },
});

const emit = defineEmits(['expand', 'edit', 'delete', 'select'])

// refs
const taskItem = ref(null);

const workflowItemType = computed(() => {
  const workflow = props.entity;
  if (workflow.task_type_id) {
    return 'Task'
  } else if (workflow.entity_type_id) {
    return 'Entity'
  } else {
    return 'Root'
  }
});

const workflowItemIcon = computed(() => {
  const workflow = props.entity;

  if (workflow.entity_type_id) {
    return collectionStore.getCollectionTypeIcon(workflow.entity_type_id)
  } else {
    return 'folder'
  }
});

const workflowTaskIcon = computed(() => {
  const workflow = props.entity;
  console.log(assetStore.getAssetTypeIcon(workflow.task_type_id))
  return assetStore.getAssetTypeIcon(workflow.task_type_id)
});

const templateIcon = computed(() => {
  const workflow = props.entity;
  return templateStore.getAssetTypeIcon(workflow.template_id)
});

// methods
const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const expandItem = () => {
  emit('expand', props.entity.id)
};

const editWorkflowItem = () => {
  emit('edit', props.entity.id)
};

const deleteWorkflowItem = () => {
  emit('delete', props.entity.id)
};

const selectWorkflowItem = () => {
  emit('select', props.entity.id)
};

onMounted(() => {
});

onBeforeUnmount(() => {
});

</script>

<style scoped>
@import "@/assets/desktop.css";

.single-action-button-disabled {
  pointer-events: none;
}

.entity-collapsed {
  transform: rotate(-90deg);
}

.entity-expanded {
  transform: rotate(0deg);
}

.chevron-inactive {
  opacity: .2;
}

.entity-item-main {
  display: flex;
  gap: .2rem;
  color: var(--white);
  align-items: center;
  padding-left: .5rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: flex-end;
  align-items: flex-start;
  background-color: var(--dark-steel);
  border-radius: 10px;
  overflow: hidden;
  padding-right: 0px;

  outline: var(--transparent-line);
  outline-offset: -1px;
  /* margin-bottom: 5px; */

}

.entity-item-main:hover {
  outline: var(--transparent-line);
  outline: 1px solid rgb(255, 255, 255);
  outline-offset: -1.5px;
  /* background-color: var(--steel); */
}

.entity-item-main-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--black-steel);
  background-color: var(--entity-item-selected);
}

.entity-item-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--blue-steel);
}

.entity-item-last-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--solid-blue-steel);
}

.entity-item-only-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--solid-blue-steel);
}

.entity-item-main-selected:hover {
  outline: 1px solid rgb(255, 255, 255);
  outline-offset: -1px;
}


.entity-drop-zone-hovered {
  width: 100%;
  height: 100%;
  position: absolute;
  opacity: .3;
  background-color: var(--drop-hover);
}

.alt-item {
  /* background-color: red; */
}

.entity-item-root {

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
  border-radius: 10px;
  overflow: hidden;
  padding-right: 0px;

}

.entity-item-container {
  display: flex;
  gap: .5rem;
  color: var(--white);
  align-items: center;
  padding: .4rem;
  box-sizing: border-box;
  width: 100%;
  height: 50px;
  justify-content: space-between;
  transition: all .3s ease-out;
  /* background-color: firebrick */
}

.entity-child-root {
  display: flex;
  flex-direction: column;
  gap: .2rem;
  color: var(--white);
  align-items: center;
  box-sizing: border-box;
  width: 100%;
  /* background-color: violet; */
  overflow: hidden;
}


.entity-child-root-collapsed {
  height: 0px;
}

.entity-spacer {
  position: relative;
  width: min-content;
  width: 25px;
  height: 60px;
  display: flex;
  box-sizing: border-box;
  align-items: center;
}

.entity-spacer-empty {
  background-color: moccasin;
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
  /* background-color: royalblue; */
}

.entity-item-preview-container {

  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  padding: .1rem;
  overflow: hidden;
  min-width: 60px;
  height: 100%;
  aspect-ratio: 16 / 9;
  /* background-color: firebrick; */
}

.entity-item-preview-image {
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

.entity-item-icon-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  min-width: min-content;
  padding: .1rem;
  overflow: hidden;
  height: 100%;
  /* background-color: firebrick; */
}

.entity-item-content {
  gap: .4rem;
  /* flex-direction: column; */
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  /* background-color: indigo; */
  width: 100%;
  overflow: hidden;
}


.entity-item-meta-container {
  /* background-color: firebrick; */
  width: 100%;
  display: flex;
  justify-content: flex-end;
}

.entity-item-meta {
  display: flex;
  padding: .2rem;
  box-sizing: border-box;
  align-items: center;
  gap: .4rem;
  /* width: 100%; */
  height: 100%;
  overflow: hidden;
  /* background-color: forestgreen; */
  font-weight: 100;
  font-size: 14px;
}

.entity-item-details {
  /* display: flex; */
  padding: .2rem;
  flex-wrap: nowrap;
  overflow: hidden;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  /* width: 50%; */
  height: 100%;
  height: min-content;
  white-space: nowrap;
  text-overflow: ellipsis;

  /* background-color: forestgreen; */
}


.entity-item-tag {
  display: flex;
  box-sizing: border-box;
  overflow: hidden;
  padding: .1rem .4rem;
  font-size: 12px;
  background-color: black;
  border-radius: 20px;
}


.entity-item-status-container {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  padding: .4rem;
  height: 100%;
  /* background-color: darkorange; */
  /* flex: 1; */
}

.entity-item-status {
  display: flex;
  /* border-radius: var(--normal-radius); */
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  width: 60px;
  padding: .4rem .4rem;
  height: max-content;
  height: 100%;
  background-color: firebrick;
  font-size: 12px;
  text-transform: uppercase;
  font-weight: 700;
  color: black;
  transition: all 0.2s ease-out;
}

/* .entity-item-status:hover {
  border-radius: 10px;
  transform: scale(1.03);
} */

.entity-item-actions {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: min-content;
  min-width: max-content;
  /* background-color: darkcyan; */
  min-width: var(--actions-width);
  gap: .7rem;
  height: 100%;
  justify-content: flex-end;
  /* flex: 1; */
}

.file-state {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: min-content;
  min-width: max-content;
  gap: .7rem;
  height: 100%;
}

.entity-item-assignee {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: min-content;
  min-width: max-content;
  gap: .7rem;
  height: 100%;
  /* background-color: darkcyan; */
  /* flex: 1; */
}

.profile-spacer {
  /* background-color: red; */
  overflow: hidden;
  display: flex;
  align-items: center;
  border-radius: 20px;
  height: 24px;
  width: 24px;
  /* padding: 5px; */
}
</style>