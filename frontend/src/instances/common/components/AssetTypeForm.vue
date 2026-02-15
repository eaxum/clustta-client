<template>
  <div class="type-form-container">
    <div class="input-section">
      <input v-model="typeName" class="input-short" type="text" :placeholder="$t('components.assetTypeForm.assetTypeNamePlaceholder')" v-focus />
      <InputAlert :show="isNameTaken" :message="$t('components.assetTypeForm.nameAlreadyExists')" />
    </div>

    <IconGrid @iconSelected="setIcon" :icons="availableIcons" />

    <div class="pop-up-actions">
      <GeneralButton :label="$t('components.assetTypeForm.cancel')" :fullWidth="true" :buttonFunction="handleCancel" :colored="false" />
      <GeneralButton :label="submitLabel" :fullWidth="true" :buttonFunction="handleSubmit" :isActive="isFormValid" :loading="isSubmitting" />
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import iconData from '@/data/iconData.json';

// components
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import IconGrid from '@/instances/desktop/components/IconGrid.vue';
import InputAlert from '@/instances/common/components/InputAlert.vue';

// services
import { AssetService } from '@/services';

// stores
import { useAssetStore } from '@/stores/assets';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';

const assetStore = useAssetStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const { t } = useI18n();

// props
const props = defineProps({
  mode: { type: String, default: 'create', validator: (v) => ['create', 'edit'].includes(v) },
  initialName: { type: String, default: '' },
  initialIcon: { type: String, default: 'generic' },
  typeId: { type: String, default: '' },
});

// emits
const emit = defineEmits(['created', 'updated', 'cancel', 'iconChange']);

// refs
const isSubmitting = ref(false);
const typeIcon = ref('generic');
const typeName = ref('');

// computed
// Returns available icons excluding already used ones (except current in edit mode).
const availableIcons = computed(() => {
  const allIcons = iconData.icons.filter((item) => item !== 'weblink');
  const usedIcons = assetStore.assetTypes
    .filter((item) => props.mode === 'edit' ? item.id !== props.typeId : true)
    .map((item) => item.icon);
  return allIcons.filter((icon) => !usedIcons.includes(icon));
});

// Returns whether the type name already exists.
const isNameTaken = computed(() => {
  const existingNames = assetStore.assetTypes
    .filter((item) => props.mode === 'edit' ? item.id !== props.typeId : true)
    .map((item) => item.name.toLowerCase());
  return existingNames.includes(typeName.value.trim().toLowerCase());
});

// Returns whether the form is valid for submission (depends on isNameTaken).
const isFormValid = computed(() => {
  return !!typeName.value.trim() && typeIcon.value !== 'generic' && !isNameTaken.value;
});

// Returns the submit button label based on mode.
const submitLabel = computed(() => {
  return props.mode === 'create' ? t('components.assetTypeForm.create') : t('components.assetTypeForm.update');
});

// methods
// Creates a new asset type.
const createType = async () => {
  isSubmitting.value = true;
  await AssetService.CreateAssetType(projectStore.activeProject.uri, typeName.value, typeIcon.value)
    .then((response) => {
      notificationStore.addNotification(t('components.assetTypeForm.assetTypeCreated'), '', 'success');
      assetStore.assetTypes.push(response);
      emit('created', response);
    })
    .catch((error) => {
      notificationStore.errorNotification(t('components.assetTypeForm.errorCreatingAssetType'), error);
    })
    .finally(() => {
      isSubmitting.value = false;
    });
};

// Handles cancel button click.
const handleCancel = () => {
  emit('cancel');
};

// Handles form submission based on mode.
const handleSubmit = () => {
  if (props.mode === 'create') {
    createType();
  } else {
    updateType();
  }
};

// Sets the selected icon.
const setIcon = (icon) => {
  typeIcon.value = icon;
  emit('iconChange', icon);
};

// Updates an existing asset type.
const updateType = async () => {
  isSubmitting.value = true;
  await AssetService.UpdateAssetType(projectStore.activeProject.uri, props.typeId, typeName.value, typeIcon.value)
    .then((response) => {
      notificationStore.addNotification(t('components.assetTypeForm.assetTypeUpdated'), '', 'success');
      const index = assetStore.assetTypes.findIndex((t) => t.id === props.typeId);
      if (index !== -1) {
        assetStore.assetTypes[index] = response;
      }
      emit('updated', response);
    })
    .catch((error) => {
      notificationStore.errorNotification(t('components.assetTypeForm.errorUpdatingAssetType'), error);
    })
    .finally(() => {
      isSubmitting.value = false;
    });
};

// lifecycle hooks
onMounted(() => {
  typeName.value = props.initialName;
  typeIcon.value = props.initialIcon;
});

// expose
defineExpose({ typeIcon, typeName });
</script>

<style scoped>
@import "@/assets/desktop.css";

.type-form-container {
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
