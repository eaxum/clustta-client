<template>
  <div :class="['hierarchy-item', { 'is-directory': item.type === 'entity', 'hierarchy-item-root': isHierarchyRoot }]">
    <div class="item-header" :class="{ 'item-header-selected': isItemSelected }">

      <span class="hierarchy-item-spacer single-action-button" @click="toggleExpand"
        :class="{ 'no-expand': item.type !== 'entity' || !item.children.length }">
        <img v-if="item.type === 'entity' && item.children.length" class="large-icons hierarchy-entity-collapsed"
          :class="{ 'hierarchy-entity-expanded': isExpanded }" :src="getAppIcon('chevron-right')">
      </span>

      <span v-if="!isHierarchyRoot" class="hierarchy-item-spacer single-action-button"
        @click="stage.selectPreviewItem(item)">
        <img class="large-icons"
          :src="isItemSelected ? getAppIcon('checkbox-selected') : getAppIcon('checkbox-unselected')">
      </span>



      <div class="hierarchy-item-title" @click="handleClick($event, item)">

        <div class="hierarchy-item-type-icon-container">
          <img v-if="itemIcon" class="large-icons" :src="getAppIcon(itemIcon)">
        </div>

        <div v-if="item.icon" class="hierarchy-item-icon-container">
          <img class="large-icons" :src="item.icon" @error="$event.target.src = '/file-icons/default.svg'">
        </div>

        <div class="hierarchy-item-name">{{ item.name }}</div>

      </div>




      <div v-if="!isHierarchyRoot && !item.is_tracked_parent" class="hierarchy-item-config">

        <DropDownBox v-if="item.entity_type_id" :items="entityStore.getEntityTypesNames" :selectedItem="entityType"
          :onSelect="selectEntityType" :fullWidth="false" />

        <DropDownBox v-else :items="itemTypes" :selectedItem="itemType" :onSelect="changeItemType" :fullWidth="false" />

        <div v-if="!item.entity_type_id" class="hierarchy-item-type-options">
          <DropDownBox :items="taskTypeNames" :selectedItem="taskType" :onSelect="selectTaskType" :fullWidth="false" />
        </div>

        <ActionButton :icon="getAppIcon('trash')" v-tooltip="'Remove'" @click="removeItem(item)" />
      </div>


    </div>

    <div v-if="item.type === 'entity' && isExpanded" class="item-children">
      <HierarchyItem v-for="child in item.children" :key="child.name" :item="child" :isExpanded="child.is_expanded" />
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
import { ref, computed } from 'vue';
import utils from '@/services/utils';

// state imports
import { useMenu } from '@/stores/menu';
import { useDndStore } from '@/stores/dnd';
import { useTaskStore } from '@/stores/task';
import { useStageStore } from '@/stores/stages';
import { useEntityStore } from '@/stores/entity';
import { useTemplateStore } from '@/stores/template';

// stores
const menu = useMenu();
const stage = useStageStore();
const dndStore = useDndStore();
const taskStore = useTaskStore();
const entityStore = useEntityStore();
const templateStore = useTemplateStore();

// components
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// Define props
const props = defineProps({
  item: { type: Object, required: true },
  isExpanded: { type: Boolean, default: false }
});

// refs
const isExpanded = ref(props.isExpanded);
const itemType = computed(() => {
  return !props.item.is_resource ? 'Task' : 'Resource';
});


// computed
const isItemSelected = computed(() => {
  return stage.markedItems?.includes(props.item.id)
  return dndStore.selectedPreviewItemIds?.includes(props.item.id);
});

const isHierarchyRoot = computed(() => {
  return props.item.root
});

const entityType = computed(() => {
  return props.item.entity_type_id ? props.item.entity_type_name : '';
});

const itemName = computed(() => {
  return props.item.type === 'entity' ? props.item.name : getItemName(props.item.name);
});

const taskType = computed(() => {
  return props.item.task_type_name;
});

const resourceType = computed(() => {
  return props.item.resource_type_name;
});


const taskTypeNames = computed(() => {
  return taskStore.getTaskTypesNames.filter((item) => item !== taskType.value);
});

const itemTypes = computed(() => {

  const allItemTypes = ['task', 'resource'];
  return allItemTypes.filter((item) => item !== itemType.value?.toLowerCase());

});

const itemIcon = computed(() => {

  const item = props.item;
  const isProjectRoot = !item.entity_type_id && item.root;

  if (item.root) {
    if (isProjectRoot) {
      return 'home';
    } else {
      return item.entity_type_icon
    }
  } else if (item.entity_type_id) {
    return item.entity_type_icon
  } else if (item.task_type_id) {
    return item.task_type_icon;
  } else {
    return item.resource_type_icon;
  }

});

// methods
const handleClick = (event, data) => {

  const allData = dndStore.previewData;
  const parentId = data.parent_id || data.entity_id;

  const orphanEntities = allData['entities'].filter((item) => !item.parent_id);
  const orphanTasks = allData['tasks'].filter((item) => !item.entity_id);
  const orphanItems = [...orphanEntities, ...orphanTasks];

  const siblingEntities = allData['entities'].filter((item) => item.parent_id === parentId);
  const siblingTasks = allData['tasks'].filter((item) => item.entity_id === parentId);
  const siblingItems = [...siblingEntities, ...siblingTasks];

  const allItems = parentId ? siblingItems : orphanItems;

  stage.handlePreviewClick(event, data, allItems, true);

};

