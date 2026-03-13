<template>

  <div class="config-bar" v-if="formattedData">
    <!-- {{ stage.markedItems }}
    {{ stage.firstSelectedItemId }} -->


    <ActionButton v-if="itemsSelected" :icon="getAppIcon('close')" v-tooltip="$t('components.importPreview.deselectAll')" @click="deselectItems" />

    <div v-if="collections.length || assets.length" class="selected-items-meta">
      {{ message }}
    </div>

    <div v-else class="selected-items-meta">
      {{ totalCountMessage }}
    </div>

    <div class="hierarchy-item-config">

      <DropDownBox v-if="collectionsSelected" :items="collectionStore.getCollectionTypesNames" :selectedItem="collectionType"
        :onSelect="selectCollectionType" :fullWidth="false" />

      <DropDownBox v-if="assetsSelected" :items="itemTypes" :selectedItem="itemType" :onSelect="changeItemType"
        :fullWidth="false" />

      <div v-if="assetsSelected" class="hierarchy-item-type-options">
        <DropDownBox :items="assetStore.getAssetTypesNames" :selectedItem="assetType" :onSelect="selectAssetType"
          :fullWidth="false" />
      </div>

      <ActionButton v-if="itemsSelected" :icon="getAppIcon('trash')" :iconAfter="true"
        :label="(collectionsSelected || assetsSelected) ? '' : $t('components.importPreview.removeSelected')" v-tooltip="$t('components.importPreview.removeSelectedTooltip')"
        @click="removeItems" />

      <ActionButton v-if="!itemsSelected && emptyCollectionIds?.length" :icon="getAppIcon('trash')" :iconAfter="true"
        :label="$t('components.importPreview.removeEmptyFolders')" v-tooltip="$t('components.importPreview.removeSelectedTooltip')" @click="removeEmptyFolders" />

    </div>


  </div>

  <AssetListSkeleton v-if="!formattedData" :forModal="true" />

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
import { useI18n } from 'vue-i18n';
import { useDndStore } from '@/stores/dnd';
import { useCollectionStore } from '@/stores/collections';
import { useTemplateStore } from '@/stores/template';
import { useMenu } from '@/stores/menu';
import { useAssetStore } from '@/stores/assets';
import { useStageStore } from '@/stores/stages';

// components
import AssetListSkeleton from '@/instances/desktop/components/AssetListSkeleton.vue'
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

const { t } = useI18n();

// refs
const itemTypes = ref(['asset', 'Resource']);
const collectionType = ref('');
const assetType = ref('');
const itemType = ref(itemTypes.value[0]);

// computed
const itemsSelected = computed(() => {
  return allSelectedItems.value?.length;
});

const collections = computed(() => {
  return dndStore.previewDataSelectedItems['collections'] ? dndStore.previewDataSelectedItems['collections'] : [];
});

const totalCollections = computed(() => {
  return dndStore.previewData['collections'] ? dndStore.previewData['collections'] : [];
});

const assets = computed(() => {
  return dndStore.previewDataSelectedItems['assets'] ? dndStore.previewDataSelectedItems['assets'] : [];
});

const totalAssets = computed(() => {
  return dndStore.previewData['assets'] ? dndStore.previewData['assets'] : [];
});

const totalCountMessage = computed(() => {
  const noOfCollections = totalCollections.value?.length;
  const collectionsMsg = noOfCollections < 2 ? ' ' + t('components.importPreview.collection') : ' ' + t('components.importPreview.collections');
  const noOfAssets = totalAssets.value?.length;
  const assetsMsg = noOfAssets < 2 ? ' ' + t('components.importPreview.asset') : ' ' + t('components.importPreview.assets');
  if (noOfAssets && noOfCollections) {
    return noOfCollections + collectionsMsg + ' ' + t('components.importPreview.and') + ' ' + noOfAssets + assetsMsg + ' ' + t('components.importPreview.toImport');
  } else if (noOfAssets) {
    return noOfAssets + assetsMsg + ' ' + t('components.importPreview.toImport');
  } else if (noOfCollections) {
    return noOfCollections + collectionsMsg + ' ' + t('components.importPreview.toImport');
  }
  else return ''
});

const message = computed(() => {
  const noOfCollections = collections.value?.length;
  const collectionsMsg = noOfCollections < 2 ? ' ' + t('components.importPreview.collection') : ' ' + t('components.importPreview.collections');
  const noOfAssets = assets.value?.length;
  const assetsMsg = noOfAssets < 2 ? ' ' + t('components.importPreview.asset') : ' ' + t('components.importPreview.assets');
  if (noOfAssets && noOfCollections) {
    return noOfCollections + collectionsMsg + ' ' + t('components.importPreview.and') + ' ' + noOfAssets + assetsMsg + ' ' + t('components.importPreview.selected');
  } else if (noOfAssets) {
    return noOfAssets + assetsMsg + ' ' + t('components.importPreview.selected');
  } else if (noOfCollections) {
    return noOfCollections + collectionsMsg + ' ' + t('components.importPreview.selected');
  }
  else return ''
});

