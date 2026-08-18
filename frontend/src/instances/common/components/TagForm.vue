<template>
  <div class="tag-form-container">
    <div class="input-section">
      <input v-model="tagName" class="input-short" type="text" :placeholder="$t('components.tagForm.namePlaceholder')" v-focus />
      <InputAlert v-if="mode === 'create'" :show="isNameCollision" :message="$t('components.tagForm.nameAlreadyExists')" />
      <InputAlert v-else :show="isNameCollision" :message="$t('components.tagForm.mergeWarning')" />
    </div>

    <div class="pop-up-actions">
      <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="handleCancel" :colored="false" />
      <GeneralButton :label="submitLabel" :fullWidth="true" :buttonFunction="handleSubmit" :isActive="isFormValid" :loading="isSubmitting" />
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import InputAlert from '@/instances/common/components/InputAlert.vue';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useTagStore } from '@/stores/tags';
import { useTrayStates } from '@/stores/TrayStates';

const props = defineProps({
  mode: { type: String, default: 'create', validator: (value) => ['create', 'edit'].includes(value) },
  initialName: { type: String, default: '' },
  tagId: { type: String, default: '' },
});

const emit = defineEmits(['created', 'updated', 'cancel']);

const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const tagStore = useTagStore();
const trayStates = useTrayStates();
const { t } = useI18n();

const isSubmitting = ref(false);
const tagName = ref('');

const normalizedName = computed(() => tagName.value.trim().toLowerCase());
const collidingTag = computed(() => tagStore.tags.find((tag) => (
  tag.id !== props.tagId && tag.name.toLowerCase() === normalizedName.value
)));
const isNameCollision = computed(() => !!collidingTag.value);
const isUnchanged = computed(() => props.mode === 'edit' && tagName.value.trim() === props.initialName);
const isFormValid = computed(() => {
  if (!tagName.value.trim() || isUnchanged.value) return false;
  return props.mode === 'edit' || !isNameCollision.value;
});
const submitLabel = computed(() => {
  if (props.mode === 'create') return t('common.create');
  return isNameCollision.value ? t('components.tagForm.merge') : t('common.rename');
});

const createTag = async () => {
  isSubmitting.value = true;
  try {
    const tag = await tagStore.createTag(tagName.value);
    notificationStore.addNotification(t('notifications.tagCreated'), '', 'success');
    emit('created', tag);
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorCreatingTag'), error);
  } finally {
    isSubmitting.value = false;
  }
};

const updateTag = async (mergeOnCollision) => {
  isSubmitting.value = true;
  try {
    const tag = await tagStore.updateTag(props.tagId, tagName.value, mergeOnCollision);
    const notificationKey = mergeOnCollision ? 'notifications.tagsMerged' : 'notifications.tagRenamed';
    notificationStore.addNotification(t(notificationKey), '', 'success');
    emit('updated', tag);
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorUpdatingTag'), error);
    if (mergeOnCollision) throw error;
  } finally {
    isSubmitting.value = false;
  }
};

const confirmMerge = () => {
  trayStates.dangerousActionTitle = t('components.tagForm.mergeTitle');
  trayStates.dangerousActionMessage = t('components.tagForm.mergeMessage', {
    source: props.initialName,
    target: collidingTag.value.name,
  });
  trayStates.dangerousActionIcon = 'tag';
  trayStates.dangerousActionConfirmLabel = t('components.tagForm.merge');
  trayStates.dangerousActionShowInput = false;
  trayStates.dangerousActionFunction = () => updateTag(true);
  modals.setModalVisibility('confirmDangerousActionModal', true);
};

const handleCancel = () => emit('cancel');
const handleSubmit = () => {
  if (!isFormValid.value) return;
  if (props.mode === 'create') {
    createTag();
    return;
  }
  if (isNameCollision.value) {
    confirmMerge();
    return;
  }
  updateTag(false);
};

onMounted(() => {
  tagName.value = props.initialName;
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.tag-form-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  width: 100%;
}

.input-short {
  flex: 1;
  width: 100%;
}
</style>
