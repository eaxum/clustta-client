<template>
  <div v-if="supportsVersionedDependencies" class="dependency-selector" @click.stop>
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
import utils from '@/services/utils';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useUserStore } from '@/stores/users';
import { VERSIONED_DEPENDENCIES } from '@/lib/apiCapabilities';

const props = defineProps({
  edge: { type: Object, required: true },
  ownerAssetId: { type: String, required: true },
  editable: { type: Boolean, default: false },
  triggerOnBadge: { type: Boolean, default: true },
});

const emit = defineEmits(['updated']);
const { locale, t } = useI18n();
const iconStore = useIconStore();
const menu = useMenu();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const userStore = useUserStore();
const supportsVersionedDependencies = computed(
  () => projectStore.supportsCapability(VERSIONED_DEPENDENCIES),
);
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
    return props.edge.tag_name || 'Tag';
  }
  return 'Latest';
});

const selectedOptionId = computed(() => {
  if (props.edge.resolution_mode === 'pinned') return `pinned-${props.edge.checkpoint_id}`;
  if (props.edge.resolution_mode === 'tagged') return `tagged-${props.edge.asset_checkpoint_tag_id}`;
  return 'floating';
});

const compactMenuOptions = computed(() => [
  {
    id: 'floating',
    label: 'Latest',
    icon: iconStore.getAppIcon('clock'),
    mode: 'floating',
    selectorId: '',
    group: 'general',
  },
  ...options.value.tags.map(tag => ({
    id: `tagged-${tag.id}`,
    label: tag.name,
    icon: iconStore.getAppIcon('tag'),
    mode: 'tagged',
    selectorId: tag.id,
    group: 'tags',
  })),
  ...options.value.checkpoints.map((checkpoint) => {
    const author = userStore.getUserData(checkpoint.author_id);
    const authorName = author
      ? `${author.first_name} ${author.last_name}`
      : t('notifications.removedUser');
    const formattedDate = utils.formatDate(checkpoint.created_at, locale.value);
    return {
      id: `pinned-${checkpoint.id}`,
      label: checkpoint.comment || 'No message',
      description: formattedDate,
      mode: 'pinned',
      selectorId: checkpoint.id,
      searchText: `${formattedDate} ${authorName}`,
      searchTerms: [checkpoint.author_id],
      group: 'checkpoints',
    };
  }),
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
  await menu.showCheckpointSelectionMenu(event, {
    key: menuKey.value,
    loading: !optionsLoaded.value,
    searchLoading: !optionsLoaded.value,
    searchPlaceholder: 'Start typing...',
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
    menu.updateCheckpointSelectionMenu(menuKey.value, {
      loading: false,
      searchLoading: false,
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
