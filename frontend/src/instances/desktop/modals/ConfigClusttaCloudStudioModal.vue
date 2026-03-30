<template>
  <div ref="modalContainer" class="modal-container">

      <HeaderArea :title="title" :icon="getAppIcon('clustta')" :showSearch="false" />

    <div class="general-container">

      <div class="studio-info-text">
        <p>{{ $t('modals.clusttaCloudDesc') }}</p>
      </div>

      <FormInput
        v-model="studioName"
        :placeholder="$t('placeholders.studioName')"
        :error="studioNameError"
        :loading="checkingStudioNameAvailability"
        :valid="!!studioName && !studioNameError && !checkingStudioNameAvailability"
        :showValidation="!!studioName"
        @input="checkStudioName"
      />

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.back')" :fullWidth="true" :buttonFunction="goBack" :colored="false" />
        <GeneralButton 
          :label="$t('common.create')" 
          :fullWidth="true" 
          @click="createStudio" 
          :isActive="isValueChanged"
          :loading="isAwaitingResponse" 
        />
      </div>
    </div>
    
  </div>
</template>

<script setup>
// imports
import { computed, ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import FormInput from '@/instances/desktop/components/FormInput.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// services
import { StudioService } from '@/services';

// stores
const { t } = useI18n();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();

import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStudioStore } from '@/stores/studio';

// constants
const title = t('modals.newClusttaCloudStudio');

// refs
const checkingStudioNameAvailability = ref(false);
const isAwaitingResponse = ref(false);
const isStudioNameTaken = ref(false);
const modalContainer = ref(null);
const studioName = ref('');
const studioNameError = ref('');

// computed
const isValueChanged = computed(() => {
  return !studioNameEmpty.value && !studioNameInUse.value && !studioNameError.value;
});

const restrictedNames = computed(() => {
  return ['clustta', 'eaxum', 'pixar', 'disney', 'dreamworks'];
});

const studioNameEmpty = computed(() => {
  return studioName.value === '';
});

const studioNameInUse = computed(() => {
  return restrictedNames.value.includes(studioName.value.toLowerCase()) || isStudioNameTaken.value;
});

// methods

// Checks if the studio name is available.
const checkStudioName = async () => {
  if (!studioName.value) {
    studioNameError.value = '';
    isStudioNameTaken.value = false;
    return;
  }
  
  if (restrictedNames.value.includes(studioName.value.toLowerCase())) {
    studioNameError.value = t('notifications.studioNameReserved');
    isStudioNameTaken.value = true;
    return;
  }
  
  checkingStudioNameAvailability.value = true;

  try {
    const nameExists = await StudioService.CheckStudioNameExists(studioName.value.toLowerCase());
    if (nameExists) {
      studioNameError.value = t('notifications.studioNameTaken');
      isStudioNameTaken.value = true;
    } else {
      studioNameError.value = '';
      isStudioNameTaken.value = false;
    }
    checkingStudioNameAvailability.value = false;
  } catch (error) {
    studioNameError.value = '';
    isStudioNameTaken.value = false;
    console.error('Error checking studio name:', error);
    checkingStudioNameAvailability.value = false;
  }
};

// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Creates a new Clustta Cloud studio.
const createStudio = async () => {
  isAwaitingResponse.value = true;
  
  try {
    await StudioService.RegisterStudio(studioName.value, '', 'cloud');

    await projectStore.loadStudios();
    let studio = projectStore.studios.find((item) => item.name === studioName.value);
    if (studio) {
      projectStore.selectedStudio = studio;
    } else {
      projectStore.selectedStudio = projectStore.studios[0];
    }

    const studioStore = useStudioStore();
    await studioStore.getStudioUsers();
    await projectStore.loadProjects();
    closeModal();
  } catch (error) {
    console.error(error);
    notificationStore.errorNotification(t('notifications.errorCreatingStudio'), error);
  } finally {
    isAwaitingResponse.value = false;
  }
};

// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Goes back to the studio type selection modal.
const goBack = () => {
  modals.setModalVisibility('selectNewStudioTypeModal', true);
};

// Handles enter key press.
const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    createStudio();
  }
};

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";
@import "@/assets/modals.css";

.general-container{
  padding-top: 1rem;
}

.studio-info-text {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  color: var(--white);
  font-size: 14px;
  padding: .5rem 0;
  box-sizing: border-box;
}

.studio-info-text p {
  margin: 0;
}
</style>
