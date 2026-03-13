<template>

  <div class="modal-container" ref="modalContainer" v-stop-propagation v-esc="closeModal">

    <div class="general-pane-header">
      <HeaderArea :title="title" :icon="assetTypeIcon" />
    </div>


    <div class="general-container">
      <div class="input-section">
        <input v-model="assetTypeName" class="input-short" type="text" :placeholder="$t('placeholders.assetTypeName')" v-focus
          @keydown.enter="handleEnterKey" />
      </div>

      <IconGrid v-if="displayIconSelector" @iconSelected="setIcon" :icons="icons" />
      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="$t('common.update')" :fullWidth="true" @click="updateAssetType" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>

    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import iconData from "@/data/iconData.json";

// components
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import IconGrid from '@/instances/desktop/components/IconGrid.vue';

// services
import { AssetService } from "@/services";

// stores
const assetStore = useAssetStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();

import { useAssetStore } from '@/stores/assets';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';

// constants
const title = 'Edit asset type';

// refs
const displayIconSelector = ref(true);
const isAwaitingResponse = ref(false);
const assetTypeIcon = ref('generic');
const assetTypeName = ref('');

// computed
const icons = computed(() => {
  const allIcons = iconData.icons;
  const allAssetTypeIcons = assetStore.assetTypes.map((item) => item.icon);
  return allIcons.filter((icon) => !allAssetTypeIcons.includes(icon));
});

const isValueChanged = computed(() => {
  return !!assetTypeName.value && assetTypeIcon.value !== 'generic';
});

// methods

// Closes the modal.
const closeModal = () => {
  modals.setModalVisibility("editUserAssetTypeModal", false);
};

// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles enter key press.
const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    // updateAssetType();
  }
};

// Sets the selected icon.
const setIcon = (icon) => {
  assetTypeIcon.value = icon;
};

// Updates the asset type with the new values.
const updateAssetType = () => {
  AssetService.UpdateAssetType(projectStore.activeProject.uri, assetStore.selectedAssetType.id, assetTypeName.value, assetTypeIcon.value)
    .then((response) => {
      notificationStore.addNotification(t('notifications.assetTypeUpdated'), "", "success");
      const index = assetStore.assetTypes.findIndex(assetType => assetType.id === assetStore.selectedAssetType.id);
      assetStore.assetTypes[index] = response;
      closeModal();
    })
    .catch((error) => {
      notificationStore.errorNotification(t('notifications.errorUpdatingAssetType'), error);
    });
};

// lifecycle
onMounted(() => {
  // assetTypeName.value = assetStore.selectedAssetType.name;
  // assetTypeIcon.value = assetStore.selectedAssetType.icon;
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.input-short {
  flex: 1;
  width: 100%;
}
</style>


