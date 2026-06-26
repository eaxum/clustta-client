<template>
  <div class="collection-item-main" v-stop-propagation>

    <div class="collection-spacer">
      <span v-if="!collection.asset_type_id" @click="expandItem" class="single-action-button">
        <img class="small-icons collection-collapsed" :class="{ 'collection-expanded': isExpanded }"
          :src="getAppIcon('chevron-down')">
      </span>

      <span v-else class="single-action-button">
        <img class="large-icons no-filter" :src="templateIcon">
      </span>
    </div>

    <div class="collection-item-root">

      <div class="collection-item-container drop-zone">

        <div class="collection-item-icon-container">
          <img v-if="!collection.asset_type_id" class="large-icons" :src="getAppIcon(workflowItemIcon)">
          <img v-else class="large-icons" :src="getAppIcon(workflowAssetIcon)">
        </div>

        <div class="collection-item-content selection-area">
          <div class="collection-item-details">
            {{ collection.name }}
          </div>
        </div>

        <ActionButton v-if="isParent && !selectable" @click="editWorkflowItem" :icon="getAppIcon('edit')"
          v-tooltip="$t('blocks.editWorkflow')" />
        <ActionButton v-if="isParent && !selectable" @click="deleteWorkflowItem" :icon="getAppIcon('trash')"
          v-tooltip="$t('blocks.deleteWorkflow')" />
        <ActionButton v-if="selectable" :label="$t('blocks.addWorkflow')" @click="selectWorkflowItem"
          :icon="getAppIcon('plus-circle')" />

      </div>


      <transition name="expand-asset" @enter="utils.startTransition" @after-enter="utils.endTransition"
        @before-leave="utils.startTransition" @after-leave="utils.endTransition">
        <div v-if="isExpanded" class="collection-child-root">

          <WorkflowItem v-for="workflowItem in collection.links" :collection="workflowItem" />
          <WorkflowItem v-for="workflowItem in collection.collections" :collection="workflowItem" />
          <WorkflowItem v-for="workflowItem in collection.assets" :collection="workflowItem" />

        </div>
      </transition>

    </div>
  </div>
</template>

<script setup>


// imports
import { computed, ref, onMounted, onBeforeUnmount, nextTick, watch } from 'vue';
import { useI18n } from 'vue-i18n';
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

const { t } = useI18n();

// props
const props = defineProps({
  collection: Object,
  index: Number,
  isExpanded: { type: Boolean, default: false },
  isParent: { type: Boolean, default: false },
  selectable: { type: Boolean, default: false },
});

const emit = defineEmits(['expand', 'edit', 'delete', 'select'])

// refs

const workflowItemIcon = computed(() => {
  const workflow = props.collection;

  if (workflow.collection_type_id) {
    return collectionStore.getCollectionTypeIcon(workflow.collection_type_id)
  } else {
    return 'folder'
  }
});

const workflowAssetIcon = computed(() => {
  const workflow = props.collection;
  console.log(assetStore.getAssetTypeIcon(workflow.asset_type_id))
  return assetStore.getAssetTypeIcon(workflow.asset_type_id)
});

const templateIcon = computed(() => {
  const workflow = props.collection;
  return templateStore.getAssetTypeIcon(workflow.template_id)
});

// methods
const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const expandItem = () => {
  emit('expand', props.collection.id)
};

const editWorkflowItem = () => {
  emit('edit', props.collection.id)
};

const deleteWorkflowItem = () => {
  emit('delete', props.collection.id)
};

const selectWorkflowItem = () => {
  emit('select', props.collection.id)
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

.collection-collapsed {
  transform: rotate(-90deg);
}

.collection-expanded {
  transform: rotate(0deg);
}

.chevron-inactive {
  opacity: .2;
}

.collection-item-main {
  display: flex;
  gap: .2rem;
  color: var(--text);
  align-items: center;
  padding-left: .5rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: flex-end;
  align-items: flex-start;
  background-color: var(--surface-2);
  border-radius: 10px;
  overflow: hidden;
  padding-right: 0px;

  outline: var(--transparent-line);
  outline-offset: -1px;
  border-radius: var(--large-radius);
  transition: all .2s ease-out;

}

.collection-item-main:hover {
  background-color: var(--surface-3);
  border-radius: var(--small-radius);
  outline: 1px solid var(--surface-4);
}

.collection-item-main-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--surface-1);
  background-color: var(--collection-item-selected);
}

.collection-item-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--selected-soft);
}

.collection-item-last-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--selected);
}

.collection-item-only-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--selected);
}

.collection-item-main-selected:hover {
  outline: 1px solid rgb(255, 255, 255);
  outline-offset: -1px;
}


.collection-drop-zone-hovered {
  width: 100%;
  height: 100%;
  position: absolute;
  opacity: .3;
  background-color: var(--drop-hover);
}

.alt-item {
  /* background-color: red; */
}

.collection-item-root {

  display: flex;
  flex-direction: column;
  gap: .2rem;
  color: var(--text);
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

.collection-item-container {
  display: flex;
  gap: .5rem;
  color: var(--text);
  align-items: center;
  padding: .4rem;
  box-sizing: border-box;
  width: 100%;
  height: 50px;
  justify-content: space-between;
  transition: all .3s ease-out;
  /* background-color: firebrick */
}

.collection-child-root {
  display: flex;
  flex-direction: column;
  gap: .2rem;
  color: var(--text);
  align-items: center;
  box-sizing: border-box;
  width: 100%;
  /* background-color: violet; */
  overflow: hidden;
}


.collection-child-root-collapsed {
  height: 0px;
}

.collection-spacer {
  position: relative;
  width: min-content;
  width: 25px;
  height: 60px;
  display: flex;
  box-sizing: border-box;
  align-items: center;
}

.collection-spacer-empty {
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

.collection-item-preview-container {

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

.collection-item-preview-image {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  height: 100%;
  aspect-ratio: 16 / 9;
  background-color: var(--surface-4);
  border-radius: 5px;
}

.collection-item-icon-container {
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

.collection-item-content {
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


.collection-item-meta-container {
  /* background-color: firebrick; */
  width: 100%;
  display: flex;
  justify-content: flex-end;
}

.collection-item-meta {
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

.collection-item-details {
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


.collection-item-tag {
  display: flex;
  box-sizing: border-box;
  overflow: hidden;
  padding: .1rem .4rem;
  font-size: 12px;
  background-color: black;
  border-radius: 20px;
}


.collection-item-status-container {
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

.collection-item-status {
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

/* .collection-item-status:hover {
  border-radius: 10px;
  transform: scale(1.03);
} */

.collection-item-actions {
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

.collection-item-assignee {
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