<template>
  <div class="trash-item-container">
    <div class="trash-item">
      <div class="trash-item-meta">
        <img v-if="resolvedIcon" class="trash-item-icon small-icons no-filter" :src="resolvedIcon" />
        <component v-else :is="trashItemIcon(trashItem.type)" :size="20" class="trash-item-icon" />
        <div class="trash-item-label">
          <div class="trash-item-label-text" @mouseenter="trayStates.handleHover($event)" @mouseleave="trayStates.resetScroll($event)"><span>{{ trashItem.name.replace(/_/g, " ") }}</span></div>
          <div v-if="trashItem.type === 'asset'" class="trash-item-collection">{{ trashItem.collection_name }}</div>
        </div>
      </div>

      <ActionButton v-if="!trashItem.checkpoints.length" :icon="resolveIconComponent(restoringIds.has(trashItem.id) ? 'loading' : 'undo')" :reverseLoading="restoringIds.has(trashItem.id)" v-tooltip="$t('components.trashItem.restore')" :buttonFunction="() => restoreItem(trashItem.id, trashItem.type)" />
      <ActionButton v-else :icon="CiChevronRight" :class="{ 'chevron-expanded': isExpanded === trashItemIndex }" :buttonFunction="() => toggleVersions(trashItemIndex)" />
    </div>

    <transition name="expand" appear>
      <div v-if="trashItem.checkpoints.length" v-show="isExpanded === trashItemIndex" class="trash-checkpoints">
        <template v-for="(checkpoint, index) in trashItem.checkpoints" :key="index">
          <div v-if="index > 0" class="trash-child-divider"></div>
          <div class="trash-child">
            <CiCheckpointStone :size="20" class="trash-child-icon" />
            <div class="trash-child-label">{{ formatName(checkpoint.name) }}</div>
            <ActionButton :icon="resolveIconComponent(restoringIds.has(checkpoint.id) ? 'loading' : 'undo')" :reverseLoading="restoringIds.has(checkpoint.id)" v-tooltip="$t('components.trashItem.restore')" :buttonFunction="() => restoreItem(checkpoint.id, checkpoint.type)" />
          </div>
        </template>
      </div>
    </transition>
  </div>
</template>

<script setup>
// imports
import { nextTick, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import utils from '@/services/utils';
import { CiCheckpointStone, CiChevronRight, CiFolder, CiFourSquares, CiGeneric } from '@clustta/icons-vue';
import { resolveIcon as resolveIconComponent } from '@/lib/icon-map';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// services
import { TrashService } from '@/services';

// stores
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';

const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const trayStates = useTrayStates();
const { t, locale } = useI18n();

// props
const props = defineProps({
  trashItem: { type: Object, required: true },
  trashItemIndex: { type: Number, required: true },
});

// refs
const isExpanded = ref(-1);
const resolvedIcon = ref('');
const restoringIds = ref(new Set());

// methods
// Formats a checkpoint name into a readable date string.
const formatName = (name) => {
  return utils.formatDate(name.slice(-20), locale.value);
};

// Resolves the file icon for assets with an extension.
const resolveIcon = async () => {
  if (props.trashItem.extension) {
    const ext = props.trashItem.extension.startsWith('.') ? props.trashItem.extension.substring(1) : props.trashItem.extension;
    resolvedIcon.value = await iconStore.getIcon(ext) || '/file-icons/default.svg';
  }
};

// Restores a trashed item and refreshes the trash list.
const restoreItem = async (id, type) => {
  restoringIds.value.add(id);
  try {
    await TrashService.Restore(projectStore.activeProject.uri, id, type);
    trayStates.trashables = await TrashService.GetTrashs(projectStore.activeProject.uri);
    notificationStore.addNotification(t('components.trashItem.itemRestored'), '', 'success', false);
    trayStates.refreshData();
  } catch (error) {
    notificationStore.addNotification(t('components.trashItem.errorRestoringItem'), error.message, 'error', false);
  } finally {
    restoringIds.value.delete(id);
  }
};

// Toggles the expanded state for checkpoints and scrolls into view.
const toggleVersions = (index) => {
  isExpanded.value = isExpanded.value !== index ? index : -1;
  nextTick(() => {
    const element = document.querySelectorAll('.trash-item-container')[index];
    if (element) element.scrollIntoView({ behavior: 'smooth', block: 'start' });
  });
};

// Returns the icon component for a trash item based on its type.
const trashItemIcon = (type) => {
  if (type === 'collection') return CiFolder;
  if (type === 'asset') return CiGeneric;
  if (type === 'template') return CiFourSquares;
  return CiGeneric;
};

// lifecycle hooks
onMounted(resolveIcon);
</script>

<style scoped>
@import "@/assets/desktop.css";

.chevron-expanded {
  transform: rotate(90deg);
}

.expand-enter-active,
.expand-leave-active {
  transition: all .2s ease-in-out;
  max-height: 300px;
}

.expand-enter-from,
.expand-leave-to {
  max-height: 0;
  opacity: 0;
}

.trash-child {
  display: flex;
  align-items: center;
  gap: .5rem;
  padding: .3rem .5rem .3rem 1.5rem;
  min-height: 28px;
}

.trash-child-divider {
  height: 1px;
  margin: 0 .75rem 0 1.5rem;
  background-color: rgba(255, 255, 255, 0.06);
}

.trash-child-icon {
  width: 16px;
  height: 16px;
  min-width: 16px;
  object-fit: contain;
  opacity: .5;
}

.trash-child-label {
  font-size: 12px;
  font-weight: 300;
  color: var(--white);
  opacity: .7;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.trash-checkpoints {
  width: 100%;
  background-color: rgba(255, 255, 255, 0.03);
  border-bottom-left-radius: var(--large-radius);
  border-bottom-right-radius: var(--large-radius);
  overflow: hidden;
}

.trash-item {
  position: relative;
  cursor: auto;
  box-sizing: border-box;
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 0 .5rem;
  overflow: hidden;
}

.trash-item-collection {
  color: var(--white);
  opacity: .5;
  background-color: rgba(0, 0, 0, 0.2);
  padding: .15rem .4rem;
  border-radius: var(--small-radius);
  font-size: 11px;
  white-space: nowrap;
  flex-shrink: 0;
}

.trash-item-container {
  position: relative;
  cursor: auto;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  color: var(--white);
  border-radius: var(--large-radius);
  overflow: hidden;
  min-height: max-content;
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--dark-steel);
  transition: all .2s ease-in-out;
}

.trash-item-container:hover {
  border-radius: var(--normal-radius);
  background-color: var(--steel);
}

.trash-item-icon {
  width: 20px;
  height: 20px;
  min-width: 20px;
  object-fit: contain;
}

.trash-item-label {
  overflow: hidden;
  width: 100%;
  display: flex;
  align-items: center;
  white-space: nowrap;
  gap: .5rem;
}

.trash-item-label-text {
  font-size: 13px;
  font-weight: 300;
  color: var(--white);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.trash-item-meta {
  padding-left: .2rem;
  box-sizing: border-box;
  overflow: hidden;
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: .5rem;
  width: 100%;
  min-height: 40px;
}
</style>

