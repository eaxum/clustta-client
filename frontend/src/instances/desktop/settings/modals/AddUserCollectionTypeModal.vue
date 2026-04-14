<template>

  <div class="modal-container" ref="modalContainer" v-stop-propagation v-esc="closeModal">

    <div class="general-pane-header">
      <HeaderArea :title="title" :icon="collectionTypeIcon" />
    </div>

    <div class="general-container">
      <div class="input-section">
        <input v-model="collectionTypeName" class="input-short" type="text" :placeholder="$t('placeholders.collectionTypeName')" v-focus
          @keydown.enter="handleEnterKey" />
      </div>

      <IconGrid v-if="displayIconSelector" @iconSelected="setIcon" :icons="icons" />

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="$t('common.create')" :fullWidth="true" @click="createCollectionType" :isActive="isValueChanged"
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
import { CollectionService } from "@/services";

// stores
const collectionStore = useCollectionStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectTemplateStore = useProjectTemplateStore();
const { t } = useI18n();

import { useCollectionStore } from '@/stores/collections';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectTemplateStore } from '@/stores/project_template';

// constants
const title = 'Add Collection type';

// refs
const displayIconSelector = ref(true);
const collectionTypeIcon = ref('generic');
const collectionTypeName = ref('');
const isAwaitingResponse = ref(false);

// computed
const icons = computed(() => {
  const allIcons = iconData.icons;
  const allCollectionTypeIcons = collectionStore.getCollectionTypes.map((item) => item.icon);
  return allIcons.filter((icon) => !allCollectionTypeIcons.includes(icon));
});

const isValueChanged = computed(() => {
  return !!collectionTypeName.value && collectionTypeIcon.value !== 'generic';
});

// methods

// Closes the modal.
const closeModal = () => {
  modals.setModalVisibility("addUserCollectionTypeModal", false);
};

// Creates a new collection type.
const createCollectionType = () => {
  CollectionService.CreateCollectionType(projectTemplateStore.activeProjectTemplate.uri, collectionTypeName.value, collectionTypeIcon.value)
    .then(() => {
      notificationStore.addNotification(t('notifications.collectionTypeCreated'), "", "success");
      projectTemplateStore.reloadProjectTemplate();
      closeModal();
    })
    .catch((error) => {
      notificationStore.errorNotification(t('notifications.errorCreatingCollectionType'), error);
    });
};

// Handles enter key press.
const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    // createCollectionType();
  }
};

// Sets the selected icon.
const setIcon = (icon) => {
  collectionTypeIcon.value = icon;
};
</script>

<style scoped>
@import "@/assets/desktop.css";

.input-short {
  flex: 1;
  width: 100%;
}
</style>