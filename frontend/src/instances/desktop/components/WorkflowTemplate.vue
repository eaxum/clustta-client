<template>
  <div class="file-hierarchy">
    <WorkflowTemplateItem :item="workflowTemplateItemData" :is-root="true" />
  </div>
</template>

<script setup>
// imports
import { ref, computed, onMounted } from 'vue'

// components
import WorkflowTemplateItem from '@/instances/desktop/components/WorkflowTemplateItem.vue'

// props
const props = defineProps({
  workflowTemplate: {
    type: Object,
    required: true,
  }
});
// computed
const workflowTemplateItemData = computed(() => {
  const rawData = props.workflowTemplate.workflowTemplateItems;

  // Check if rawData is empty
  if (!rawData || Object.keys(rawData).length === 0) {
    return [];
  }

  const transformedData = transformData(rawData);

  console.log(rawData)
  console.log(transformedData)

  return transformedData;
});

// methods

const transformData = (data) => {
  const entitiesMap = new Map();
  let rootEntity = null;

  // Create a map of entities for easy lookup
  for (const entity of data.entities) {
    entitiesMap.set(entity.id, { ...entity, children: [] });
  }

  // Get the parent ID from either entity_id or parent_id
  const getParentId = (item) => item.entity_id || item.parent_id;

  // Process resources and tasks, assigning them to their parent entities
  const processItems = (items) => {
    for (const item of items) {
      const parentId = getParentId(item);
      if (parentId && entitiesMap.has(parentId)) {
        entitiesMap.get(parentId).children.push(item);
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

  // Build the nested structure and find root entity
  for (const entity of data.entities) {
    const parentId = getParentId(entity);
    if (!parentId || !entitiesMap.has(parentId)) {
      rootEntity = entitiesMap.get(entity.id);
    } else {
      const parent = entitiesMap.get(parentId);
      if (parent) {
        parent.children.push(entitiesMap.get(entity.id));
      }
    }
  }

  return rootEntity;
};

onMounted(async () => {
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
  padding: .2rem;
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
</style>