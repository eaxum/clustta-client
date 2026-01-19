<template>
  <div class="conflict-item-main" :class="{ 'conflict-item-resolved': isResolved }" @mouseenter="isHovered = true" @mouseleave="isHovered = false">
    <div class="conflict-item-icon-container">
      <img :src="itemIcon" :alt="conflict.name" class="conflict-item-icon" />
    </div>
    <div class="conflict-item-content">
      <RenameInput v-if="isRenaming" v-model="newName" :originalValue="conflict.name" placeholder="Enter new name" @confirm="handleRenameConfirm" @cancel="handleRenameCancel" />
      <span v-else class="conflict-item-name">{{ displayName }}</span>
    </div>
    <div v-if="!isResolved && !isRenaming" class="conflict-item-actions">
      <ActionButton :icon="getAppIcon('edit')" v-tooltip="'Rename'" :buttonFunction="startRename" />
      <ActionButton :icon="getAppIcon('merge')" v-tooltip="'Merge'" :buttonFunction="handleMerge" />
    </div>
    <ActionButton v-if="isResolved && !isRenaming" :icon="getAppIcon('circle-check')" :isDisabled="true" :useGo="true" :allowDeactivate="true" />
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import RenameInput from '@/instances/desktop/components/RenameInput.vue';

// services
import { AssetService, CollectionService } from '@/services';

// stores
const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const syncConflictStore = useSyncConflictStore();

import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useSyncConflictStore } from '@/stores/syncConflict';

// props
const props = defineProps({
  conflict: {
    type: Object,
    required: true
  },
  hideExtensions: {
    type: Boolean,
    default: false
  },
  showFullPath: {
    type: Boolean,
    default: false
  }
});

// emits
const emit = defineEmits(['resolved', 'merge']);

// refs
const currentName = ref('');
const isHovered = ref(false);
const isRenaming = ref(false);
const isResolved = ref(false);
const itemIcon = ref('');
const newName = ref('');

// computed properties
const displayName = computed(() => {
  const name = currentName.value || props.conflict.name;
  
  if (props.showFullPath && props.conflict.entity_path) {
    if (props.conflict.type === 'task' && props.conflict.extension && !props.hideExtensions) {
      return `${props.conflict.entity_path}${name}${props.conflict.extension}`;
    }
    return `${props.conflict.entity_path}${name}`;
  }
  
  if (props.conflict.type === 'task' && props.conflict.extension && !props.hideExtensions) {
    return `${name}${props.conflict.extension}`;
  }
  return name;
});

// methods
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Emits merge event and removes conflict from store.
const handleMerge = () => {
  syncConflictStore.removeConflict(props.conflict.local_id);
  emit('merge', props.conflict);
};

// Cancels rename mode and resets input.
const handleRenameCancel = () => {
  isRenaming.value = false;
  newName.value = '';
};

// Renames the conflict item via appropriate service.
// Updates local state and removes from conflict store on success.
const handleRenameConfirm = async (confirmedName) => {
  const projectUri = projectStore.activeProject?.uri;
  if (!projectUri) {
    notificationStore.errorNotification('Error', 'No active project');
    return;
  }

  try {
    if (props.conflict.type === 'entity') {
      await CollectionService.RenameCollection(projectUri, props.conflict.local_id, confirmedName);
    } else {
      await AssetService.RenameAsset(projectUri, props.conflict.local_id, confirmedName);
    }

    currentName.value = confirmedName;
    isResolved.value = true;
    isRenaming.value = false;
    
    notificationStore.addNotification(
      'Renamed Successfully',
      `${props.conflict.type === 'entity' ? 'Collection' : 'Asset'} renamed to "${confirmedName}"`,
      'success'
    );

    syncConflictStore.removeConflict(props.conflict.local_id);

    // If entity, also remove child conflicts
    if (props.conflict.type === 'entity' && props.conflict.entity_path) {
      syncConflictStore.removeChildConflicts(props.conflict.entity_path);
    }

    emit('resolved', {
      ...props.conflict,
      name: confirmedName,
      resolved: true
    });

  } catch (error) {
    console.error('Failed to rename:', error);
    notificationStore.errorNotification('Rename Failed', error.message || 'Failed to rename item');
  }
};

// Loads the appropriate icon based on conflict type.
const loadIcon = async () => {
  if (props.conflict.type === 'entity') {
    itemIcon.value = getAppIcon(props.conflict.entity_type_icon || 'folder');
  } else {
    if (props.conflict.extension) {
      const ext = props.conflict.extension.toLowerCase().replace(/^\./, '');
      const iconPath = await iconStore.getIcon(ext);
      itemIcon.value = iconPath || getAppIcon('file');
    } else {
      itemIcon.value = getAppIcon(props.conflict.task_type_icon || 'file');
    }
  }
};

// Enters rename mode with current name pre-filled.
const startRename = () => {
  newName.value = currentName.value || props.conflict.name;
  isRenaming.value = true;
};

// lifecycle hooks
onMounted(() => {
  loadIcon();
  currentName.value = props.conflict.name;
  isResolved.value = props.conflict.resolved || false;
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.conflict-item-main {
  display: flex;
  gap: 0.5rem;
  color: var(--white);
  align-items: center;
  padding: 0.5rem;
  box-sizing: border-box;
  width: 100%;
  border-radius: var(--large-radius);
  overflow: hidden;
  background-color: var(--dark-steel);
  outline: var(--transparent-line);
  outline-offset: -1px;
  transition: all 0.2s ease-out;
}

.conflict-item-main:hover {
  background-color: var(--steel);
  border-radius: var(--small-radius);
  outline: 1px solid var(--light-steel);
}

.conflict-item-resolved {
  opacity: 0.7;
}

.conflict-item-resolved .conflict-item-name {
  text-decoration: line-through;
  text-decoration-color: var(--gray);
}

.conflict-item-icon-container {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  flex-shrink: 0;
}

.conflict-item-icon {
  width: 24px;
  height: 24px;
  object-fit: contain;
}

.conflict-item-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.conflict-item-name {
  font-size: 13px;
  font-weight: 300;
  color: var(--white);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.conflict-item-actions {
  display: flex;
  gap: 0.25rem;
  align-items: center;
  opacity: 0;
  transition: opacity 0.2s ease-out;
}

.conflict-item-main:hover .conflict-item-actions {
  opacity: 1;
}
</style>
