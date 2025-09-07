<template>

  <div class="config-bar" v-if="formattedData">
    <!-- {{ stage.markedItems }}
    {{ stage.firstSelectedItemId }} -->


    <ActionButton v-if="itemsSelected" :icon="getAppIcon('close')" v-tooltip="'Deselect all'" @click="deselectItems" />

    <div v-if="entities.length || tasks.length" class="selected-items-meta">
      {{ message }}
    </div>

    <div v-else class="selected-items-meta">
      {{ totalCountMessage }}
    </div>

    <div class="hierarchy-item-config">

      <DropDownBox v-if="entitiesSelected" :items="collectionStore.getEntityTypesNames" :selectedItem="entityType"
        :onSelect="selectEntityType" :fullWidth="false" />

      <DropDownBox v-if="tasksSelected" :items="itemTypes" :selectedItem="itemType" :onSelect="changeItemType"
        :fullWidth="false" />

      <div v-if="tasksSelected" class="hierarchy-item-type-options">
        <DropDownBox :items="assetStore.getTaskTypesNames" :selectedItem="taskType" :onSelect="selectTaskType"
          :fullWidth="false" />
      </div>

      <ActionButton v-if="itemsSelected" :icon="getAppIcon('trash')" :iconAfter="true"
        :label="(entitiesSelected || tasksSelected) ? '' : 'Remove selected'" v-tooltip="'Remove Selected'"
        @click="removeItems" />

      <ActionButton v-if="!itemsSelected && emptyEntityIds?.length" :icon="getAppIcon('trash')" :iconAfter="true"
        :label="'Remove Empty Folders'" v-tooltip="'Remove Selected'" @click="removeEmptyFolders" />

    </div>


  </div>

  <TaskListSkeleton v-if="!formattedData" :forModal="true" />

  <div v-else class="file-hierarchy" @scroll="disableListBox">
    <HierarchyItem :item="formattedData" :isExpanded="true" />
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

import emitter from '@/lib/mitt';

import { ref, computed, onMounted } from 'vue'
import { useDndStore } from '@/stores/dnd';
import { useCollectionStore } from '@/stores/collections';
import { useTemplateStore } from '@/stores/template';
import { useMenu } from '@/stores/menu';
import { useAssetStore } from '@/stores/assets';
import { useStageStore } from '@/stores/stages';

// components
import TaskListSkeleton from '@/instances/desktop/components/TaskListSkeleton.vue'
import HierarchyItem from '@/instances/desktop/components/HierarchyItem.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// state imports
const dndStore = useDndStore();
const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const templateStore = useTemplateStore();
const menu = useMenu();
const stage = useStageStore();

// refs
const itemTypes = ref(['task', 'Resource']);
const entityType = ref('');
const taskType = ref('');
const itemType = ref(itemTypes.value[0]);

// computed
const itemsSelected = computed(() => {
  return allSelectedItems.value?.length;
});

const entities = computed(() => {
  return dndStore.previewDataSelectedItems['entities'] ? dndStore.previewDataSelectedItems['entities'] : [];
});

const totalEntities = computed(() => {
  return dndStore.previewData['entities'] ? dndStore.previewData['entities'] : [];
});

const tasks = computed(() => {
  return dndStore.previewDataSelectedItems['tasks'] ? dndStore.previewDataSelectedItems['tasks'] : [];
});

const totalTasks = computed(() => {
  return dndStore.previewData['tasks'] ? dndStore.previewData['tasks'] : [];
});

const totalCountMessage = computed(() => {
  const noOfEntities = totalEntities.value?.length;
  const entitiesMsg = noOfEntities < 2 ? ' entity' : ' entities';
  const noOfTasks = totalTasks.value?.length;
  const tasksMsg = noOfTasks < 2 ? ' task' : ' tasks';
  if (noOfTasks && noOfEntities) {
    return noOfEntities + entitiesMsg + ' and ' + noOfTasks + tasksMsg + ' to import';
  } else if (noOfTasks) {
    return noOfTasks + tasksMsg + ' to import';
  } else if (noOfEntities) {
    return noOfEntities + entitiesMsg + ' to import';
  }
  else return ''
});

