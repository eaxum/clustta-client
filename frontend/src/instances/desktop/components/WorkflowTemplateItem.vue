<template>
  <div :class="['hierarchy-item', { 'is-directory': item.type === 'entity', 'hierarchy-item-root': isHierarchyRoot }]"
    @click="console.log()">
    <div class="item-header">

      <span class="hierarchy-item-spacer single-action-button" @click="toggleExpand"
        :class="{ 'no-expand': item.type !== 'entity' || !item.children.length }">
        <img v-if="item.type === 'entity' && item.children.length" class="large-icons hierarchy-entity-collapsed"
          :class="{ 'hierarchy-entity-expanded': isExpanded }" :src="getAppIcon('chevron_down_white_slim')">
      </span>

      <div class="hierarchy-item-type-icon-container">
        <img v-if="itemIcon" class="large-icons" :src="itemIcon">
      </div>

      <div v-if="item.icon" class="hierarchy-item-icon-container">
        <img v-if="item.icon" class="large-icons" :src="item.icon">
      </div>

      <span class="item-name">{{ item.name }}</span>

      <div v-if="isHierarchyRoot" class="hierarchy-header-config">
        <ActionButton v-if="isHierarchyRoot" :icon="getAppIcon('plus-circle')" v-tooltip="'Remove'" @click="addWorkflow()" />
      </div>

      <div v-else class="hierarchy-item-config">
        <div v-if="!item.entity_type_id" class="hierarchy-item-type-options">
          <ListBox v-if="item.task_type_id" :items="assetStore.getAssetTypesNames" :selectedItem="taskType"
            :onSelect="selectTaskType" />
        </div>
        <ActionButton :icon="getAppIcon('trash')" v-tooltip="'Remove'" @click="removeItem(item)" />
      </div>


    </div>

    <div v-if="item.type === 'entity' && isExpanded" class="item-children">
      <WorkflowTemplateItem v-for="child in item.children" :key="child.name" :item="child" />
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
import { ref, computed } from 'vue'

// state imports
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useDndStore } from '@/stores/dnd';

// stores
const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const dndStore = useDndStore();

// components
import ListBox from '@/instances/common/components/ListBox.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue'

// Define props
const props = defineProps({
  item: { type: Object, required: true },
  isExpanded: { type: Boolean, default: false }
});

// refs
const isExpanded = ref(props.isExpanded);
const entityType = ref(props.item.entity_type_name);
const itemType = ref(props.item.type);

const itemTypes = ref(['Task', 'Resource']);

// computed
const isHierarchyRoot = computed(() => {
  return props.item.is_root
});

const taskType = computed(() => {
  return props.item.task_type_name;
});

const resourceType = computed(() => {
  return props.item.resource_type_name;
});

const itemIcon = computed(() => {

  const item = props.item;
  const isProjectRoot = !item.entity_type_id && item.is_root;

  if (item.is_root) {
    if (isProjectRoot) {
      return '/icons/home.svg';
    } else {
      return '/types-icons/' + item.entity_type_icon + '.svg'
    }
  } else if (item.entity_type_id) {
    return '/types-icons/' + item.entity_type_icon + '.svg'
  } else if (item.task_type_id) {
    return '/types-icons/' + item.task_type_icon + '.svg';
  } else {
    return '/types-icons/' + item.resource_type_icon + '.svg';
  }

});


const toggleExpand = () => {
  if (props.item.type === 'entity' && props.item.children.length) {
    isExpanded.value = !isExpanded.value
  }
};

// methods

const addWorkflow = () => {
  console.log('added')
};



const changeItemType = (newItemTypeName) => {

  const itemTypeName = itemType.value.toLowerCase() + 's';
  const lowerNewItemTypeName = newItemTypeName.toLowerCase();
  const formattedNewItemTypeName = newItemTypeName.toLowerCase() + 's';

  if (itemTypeName === formattedNewItemTypeName) {
    return
  }

  let previewDataBefore = dndStore.previewData[itemTypeName];
  const itemId = props.item.id;

  const selectedItem = previewDataBefore.find(item => item.id === itemId);

  const updatedItem = {
    name: selectedItem.name,
    id: selectedItem.id,
    [lowerNewItemTypeName + '_type_name']: "generic",
    [lowerNewItemTypeName + '_type_icon']: "generic",
    [lowerNewItemTypeName + '_type_id']: "c432bcb8-204a-4a95-878f-8613b7cb9b05",
    entity_id: selectedItem.entity_id,
  };

  let reducedPreviewData = previewDataBefore.filter(item => item.id !== itemId);

  dndStore.previewData[itemTypeName] = [...reducedPreviewData];
  dndStore.previewData[formattedNewItemTypeName].push(updatedItem);
  itemType.value = newItemTypeName;

};

