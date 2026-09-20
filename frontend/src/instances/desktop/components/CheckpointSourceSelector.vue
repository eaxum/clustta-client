<template>
  <div class="checkpoint-source-selector">
    <div class="source-row">
      <DropDownBox :items="assetOptions" :selectedItem="sourceAssetId" :onSelect="selectAsset"
        placeHolder="Source" searchable searchPlaceholder="Start typing..." :searchLoading="assetsLoading"
        :disabled="disabled || loading">
        <template #item="{ item, close }">
          <AssetItem :item="item.asset"
            :hideExtension="!showExtensions" :tooltip="showExtensions ? item.searchText : ''"
            variant="compact" :showBadge="false" :disableOutline="true">
            <template v-if="item.id === sourceAssetId" #actions>
              <ActionButton :icon="iconStore.getAppIcon('close')" :isDisabled="disabled || loading"
                :buttonFunction="() => clearSource(close)" v-tooltip="t('common.remove')" />
            </template>
          </AssetItem>
        </template>
      </DropDownBox>
      <div class="latest-control">
        <CheckBox v-model="latest" ariaLabel="Use latest source checkpoint" :disabled="disabled || loading" />
        <span @click="toggleLatest">Latest</span>
      </div>
    </div>
    <DropDownBox v-if="sourceAssetId && !latest" :items="checkpointOptions" :selectedItem="checkpointId"
      :onSelect="selectCheckpoint" placeHolder="Checkpoint" searchable searchPlaceholder="Start typing..."
      :searchLoading="loading" :disabled="disabled || loading">
      <template #item="{ item }">
        <div class="source-option" :title="item.name">
          <span class="checkpoint-option-message">{{ item.name }}</span>
          <span class="checkpoint-option-date">{{ item.formattedDate }}</span>
        </div>
      </template>
    </DropDownBox>
    <small v-if="error" class="source-error">{{ error }}</small>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import CheckBox from '@/instances/common/components/CheckBox.vue';
import AssetItem from '@/instances/desktop/components/AssetItem.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import utils from '@/services/utils';
import { AssetService, CheckpointService } from '@/services';
import { useAssetStore } from '@/stores/assets';
import { useProjectStore } from '@/stores/projects';
import { useIconStore } from '@/stores/icons';
import { useUserStore } from '@/stores/users';

const props = defineProps({
  sourceCheckpointId: { type: String, default: '' },
  checkpointId: { type: String, default: '' },
  disabled: { type: Boolean, default: false },
  showExtensions: { type: Boolean, default: false },
});
const emit = defineEmits(['validityChange']);
const assetStore = useAssetStore();
const iconStore = useIconStore();
const projectStore = useProjectStore();
const userStore = useUserStore();
const { locale, t } = useI18n();
const sourceAssetId = ref('');
const checkpointId = ref(props.sourceCheckpointId);
const latest = ref(!props.sourceCheckpointId);
const checkpoints = ref([]);
const assets = ref([]);
const assetsLoading = ref(false);
const loading = ref(false);
const error = ref('');
const changed = ref(false);
const isValid = computed(() => !sourceAssetId.value || latest.value || !!checkpointId.value);
const assetOptions = computed(() => (
  assets.value.filter(asset => !asset.is_link && !asset.trashed).map(asset => ({
    id: asset.id,
    name: props.showExtensions ? `${asset.name}${asset.extension || ''}` : asset.name,
    selectionValue: asset.id,
    searchText: asset.asset_path || asset.collection_path || '',
    asset,
  }))
));
const checkpointOptions = computed(() => checkpoints.value
  .filter(checkpoint => checkpoint.id !== props.checkpointId)
  .map(checkpoint => {
    const author = userStore.getUserData(checkpoint.author_id);
    const authorName = author ? `${author.first_name} ${author.last_name}` : t('notifications.removedUser');
    const formattedDate = utils.formatDate(checkpoint.created_at, locale.value);
    return {
      id: checkpoint.id,
      name: checkpoint.comment || 'No message',
      selectionValue: checkpoint.id,
      formattedDate,
      searchText: `${formattedDate} ${authorName}`,
      searchTerms: [checkpoint.author_id],
    };
  }));