const message = computed(() => {
  const noOfEntities = entities.value?.length;
  const entitiesMsg = noOfEntities < 2 ? ' entity' : ' entities';
  const noOfTasks = tasks.value?.length;
  const tasksMsg = noOfTasks < 2 ? ' task' : ' tasks';
  if (noOfTasks && noOfEntities) {
    return noOfEntities + entitiesMsg + ' and ' + noOfTasks + tasksMsg + ' selected';
  } else if (noOfTasks) {
    return noOfTasks + tasksMsg + ' selected';
  } else if (noOfEntities) {
    return noOfEntities + entitiesMsg + ' selected';
  }
  else return ''
});

const allSelectedItems = computed(() => {
  return [...entities.value, ...tasks.value]
});

const entitiesSelected = computed(() => {
  const onlyEntitiesSelected = allSelectedItems.value?.every((item) => item.type === 'entity');
  return itemsSelected.value && onlyEntitiesSelected;
});

const tasksSelected = computed(() => {
  const onlyTasksSelected = allSelectedItems.value?.every((item) => item.type === 'task');
  return itemsSelected.value && onlyTasksSelected;
});

const targetEntity = computed(() => {
  if (!dndStore.targetItemId) {
    return null;
  }
  return collectionStore.getEntities.find(entity => entity.id === dndStore.targetItemId);
});

const previewData = computed(() => {
  const rawData = dndStore.previewData
  if (!rawData || Object.keys(rawData).length === 0) {
    return [];
  }

  const simplifiedResponse = simplifyObject(rawData);
  const transformedData = transformData(simplifiedResponse);

  const rootEntity = targetEntity.value;

  const formattedData = {

    name: rootEntity ? rootEntity.name : 'Project Root',
    root: true,
    is_tracked_parent: true,
    type: "entity",
    entity_type_name: rootEntity ? rootEntity.entity_type_name : '',
    entity_type_icon: rootEntity ? rootEntity.entity_type_icon : '',
    entity_type_id: rootEntity ? rootEntity.entity_type_id : '',
    children: transformedData.rootItems,

  };
  const emptyEntities = transformedData.emptyEntities;

  return {
    formattedData,
    emptyEntities
  };
  return formattedData;
});

const formattedData = computed(() => {
  return previewData.value?.formattedData;
});

const emptyEntityIds = computed(() => {
  return previewData.value?.emptyEntities;
});

// methods
const disableListBox = () => {
  emitter.emit('disableListBoxOnScroll');
};

const getIconPath = (filePath) => {
  const extension = filePath.split('.').pop().toLowerCase();
  return `/file-icons/${extension}.svg`;
};

const simplifyObject = (data) => {
  const simplifiedData = {};

  const necessaryKeys = ['id', 'name', 'entity_id', 'parent_id', 'icon', 'file_path', 'is_resource', 'is_expanded', 'is_tracked_parent'];

  for (const key in data) {
    if (Array.isArray(data[key])) {
      simplifiedData[key] = data[key].map(item => {
        const simplifiedItem = {};
        simplifiedItem.type = key === 'entities' ? 'entity' : key.slice(0, -1);
        simplifiedItem.icon = key === 'entities' ? '' : getIconPath(item.file_path);

        for (const itemKey in item) {
          if (necessaryKeys.includes(itemKey) || itemKey.includes('type')) {
            simplifiedItem[itemKey] = item[itemKey];
          }
        }
        return simplifiedItem;
      });
    }
  }

  return simplifiedData;
};

