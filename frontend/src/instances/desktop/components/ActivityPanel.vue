<template>
  <section class="activity-panel expandable-panel" :class="{ 'panel-maximized': maximized }" :aria-label="$t('activity.title')">
    <ExpandablePanelHeader :inert="!!detailsOperation" v-model="searchQuery" :title="$t('activity.title')" icon="activity" closeIcon="chevron-down"
      :count="searchQuery ? `${filteredOperations.length}/${activity.operations.length}` : activity.operations.length"
      :filterPlaceholder="$t('activity.filterPlaceholder')" :maximized="maximized"
      @toggle-maximize="$emit('toggle-maximize')" @close="activity.expanded = false">
      <ActionButton :icon="icons.getAppIcon('broom')" v-tooltip="$t('activity.clearFinished')"
        :allowDeactivate="activity.finished.length > 0" :isDisabled="!activity.finished.length" :buttonFunction="() => activity.clearFinished()" />
    </ExpandablePanelHeader>
    <div class="activity-content">
      <div class="activity-list app-scrollbar" :inert="!!detailsOperation">
        <p v-if="activity.connectionError" class="activity-error" role="alert">{{ activity.connectionError }}</p>
        <article v-for="item in filteredOperations" :key="item.operation_id" class="activity-row">
          <div class="activity-row-header">
            <div class="activity-label">
              <strong>{{ activityDisplayName(item) }}</strong>
              <span class="activity-project-separator" aria-hidden="true"></span>
              <span class="activity-project-name">{{ sentenceCase(item.project_name) }}</span>
            </div>
            <span v-if="item.status === 'completed'" class="activity-status">
              {{ item.transferSizes ? $t('activity.downloadedTooltip', { size: utils.formatBytes(item.transfer.downloaded, 2) }) : $t('activity.completed') }}
            </span>
            <strong v-else-if="item.status === 'running' && item.total > 0 && (item.transfer.downloading || Math.round(item.percentage) > 0)" class="activity-metric">
              {{ item.transfer.downloading ? item.transferSizes : `${Math.round(item.percentage)}%` }}
            </strong>
            <span v-else-if="item.status !== 'running'" class="activity-status"
              v-tooltip="item.transferSizes">{{ $t(`activity.${item.status}`) }}</span>
            <ActionButton v-if="item.collections?.length || item.assets?.length" :icon="icons.getAppIcon('file-search')"
              v-tooltip="$t(item.collections?.length ? 'menus.goToCollection' : 'menus.goToAsset')"
              :allowDeactivate="true" :buttonFunction="() => goToItem(item, item.collections?.[0] || item.assets[0], item.collections?.length ? 'collection' : 'asset')" />
            <ActionButton :icon="icons.getAppIcon('info')" v-tooltip="$t('activity.details')"
              :allowDeactivate="true" :buttonFunction="() => detailsId = item.operation_id" />
            <ActionButton v-if="item.status === 'running' || item.status === 'queued'" :icon="icons.getAppIcon('close-circle')"
              v-tooltip="$t(item.cancel_requested ? 'activity.cancelling' : 'activity.cancel')"
              :allowDeactivate="!item.cancel_requested" :isDisabled="item.cancel_requested" :buttonFunction="() => activity.cancel(item.operation_id)" />
            <ActionButton v-else :icon="icons.getAppIcon('close')" v-tooltip="$t('activity.dismiss')"
              :allowDeactivate="true" :buttonFunction="() => activity.dismiss(item.operation_id)" />
          </div>
          <div v-if="item.status === 'running'" class="activity-progress-bar" role="progressbar"
            :aria-valuetext="transferTooltip(item)"
            :aria-label="item.title" :aria-valuenow="item.total > 0 ? item.percentage : undefined" aria-valuemin="0" aria-valuemax="100">
            <span v-if="item.transfer.savedPercentage > 0" class="activity-progress-saved"
              v-tooltip="$t('activity.savedTooltip', { size: utils.formatBytes(item.transfer.saved, 2) })" :style="{ width: item.transfer.savedPercentage + '%' }"></span>
            <span v-if="item.transfer.downloadedPercentage > 0" class="activity-progress-downloaded"
              v-tooltip="item.transfer.downloading ? $t('activity.downloadedTooltip', { size: utils.formatBytes(item.transfer.downloaded, 2) }) : $t('activity.rebuildingTooltip', { percent: Math.round(item.percentage) })" :style="{ width: item.transfer.downloadedPercentage + '%' }"></span>
          </div>
          <p v-if="item.error || activity.errors[item.operation_id]" class="activity-error" role="alert">{{ item.error || activity.errors[item.operation_id] }}</p>
        </article>
        <PageState v-if="!filteredOperations.length" class="activity-empty"
          :message="$t(searchQuery ? 'activity.noMatches' : 'activity.noActivity')"
          :illustration="`/page-states/workflow.png`" />
      </div>
    </div>
    <div v-if="detailsOperation" class="activity-details-overlay" @click.self="detailsId = null">
      <ActivityDetails :operation="detailsOperation" :error="activity.errors[detailsId]"
        @close="detailsId = null" @navigate="goToItem" />
    </div>
  </section>
</template>
<script setup>
import { computed, nextTick, ref } from 'vue';
import { useActivityStore } from '@/stores/activity';
import { useIconStore } from '@/stores/icons';
import { useCommonStore } from '@/stores/common';
import { useProjectStore } from '@/stores/projects';
import { useCollectionStore } from '@/stores/collections';
import { usePaneStore } from '@/stores/panes';
import { useStageStore } from '@/stores/stages';
import { AssetService, CollectionService } from '@/services';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import ActivityDetails from '@/instances/desktop/components/ActivityDetails.vue';
import ExpandablePanelHeader from '@/instances/desktop/components/ExpandablePanelHeader.vue';
import { activityDisplayName, activityTransferSizes, activityTransferProgress, sentenceCase, filterActivityOperations } from '@/lib/activity';
import PageState from '@/instances/common/components/PageState.vue';
import { formatError } from '@/lib/errors';
import emitter from '@/lib/mitt';
import utils from '@/services/utils';
import { useI18n } from 'vue-i18n';

