<template>
  <div ref="manageTagsMenuRoot" class="filter-menu-container" v-stop-propagation>

    <div class="input-section">
      <div class="horizontal-flex search-row">
        <input ref="searchTagInput" v-model="searchTerm" class="input-short" type="text" :maxlength="TAG_MAX_LENGTH"
          :placeholder="$t('menus.manageTags.searchPlaceholder')" @keydown.enter="addNewTag" />
        <ActionButton v-if="searchTerm.length" :icon="iconStore.getAppIcon('close')" :allowDeactivate="true"
          v-tooltip="$t('components.searchBar.clearSearch')" :buttonFunction="clearSearch" />
      </div>
    </div>

    <span v-if="canAddNewTag" class="filter-menu-item add-tag-row" @click="addNewTag">
      <img class="small-icons" :src="iconStore.getAppIcon('plus-circle')">
      <div class="horizontal-flex">
        <div class="tag-label">{{ $t('menus.manageTags.addTag', { name: searchTerm.trim() }) }}</div>
      </div>
    </span>

    <span v-for="tag in filteredTags" :key="tag.id" class="filter-menu-item" @click="toggleTag(tag)">
      <img class="small-icons" src="/icons/tags.svg">
      <div class="horizontal-flex">
        <div class="tag-label"> {{ utils.capitalizeStr(tag.name) }} </div>
        <CheckBox :modelValue="isTagOnAll(tag)" :indeterminate="isTagOnSome(tag)" :disabled="isProcessing || !canUpdateAssets"
          :ariaLabel="`Assign ${tag.name}`" @click.stop @change="toggleTag(tag)" />
      </div>
    </span>

    <span v-if="!tagStore.tags.length && !searchTerm.trim()" class="filter-menu-item disabled">
      <div>{{ $t('menus.manageTags.empty') }}</div>
    </span>

    <span v-else-if="!filteredTags.length && !canAddNewTag" class="filter-menu-item disabled">
      <div>{{ $t('menus.noResults') }}</div>
    </span>

  </div>

</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';
import { canActOnAssets } from '@/lib/permissions';
import utils from '@/services/utils';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import CheckBox from '@/instances/common/components/CheckBox.vue';

// stores
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useStageStore } from '@/stores/stages';
import { useTagStore } from '@/stores/tags';

const iconStore = useIconStore();
const menu = useMenu();
const notificationStore = useNotificationStore();
const stage = useStageStore();
const tagStore = useTagStore();
const { t } = useI18n();

// refs
const isProcessing = ref(false);
const manageTagsMenuRoot = ref(null);
const searchTagInput = ref(null);
const searchTerm = ref('');

const selectedAssets = computed(() => stage.selectedItems.filter(item => item?.type === 'asset'));
const canUpdateAssets = computed(() => canActOnAssets('update_asset', selectedAssets.value));

// constants
const TAG_MAX_LENGTH = 32;

// computed properties
// Returns true when the search term is non-empty and doesn't match an existing tag.
const canAddNewTag = computed(() => {
  const term = searchTerm.value.trim().toLowerCase();
  if (!term) return false;
  return !tagStore.tags.some((tag) => tag.name.toLowerCase() === term);
});

// Returns tags filtered by the current search term.
const filteredTags = computed(() => {
  const term = searchTerm.value.trim().toLowerCase();
  if (!term) return tagStore.tags;
  return tagStore.tags.filter((tag) => tag.name.toLowerCase().includes(term));
});

// methods
// Clears the search input and re-focuses it.
const clearSearch = () => {
  searchTerm.value = '';
  searchTagInput.value?.focus();
};

