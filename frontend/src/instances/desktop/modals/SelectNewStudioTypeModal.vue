<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>
    <HeaderArea :title="title" :icon="'stall'" :showSearch="showSearch" />
    <div class="general-container">

      <div class="studio-info-text">
        <p>{{ $t('modals.studioDescription') }}</p>
      </div>

      <div class="studio-types-container">
        <OptionCard v-for="studioType in studioTypes" :key="studioType.type" :icon="getAppIcon(studioType.icon)" :title="$t(studioType.titleKey)" :description="$t(studioType.messageKey)" @select="selectStudioType(studioType.type)" />
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="$t('common.next')" :fullWidth="true" :buttonFunction="handleNext" :isActive="true" />
      </div>

    </div>
  </div>
</template>

<script setup>
// imports
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import OptionCard from '@/instances/common/components/OptionCard.vue';

// stores
const { t } = useI18n();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const templateStore = useTemplateStore();
const trayStates = useTrayStates();

import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useTemplateStore } from '@/stores/template';
import { useTrayStates } from '@/stores/TrayStates';

// constants
const showSearch = false;
const title = t('modals.newStudio');

// studio types data
const studioTypes = ref([
  {
    type: 'clustta-cloud',
    titleKey: 'modals.newClusttaCloudStudio',
    messageKey: 'modals.clusttaCloudDesc',
    icon: 'clustta'
  },
  {
    type: 'self-managed',
    titleKey: 'modals.newSelfManagedStudio',
    messageKey: 'modals.selfManagedDesc',
    icon: 'two-drives'
  }
]);

// refs
const modalContainer = ref(null);
const selectedStudioType = ref('clustta-cloud');
const selectedTemplate = ref('');

// methods

// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles the next button click to open the appropriate config modal.
const handleNext = () => {
  if (selectedStudioType.value === 'self-managed') {
    modals.setModalVisibility('configSelfManagedStudioModal', true);
  } else if (selectedStudioType.value === 'clustta-cloud') {
    modals.setModalVisibility('configClusttaCloudStudioModal', true);
  }
};

// Selects a studio type.
const selectStudioType = (type) => {
  selectedStudioType.value = type;
  handleNext();
};

// lifecycle
onMounted(() => {
  trayStates.tagSearchQuery = '';
  trayStates.itemTags = [];

  if (templateStore.lastUsedTemplate) {
    selectedTemplate.value = templateStore.lastUsedTemplate;
  } else {
    selectedTemplate.value = templateStore.templates[0]?.name;
  }
});
</script>


<style scoped>
@import "@/assets/desktop.css";

.studio-info-text {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  color: var(--text);
  font-size: 14px;
  padding: .5rem 0;
  box-sizing: border-box;
}

.studio-info-text p {
  margin: 0;
}

.studio-types-container {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: .5rem;
  box-sizing: border-box;
  overflow: hidden;
}

.general-container{
  gap: 1rem;
  /* padding: 1rem 0; */
  /* padding-bottom: 1rem; */
}

.pop-up-actions {
  padding: 0px;
  margin-top: 0;
}
</style>

