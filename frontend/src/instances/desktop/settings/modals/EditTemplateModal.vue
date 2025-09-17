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

import { useTrayStates } from '@/stores/TrayStates';
import { ref, onMounted } from 'vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import { TemplateService } from "@/../bindings/clustta/services";
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useTemplateStore } from '@/stores/template';
import { useProjectStore } from '@/stores/projects';
import { DialogService } from '@/../bindings/clustta/services/index';

let title = 'Edit Template';
let showSearch = false;

const fileIsSelected = ref(false);
const templateName = ref('');
const templateFullName = ref('');
const templatePath = ref("");

const trayStates = useTrayStates();
const templateStore = useTemplateStore();
const projectStore = useProjectStore();

const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();

const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    editTemplate();
  }
};

const escape = () => {
  modals.setModalVisibility('editTemplateModal', false);
};

const closeModal = () => {
  modals.setModalVisibility("editTemplateModal", false);
};

const selectFile = async () => {
  if (!trayStates.userPin) {
    await trayStates.togglePin()
  }
  let selectedTemplate = templateStore.selectedTemplate;
  let extension = selectedTemplate.extension.replace('.', '');
  const result = await DialogService.SelectFileDialog("Select Image File", "*." + extension);
  if (result) {
    let filePath = result.replace(/\\/g, '/');
    let fileName = filePath.split('/').pop();
    templatePath.value = filePath;
    templateFullName.value = fileName
    if (!templateName.value) {
      templateName.value = fileName.split('.').slice(0, -1).join('.');
    }
    fileIsSelected.value = true;
  }

  if (!trayStates.userPin) {
    await trayStates.togglePin()
  }
};

const clearSelectedFile = () => {
  fileIsSelected.value = false;
  templatePath.value = "";
  templateFullName.value = "";
};

const editTemplate = async () => {
  let selectedTemplate = templateStore.selectedTemplate;
  if (!templateName.value) {
    notificationStore.addNotification('Error editing template', 'Invalid template name', "error");
    return;
  }

  try {
    if (templateName.value !== selectedTemplate.name) {
      await TemplateService.RenameTemplate(projectStore.activeProject.uri, selectedTemplate.name, templateName.value)
    }

    if (fileIsSelected.value) {
      await TemplateService.ChangeTemplateFile(projectStore.activeProject.uri, selectedTemplate.name, templatePath.value)
    }
  } catch (error) {
    notificationStore.errorNotification('Error editing template', error);
    return;
  }
  trayStates.refreshData();
  closeModal();
};

onMounted(() => {
  if (templateStore.selectedTemplate) {
    templateName.value = templateStore.selectedTemplate.name;
  }
});


</script>

<style scoped>
@import "@/assets/desktop.css";

.add-category {

  display: flex;
  gap: .5rem;
  flex-direction: row;
  /* background-color: chocolate; */
}

.input-short {
  flex: 1;
  width: 100%;
}

.listbox-short {

  flex: 1;
  width: 130px;
}

.input-label {

  font-family: Inter, sans-serif;
  color: white;
  font-size: 16px;
  white-space: nowrap;
  flex: 1;

}

.category-area {
  box-sizing: border-box;
  /* margin-top: 1rem; */
  display: flex;
  /* align-items: center; */
  flex-direction: column;
  /* justify-content: space-between; */
  gap: 1rem;
  /* background-color: darkkhaki; */
  color: white;
  width: 98%;
}

.category-list {
  box-sizing: border-box;
  /* margin-top: 1rem; */
  display: flex;
  padding: .5rem;
  align-items: center;
  flex-direction: column;
  /* justify-content: space-between; */
  gap: .2rem;
  /* background-color: rgb(57, 122, 108); */

  background-color: rgba(0, 0, 0, 0.144);
  height: 290px;
  /* height: 90%; */
  overflow: hidden;
  overflow-y: scroll;
  width: 100%;
  border-radius: 10px;
}

.category-list::-webkit-scrollbar {
  width: 4px;
}

.category-list::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: rgba(255, 255, 255, 0.295);
}

.category-list::-webkit-scrollbar-track {
  border-radius: 10px;
  /* background-color: rgba(0, 0, 0, 0.295); */
}

.category-item {
  color: white;
  /* margin-top: 1rem; */
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .5rem;
  width: 100%;
  border-bottom: 1px solid rgba(255, 255, 255, 0.096);
  height: max-content;
  padding: .2rem;

  /* background-color: greenyellow; */
}

.category-item-actions {

  /* background-color: red; */
  display: flex;

}
</style>

