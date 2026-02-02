<template>
  <div class="modal-container" ref="modalContainer" v-stop-propagation v-esc="closeModal">
    <div class="general-pane-header">
      <HeaderArea :title="title" :icon="typeIcon" />
    </div>

    <div class="general-container">
      <CollectionTypeForm ref="typeFormRef" mode="create" @created="handleCreated" @cancel="closeModal" @iconChange="handleIconChange" />
    </div>
  </div>
</template>

<script setup>
// imports
import { ref } from 'vue';

// components
import CollectionTypeForm from '@/instances/common/components/CollectionTypeForm.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';

const modals = useDesktopModalStore();

// refs
const modalContainer = ref(null);
const typeFormRef = ref(null);
const typeIcon = ref('generic');

// constants
const title = 'Add Collection type';

// methods
// Closes the modal.
const closeModal = () => {
  modals.setModalVisibility('addCollectionTypeModal', false);
};

// Handles successful type creation.
const handleCreated = () => {
  closeModal();
};

// Handles icon change from form.
const handleIconChange = (icon) => {
  typeIcon.value = icon;
};
</script>

<style scoped>
@import "@/assets/desktop.css";

.input-short {
  flex: 1;
  width: 100%;
}
</style>