const loadCheckpoints = async () => {
  const assetId = sourceAssetId.value;
  checkpoints.value = [];
  if (!assetId) return;
  loading.value = true;
  error.value = '';
  try {
    const result = await CheckpointService.GetCheckpoints(projectStore.activeProject.uri, assetId);
    if (sourceAssetId.value === assetId) checkpoints.value = result;
  } catch (cause) {
    error.value = `Unable to load source checkpoints: ${cause}`;
  } finally {
    loading.value = false;
  }
};
const selectAsset = async (id) => {
  changed.value = true;
  sourceAssetId.value = id;
  checkpointId.value = '';
  error.value = '';
  await loadCheckpoints();
};
const selectCheckpoint = (id) => {
  changed.value = true;
  checkpointId.value = id;
};
const clearSource = (closeList) => {
  if (props.disabled || loading.value) return;
  changed.value = true;
  sourceAssetId.value = '';
  checkpointId.value = '';
  checkpoints.value = [];
  latest.value = true;
  error.value = '';
  closeList();
};
const toggleLatest = () => {
  if (!props.disabled && !loading.value) latest.value = !latest.value;
};
const resolve = async () => {
  if (loading.value) throw new Error('Source checkpoints are still loading');
  if (!changed.value && !latest.value && props.sourceCheckpointId) return props.sourceCheckpointId;
  if (!sourceAssetId.value) return '';
  if (!latest.value && !checkpointId.value) throw new Error('Select a source checkpoint');
  return CheckpointService.ResolveCheckpointSource(
    projectStore.activeProject.uri, sourceAssetId.value, latest.value ? '' : checkpointId.value,
  );
};
const loadAssets = async () => {
  assetsLoading.value = true;
  try {
    assets.value = await AssetService.GetAssets(projectStore.activeProject.uri);
  } catch (cause) {
    assets.value = assetStore.assets;
    error.value = `Unable to refresh project assets: ${cause}`;
  } finally {
    assetsLoading.value = false;
  }
};

watch(isValid, value => emit('validityChange', value), { immediate: true });

onMounted(async () => {
  await loadAssets();
  if (!props.sourceCheckpointId) return;
  loading.value = true;
  try {
    const source = await CheckpointService.GetCheckpoint(projectStore.activeProject.uri, props.sourceCheckpointId);
    sourceAssetId.value = source.asset_id;
    await loadCheckpoints();
  } catch (cause) {
    error.value = `Source unavailable. Select a replacement or remove the source. ${cause}`;
  } finally {
    loading.value = false;
  }
});
defineExpose({ resolve });
</script>

<style scoped>
.checkpoint-source-selector {
  display: flex;
  flex-direction: column;
  gap: .75rem;
  width: 100%;
}
.source-row {
  display: flex;
  align-items: center;
  gap: 1rem;
}
.source-row > :first-child {
  flex: 1;
  min-width: 0;
}
.latest-control {
  display: flex;
  align-items: center;
  gap: .5rem;
  flex-shrink: 0;
}
.latest-control span {
  cursor: pointer;
}
.source-option {
  display: flex;
  flex-direction: column;
  min-width: 0;
  width: 100%;
  gap: .25rem;
  padding: .5rem;
  box-sizing: border-box;
}
.checkpoint-option-message,
.checkpoint-option-date {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.checkpoint-option-message {
  color: var(--text-muted);
  font-family: 'Inter', sans-serif;
  font-size: 14px;
  font-weight: 400;
}

.checkpoint-option-date {
  color: var(--text-muted);
  font-family: 'Inter', sans-serif;
  font-size: 12px;
  font-weight: 400;
}
.source-error {
  color: var(--error-color, #e78383);
}
</style>
