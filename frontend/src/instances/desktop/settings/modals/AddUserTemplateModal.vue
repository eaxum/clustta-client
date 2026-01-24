<template>

  <div class="modal-container" ref="modalContainer" v-stop-propagation v-esc="closeModal">
    <HeaderArea :title="title" :icon="'file-plus'" :showSearch="showSearch" />

    <div class="general-container">
      <div class="input-section">
        <input v-model="templateName" class="input-short" type="text" placeholder="Template Name" v-focus
          @keydown.enter="handleEnterKey" />
      </div>

      <div v-if="!fileIsSelected" class="category-item">
        <span @click="selectFile()" class="single-action-button"><img class="small-icons" src="/icons/add.svg">Select
          a file</span>
      </div>

      <div v-else class="category-item">
        <label class="category-name">{{ templateFullName }}</label>
        <div class="category-item-actions">
          <span class="single-action-button" @click="clearSelectedFile()"><img class="small-icons"
              src="/icons/delete.svg"></span>
        </div>
      </div>


      <div class="pop-up-actions">
        <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Create'" :fullWidth="true" @click="createTemplate" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>


    </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref } from 'vue';

// components
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// services
import { DialogService, TemplateService } from "@/services";

// stores
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectTemplateStore = useProjectTemplateStore();
const trayStates = useTrayStates();

import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectTemplateStore } from '@/stores/project_template';
import { useTrayStates } from '@/stores/TrayStates';

// constants
const showSearch = false;
const title = 'Add Template';

// refs
const fileIsSelected = ref(false);
const isAwaitingResponse = ref(false);
const templateFullName = ref('');
const templateName = ref('');
const templatePath = ref("");

// computed
const isValueChanged = computed(() => {
  return templateName.value && fileIsSelected.value;
});

// methods

// Clears the selected file.
const clearSelectedFile = () => {
  fileIsSelected.value = false;
  templatePath.value = "";
  templateFullName.value = "";
};

// Closes the modal.
const closeModal = () => {
  modals.setModalVisibility("addUserTemplateModal", false);
};

// Creates a new template from the selected file.
const createTemplate = () => {
  if (!templateName.value) {
    notificationStore.addNotification('Template Name is Required', 'Template name is required', "error");
    return;
  }
  if (!fileIsSelected.value) {
    notificationStore.addNotification('Template File is Required', 'Template file is required', "error");
    return;
  }

  TemplateService.CreateTemplate(projectTemplateStore.activeProjectTemplate.uri, templateName.value, templatePath.value)
    .then(() => {
      projectTemplateStore.reloadProjectTemplate();
    })
    .catch((error) => {
      notificationStore.errorNotification('Error creating template', error);
    });

  closeModal();
};

// Handles enter key press.
const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    createTemplate();
  }
};

// Opens a file selection dialog and sets the selected file.
const selectFile = async () => {
  if (!trayStates.userPin) {
    await trayStates.togglePin();
  }
  
  const result = await DialogService.SelectFileDialog("Select Image File", "*");
  if (result) {
    let filePath = result.replace(/\\/g, '/');
    let fileName = filePath.split('/').pop();
    templatePath.value = filePath;
    templateFullName.value = fileName;
    if (!templateName.value) {
      templateName.value = fileName.split('.').slice(0, -1).join('.');
    }
    fileIsSelected.value = true;
  }

  if (!trayStates.userPin) {
    await trayStates.togglePin();
  }
};
</script>

<style scoped>
@import "@/assets/desktop.css";

.category-item {
  color: white;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .5rem;
  width: 100%;
  border-bottom: 1px solid rgba(255, 255, 255, 0.096);
  height: max-content;
  padding: .2rem;
}

.category-item-actions {
  display: flex;
}

.input-short {
  flex: 1;
  width: 100%;
}
</style>

