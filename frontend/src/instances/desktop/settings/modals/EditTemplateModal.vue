<template>

  <div class="modal-container" v-esc="escape" @keydown.enter="handleEnterKey">
    <HeaderArea :title="title" :icon="'file'" :showSearch="showSearch" />

    <div class="general-container">
      <div class="input-section">
        <input v-model="templateName" class="input-short" type="text" placeholder="Template Name" v-focus />
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
        <button class="button default" @click="closeModal()" v-stop-propagation>Cancel</button>
        <button class="button colored" @click="editTemplate()" v-stop-propagation>Confirm</button>
      </div>


    </div>
  </div>
</template>

<script setup>
// imports
import { onMounted, ref } from 'vue';

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// services
import { DialogService, TemplateService } from "@/services";

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useTemplateStore } from '@/stores/template';
import { useTrayStates } from '@/stores/TrayStates';

const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const templateStore = useTemplateStore();
const trayStates = useTrayStates();

// refs
const fileIsSelected = ref(false);
const templateFullName = ref('');
const templateName = ref('');
const templatePath = ref("");

// constants
const showSearch = false;
const title = 'Edit Template';

// methods
// Clears the selected file.
const clearSelectedFile = () => {
  fileIsSelected.value = false;
  templatePath.value = "";
  templateFullName.value = "";
};

// Closes the modal.
const closeModal = () => {
  modals.setModalVisibility("editTemplateModal", false);
};

// Edits the template with updated values.
const editTemplate = async () => {
  let selectedTemplate = templateStore.selectedTemplate;
  if (!templateName.value) {
    notificationStore.addNotification('Error editing template', 'Invalid template name', "error");
    return;
  }

  try {
    if (templateName.value !== selectedTemplate.name) {
      await TemplateService.RenameTemplate(projectStore.activeProject.uri, selectedTemplate.name, templateName.value);
    }

    if (fileIsSelected.value) {
      await TemplateService.ChangeTemplateFile(projectStore.activeProject.uri, selectedTemplate.name, templatePath.value);
    }
  } catch (error) {
    notificationStore.errorNotification('Error editing template', error);
    return;
  }
  trayStates.refreshData();
  closeModal();
};

// Closes the modal on escape key.
const escape = () => {
  modals.setModalVisibility('editTemplateModal', false);
};

// Handles enter key press to trigger edit.
const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    editTemplate();
  }
};

// Opens dialog to select a new file.
const selectFile = async () => {
  if (!trayStates.userPin) {
    await trayStates.togglePin();
  }
  let selectedTemplate = templateStore.selectedTemplate;
  let extension = selectedTemplate.extension.replace('.', '');
  const result = await DialogService.SelectFileDialog("Select Image File", "*." + extension);
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

// lifecycle
onMounted(() => {
  if (templateStore.selectedTemplate) {
    templateName.value = templateStore.selectedTemplate.name;
  }
});
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

