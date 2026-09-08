<template>
  <div class="modal-container dependency-build-modal" v-stop-propagation>
    <HeaderArea title="Download" :icon="getAppIcon('download')" />
    <div class="general-container build-plan-container">
      <div class="build-summary">
        <span class="build-target" v-tooltip="rootAsset.name">{{ rootAsset.name || 'Download preview' }}</span>
        <span>{{ plan.entries.length }} assets</span>
        <span v-if="missingChunkCount">{{ missingChunkCount }} downloads</span>
        <span v-if="modifiedEntries.length" class="build-warning">{{ modifiedEntries.length }} modified files</span>
      </div>

      <div class="build-preview-scroll">
      <div v-if="plan.conflicts.length" class="build-alert build-conflict" role="alert">
        <strong>Dependency conflicts</strong>
        <div v-for="conflict in plan.conflicts" :key="conflict.asset_id || conflict.message">
          {{ conflict.message }}
        </div>
      </div>

      <div v-if="plan.warnings.length" class="build-alert">
        <div v-for="warning in plan.warnings" :key="warning">{{ warning }}</div>
      </div>

      <div class="build-list">
        <AssetItem v-for="entry in plan.entries" :key="`${entry.asset_id}-${entry.checkpoint_id}`"
          :item="assetItem(entry.asset_id)" :showBadge="false" :hideExtension="true">
          <template #persistent>
            <span v-if="entry.requires_overwrite" class="build-status build-warning" v-tooltip="'Locally modified - will be overwritten'" aria-label="Locally modified">
              <img class="small-icons" :src="getAppIcon('alert')" alt="" />
            </span>
            <span v-if="entry.missing_chunks" class="build-status" v-tooltip="'Checkpoint will be downloaded'" aria-label="Download required">
              <img class="small-icons" :src="getAppIcon('download')" alt="" />
            </span>
            <span class="build-version" v-tooltip="entryRequirement(entry)">
              <img class="small-icons" :src="getAppIcon(entryIcon(entry))" alt="" />
              <span>{{ entryRequirement(entry) }}</span>
            </span>
          </template>
        </AssetItem>
      </div>
      </div>

      <div v-if="modifiedEntries.length" class="overwrite-confirmation">
        <CheckBox v-model="allowModified" :disabled="isBuilding" ariaLabel="Overwrite locally modified files" />
        <span>Overwrite {{ modifiedEntries.length }} locally modified {{ modifiedEntries.length === 1 ? 'file' : 'files' }}</span>
      </div>

      <div class="pop-up-actions">
        <GeneralButton label="Cancel" :fullWidth="true" :buttonFunction="closeModal" :colored="false" :isActive="!isBuilding" />
        <GeneralButton label="Download" :fullWidth="true" :isActive="canBuild" :loading="isBuilding" @click="executeBuild" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import emitter from '@/lib/mitt';
import { AssetService, CheckpointService } from '@/services';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import CheckBox from '@/instances/common/components/CheckBox.vue';
import AssetItem from '@/instances/desktop/components/AssetItem.vue';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';

const { t } = useI18n();
const modals = useDesktopModalStore();
const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const allowModified = ref(false);
const assetsById = ref(new Map());
const dependencyEdges = ref(new Map());
const isBuilding = ref(false);
const plan = computed(() => modals.dependencyBuildPlan.plan || { entries: [], warnings: [], conflicts: [] });

const modifiedEntries = computed(() => plan.value.entries.filter(entry => entry.requires_overwrite));
const missingChunkCount = computed(() => plan.value.entries.filter(entry => entry.missing_chunks).length);
const canBuild = computed(() => {
  return !isBuilding.value
    && plan.value.conflicts.length === 0
    && (modifiedEntries.value.length === 0 || allowModified.value);
});

const rootAsset = computed(() => assetItem(modals.dependencyBuildPlan.rootAssetId));
const getAppIcon = iconName => iconStore.getAppIcon(iconName);
const assetItem = assetId => assetsById.value.get(assetId) || { name: 'Loading asset…' };
const entryIcon = entry => ({ pinned: 'pin', tagged: 'tag' }[entry.resolution_mode] || 'clock');
const entryRequirement = (entry) => {
  const edge = dependencyEdges.value.get(entry.dependency_edge_id);
  if (entry.resolution_mode === 'tagged') return edge?.tag_name || 'Tagged';
  if (entry.resolution_mode === 'pinned') {
    // Edge labels may have changed since this frozen plan was resolved.
    return edge?.resolved_checkpoint_id === entry.checkpoint_id && edge.resolved_checkpoint_label
      ? `Pinned · ${edge.resolved_checkpoint_label}`
      : 'Pinned';
  }
  return 'Latest';
};
const closeModal = () => modals.disableAllModals();

