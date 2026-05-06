<template>
  <div class="conflict-item-main" :class="{ 'conflict-item-resolved': isResolved }" @mouseenter="isHovered = true" @mouseleave="isHovered = false">
    <div class="conflict-item-icon-container">
      <img :src="itemIcon" :alt="conflict.name" class="conflict-item-icon small-icons" :class="{ 'no-filter': isCustomIcon }" />
    </div>

    <div class="conflict-item-content">
      <RenameInput v-if="isRenaming" v-model="newName" :originalValue="conflict.name" :placeholder="$t('components.conflictItem.enterNewName')" @confirm="handleRenameConfirm" @cancel="handleRenameCancel" />
      <span v-else class="conflict-item-name">{{ displayName }}</span>
      <!-- <span v-else class="conflict-item-name">{{ conflict.collection_path }}</span> -->
    </div>

    <div v-if="!isResolved && !isRenaming" class="conflict-item-actions">
      <ActionButton :icon="CiEdit" v-tooltip="$t('components.conflictItem.rename')" :buttonFunction="startRename" />
      <ActionButton :icon="CiMerge" v-tooltip="$t('components.conflictItem.merge')" :buttonFunction="handleMerge" />
    </div>

    <ActionButton v-if="isResolved && !isRenaming" :icon="CiCircleCheck" :isDisabled="true" :useGo="true" :allowDeactivate="true" />
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { CiCircleCheck, CiEdit, CiMerge } from '@clustta/icons-vue';
import { resolveIcon } from '@/lib/icon-map';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import RenameInput from '@/instances/desktop/components/RenameInput.vue';

// services
import { AssetService, CollectionService } from '@/services';

// stores
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';

const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();

const { t } = useI18n();

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
const isCustomIcon = ref(false);
const isHovered = ref(false);
const isRenaming = ref(false);
const isResolved = ref(false);
const itemIcon = ref('');
const newName = ref('');

// computed properties
const displayName = computed(() => {
  const name = currentName.value || props.conflict.name;
  
  if (props.showFullPath && props.conflict.collection_path) {
    if (props.conflict.type === 'collection') {
      return `${props.conflict.collection_path}`;
    }
    if (props.conflict.type === 'asset' && props.conflict.extension && !props.hideExtensions) {
      return `${props.conflict.collection_path}${name}${props.conflict.extension}`;
    }
    return `${props.conflict.collection_path}${name}`;
  }
  
  if (props.conflict.type === 'asset' && props.conflict.extension && !props.hideExtensions) {
    return `${name}${props.conflict.extension}`;
  }
  return name;
});

// methods
// Returns icon path from icon store.
const getAppIcon = (iconName) => {
  return iconStore.resolveIcon(iconName);
};

// Emits merge event to parent for handling.
const handleMerge = () => {
  emit('merge', props.conflict);
};

// Cancels rename mode and resets input.
const handleRenameCancel = () => {
  isRenaming.value = false;
  newName.value = '';
};

// Renames the conflict item via appropriate service.
// On success, emits resolved event for parent to handle cleanup.
const handleRenameConfirm = async (confirmedName) => {
  const projectUri = projectStore.activeProject?.uri;
  if (!projectUri) {
    notificationStore.errorNotification('Error', t('components.conflictItem.noActiveProject'));
    return;
  }

  try {
    if (props.conflict.type === 'collection') {
      await CollectionService.RenameCollection(projectUri, props.conflict.local_id, confirmedName);
    } else {
      await AssetService.RenameAsset(projectUri, props.conflict.local_id, confirmedName);
    }

    currentName.value = confirmedName;
    isResolved.value = true;
    isRenaming.value = false;
    
    emit('resolved', {
      ...props.conflict,
      name: confirmedName,
      resolved: true
    });
  } catch (error) {
    console.error('Failed to rename:', error);
    notificationStore.errorNotification(t('components.conflictItem.renameFailed'), error.message || 'Failed to rename item');
  }
};

// Loads the appropriate icon based on conflict type.
const loadIcon = async () => {
  isCustomIcon.value = false;
  
  if (props.conflict.type === 'collection') {
    itemIcon.value = resolveIcon(props.conflict.collection_type_icon || 'folder');
  } else {
    if (props.conflict.extension) {
      const ext = props.conflict.extension.toLowerCase().replace(/^\./, '');
      const iconPath = await iconStore.getIcon(ext);
      if (iconPath) {
        itemIcon.value = iconPath;
        isCustomIcon.value = true;
      } else {
        itemIcon.value = resolveIcon(props.conflict.asset_type_icon || 'file');
      }
    } else {
      itemIcon.value = resolveIcon(props.conflict.asset_type_icon || 'file');
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
