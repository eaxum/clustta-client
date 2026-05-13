<template>
  <div :class="['hierarchy-item', { 'is-directory': item.type === 'collection', 'hierarchy-item-root': isHierarchyRoot }]">
    <div class="item-header" :class="{ 'item-header-selected': isItemSelected }">

      <span class="hierarchy-item-spacer single-action-button" @click="toggleExpand"
        :class="{ 'no-expand': item.type !== 'collection' || !item.children.length }">
        <img v-if="item.type === 'collection' && item.children.length" class="large-icons hierarchy-collection-collapsed"
          :class="{ 'hierarchy-collection-expanded': isExpanded }" :src="getAppIcon('chevron-right')">
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

        <DropDownBox v-if="item.collection_type_id" :items="collectionTypeOptions" :selectedItem="collectionType"
          :onSelect="selectCollectionType" :fullWidth="false" />

        <DropDownBox v-else :items="itemTypes" :selectedItem="itemType" :onSelect="changeItemType" :fullWidth="false" />

        <div v-if="!item.collection_type_id" class="hierarchy-item-type-options">
          <DropDownBox :items="assetTypeOptions" :selectedItem="assetType" :onSelect="selectAssetType" :fullWidth="false" />
        </div>

        <ActionButton :icon="getAppIcon('trash')" v-tooltip="$t('components.hierarchyItem.remove')" @click="removeItem(item)" />
      </div>


    </div>

    <div v-if="item.type === 'collection' && isExpanded" class="item-children">
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
import { useI18n } from 'vue-i18n';
import utils from '@/services/utils';

const { t } = useI18n();

// state imports
import { useMenu } from '@/stores/menu';
import { useDndStore } from '@/stores/dnd';
import { useAssetStore } from '@/stores/assets';
import { useStageStore } from '@/stores/stages';
import { useCollectionStore } from '@/stores/collections';
import { useTemplateStore } from '@/stores/template';

// stores
const menu = useMenu();
const stage = useStageStore();
const dndStore = useDndStore();
const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
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
  return !props.item.is_resource ? t('components.hierarchyItem.asset') : t('components.hierarchyItem.resource');
});


// computed
const isItemSelected = computed(() => {
  return stage.markedItems?.includes(props.item.id)
  return dndStore.selectedPreviewItemIds?.includes(props.item.id);
});

const isHierarchyRoot = computed(() => {
  return props.item.root
});

const collectionType = computed(() => {
  return props.item.collection_type_id ? props.item.collection_type_name : '';
});

const itemName = computed(() => {
  return props.item.type === 'collection' ? props.item.name : getItemName(props.item.name);
});

const assetType = computed(() => {
  return props.item.asset_type_name;
});

const resourceType = computed(() => {
  return props.item.resource_type_name;
});


const assetTypeNames = computed(() => {
  return assetStore.getAssetTypesNames.filter((item) => item !== assetType.value);
});

const assetTypeOptions = computed(() => {
  return assetStore.getAssetTypes
    .filter((type) => type.name !== assetType.value)
    .map((type) => ({
      id: type.id,
      name: type.name,
      icon: type.icon ? getAppIcon(type.icon) : null,
    }));
});

const collectionTypeOptions = computed(() => {
  return collectionStore.getCollectionTypes.map((type) => ({
    id: type.id,
    name: type.name,
    icon: type.icon ? getAppIcon(type.icon) : null,
  }));
});

const itemTypes = computed(() => {

  const allItemTypes = ['asset', 'resource'];
  return allItemTypes.filter((item) => item !== itemType.value?.toLowerCase());

});

const itemIcon = computed(() => {

  const item = props.item;
  const isProjectRoot = !item.collection_type_id && item.root;

  if (item.root) {
    if (isProjectRoot) {
      return 'home';
    } else {
      return item.collection_type_icon
    }
  } else if (item.collection_type_id) {
    return item.collection_type_icon
  } else if (item.asset_type_id) {
    return item.asset_type_icon;
  } else {
    return item.resource_type_icon;
  }

});

// methods
const handleClick = (event, data) => {

  const allData = dndStore.previewData;
  const parentId = data.parent_id || data.collection_id;

  const orphanCollections = allData['collections'].filter((item) => !item.parent_id);
  const orphanAssets = allData['assets'].filter((item) => !item.collection_id);
  const orphanItems = [...orphanCollections, ...orphanAssets];

  const siblingCollections = allData['collections'].filter((item) => item.parent_id === parentId);
  const siblingAssets = allData['assets'].filter((item) => item.collection_id === parentId);
  const siblingItems = [...siblingCollections, ...siblingAssets];

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
    'collection': 'collections',
    'asset': 'assets',
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
  if (props.item.type === 'collection' && props.item.children.length) {
    isExpanded.value = !isExpanded.value
  }
};

const changeItemType = (newItemTypeName) => {
  console.log(newItemTypeName)

  const itemTypeName = itemType.value.toLowerCase() + 's';

  let previewData = dndStore.previewData['assets'];
  const itemId = props.item.id;
  const selectedItem = previewData.find(item => item.id === itemId);


  if (selectedItem) {
    console.log(selectedItem)
    selectedItem.is_resource = itemTypeName === 'assets';
    dndStore.previewData['assets'] = [...previewData];
    // itemType.value = newItemTypeName;
  }

};

const selectAssetType = (assetTypeName) => {

  let newAssetType;
  const assetTypes = assetStore.getAssetTypes;
  newAssetType = assetTypes.find((item) => item.name === assetTypeName);

  let previewData = dndStore.previewData['assets'];
  const itemId = props.item.id;
  const selectedItem = previewData.find(item => item.id === itemId);


  if (selectedItem) {
    selectedItem.asset_type_name = newAssetType.name;
    selectedItem.asset_type_icon = newAssetType.icon;
    selectedItem.asset_type_id = newAssetType.id;
    dndStore.previewData['assets'] = [...previewData];
  }

};

const selectCollectionType = (collectionTypeName) => {

  let newCollectionType;
  const collectionTypes = collectionStore.getCollectionTypes;
  newCollectionType = collectionTypes.find((item) => item.name === collectionTypeName);

  let previewData = dndStore.previewData['collections'];
  const itemId = props.item.id;
  const selectedItem = previewData.find(item => item.id === itemId);


  if (selectedItem) {
    selectedItem.collection_type_name = newCollectionType.name;
    selectedItem.collection_type_icon = newCollectionType.icon;
    selectedItem.collection_type_id = newCollectionType.id;
    dndStore.previewData['collections'] = [...previewData];
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

.hierarchy-collection-collapsed {
  transform: rotate(0deg);
}

.hierarchy-collection-expanded {
  transform: rotate(90deg);
}
</style>