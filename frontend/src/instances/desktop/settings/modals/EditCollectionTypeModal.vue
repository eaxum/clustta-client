<template>
  <div class="modal-container" ref="modalContainer" v-stop-propagation v-esc="closeModal">
      <HeaderArea :title="title" :icon="typeIcon" />

    <div class="general-container">
      <CollectionTypeForm ref="typeFormRef" mode="edit" :initialName="initialName" :initialIcon="initialIcon" :typeId="typeId" @updated="handleUpdated" @cancel="closeModal" @iconChange="handleIconChange" />
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import CollectionTypeForm from '@/instances/common/components/CollectionTypeForm.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// stores
import { useCollectionStore } from '@/stores/collections';
import { useDesktopModalStore } from '@/stores/desktopModals';

const collectionStore = useCollectionStore();
const modals = useDesktopModalStore();

const { t } = useI18n();

// refs
const modalContainer = ref(null);
const typeFormRef = ref(null);
const typeIcon = ref('generic');

// constants
const title = t('modals.editCollectionType');

// computed
// Returns the initial icon from the selected collection type.
const initialIcon = computed(() => {
  return collectionStore.selectedCollectionType?.icon || 'generic';
});

// Returns the initial name from the selected collection type.
const initialName = computed(() => {
  return collectionStore.selectedCollectionType?.name || '';
});

// Returns the type ID from the selected collection type.
const typeId = computed(() => {
  return collectionStore.selectedCollectionType?.id || '';
});

// methods
// Closes the modal.
const closeModal = () => {
  modals.setModalVisibility('editCollectionTypeModal', false);
};

// Handles icon change from form.
const handleIconChange = (icon) => {
  typeIcon.value = icon;
};

// Handles successful type update.
const handleUpdated = () => {
  closeModal();
};
</script>

<style scoped>
@import "@/assets/desktop.css";

.input-short {
  flex: 1;
  width: 100%;
}
</style>