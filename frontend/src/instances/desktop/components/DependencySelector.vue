<template>
  <div class="dependency-selector" @click.stop>
    <button class="selector-badge" :class="[`selector-${edge.resolution_mode}`, { 'selector-broken': isBroken }]"
      :disabled="!editable || !triggerOnBadge" type="button" @click="openEditor">
      {{ selectorLabel }}
    </button>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { AssetService } from '@/services';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';

const props = defineProps({
  edge: { type: Object, required: true },
  ownerAssetId: { type: String, required: true },
  editable: { type: Boolean, default: false },
  triggerOnBadge: { type: Boolean, default: true },
});

const emit = defineEmits(['updated']);
const { t } = useI18n();
const iconStore = useIconStore();
const menu = useMenu();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const options = ref({ checkpoints: [], tags: [] });
const optionsLoaded = ref(false);

const menuKey = computed(() => `dependency-selector-${props.edge.id}`);
const isBroken = computed(() => props.edge.resolution_status && props.edge.resolution_status !== 'ready');

const selectorLabel = computed(() => {
  if (isBroken.value) return 'Fix selector';
  if (props.edge.resolution_mode === 'pinned') {
    return `Pinned ${props.edge.resolved_checkpoint_label || 'checkpoint'}`;
  }
  if (props.edge.resolution_mode === 'tagged') {
    const version = props.edge.resolved_checkpoint_label ? ` -> ${props.edge.resolved_checkpoint_label}` : '';
    return `${props.edge.tag_name || 'Tag'}${version}`;
  }
  return 'Latest';
});

const selectedOptionId = computed(() => {
  if (props.edge.resolution_mode === 'pinned') return `pinned-${props.edge.checkpoint_id}`;
  if (props.edge.resolution_mode === 'tagged') return `tagged-${props.edge.checkpoint_tag_id}`;
  return 'floating';
});

const formatPinnedDate = (checkpoint) => {
  if (!checkpoint?.created_at) return 'Pinned checkpoint';
  const date = new Date(checkpoint.created_at);
  if (Number.isNaN(date.getTime())) return 'Pinned checkpoint';
  return `Pinned at ${new Intl.DateTimeFormat('en-GB', {
    day: '2-digit',
    month: '2-digit',
    year: '2-digit',
  }).format(date)}`;
};

const currentPinnedCheckpoint = computed(() => (
  options.value.checkpoints.find(checkpoint => checkpoint.id === props.edge.checkpoint_id)
));

const compactMenuOptions = computed(() => [
  {
    id: 'floating',
    label: 'Latest',
    icon: iconStore.getAppIcon('clock'),
    mode: 'floating',
    selectorId: '',
  },
  ...(props.edge.resolution_mode === 'pinned' && props.edge.checkpoint_id ? [{
    id: `pinned-${props.edge.checkpoint_id}`,
    label: formatPinnedDate(currentPinnedCheckpoint.value),
    icon: iconStore.getAppIcon('pin'),
    mode: 'pinned',
    selectorId: props.edge.checkpoint_id,
  }] : []),
  ...options.value.tags.map(tag => ({
    id: `tagged-${tag.id}`,
    label: tag.name,
    icon: iconStore.getAppIcon('tag'),
    mode: 'tagged',
    selectorId: tag.id,
  })),
]);

const updateSelector = async (option) => {
  try {
    const updatedEdge = await AssetService.UpdateAssetDependencySelector(
      projectStore.activeProject.uri,
      props.ownerAssetId,
      props.edge.id,
      option.mode,
      option.mode === 'pinned' ? option.selectorId : '',
      option.mode === 'tagged' ? option.selectorId : '',
    );
    emit('updated', updatedEdge);
    return true;
  } catch (error) {
    notificationStore.errorNotification('Unable to update dependency version', error);
    return false;
  }
};

const openEditor = async (event) => {
  if (!props.editable) return;
  await menu.showCompactEditMenu(event, {
    key: menuKey.value,
    title: 'Dependency version',
    loading: !optionsLoaded.value,
    options: optionsLoaded.value ? compactMenuOptions.value : [],
    selectedId: selectedOptionId.value,
    onSelect: updateSelector,
  });
  if (optionsLoaded.value) return;

  try {
    options.value = await AssetService.GetDependencySelectorOptions(
      projectStore.activeProject.uri,
      props.edge.dependency_id,
    );
    optionsLoaded.value = true;
    menu.updateCompactEditMenu(menuKey.value, {
      loading: false,
      options: compactMenuOptions.value,
    });
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorLoadingProjectData'), error);
    menu.hideContextMenu();
  }
};

defineExpose({ openEditor });
</script>

<style scoped>
.dependency-selector {
  position: relative;
}

.selector-badge {
  max-width: 150px;
  padding: .2rem .45rem;
  overflow: hidden;
  border: 1px solid var(--border-color);
  border-radius: .35rem;
  background: var(--surface-2);
  color: var(--text);
  cursor: pointer;
  font-size: .68rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.selector-badge:disabled {
  cursor: default;
}

.selector-pinned {
  border-color: var(--warning);
}

.selector-tagged {
  border-color: var(--selected);
}

.selector-broken {
  border-color: var(--danger);
  color: var(--danger);
}
</style>