const executeBuild = async () => {
  if (!canBuild.value) return;
  isBuilding.value = true;
  try {
    await CheckpointService.ExecuteDependencyBuildPlan(
      projectStore.activeProject.uri,
      projectStore.getActiveProjectUrl,
      modals.dependencyBuildPlan.rootAssetId,
      plan.value.fingerprint,
      allowModified.value,
    );
    emitter.emit('refresh-browser');
    closeModal();
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorRevertingAssets'), error);
  } finally {
    isBuilding.value = false;
  }
};

onMounted(async () => {
  try {
    const assets = await AssetService.GetAssets(projectStore.activeProject.uri);
    assetsById.value = new Map(assets.map(asset => [asset.id, asset]));
    const ownerIds = [...new Set(plan.value.entries.map(entry => entry.requested_by_asset_id).filter(Boolean))];
    const edgeGroups = await Promise.all(ownerIds.map(assetId => (
      AssetService.GetAssetDependencyEdges(projectStore.activeProject.uri, assetId)
    )));
    dependencyEdges.value = new Map(edgeGroups.flat().map(edge => [edge.id, edge]));
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorLoadingProjectData'), error);
  }
});
</script>

<style scoped>
.dependency-build-modal {
  width: 840px;
  min-width: min(840px, 90vw);
  max-width: 90vw;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
}
.build-plan-container {
  width: 100%;
  min-width: 0;
  max-width: none;
  min-height: 0;
  align-items: stretch;
  gap: 1rem;
  overflow: hidden;
  color: var(--text);
}
.build-summary {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: .5rem 1rem;
  padding-bottom: .75rem;
  border-bottom: 1px solid var(--border);
  color: var(--text-muted);
  font-size: .8rem;
}
.build-target {
  flex: 1 1 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 1rem;
  font-weight: 500;
  color: var(--text);
}
.build-preview-scroll {
  min-width: 50vw;
  min-height: 0;
  max-height: min(360px, 45vh);
  overflow-y: auto;
  overflow-x: hidden;
  scrollbar-gutter: stable;
  scrollbar-width: thin;
  scrollbar-color: var(--surface-4) transparent;
}
.build-preview-scroll::-webkit-scrollbar {
  width: 4px;
}
.build-preview-scroll::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--surface-4);
}
.build-preview-scroll::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}
.build-list {
  display: flex;
  flex-direction: column;
  gap: .5rem;
}
.build-version, .build-status {
  display: inline-flex;
  align-items: center;
  gap: .35rem;
  color: var(--text-muted);
}
.build-version {
  max-width: 180px;
  padding: .25rem .5rem;
  border: 1px solid var(--border);
  border-radius: 999px;
  color: var(--text);
  font-size: .8rem;
}
.build-version span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.build-version img, .build-status img {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}
.build-status {
  padding: .25rem;
}
.build-alert {
  padding: .75rem;
  margin-bottom: .75rem;
  border: 1px solid color-mix(in srgb, var(--warning) 45%, transparent);
  border-radius: var(--normal-radius);
  background: color-mix(in srgb, var(--warning) 10%, transparent);
  color: var(--warning);
  font-size: .85rem;
  overflow-wrap: anywhere;
}
.build-warning {
  color: var(--warning);
}
.build-conflict {
  color: var(--danger);
  border-color: var(--danger);
  background: color-mix(in srgb, var(--danger) 10%, transparent);
}
.overwrite-confirmation {
  display: flex;
  align-items: center;
  gap: .5rem;
  font-size: .85rem;
  color: var(--warning);
}
.overwrite-confirmation :deep(.checkbox-container) {
  flex-shrink: 0;
}
.pop-up-actions {
  flex-shrink: 0;
  margin-top: auto;
}
@media (max-width: 480px) {
  .build-version { max-width: 110px; }
}
</style>
