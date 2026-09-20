<template>
  <div class="modal-container edit-checkpoint-modal" v-esc="closeModal" v-stop-propagation>
    <HeaderArea title="Edit Checkpoint" icon="edit" />
    <div class="general-container">
      <textarea v-model="comment" class="desktop-input-long" placeholder="Checkpoint comment"
        :disabled="saving" />
      <div class="checkpoint-edit-tags">
        <div v-for="name in tagNames" :key="name" class="checkpoint-edit-tag">
          <span>{{ name }}</span>
          <ActionButton v-if="canManageTags" :icon="iconStore.getAppIcon('close')"
            :isDisabled="saving" :buttonFunction="() => removeTag(name)" />
        </div>
        <CheckpointTagSelector v-if="canManageTags" v-model="selectedTag" :assetIds="[assetId]"
          :disabled="loading || saving" />
      </div>
      <CheckpointSourceSelector v-if="checkpoint" ref="sourceSelector"
        :checkpointId="checkpoint.id" :sourceCheckpointId="checkpoint.source_checkpoint_id || ''"
        :disabled="saving" @validityChange="sourceSelectionValid = $event" />
      <div class="pop-up-actions">
        <GeneralButton label="Close" :fullWidth="true" :colored="false" :isActive="!saving" :buttonFunction="closeModal" />
        <GeneralButton label="Update" :fullWidth="true"
          :isActive="!loading && !saving && !!checkpoint && sourceSelectionValid"
          :loading="saving" :buttonFunction="save" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue';
import { CheckpointService } from '@/services';
import emitter from '@/lib/mitt';
import { canActOnAsset } from '@/lib/permissions';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import CheckpointSourceSelector from '@/instances/desktop/components/CheckpointSourceSelector.vue';
import CheckpointTagSelector from '@/instances/desktop/components/CheckpointTagSelector.vue';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useAssetStore } from '@/stores/assets';
import { useProjectStore } from '@/stores/projects';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';

const modals = useDesktopModalStore();
const assetStore = useAssetStore();
const projectStore = useProjectStore();
const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const projectPath = projectStore.activeProject.uri;
const selectedCheckpoint = modals.editCheckpoint;
const assetId = selectedCheckpoint.ownerId;
const checkpoint = ref(null);
const sourceSelector = ref(null);
const sourceSelectionValid = ref(true);
const comment = ref(selectedCheckpoint.comment);
const tagNames = ref([]);
const originalTags = ref([]);
const selectedTag = ref('');
const loading = ref(true);
const saving = ref(false);
const canManageTags = computed(() => canActOnAsset('manage_dependencies', assetStore.findAsset(assetId) || assetStore.selectedAsset));
const removeTag = (name) => {
  if (!saving.value) tagNames.value = tagNames.value.filter(tag => tag !== name);
};
watch(selectedTag, (name) => {
  if (name && !tagNames.value.includes(name)) tagNames.value.push(name);
  selectedTag.value = '';
});
const closeModal = () => {
  if (!saving.value) modals.setModalVisibility('editCheckpointModal', false);
};
const save = async () => {
  if (saving.value || loading.value || !checkpoint.value || !sourceSelectionValid.value) return;
  saving.value = true;
  try {
    const sourceId = await sourceSelector.value.resolve();
    const tagsChanged = JSON.stringify([...tagNames.value].sort()) !== JSON.stringify([...originalTags.value].sort());
    await CheckpointService.UpdateCheckpoint(projectPath, checkpoint.value.id, comment.value, sourceId,
      canManageTags.value && tagsChanged ? tagNames.value : null);
    emitter.emit('update-checkpoints');
    emitter.emit('refresh-browser');
    emitter.emit('get-project-data');
    modals.setModalVisibility('editCheckpointModal', false);
  } catch (error) {
    notificationStore.errorNotification('Unable to edit checkpoint', error);
  } finally {
    saving.value = false;
  }
};
onMounted(async () => {
  try {
    const result = await CheckpointService.GetCheckpoint(projectPath, selectedCheckpoint.checkpoint_id);
    const tags = await CheckpointService.GetCheckpointTags(projectPath, assetId);
    checkpoint.value = result;
    comment.value = result.comment;
    tagNames.value = tags.filter(tag => tag.checkpoint_id === result.id).map(tag => tag.name);
    originalTags.value = [...tagNames.value];
  } catch (error) {
    notificationStore.errorNotification('Unable to load checkpoint', error);
  } finally {
    loading.value = false;
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";
.desktop-input-long {
  width: 100%;
  margin: 0;
  min-height: 9rem;
  color: var(--text);
  background-color: var(--bg);
}
.checkpoint-edit-tags {
  display: flex;
  flex-direction: column;
  gap: .5rem;
  width: 100%;
}
.checkpoint-edit-tag {
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-sizing: border-box;
  width: 100%;
  padding: .35rem .6rem;
  border-radius: var(--normal-radius);
  background-color: var(--surface-2);
}
.checkpoint-edit-tag span {
  overflow: hidden;
  color: var(--text);
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
