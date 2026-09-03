<template>
  <div ref="menuRoot" class="filter-menu-container" v-stop-propagation>
    <div class="input-section">
      <div class="horizontal-flex search-row">
        <input ref="searchInput" v-model="searchTerm" class="input-short" type="text" :maxlength="TAG_MAX_LENGTH"
          placeholder="Search or add tags" @keydown.enter="addNewTag" />
        <ActionButton v-if="searchTerm.length" :icon="iconStore.getAppIcon('close')" :allowDeactivate="true"
          v-tooltip="$t('components.searchBar.clearSearch')" :buttonFunction="clearSearch" />
      </div>
    </div>

    <span v-if="canAddNewTag" class="filter-menu-item add-tag-row" @click="addNewTag">
      <img class="small-icons" :src="iconStore.getAppIcon('plus-circle')">
      <div class="tag-label">Add "{{ searchTerm.trim() }}"</div>
    </span>

    <span v-for="tag in filteredTags" :key="tag.id" class="filter-menu-item" @click="toggleTag(tag)">
      <img class="small-icons" :src="iconStore.getAppIcon('tag')">
      <div class="horizontal-flex">
        <div class="tag-label">{{ utils.capitalizeStr(tag.name) }}</div>
        <CheckBox :modelValue="isAssigned(tag)" :disabled="isProcessing"
          :ariaLabel="`${isAssigned(tag) ? 'Remove' : 'Assign'} ${tag.name}`"
          @click.stop @change="toggleTag(tag)" />
      </div>
    </span>

    <span v-if="!tagStore.tags.length && !searchTerm.trim()" class="filter-menu-item disabled">
      <div>No tags</div>
    </span>
    <span v-else-if="!filteredTags.length && !canAddNewTag" class="filter-menu-item disabled">
      <div>{{ $t('menus.noResults') }}</div>
    </span>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import emitter from '@/lib/mitt';
import { canActOnAsset } from '@/lib/permissions';
import utils from '@/services/utils';
import { CheckpointService } from '@/services';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import CheckBox from '@/instances/common/components/CheckBox.vue';
import { useAssetStore } from '@/stores/assets';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useTagStore } from '@/stores/tags';

const TAG_MAX_LENGTH = 32;

const { t } = useI18n();
const assetStore = useAssetStore();
const iconStore = useIconStore();
const menu = useMenu();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const tagStore = useTagStore();

const isProcessing = ref(false);
const menuRoot = ref(null);
const searchInput = ref(null);
const searchTerm = ref('');

const checkpoint = computed(() => menu.checkpointTagMenuData.checkpoint);
const canManageTags = computed(() => canActOnAsset('manage_dependencies', assetStore.selectedAsset));
const assignmentsByTagId = computed(() => new Map(
  (checkpoint.value?.tags || []).map(assignment => [assignment.tag_id, assignment]),
));
const normalizedSearchTerm = computed(() => searchTerm.value.trim().toLowerCase());
const canAddNewTag = computed(() => (
  normalizedSearchTerm.value
  && !tagStore.tags.some(tag => tag.name.toLowerCase() === normalizedSearchTerm.value)
));
const filteredTags = computed(() => {
  if (!normalizedSearchTerm.value) return tagStore.tags;
  return tagStore.tags.filter(tag => tag.name.toLowerCase().includes(normalizedSearchTerm.value));
});

const clearSearch = () => {
  searchTerm.value = '';
  searchInput.value?.focus();
};

const isAssigned = (tag) => assignmentsByTagId.value.has(tag.id);

const refreshCheckpointData = () => {
  emitter.emit('update-checkpoints');
  emitter.emit('refresh-browser');
};

const updateSelectedAssetTags = (tagName, assigned) => {
  const asset = assetStore.selectedAsset;
  if (!asset) return;
  const tags = Array.isArray(asset.tags) ? asset.tags : [];
  asset.tags = assigned
    ? [...new Set([...tags, tagName])]
    : tags.filter(name => name !== tagName);
  emitter.emit('update-root-data', {
    itemId: asset.id,
    updates: [{ property: 'tags', value: asset.tags }],
  });
};

const assignTag = async (tag) => {
  const assignment = await CheckpointService.SetCheckpointTag(
    projectStore.activeProject.uri,
    tag.id,
    tag.name,
    checkpoint.value.checkpoint_id,
  );
  checkpoint.value.tags = [
    ...(checkpoint.value.tags || []).filter(item => item.tag_id !== tag.id),
    assignment,
  ];
  updateSelectedAssetTags(tag.name, true);
};

const addNewTag = async () => {
  if (!canManageTags.value || !canAddNewTag.value || isProcessing.value || !checkpoint.value) return;
  isProcessing.value = true;
  try {
    const assignment = await CheckpointService.SetCheckpointTag(
      projectStore.activeProject.uri,
      '',
      searchTerm.value.trim(),
      checkpoint.value.checkpoint_id,
    );
    checkpoint.value.tags = [...(checkpoint.value.tags || []), assignment];
    updateSelectedAssetTags(assignment.name, true);
    await tagStore.reloadTags();
    clearSearch();
    refreshCheckpointData();
  } catch (error) {
    notificationStore.errorNotification('Unable to add checkpoint tag', error);
  } finally {
    isProcessing.value = false;
  }
};

const toggleTag = async (tag) => {
  if (!canManageTags.value || isProcessing.value || !checkpoint.value) return;
  isProcessing.value = true;
  try {
    const assignment = assignmentsByTagId.value.get(tag.id);
    if (assignment) {
      await CheckpointService.DeleteCheckpointTag(projectStore.activeProject.uri, assignment.id);
      checkpoint.value.tags = (checkpoint.value.tags || []).filter(item => item.id !== assignment.id);
      updateSelectedAssetTags(tag.name, false);
    } else {
      await assignTag(tag);
    }
    refreshCheckpointData();
  } catch (error) {
    notificationStore.errorNotification(t('menus.manageTags.error'), error);
  } finally {
    isProcessing.value = false;
  }
};

onMounted(async () => {
  if (!tagStore.tags.length) await tagStore.reloadTags();
  menu.assetMenuWidth = menuRoot.value.getBoundingClientRect().width;
  menu.collectionMenu = menuRoot.value;
  searchInput.value?.focus();
});

onBeforeUnmount(() => {
  if (!menuRoot.value) return;
  menu.assetMenuWidth = menuRoot.value.getBoundingClientRect().width;
  menu.assetMenuHeight = menuRoot.value.getBoundingClientRect().height;
});
</script>

<style scoped>
@import "@/assets/desktop.css";
@import "@/assets/menu.css";

.filter-menu-item.disabled {
  opacity: .6;
  cursor: default;
}

.input-section {
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