const selectTaskType = (taskTypeName) => {

  let previewData = dndStore.previewData['tasks'];
  const itemId = props.item.id;
  const selectedItem = previewData.find(item => item.id === itemId);


  if (selectedItem) {
    selectedItem.task_type_name = taskTypeName;
    selectedItem.task_type_icon = taskTypeName;
    dndStore.previewData['tasks'] = [...previewData];
  }

};

const selectResourceType = (resourceTypeName) => {

  let previewData = dndStore.previewData['resources'];
  const itemId = props.item.id;
  const selectedItem = previewData.find(item => item.id === itemId);


  if (selectedItem) {
    selectedItem.resource_type_name = resourceTypeName;
    selectedItem.resource_type_icon = resourceTypeName;
    dndStore.previewData['resources'] = [...previewData];
  }

};

const selectEntityType = (entityTypeName) => {

  let previewData = dndStore.previewData['entities'];
  const itemId = props.item.id;
  const selectedItem = previewData.find(item => item.id === itemId);


  if (selectedItem) {
    selectedItem.entity_type_name = entityTypeName;
    selectedItem.entity_type_icon = entityTypeName;
    dndStore.previewData['entities'] = [...previewData];
    entityType.value = entityTypeName;
  }

};

</script>

<style scoped>
.large-icons {
  transition: transform 0s;
}

.hierarchy-item-icon-container {
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

.hierarchy-item-config {
  display: flex;
  gap: .5rem;
  flex: 1;
}

.hierarchy-header-config {
  /* background-color: crimson; */
  display: flex;
  gap: .5rem;
  flex: 1;
  justify-content: flex-end;
}

.hierarchy-item-type-options {
  /* min-width: 400px; */
  flex: 1;
  width: 100%;
  display: flex;
  gap: .5rem;
}

.hierarchy-item-type-icon-container {
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

.file-hierarchy {
  /* font-family: Arial, sans-serif; */
  user-select: none;
}

.item-children {
  /* background-color: firebrick; */
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.hierarchy-item {
  display: flex;
  flex-direction: column;
  width: 100%;
  /* background-color: rgba(61, 61, 61, 0.377); */
  padding-left: 40px;
  gap: 5px;
  box-sizing: border-box;
  border-radius: var(--small-radius);
  overflow: hidden;
  text-wrap: nowrap;
}

.hierarchy-item-root {
  padding-left: 0px;
}

.item-header {
  display: flex;
  align-items: center;
  /* justify-content: space-between; */
  cursor: default;
  padding: 5px;
  padding-left: .5rem;
  gap: .5rem;
  border-radius: var(--small-radius);
  background-color: rgba(61, 61, 61, 0.377);
  background-color: var(--steel);
  min-height: 40px;
}

.no-expand {
  cursor: pointer;
  /* background-color: white; */
  pointer-events: none;
}

.item-header:hover.no-expand {
  /* background-color: #f0f0f0; */
  outline: var(--transparent-line);
}

.hierarchy-item-spacer {
  position: relative;
  width: min-content;
  min-width: 30px;
  /* height: 60px; */
  display: flex;
  box-sizing: border-box;
  align-items: center;
  transition: 0s;
  /* background-color: salmon; */
}

.hierarchy-entity-collapsed {
  transform: rotate(-90deg);
}

.hierarchy-entity-expanded {
  transform: rotate(0deg);
}

.item-name {
  /* flex-grow: 1; */
}

.is-directory>.item-header {
  /* font-weight: bold; */
}

.is-root {
  /* background-color: firebrick; */
  /* padding-left: 0px; */
}
</style>