const allSelectedItems = computed(() => {
  return [...collections.value, ...assets.value]
});

const collectionsSelected = computed(() => {
  const onlyCollectionsSelected = allSelectedItems.value?.every((item) => item.type === 'collection');
  return itemsSelected.value && onlyCollectionsSelected;
});

const assetsSelected = computed(() => {
  const onlyAssetsSelected = allSelectedItems.value?.every((item) => item.type === 'asset');
  return itemsSelected.value && onlyAssetsSelected;
});

const targetCollection = computed(() => {
  if (!dndStore.targetItemId) {
    return null;
  }
  return collectionStore.getCollections.find(collection => collection.id === dndStore.targetItemId);
});

const previewData = computed(() => {
  const rawData = dndStore.previewData
  if (!rawData || Object.keys(rawData).length === 0) {
    return [];
  }

  const simplifiedResponse = simplifyObject(rawData);
  const transformedData = transformData(simplifiedResponse);

  const rootCollection = targetCollection.value;

  const formattedData = {

    name: rootCollection ? rootCollection.name : t('components.importPreview.projectRoot'),
    root: true,
    is_tracked_parent: true,
    type: "collection",
    collection_type_name: rootCollection ? rootCollection.collection_type_name : '',
    collection_type_icon: rootCollection ? rootCollection.collection_type_icon : '',
    collection_type_id: rootCollection ? rootCollection.collection_type_id : '',
    children: transformedData.rootItems,

  };
  const emptyCollections = transformedData.emptyCollections;

  return {
    formattedData,
    emptyCollections
  };
  return formattedData;
});

const formattedData = computed(() => {
  return previewData.value?.formattedData;
});

