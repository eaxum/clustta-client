<template>

  <div class="modal-container" ref="modalContainer" v-stop-propagation v-esc="closeModal">
    <!-- <HeaderArea :title="title" :icon="'folder'" :showSearch="showSearch" /> -->

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
        <GeneralButton :label="$t('common.close')" :fullWidth="true" :buttonFunction="closeModal" :isActive="!isAwaitingResponse" :colored="false" />
        <GeneralButton :label="$t('common.update')" :fullWidth="true" @click="updateCollectionType" :isActive="isValueChanged"
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
import { CollectionService } from "@/services";

// stores
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectTemplateStore = useProjectTemplateStore();
const { t } = useI18n();

import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectTemplateStore } from '@/stores/project_template';

// constants
const title = 'Edit Collection type';

// refs
const displayIconSelector = ref(true);
const collectionTypeIcon = ref('');
const collectionTypeName = ref('');
const isAwaitingResponse = ref(false);

// computed
const icons = computed(() => {
  const allIcons = iconData.icons;
  const allCollectionTypeIcons = projectTemplateStore.collectionTypes.map((item) => item.icon);
  return allIcons.filter((icon) => !allCollectionTypeIcons.includes(icon));
});

const isValueChanged = computed(() => {
  return !!collectionTypeName.value && collectionTypeIcon.value !== 'generic';
});

// methods

// Closes the modal.
const closeModal = () => {
  modals.setModalVisibility("editUserCollectionTypeModal", false);
};

// Handles enter key press.
const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    // updateCollectionType();
  }
};

// Sets the selected icon.
const setIcon = (icon) => {
  collectionTypeIcon.value = icon;
};

// Updates the collection type with the new values.
const updateCollectionType = () => {
  const typeId = projectTemplateStore.selectedCollectionTypeId;
  CollectionService.UpdateCollectionType(projectTemplateStore.activeProjectTemplate.uri, typeId, collectionTypeName.value, collectionTypeIcon.value)
    .then(() => {
      notificationStore.addNotification(t('notifications.collectionTypeUpdated'), "", "success");
      projectTemplateStore.reloadProjectTemplate();
      closeModal();
    })
    .catch((error) => {
      notificationStore.errorNotification(t('notifications.errorUpdatingCollectionType'), error);
    });
};

// lifecycle
onMounted(() => {
  const selectedType = projectTemplateStore.collectionTypes.find(t => t.id === projectTemplateStore.selectedCollectionTypeId);
  if (selectedType) {
    collectionTypeName.value = selectedType.name;
    collectionTypeIcon.value = selectedType.icon;
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.input-short {
  flex: 1;
  width: 100%;
}
</style>