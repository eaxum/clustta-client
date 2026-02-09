<template>
  <div class="type-form-container">
    <div class="input-section">
      <input v-model="typeName" class="input-short" type="text" placeholder="Collection Type Name" v-focus />
      <InputAlert :show="isNameTaken" message="A collection type with this name already exists." />
    </div>

    <IconGrid @iconSelected="setIcon" :icons="availableIcons" />

    <div class="pop-up-actions">
      <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="handleCancel" :colored="false" />
      <GeneralButton :label="submitLabel" :fullWidth="true" :buttonFunction="handleSubmit" :isActive="isFormValid" :loading="isSubmitting" />
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';
import iconData from '@/data/iconData.json';

// components
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import IconGrid from '@/instances/desktop/components/IconGrid.vue';
import InputAlert from '@/instances/common/components/InputAlert.vue';

// services
import { CollectionService } from '@/services';

// stores
import { useCollectionStore } from '@/stores/collections';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';

const collectionStore = useCollectionStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();

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
  const allIcons = iconData.icons;
  const usedIcons = collectionStore.collectionTypes
    .filter((item) => props.mode === 'edit' ? item.id !== props.typeId : true)
    .map((item) => item.icon);
  return allIcons.filter((icon) => !usedIcons.includes(icon));
});

// Returns whether the type name already exists.
const isNameTaken = computed(() => {
  const existingNames = collectionStore.collectionTypes
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
  return props.mode === 'create' ? 'Create' : 'Update';
});

// methods
// Creates a new collection type.
const createType = async () => {
  isSubmitting.value = true;
  await CollectionService.CreateCollectionType(projectStore.activeProject.uri, typeName.value, typeIcon.value)
    .then((response) => {
      notificationStore.addNotification('Collection type created', '', 'success');
      collectionStore.collectionTypes.push(response);
      emit('created', response);
    })
    .catch((error) => {
      notificationStore.errorNotification('Error creating collection type', error);
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

// Updates an existing collection type.
const updateType = async () => {
  isSubmitting.value = true;
  await CollectionService.UpdateCollectionType(projectStore.activeProject.uri, props.typeId, typeName.value, typeIcon.value)
    .then((response) => {
      notificationStore.addNotification('Collection type updated', '', 'success');
      const index = collectionStore.collectionTypes.findIndex((t) => t.id === props.typeId);
      if (index !== -1) {
        collectionStore.collectionTypes[index] = response;
      }
      emit('updated', response);
    })
    .catch((error) => {
      notificationStore.errorNotification('Error updating collection type', error);
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
