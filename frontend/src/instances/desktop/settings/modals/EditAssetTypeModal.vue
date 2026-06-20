<template>
  <div class="modal-container" ref="modalContainer" v-stop-propagation v-esc="closeModal">
      <HeaderArea :title="title" :icon="typeIcon" />

    <div class="general-container">
      <AssetTypeForm ref="typeFormRef" mode="edit" :initialName="initialName" :initialIcon="initialIcon" :typeId="typeId" @updated="handleUpdated" @cancel="closeModal" @iconChange="handleIconChange" />
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import AssetTypeForm from '@/instances/common/components/AssetTypeForm.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// stores
import { useAssetStore } from '@/stores/assets';
import { useDesktopModalStore } from '@/stores/desktopModals';

const assetStore = useAssetStore();
const modals = useDesktopModalStore();

const { t } = useI18n();

// refs
const modalContainer = ref(null);
const typeFormRef = ref(null);
const selectedIcon = ref('');

// constants
const title = t('modals.editAssetType');

// computed
// Returns the initial icon from the selected asset type.
const initialIcon = computed(() => {
  return assetStore.selectedAssetType?.icon || 'generic';
});

// Returns the currently displayed icon, using the saved icon until a new one is selected.
const typeIcon = computed(() => {
  return selectedIcon.value || initialIcon.value;
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
  selectedIcon.value = icon;
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

