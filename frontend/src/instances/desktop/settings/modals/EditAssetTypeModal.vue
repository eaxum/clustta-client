<template>
  <div class="modal-container" ref="modalContainer" v-stop-propagation v-esc="closeModal">
    <div class="general-pane-header">
      <HeaderArea :title="title" :icon="typeIcon" />
    </div>

    <div class="general-container">
      <AssetTypeForm ref="typeFormRef" mode="edit" :initialName="initialName" :initialIcon="initialIcon" :typeId="typeId" @updated="handleUpdated" @cancel="closeModal" @iconChange="handleIconChange" />
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref } from 'vue';

// components
import AssetTypeForm from '@/instances/common/components/AssetTypeForm.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// stores
import { useAssetStore } from '@/stores/assets';
import { useDesktopModalStore } from '@/stores/desktopModals';

const assetStore = useAssetStore();
const modals = useDesktopModalStore();

// refs
const modalContainer = ref(null);
const typeFormRef = ref(null);
const typeIcon = ref('generic');

// constants
const title = 'Edit asset type';

// computed
// Returns the initial icon from the selected asset type.
const initialIcon = computed(() => {
  return assetStore.selectedAssetType?.icon || 'generic';
});

// Returns the initial name from the selected asset type.
const initialName = computed(() => {
  return assetStore.selectedAssetType?.name || '';
});

// Returns the type ID from the selected asset type.
const typeId = computed(() => {
  return assetStore.selectedAssetType?.id || '';
});

// methods
// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
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