const activity = useActivityStore();
const { t } = useI18n();
const transferTooltip = (item) => {
  const { saved, downloaded, downloading } = item.transfer;
  if (!downloading) return t('activity.rebuildingTooltip', { percent: Math.round(item.percentage) });
  return t('activity.transferTooltip', { saved: utils.formatBytes(saved, 2), downloaded: utils.formatBytes(downloaded, 2) });
};
defineProps({ maximized: { type: Boolean, default: false } });
const emit = defineEmits(['toggle-maximize', 'restore']);
const icons = useIconStore();
const commonStore = useCommonStore();
const projectStore = useProjectStore();
const collectionStore = useCollectionStore();
const panes = usePaneStore();
const stage = useStageStore();
const detailsId = ref(null);
const searchQuery = ref('');
const detailsOperation = computed(() => activity.operations.find((item) => item.operation_id === detailsId.value));
const filteredOperations = computed(() => filterActivityOperations(activity.sortedOperations, searchQuery.value)
  .map((item) => ({ ...item, transfer: activityTransferProgress(item), transferSizes: activityTransferSizes(item) })));

const goToItem = async (operation, requestedItem, type = 'asset') => {
  const viewMode = commonStore.viewMode;
  try {
    const project = [projectStore.activeProject, ...projectStore.projects]
      .find((item) => item?.id === operation.project_id && item?.uri === operation.project_uri);
    if (!project) throw new Error('Open the original project to locate this item.');
    const target = type === 'asset'
      ? await AssetService.GetAssetByID(project.uri, requestedItem.id)
      : requestedItem.id ? await CollectionService.GetCollectionByID(project.uri, requestedItem.id) : null;
    const collection = type === 'collection' ? target : target.collection_id
      ? await CollectionService.GetCollectionByID(project.uri, target.collection_id)
      : null;
    if (projectStore.activeProject?.uri !== project.uri) await projectStore.gotoProject(project);
    if (projectStore.activeProject?.uri !== project.uri) return;
    commonStore.activeWorkspace = 'Project';
    commonStore.viewSearchQuery = '';
    commonStore.resetFilters();
    commonStore.applyViewMode(viewMode);
    emit('restore');
    await nextTick();
    collectionStore.navigateToCollection(collection);
    commonStore.navigatorMode = true;
    stage.deselectAllItems();
    const selectedItem = target ? { ...target, type } : null;
    if (selectedItem) stage.selectItem(selectedItem, type);
    stage.selectedItem = selectedItem;
    stage.selectedItems = selectedItem ? [selectedItem] : [];
    stage.firstSelectedItemId = target?.id || '';
    stage.lastSelectedItemId = target?.id || '';
    stage.markedItems = target ? [target.id] : [];
    panes.showDetailsPane = true;
    panes.setPaneVisibility(target ? `${type}Details` : 'projectDetails', true);
    detailsId.value = null;
    await nextTick();
    emitter.emit('view-details');
    delete activity.errors[operation.operation_id];
  } catch (error) {
    activity.errors[operation.operation_id] = formatError(error);
  }
};
</script>
<style scoped>

.activity-panel {
  font-size: 13px;
  border-radius: var(--very-large-radius);
}

.activity-list {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow-y: auto;
  padding: .5rem;
  gap: .2rem;
}
.activity-row {
  padding: .5rem;
  border-radius: var(--large-radius);
  outline: var(--transparent-line);
  background-color: var(--surface-2);
  transition: all .2s ease-out;
}

.activity-row:hover {
  background: var(--hover);
  border-radius: var(--normal-radius);
}
.activity-row-header {
  display: flex;
  align-items: center;
  gap: .5rem;
}
.activity-label {
  display: flex;
  flex: 1;
  min-width: 0;
  align-items: baseline;
  gap: .4rem;
  overflow: hidden;
  white-space: nowrap;
}
.activity-metric {
  white-space: nowrap;
}
.activity-label strong, .activity-metric {
  font-weight: 400;
}
.activity-label span, .activity-status, .activity-empty {
  color: var(--text-muted);
  font-size: 12px;
}
.activity-status {
  white-space: nowrap;
  font-weight: 500;
}
.activity-label strong, .activity-project-name {
  overflow: hidden;
  text-overflow: ellipsis;
}
.activity-project-name {
  font-weight: 500;
}
.activity-project-separator {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: currentColor;
  flex-shrink: 0;
  align-self: center;
}
.activity-progress-bar {
  display: flex;
  width: 100%;
  height: .3rem;
  margin-top: .35rem;
  margin-bottom: .5rem;
  background: var(--surface-4);
  border-radius: var(--small-radius);
  overflow: hidden;
}
.activity-progress-bar span {
  height: 100%;
  border-radius: var(--small-radius);
}
.activity-progress-saved + .activity-progress-downloaded {
  margin-left: .2rem;
  box-sizing: border-box;
}
.activity-progress-saved {
  background: var(--info);
}
.activity-progress-downloaded {
  background: rgb(67, 210, 67);
}
.activity-error {
  color: var(--red);
  overflow-wrap: anywhere;
}

.activity-content {
  display: flex;
  flex: 1;
  min-height: 0;
}
.activity-details-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  justify-content: flex-end;
  background: #0006;
  z-index: 1;
}
</style>