const getItemName = (itemName) => {
  if (!itemName) {
    return ''
  }

  if (!itemName.includes('.')) {
    return itemName;
  }

  return itemName.split('.').slice(0, -1).join('.');
};

const pluralize = (word) => {
  const pluralRules = {
    'entity': 'entities',
    'task': 'tasks',
    'resource': 'resources'
  };

  return pluralRules[word] || `${word}s`;
};

const removeItem = (item) => {

  const processedIds = new Set();

  const removeItemAndChildren = (currentItem) => {
    if (processedIds.has(currentItem.id)) {
      return;
    }
    processedIds.add(currentItem.id);

    if (currentItem.children?.length > 0) {
      const childrenCopy = [...currentItem.children];

      childrenCopy.forEach(child => {
        if (child.id) {
          removeItemAndChildren(child);
        }
      });
    }

    const itemTypePlural = pluralize(currentItem.type);
    const previewData = dndStore.previewData[itemTypePlural];

    if (Array.isArray(previewData)) {
      dndStore.previewData[itemTypePlural] = previewData.filter(
        item => item.id !== currentItem.id
      );
    } else {
      console.warn(`No preview data found for type: ${itemTypePlural}`);
    }
  };

  removeItemAndChildren(item);
};

const toggleExpand = () => {
  if (props.item.type === 'entity' && props.item.children.length) {
    isExpanded.value = !isExpanded.value
  }
};

const changeItemType = (newItemTypeName) => {
  console.log(newItemTypeName)

  const itemTypeName = itemType.value.toLowerCase() + 's';

  let previewData = dndStore.previewData['tasks'];
  const itemId = props.item.id;
  const selectedItem = previewData.find(item => item.id === itemId);


  if (selectedItem) {
    console.log(selectedItem)
    selectedItem.is_resource = itemTypeName === 'tasks';
    dndStore.previewData['tasks'] = [...previewData];
    // itemType.value = newItemTypeName;
  }

};

const selectTaskType = (taskTypeName) => {

  let newTaskType;
  const taskTypes = taskStore.getTaskTypes;
  newTaskType = taskTypes.find((item) => item.name === taskTypeName);

  let previewData = dndStore.previewData['tasks'];
  const itemId = props.item.id;
  const selectedItem = previewData.find(item => item.id === itemId);


  if (selectedItem) {
    selectedItem.task_type_name = newTaskType.name;
    selectedItem.task_type_icon = newTaskType.icon;
    selectedItem.task_type_id = newTaskType.id;
    dndStore.previewData['tasks'] = [...previewData];
  }

};

const selectEntityType = (entityTypeName) => {

  let newEntityType;
  const entityTypes = entityStore.getEntityTypes;
  newEntityType = entityTypes.find((item) => item.name === entityTypeName);

  let previewData = dndStore.previewData['entities'];
  const itemId = props.item.id;
  const selectedItem = previewData.find(item => item.id === itemId);


  if (selectedItem) {
    selectedItem.entity_type_name = newEntityType.name;
    selectedItem.entity_type_icon = newEntityType.icon;
    selectedItem.entity_type_id = newEntityType.id;
    dndStore.previewData['entities'] = [...previewData];
  }

};

</script>

<style scoped>
.large-icons {
  transition: transform 0s;
}

.hierarchy-item-meta {
  color: rgb(219, 219, 219);
  color: white;
  background-color: var(--black-steel);
  padding: .3rem;
  border-radius: 5px;
  font-size: 13px;
  /* font-weight: 200; */
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

.hierarchy-item-title {
  display: flex;
  gap: .2rem;
  padding: .5rem .2rem;
  flex-wrap: nowrap;
  overflow: hidden;
  box-sizing: border-box;
  width: 100%;
  /* width: 50%; */
  height: 100%;
  /* background-color: forestgreen; */
}

.hierarchy-item-name {
  padding: .2rem 0;
  flex-wrap: nowrap;
  overflow: hidden;
  box-sizing: border-box;
  /* width: 50%; */
  height: 100%;
  height: min-content;
  white-space: nowrap;
  text-overflow: ellipsis;
  /* background-color: crimson; */
}

.hierarchy-item-config {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: .5rem;
  flex: 1;
  min-width: min-content;
  /* background-color: forestgreen; */
}

.meta-config {
  padding-right: .8rem;
}

.hierarchy-item-type-options {
  /* background-color: crimson; */
  /* min-width: 400px; */
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
  padding-left: 20px;
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
  cursor: default;
  padding: 5px;
  padding-left: .5rem;
  gap: .5rem;
  border-radius: var(--small-radius);
  background-color: var(--steel);
  min-height: 40px;
}

.item-header-selected {
  outline: 1px solid rgb(255, 255, 255);
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--blue-steel);
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
  transform: rotate(0deg);
}

.hierarchy-entity-expanded {
  transform: rotate(90deg);
}
</style>