const transformData2 = (data) => {
  const entitiesMap = new Map();
  const rootItems = [];
  const emptyEntities = new Set(); // Using Set for efficient deletion/lookup

  // Create a map of entities for easy lookup and collect all entity IDs initially
  for (const entity of data.entities) {
    entitiesMap.set(entity.id, { ...entity, children: [] });
    emptyEntities.add(entity.id); // Initially add all entity IDs
  }

  // Get the parent ID from either entity_id or parent_id
  const getParentId = (item) => item.entity_id || item.parent_id;

  // Process resources and tasks, assigning them to their parent entities or root
  const processItems = (items) => {
    for (const item of items) {
      const parentId = getParentId(item);
      if (parentId && entitiesMap.has(parentId)) {
        entitiesMap.get(parentId).children.push(item);
        emptyEntities.delete(parentId); // Remove parent ID since it has children
      } else {
        rootItems.push(item);
      }
    }
  };

  // Process resources and tasks if they exist
  if (data.resources) {
    processItems(data.resources);
  }
  if (data.tasks) {
    processItems(data.tasks);
  }

  // Build the nested structure
  for (const entity of data.entities) {
    const parentId = getParentId(entity);
    if (!parentId || !entitiesMap.has(parentId)) {
      rootItems.push(entitiesMap.get(entity.id));
    } else {
      const parent = entitiesMap.get(parentId);
      if (parent) {
        parent.children.push(entitiesMap.get(entity.id));
        emptyEntities.delete(parentId); // Remove parent ID since it has children
      }
    }
  }

  // const emptyEntitiesArray = Array.from(emptyEntities);
  // return rootItems;

  return {
    rootItems,
    emptyEntities: Array.from(emptyEntities) // Convert Set to Array
  };
};

const transformData = (data) => {
  data.entities.sort((a, b) => a.name.localeCompare(b.name));
  data.tasks.sort((a, b) => a.name.localeCompare(b.name));
  const entitiesMap = new Map();
  const rootItems = [];
  const emptyEntities = new Set();
  const childToParent = new Map(); // Track parent relationships

  // Create a map of entities for easy lookup and collect all entity IDs initially
  for (const entity of data.entities) {
    entitiesMap.set(entity.id, { ...entity, children: [] });
    emptyEntities.add(entity.id);
  }

  // Get the parent ID from either entity_id or parent_id
  const getParentId = (item) => item.entity_id || item.parent_id;

  // Process resources and tasks, assigning them to their parent entities or root
  const processItems = (items) => {
    for (const item of items) {
      const parentId = getParentId(item);
      if (parentId && entitiesMap.has(parentId)) {
        entitiesMap.get(parentId).children.push(item);
        emptyEntities.delete(parentId);
      } else {
        rootItems.push(item);
      }
    }
  };

  // Process resources and tasks if they exist
  if (data.resources) {
    processItems(data.resources);
  }
  if (data.tasks) {
    processItems(data.tasks);
  }

  // Build the nested structure and track parent relationships
  for (const entity of data.entities) {
    const parentId = getParentId(entity);
    if (!parentId || !entitiesMap.has(parentId)) {
      rootItems.push(entitiesMap.get(entity.id));
    } else {
      const parent = entitiesMap.get(parentId);
      if (parent) {
        parent.children.push(entitiesMap.get(entity.id));
        emptyEntities.delete(parentId);
        childToParent.set(entity.id, parentId); // Track parent relationship
      }
    }
  }

  // Function to get all siblings of an entity
  const getSiblings = (entityId) => {
    const parentId = childToParent.get(entityId);
    if (!parentId) return [];
    const parent = entitiesMap.get(parentId);
    return parent.children.map(child => child.id || child.entity_id).filter(id => id !== entityId);
  };

  // Function to recursively add parents with no other children
  const addLonelyParents = (entityId, addedParents = new Set()) => {
    const parentId = childToParent.get(entityId);
    if (!parentId || addedParents.has(parentId)) return;

    const siblings = getSiblings(entityId);
    if (siblings.length === 0) {
      emptyEntities.add(parentId);
      addedParents.add(parentId);
      addLonelyParents(parentId, addedParents);
    }
  };

  // Process each childless entity to include its lonely parents
  for (const entityId of emptyEntities) {
    addLonelyParents(entityId);
  }

  return {
    rootItems,
    emptyEntities: Array.from(emptyEntities) // Convert Set to Array
  }

  return {
    rootItems,
    entitiesWithoutChildren: Array.from(entitiesWithoutChildren)
  };
};

// change types
const pluralize = (word) => {
  const pluralRules = {
    'entity': 'entities',
    'task': 'tasks',
    'resource': 'resources'
  };

  return pluralRules[word] || `${word}s`;
};


