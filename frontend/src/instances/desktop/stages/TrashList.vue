<template>
  <div class="settings-component-root">
    <div class="trash-list-header">
      <HeaderTabs :useTooltip="false" :dataTypes="trashTypes" @filter="filterList" :fullWidth="true" />
    </div>

    <div class="settings-component-container">
      <PageState v-if="!filteredTrashItems.length" :message="emptyMessage" :illustration="emptyIllustration" />

      <div v-else class="trash-list-body">
        <TrashItem v-for="(trashItem, index) in filteredTrashItems" :key="index" :trashItem="trashItem" :trashItemIndex="index" :style="{ animationDelay: index < 10 ? `${(index - 1) * 0.05}s` : '0s' }" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onBeforeMount, onBeforeUnmount, ref } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import HeaderTabs from '@/instances/common/components/HeaderTabs.vue';
import PageState from '@/instances/common/components/PageState.vue';
import TrashItem from '@/instances/desktop/components/TrashItem.vue';

// services
import { TrashService } from '@/services';

// stores
import { useAssetStore } from '@/stores/assets';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';

const assetStore = useAssetStore();
const projectStore = useProjectStore();
const trayStates = useTrayStates();
const { t } = useI18n();

// refs
const assets = ref([]);
const trashTypeFilter = ref('all');

// computed properties
// Returns the empty state illustration based on the current filter.
const emptyIllustration = computed(() => {
  const illustrations = {
    collection: '/page-states/collections.png',
    asset: '/page-states/assets.png',
    template: '/page-states/template.png',
    all: '/page-states/resources.png',
  };
  return illustrations[trashTypeFilter.value] || '/page-states/resources.png';
});

// Returns the empty state message based on the current filter.
const emptyMessage = computed(() => {
  if (trashTypeFilter.value === 'all') return t('stages.noDeletedItems');
  return t('stages.noDeletedItemsByType', { type: trashTypeFilter.value });
});

// Returns trash items filtered by type and search query.
const filteredTrashItems = computed(() => {
  const data = sortedTrashItems.value;
  const query = trayStates.trashSearchQuery?.toLowerCase() || '';
  const hasSearch = trayStates.showTraySearch && query !== '';
  const typeFilter = trashTypeFilter.value;

  return data.filter((item) => {
    const matchesType = typeFilter === 'all' || item.type === typeFilter;
    const matchesSearch = !hasSearch || item.name.toLowerCase().includes(query);
    return matchesType && matchesSearch;
  });
});

// Returns trash items sorted by name with checkpoints nested under parents.
const sortedTrashItems = computed(() => {
  const trashables = trayStates.trashables;
  const allTrash = [];

  for (const key in trashables) {
    if (trashables.hasOwnProperty(key)) {
      const item = trashables[key];
      const collectionName = getMeta(item.type, item.id, item.parent_id);
      allTrash.push({ ...item, collection_name: collectionName });
    }
  }

  const ids = new Set(allTrash.map(item => item.id));
  const orphanItems = allTrash.filter(item => !ids.has(item.parent_id));
  const parentMap = {};

  orphanItems.forEach(item => {
    if (item.type.includes('checkpoint')) {
      if (!parentMap[item.parent_id]) {
        parentMap[item.parent_id] = {
          name: item.name.slice(0, -21),
          type: item.type.replace('_checkpoint', ''),
          id: item.parent_id,
          collection_name: item.collection_name,
          asset_name: item.asset_name,
          parent_id: item.parent_id,
          checkpoints: [],
        };
      }
      parentMap[item.parent_id].checkpoints.push(item);
    } else {
      parentMap[item.id] = { ...item, checkpoints: [] };
    }
  });

  const result = Object.values(parentMap).map(parent => {
    const originalItem = orphanItems.find(item => item.id === parent.id);
    return originalItem ? { ...originalItem, checkpoints: parent.checkpoints } : parent;
  });

  return result.sort((a, b) => a.name.localeCompare(b.name));
});

// Returns trash type filters excluding checkpoints.
const trashTypes = computed(() => {
  return trayStates.trashTypes.filter(trashType => !trashType.name.includes('checkpoint'));
});

// methods
// Filters the trash list by type.
const filterList = (trashType) => {
  trashTypeFilter.value = trashType === 'collections' ? 'collection' : trashType;
};

// Returns the collection name for a trash item.
const getMeta = (type, id, parentId) => {
  if (type === 'asset' || type === 'asset_checkpoint') {
    const lookupId = type === 'asset_checkpoint' ? parentId : id;
    return assets.value.find(item => item.id === lookupId)?.collection_name || '';
  }
  return '';
};

// lifecycle hooks
onBeforeMount(async () => {
  trayStates.showMeta = false;
  assets.value = assetStore.assets;
  trayStates.trashables = await TrashService.GetTrashs(projectStore.activeProject.uri);
});

onBeforeUnmount(() => {
  trayStates.trashables = [];
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.settings-component-root {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
}

.settings-component-container {
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  height: 100%;
  overflow: hidden;
  width: 96%;
  gap: .5rem;
  align-items: center;
  color: white;
  padding: 1rem;
  background-color: var(--black-steel);
  border-radius: var(--very-large-radius);
}

.trash-list-header {
  width: 100%;
  display: flex;
  box-sizing: border-box;
  align-items: flex-start;
  justify-content: flex-start;
}

.trash-list-body {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  gap: .5rem;
  overflow: hidden;
  overflow-y: scroll;
  padding-right: .4rem;
}

.trash-list-body::-webkit-scrollbar {
  width: 4px;
}

.trash-list-body::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--light-steel);
}

.trash-list-body::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}
</style>





