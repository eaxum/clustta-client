<template>
  <div class="modal-container dependency-build-modal" v-stop-propagation>
    <HeaderArea title="Build with dependencies" :icon="getAppIcon('revert')" />
    <div class="general-container build-plan-container">
      <div class="build-summary">
        <span>{{ plan.entries.length }} checkpoints</span>
        <span v-if="missingChunkCount">{{ missingChunkCount }} downloads</span>
        <span v-if="modifiedEntries.length" class="build-warning">{{ modifiedEntries.length }} modified files</span>
      </div>

      <div v-if="plan.conflicts.length" class="build-alert">
        <strong>Build conflicts</strong>
        <div v-for="conflict in plan.conflicts" :key="conflict.asset_id || conflict.message">
          {{ conflict.message }}
        </div>
      </div>

      <div v-if="plan.warnings.length" class="build-alert">
        <div v-for="warning in plan.warnings" :key="warning">{{ warning }}</div>
      </div>

      <div class="build-groups">
        <div v-for="group in resolutionGroups" :key="group.mode" class="build-group">
          <div class="build-group-title">{{ group.label }} - {{ group.entries.length }}</div>
          <div v-for="entry in group.entries" :key="`${entry.asset_id}-${entry.checkpoint_id}`" class="build-entry">
            <span>{{ assetName(entry.asset_id) }}</span>
            <span class="build-checkpoint">{{ entryRequirement(entry) }}</span>
          </div>
        </div>
      </div>

      <label v-if="modifiedEntries.length" class="overwrite-confirmation">
        <input v-model="allowModified" type="checkbox" />
        Overwrite locally modified dependency files
      </label>

      <div class="pop-up-actions">
        <GeneralButton label="Cancel" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton label="Build" :fullWidth="true" :isActive="canBuild" :loading="isBuilding" @click="executeBuild" />
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
const assetNames = ref(new Map());
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

const resolutionGroups = computed(() => {
  const labels = { floating: 'Latest', pinned: 'Exact pins', tagged: 'Tags', root: 'Root asset' };
  const grouped = new Map();
  plan.value.entries.forEach(entry => {
    const mode = entry.resolution_mode || 'floating';
    if (!grouped.has(mode)) grouped.set(mode, []);
    grouped.get(mode).push(entry);
  });
  return [...grouped.entries()].map(([mode, entries]) => ({ mode, entries, label: labels[mode] || mode }));
});

const getAppIcon = iconName => iconStore.getAppIcon(iconName);
const assetName = assetId => assetNames.value.get(assetId) || assetId.slice(0, 8);
const entryRequirement = (entry) => {
  const edge = dependencyEdges.value.get(entry.dependency_edge_id);
  if (edge?.resolution_mode === 'tagged') return `${edge.tag_name} -> ${edge.resolved_checkpoint_label}`;
  if (edge?.resolution_mode === 'pinned') return edge.resolved_checkpoint_label || entry.checkpoint_id.slice(0, 8);
  return edge?.resolved_checkpoint_label || entry.checkpoint_id.slice(0, 8);
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
    assetNames.value = new Map(assets.map(asset => [asset.id, asset.name]));
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
  width: min(620px, 90vw);
}

.build-plan-container,
.build-groups,
.build-group {
  display: flex;
  flex-direction: column;
  gap: .5rem;
}

.build-groups {
  max-height: 360px;
  overflow-y: auto;
}

.build-summary,
.build-entry {
  display: flex;
  justify-content: space-between;
  gap: .75rem;
}

.build-summary,
.build-group-title {
  font-size: .75rem;
  font-weight: 600;
}

.build-group {
  padding: .5rem;
  border-radius: .4rem;
  background: var(--surface-2);
}

.build-entry {
  color: var(--text-muted);
  font-size: .72rem;
}

.build-checkpoint {
  font-family: monospace;
}

.build-alert,
.build-warning {
  color: var(--danger);
}

.build-alert {
  padding: .5rem;
  border: 1px solid var(--danger);
  border-radius: .4rem;
  font-size: .72rem;
}

.overwrite-confirmation {
  display: flex;
  align-items: center;
  gap: .4rem;
  color: var(--text);
  font-size: .75rem;
}
</style>
