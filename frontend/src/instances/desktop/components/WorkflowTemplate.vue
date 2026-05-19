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
  const collectionsMap = new Map();
  let rootCollection = null;

  // Create a map of collections for easy lookup
  for (const collection of data.collections) {
    collectionsMap.set(collection.id, { ...collection, children: [] });
  }

  // Get the parent ID from either collection_id or parent_id
  const getParentId = (item) => item.collection_id || item.parent_id;

  // Process resources and assets, assigning them to their parent collections
  const processItems = (items) => {
    for (const item of items) {
      const parentId = getParentId(item);
      if (parentId && collectionsMap.has(parentId)) {
        collectionsMap.get(parentId).children.push(item);
      }
    }
  };

  // Process resources and assets if they exist
  if (data.resources) {
    processItems(data.resources);
  }
  if (data.assets) {
    processItems(data.assets);
  }

  // Build the nested structure and find root collection
  for (const collection of data.collections) {
    const parentId = getParentId(collection);
    if (!parentId || !collectionsMap.has(parentId)) {
      rootCollection = collectionsMap.get(collection.id);
    } else {
      const parent = collectionsMap.get(parentId);
      if (parent) {
        parent.children.push(collectionsMap.get(collection.id));
      }
    }
  }

  return rootCollection;
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
  background-color: var(--surface-4);
}

.file-hierarchy::-webkit-scrollbar-track {
  border-radius: 10px;
}
</style>

