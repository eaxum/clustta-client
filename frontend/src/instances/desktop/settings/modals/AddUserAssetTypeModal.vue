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
        <GeneralButton :label="$t('common.create')" :fullWidth="true" @click="createAssetType" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>

    </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref } from 'vue';
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
const projectTemplateStore = useProjectTemplateStore();
const { t } = useI18n();

import { useAssetStore } from '@/stores/assets';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectTemplateStore } from '@/stores/project_template';

// constants
const title = 'Add Asset type';

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
  modals.setModalVisibility("addUserAssetTypeModal", false);
};

// Creates a new asset type.
const createAssetType = () => {
  AssetService.CreateAssetType(projectTemplateStore.activeProjectTemplate.uri, assetTypeName.value, assetTypeIcon.value)
    .then(() => {
      notificationStore.addNotification(t('notifications.assetTypeCreated'), "", "success");
      projectTemplateStore.reloadProjectTemplate();
      closeModal();
    })
    .catch((error) => {
      notificationStore.errorNotification(t('notifications.errorCreatingAssetType'), error);
    });
};

// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles enter key press.
const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    // createAssetType();
  }
};

// Sets the selected icon.
const setIcon = (icon) => {
  assetTypeIcon.value = icon;
};
</script>

<style scoped>
@import "@/assets/desktop.css";

.input-short {
  flex: 1;
  width: 100%;
}
</style>


