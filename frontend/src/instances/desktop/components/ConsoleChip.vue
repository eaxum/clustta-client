<template>
  <span class="console-chip" :class="chipClasses" :tabindex="isClickable ? 0 : -1" :role="isClickable ? 'button' : null"
    v-tooltip="tooltipText" @click="handleClick" @keydown.enter.prevent="handleClick" @keydown.space.prevent="handleClick">

    <span class="console-chip-leading">
      <ProfilePhoto v-if="type === 'user'" :assigneeId="entityId" :userPhoto="userPhoto" />

      <img v-else-if="useRawIcon" class="console-chip-icon no-filter" :src="iconSrc">

      <img v-else class="console-chip-icon" :src="iconSrc">
    </span>

    <span class="console-chip-label">{{ displayLabel }}</span>
  </span>
</template>

<script setup>
// imports
import { computed, onMounted, ref, watch } from 'vue';
import { Browser } from '@wailsio/runtime';
import emitter from '@/lib/mitt';
import { FSService } from '@/services';

// components
import ProfilePhoto from '@/instances/common/components/ProfilePhoto.vue';

// stores
import { useAgentEntityCacheStore } from '@/stores/agentEntityCache';
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useIconStore } from '@/stores/icons';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';

const cacheStore = useAgentEntityCacheStore();
const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const iconStore = useIconStore();
const projectStore = useProjectStore();
const stage = useStageStore();

// props
const props = defineProps({
  type: { type: String, required: true, validator: (v) => ['asset', 'collection', 'untracked_asset', 'untracked_collection', 'user'].includes(v) },
  entityId: { type: String, required: true },
  fallbackLabel: { type: String, default: '' },
});

// refs
const resolved = ref(cacheStore.get(props.type, props.entityId));
const resolvedAssetIcon = ref('');

// computed properties
const isMissing = computed(() => cacheStore.isMissing(props.type, props.entityId));

const isClickable = computed(() => !isMissing.value);

const displayLabel = computed(() => {
  const r = resolved.value;
  if (props.type === 'user' && r) return r.full_name || r.username || props.fallbackLabel;
  if (r && r.name) return r.name;
  return props.fallbackLabel || '';
});

const userPhoto = computed(() => (props.type === 'user' && resolved.value ? resolved.value.photo || '' : ''));

// Renders the asset's own icon (file type, app icon) without theme tinting.
const useRawIcon = computed(() => ['asset', 'untracked_asset'].includes(props.type) && !!resolvedAssetIcon.value);

const iconSrc = computed(() => {
  if (props.type === 'asset' || props.type === 'untracked_asset') {
    return resolvedAssetIcon.value || iconStore.getAppIcon('file');
  }
  if (props.type === 'collection') return iconStore.getAppIcon('folder');
  if (props.type === 'untracked_collection') return iconStore.getAppIcon('folder');
  return iconStore.getAppIcon('person');
});

const tooltipText = computed(() => {
  if (isMissing.value) return 'No longer available';
  return displayLabel.value;
});

const chipClasses = computed(() => ({
  'console-chip-clickable': isClickable.value,
  'console-chip-dimmed': isMissing.value,
  [`console-chip-${props.type}`]: true,
}));

// methods

// Reads the entity from the cache, fetching it on first sight.
const resolveEntity = async () => {
  const cached = cacheStore.get(props.type, props.entityId);
  if (cached) {
    resolved.value = cached;
    await resolveAssetIcon(cached);
    return;
  }
  const result = await cacheStore.ensure(props.type, props.entityId);
  resolved.value = result || cacheStore.get(props.type, props.entityId);
  await resolveAssetIcon(resolved.value);
};

// Mirrors Browser.softRefresh: resolve the actual file icon through the icon
// store using the normalized extension for tracked and untracked assets.
const resolveAssetIcon = async (entity) => {
  if (!['asset', 'untracked_asset'].includes(props.type)) {
    resolvedAssetIcon.value = '';
    return;
  }
  if (entity?.icon) {
    resolvedAssetIcon.value = entity.icon;
    return;
  }
  const extension = String(entity?.extension || '').toLowerCase().replace(/^\./, '');
  resolvedAssetIcon.value = extension ? ((await iconStore.getIcon(extension)) || '') : '';
};

// Opens the asset's parent collection in the browser and selects the asset.
const revealAsset = async (asset) => {
  commonStore.activeWorkspace = 'Default';
  commonStore.viewSearchQuery = '';
  commonStore.resetFilters();
  commonStore.navigatorMode = true;

  if (asset.collection_id) {
    const parent = await cacheStore.ensure('collection', asset.collection_id);
    if (parent) collectionStore.navigateToCollection(parent);
  } else {
    collectionStore.navigateToCollection(null);
  }

  stage.deselectAllItems();
  assetStore.selectAsset(asset);
  stage.firstSelectedItemId = asset.id;
  stage.markedItems = [asset.id];
  emitter.emit('refresh-browser');
};

// Opens the collection in the browser and marks it selected.
const revealCollection = (collection) => {
  commonStore.activeWorkspace = 'Default';
  commonStore.viewSearchQuery = '';
  commonStore.resetFilters();
  commonStore.navigatorMode = true;
  collectionStore.navigateToCollection(collection);
  collectionStore.selectedCollection = collection;
  emitter.emit('refresh-browser');
};

// Opens the user's public Clustta profile page in the system browser.
const openUserProfile = (user) => {
  if (!user.username) return;
  const profileUrl = `https://app.clustta.com/user/${user.username}`;
  Browser.OpenURL(profileUrl);
};

// Resolves the entity if needed, then dispatches to the right reveal action.
const handleClick = async () => {
  if (!isClickable.value) return;
  if (!projectStore.activeProject && props.type !== 'user') return;

  let entity = resolved.value;
  if (!entity) {
    entity = await cacheStore.ensure(props.type, props.entityId);
  }
  if (!entity) return;

  if (props.type === 'asset') return revealAsset(entity);
  if (props.type === 'collection') return revealCollection(entity);
  if (props.type === 'untracked_asset' || props.type === 'untracked_collection') {
    if (entity.path) await FSService.RevealInExplorer(entity.path);
    return;
  }
  if (props.type === 'user') return openUserProfile(entity);
};

// watchers
watch(() => `${props.type}:${props.entityId}`, () => {
  resolved.value = cacheStore.get(props.type, props.entityId);
  resolvedAssetIcon.value = '';
  resolveEntity();
});

// lifecycle hooks
onMounted(() => {
  resolveEntity();
});
</script>

<style scoped>
.console-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.15rem 0.6rem 0.15rem 0.35rem;
  margin: 2px 0;
  background-color: var(--surface-3);
  border-radius: var(--large-radius);
  font-size: 0.875rem;
  font-weight: 400;
  color: var(--text);
  max-width: 220px;
  vertical-align: middle;
  line-height: 1.2;
  user-select: none;
  outline: none;
  transition: background-color 0.15s ease;
}

.console-chip-clickable {
  cursor: pointer;
}

.console-chip-clickable:hover,
.console-chip-clickable:focus-visible {
  background-color: var(--surface-4);
  outline: var(--transparent-line);
}

.console-chip-dimmed {
  opacity: 0.5;
  text-decoration: line-through;
  cursor: default;
}

.console-chip-leading {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.console-chip-icon {
  width: 16px;
  height: 16px;
  object-fit: contain;
  display: block;
}

.console-chip-label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}
</style>