const changeItemType = (newItemTypeName) => {

  const itemTypeName = newItemTypeName.toLowerCase() + 's';

  let previewData = dndStore.previewData['tasks'];
  const tasks = allSelectedItems.value;

  for (const task of tasks) {

    const taskId = task.id;
    const selectedItem = previewData.find(item => item.id === taskId);


    if (selectedItem) {
      selectedItem.is_resource = itemTypeName !== 'tasks';
      dndStore.previewData['tasks'] = [...previewData];
      itemType.value = newItemTypeName;
    }
  }

};

const selectTaskType = (taskTypeName) => {

  let newTaskType;
  const taskTypes = assetStore.getTaskTypes;
  newTaskType = taskTypes.find((item) => item.name === taskTypeName);


  let previewData = dndStore.previewData['tasks'];
  const tasks = allSelectedItems.value;

  for (const task of tasks) {

    const taskId = task.id;
    const selectedTask = previewData.find(item => item.id === taskId);

    if (selectedTask) {
      selectedTask.task_type_name = newTaskType.name;
      selectedTask.task_type_icon = newTaskType.icon;
      selectedTask.task_type_id = newTaskType.id;
      dndStore.previewData['tasks'] = [...previewData];
    }

  }

  taskType.value = taskTypeName;

};

const selectEntityType = (entityTypeName) => {

  let newEntityType;
  const entityTypes = collectionStore.getEntityTypes;
  newEntityType = entityTypes.find((item) => item.name === entityTypeName);


  let previewData = dndStore.previewData['entities'];
  const entities = allSelectedItems.value;


  for (const entity of entities) {

    const entityId = entity.id;
    const selectedEntity = previewData.find(item => item.id === entityId);

    if (selectedEntity) {
      selectedEntity.entity_type_name = newEntityType.name;
      selectedEntity.entity_type_icon = newEntityType.icon;
      selectedEntity.entity_type_id = newEntityType.id;
      dndStore.previewData['entities'] = [...previewData];
    }

  }

  entityType.value = entityTypeName;


};

const deselectItems = () => {
  dndStore.previewDataActiveItem = null;
  dndStore.previewDataSelectedItems = {};
  stage.markedItems = [];
};

const removeItems = () => {
  const entities = dndStore.previewDataSelectedItems['entities'] ? dndStore.previewDataSelectedItems['entities'] : [];
  const tasks = dndStore.previewDataSelectedItems['tasks'] ? dndStore.previewDataSelectedItems['tasks'] : [];
  const items = [...entities, ...tasks]
  for (const item of items) {
    removeItem(item);
  }
  dndStore.previewDataSelectedItems = {}
};

const removeEmptyFolders = () => {
  const previewData = dndStore.previewData['entities'];
  dndStore.previewData['entities'] = previewData.filter(item => !emptyEntityIds.value.includes(item.id));

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

onMounted(async () => {
  entityType.value = collectionStore.getEntityTypesNames[0];
  taskType.value = assetStore.getTaskTypesNames[0];
});


</script>

<style scoped>
.file-hierarchy {
  /* background-color: crimson; */
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  max-height: 50vh;
  overflow: hidden;
  overflow-y: scroll;
  padding: 10px;
  border-radius: 5px;
}

.file-hierarchy::-webkit-scrollbar {
  width: 4px;
}

.file-hierarchy::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--light-steel);
}

.file-hierarchy::-webkit-scrollbar-track {
  border-radius: 10px;
}

.config-bar {
  display: flex;
  box-sizing: border-box;
  width: 100%;
  padding: .5rem;
  padding-right: 1rem;
  /* flex-direction: column; */
  justify-content: space-between;
  align-items: center;
  gap: .5rem;
  height: 60px;
  /* background-color: cornflowerblue; */
}

.hierarchy-item-config {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: .5rem;
  flex: 1;
  min-width: min-content;
  /* background-color: hotpink; */
}

.selected-items-meta {
  /* background-color: crimson; */
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: .5rem;
  flex: 1;
  min-width: min-content;
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
</style>