const emptyCollectionIds = computed(() => {
  return previewData.value?.emptyCollections;
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

  const necessaryKeys = ['id', 'name', 'collection_id', 'parent_id', 'icon', 'file_path', 'is_resource', 'is_expanded', 'is_tracked_parent'];

  for (const key in data) {
    if (Array.isArray(data[key])) {
      simplifiedData[key] = data[key].map(item => {
        const simplifiedItem = {};
        simplifiedItem.type = key === 'collections' ? 'collection' : key.slice(0, -1);
        simplifiedItem.icon = key === 'collections' ? '' : getIconPath(item.file_path);

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
  const collectionsMap = new Map();
  const rootItems = [];
  const emptyCollections = new Set(); // Using Set for efficient deletion/lookup

  // Create a map of collections for easy lookup and collect all collection IDs initially
  for (const collection of data.collections) {
    collectionsMap.set(collection.id, { ...collection, children: [] });
    emptyCollections.add(collection.id); // Initially add all collection IDs
  }

  // Get the parent ID from either collection_id or parent_id
  const getParentId = (item) => item.collection_id || item.parent_id;

  // Process resources and assets, assigning them to their parent collections or root
  const processItems = (items) => {
    for (const item of items) {
      const parentId = getParentId(item);
      if (parentId && collectionsMap.has(parentId)) {
        collectionsMap.get(parentId).children.push(item);
        emptyCollections.delete(parentId); // Remove parent ID since it has children
      } else {
        rootItems.push(item);
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

  // Build the nested structure
  for (const collection of data.collections) {
    const parentId = getParentId(collection);
    if (!parentId || !collectionsMap.has(parentId)) {
      rootItems.push(collectionsMap.get(collection.id));
    } else {
      const parent = collectionsMap.get(parentId);
      if (parent) {
        parent.children.push(collectionsMap.get(collection.id));
        emptyCollections.delete(parentId); // Remove parent ID since it has children
      }
    }
  }

  // const emptyCollectionsArray = Array.from(emptyCollections);
  // return rootItems;

  return {
    rootItems,
    emptyCollections: Array.from(emptyCollections) // Convert Set to Array
  };
};

const transformData = (data) => {
  data.collections.sort((a, b) => a.name.localeCompare(b.name));
  data.assets.sort((a, b) => a.name.localeCompare(b.name));
  const collectionsMap = new Map();
  const rootItems = [];
  const emptyCollections = new Set();
  const childToParent = new Map(); // Track parent relationships

  // Create a map of collections for easy lookup and collect all collection IDs initially
  for (const collection of data.collections) {
    collectionsMap.set(collection.id, { ...collection, children: [] });
    emptyCollections.add(collection.id);
  }

  // Get the parent ID from either collection_id or parent_id
  const getParentId = (item) => item.collection_id || item.parent_id;

  // Process resources and assets, assigning them to their parent collections or root
  const processItems = (items) => {
    for (const item of items) {
      const parentId = getParentId(item);
      if (parentId && collectionsMap.has(parentId)) {
        collectionsMap.get(parentId).children.push(item);
        emptyCollections.delete(parentId);
      } else {
        rootItems.push(item);
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

  // Build the nested structure and track parent relationships
  for (const collection of data.collections) {
    const parentId = getParentId(collection);
    if (!parentId || !collectionsMap.has(parentId)) {
      rootItems.push(collectionsMap.get(collection.id));
    } else {
      const parent = collectionsMap.get(parentId);
      if (parent) {
        parent.children.push(collectionsMap.get(collection.id));
        emptyCollections.delete(parentId);
        childToParent.set(collection.id, parentId); // Track parent relationship
      }
    }
  }

  // Function to get all siblings of an collection
  const getSiblings = (collectionId) => {
    const parentId = childToParent.get(collectionId);
    if (!parentId) return [];
    const parent = collectionsMap.get(parentId);
    return parent.children.map(child => child.id || child.collection_id).filter(id => id !== collectionId);
  };

  // Function to recursively add parents with no other children
  const addLonelyParents = (collectionId, addedParents = new Set()) => {
    const parentId = childToParent.get(collectionId);
    if (!parentId || addedParents.has(parentId)) return;

    const siblings = getSiblings(collectionId);
    if (siblings.length === 0) {
      emptyCollections.add(parentId);
      addedParents.add(parentId);
      addLonelyParents(parentId, addedParents);
    }
  };

  // Process each childless collection to include its lonely parents
  for (const collectionId of emptyCollections) {
    addLonelyParents(collectionId);
  }

  return {
    rootItems,
    emptyCollections: Array.from(emptyCollections) // Convert Set to Array
  }

  return {
    rootItems,
    collectionsWithoutChildren: Array.from(collectionsWithoutChildren)
  };
};

// change types
const pluralize = (word) => {
  const pluralRules = {
    'collection': 'collections',
    'asset': 'assets',
    'resource': 'resources'
  };

  return pluralRules[word] || `${word}s`;
};


const changeItemType = (newItemTypeName) => {

  const itemTypeName = newItemTypeName.toLowerCase() + 's';

  let previewData = dndStore.previewData['assets'];
  const assets = allSelectedItems.value;

  for (const asset of assets) {

    const assetId = asset.id;
    const selectedItem = previewData.find(item => item.id === assetId);


    if (selectedItem) {
      selectedItem.is_resource = itemTypeName !== 'assets';
      dndStore.previewData['assets'] = [...previewData];
      itemType.value = newItemTypeName;
    }
  }

};

const selectAssetType = (assetTypeName) => {

  let newAssetType;
  const assetTypes = assetStore.getAssetTypes;
  newAssetType = assetTypes.find((item) => item.name === assetTypeName);


  let previewData = dndStore.previewData['assets'];
  const assets = allSelectedItems.value;

  for (const asset of assets) {

    const assetId = asset.id;
    const selectedAsset = previewData.find(item => item.id === assetId);

    if (selectedAsset) {
      selectedAsset.asset_type_name = newAssetType.name;
      selectedAsset.asset_type_icon = newAssetType.icon;
      selectedAsset.asset_type_id = newAssetType.id;
      dndStore.previewData['assets'] = [...previewData];
    }

  }

  assetType.value = assetTypeName;

};

const selectCollectionType = (collectionTypeName) => {

  let newCollectionType;
  const collectionTypes = collectionStore.getCollectionTypes;
  newCollectionType = collectionTypes.find((item) => item.name === collectionTypeName);


  let previewData = dndStore.previewData['collections'];
  const collections = allSelectedItems.value;


  for (const collection of collections) {

    const collectionId = collection.id;
    const selectedCollection = previewData.find(item => item.id === collectionId);

    if (selectedCollection) {
      selectedCollection.collection_type_name = newCollectionType.name;
      selectedCollection.collection_type_icon = newCollectionType.icon;
      selectedCollection.collection_type_id = newCollectionType.id;
      dndStore.previewData['collections'] = [...previewData];
    }

  }

  collectionType.value = collectionTypeName;


};

const deselectItems = () => {
  dndStore.previewDataActiveItem = null;
  dndStore.previewDataSelectedItems = {};
  stage.markedItems = [];
};

const removeItems = () => {
  const collections = dndStore.previewDataSelectedItems['collections'] ? dndStore.previewDataSelectedItems['collections'] : [];
  const assets = dndStore.previewDataSelectedItems['assets'] ? dndStore.previewDataSelectedItems['assets'] : [];
  const items = [...collections, ...assets]
  for (const item of items) {
    removeItem(item);
  }
  dndStore.previewDataSelectedItems = {}
};

const removeEmptyFolders = () => {
  const previewData = dndStore.previewData['collections'];
  dndStore.previewData['collections'] = previewData.filter(item => !emptyCollectionIds.value.includes(item.id));

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
  collectionType.value = collectionStore.getCollectionTypesNames[0];
  assetType.value = assetStore.getAssetTypesNames[0];
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