// Creates a new tag from the search term and assigns it to all selected assets.
const addNewTag = async () => {
  if (!canUpdateAssets.value) return;
  const name = searchTerm.value.trim();
  if (!name || !canAddNewTag.value || isProcessing.value) return;
  const assets = getSelectedAssets();
  if (!assets.length) return;

  isProcessing.value = true;
  try {
    const targetIds = assets.map((a) => a.id);
    await tagStore.addTagToMultipleAssets(targetIds, name);
    for (const asset of assets) {
      const existing = Array.isArray(asset.tags) ? asset.tags : [];
      if (existing.includes(name)) continue;
      const updatedTags = [...existing, name];
      asset.tags = updatedTags;
      emitAssetUpdates(asset.id, [{ property: 'tags', value: updatedTags }]);
    }
    searchTerm.value = '';
    searchTagInput.value?.focus();
  } catch (error) {
    console.error(error);
    notificationStore.errorNotification(t('menus.manageTags.error'), error);
  } finally {
    isProcessing.value = false;
  }
};

// Emits update events so the browser/grid views refresh tags for an asset.
const emitAssetUpdates = (assetId, updates) => {
  const updateData = { itemId: assetId, updates };
  emitter.emit('update-root-data', updateData);
};

// Returns the list of currently selected asset objects from the stage.
const getSelectedAssets = () => {
  return stage.selectedItems.filter((item) => item && item.type === 'asset');
};

// Returns true if every selected asset already has the given tag.
const isTagOnAll = (tag) => {
  const assets = getSelectedAssets();
  if (!assets.length) return false;
  return assets.every((asset) => Array.isArray(asset.tags) && asset.tags.includes(tag.name));
};

// Returns true when a tag is assigned to some, but not all, selected assets.
const isTagOnSome = (tag) => {
  const assets = getSelectedAssets();
  if (!assets.length || isTagOnAll(tag)) return false;
  return assets.some((asset) => Array.isArray(asset.tags) && asset.tags.includes(tag.name));
};

// Toggles a tag on or off for all selected assets in a single batch call.
const toggleTag = async (tag) => {
  if (!canUpdateAssets.value) return;
  if (isProcessing.value) return;
  const assets = getSelectedAssets();
  if (!assets.length) return;

  isProcessing.value = true;
  try {
    if (isTagOnAll(tag)) {
      const targets = assets.filter((asset) => Array.isArray(asset.tags) && asset.tags.includes(tag.name));
      const targetIds = targets.map((a) => a.id);
      await tagStore.removeTagFromMultipleAssets(targetIds, tag.id);
      for (const asset of targets) {
        const updatedTags = (asset.tags || []).filter((name) => name !== tag.name);
        asset.tags = updatedTags;
        emitAssetUpdates(asset.id, [{ property: 'tags', value: updatedTags }]);
      }
    } else {
      const targets = assets.filter((asset) => !Array.isArray(asset.tags) || !asset.tags.includes(tag.name));
      const targetIds = targets.map((a) => a.id);
      await tagStore.addTagToMultipleAssets(targetIds, tag.name);
      for (const asset of targets) {
        const updatedTags = [...(asset.tags || []), tag.name];
        asset.tags = updatedTags;
        emitAssetUpdates(asset.id, [{ property: 'tags', value: updatedTags }]);
      }
    }
  } catch (error) {
    console.error(error);
    notificationStore.errorNotification(t('menus.manageTags.error'), error);
  } finally {
    isProcessing.value = false;
  }
};

// lifecycle hooks
onMounted(async () => {
  if (!tagStore.tags.length) await tagStore.reloadTags();
  menu.assetMenuWidth = manageTagsMenuRoot.value.getBoundingClientRect().width;
  menu.collectionMenu = manageTagsMenuRoot.value;
  searchTagInput.value?.focus();
});

onBeforeUnmount(() => {
  if (!manageTagsMenuRoot.value) return;
  menu.assetMenuWidth = manageTagsMenuRoot.value.getBoundingClientRect().width;
  menu.assetMenuHeight = manageTagsMenuRoot.value.getBoundingClientRect().height;
});
</script>

<style scoped>
@import "@/assets/desktop.css";
@import "@/assets/menu.css";

.filter-menu-item.disabled {
  opacity: 0.6;
  cursor: default;
}

.input-section{
  margin-bottom: .3rem;
}
.add-tag-row {
  font-style: italic;
}

.tag-label {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-align: left;
}

.search-row {
  align-items: center;
  gap: .25rem;
}
</style>
