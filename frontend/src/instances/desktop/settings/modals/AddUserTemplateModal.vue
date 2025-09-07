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
import { ref, onMounted, computed } from 'vue';

import { useTrayStates } from '@/stores/TrayStates';
import { TemplateService } from "@/../bindings/clustta/services";
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useProjectStore } from '@/stores/projects';
import { DialogService } from '@/../bindings/clustta/services/index';

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';

import { useProjectTemplateStore } from '@/stores/project_template';

// stores
const modals = useDesktopModalStore();
const projectStore = useProjectStore();
const trayStates = useTrayStates();
const notificationStore = useNotificationStore();
const projectTemplateStore = useProjectTemplateStore();

// header
let title = 'Add Template';
let showSearch = false;

// refs
const templateName = ref('');
const templatePath = ref("");
const templateFullName = ref('');
const isAwaitingResponse = ref(false);
const fileIsSelected = ref(false);

const isValueChanged = computed(() => {
  return templateName.value && fileIsSelected.value
});


// methods
const closeModal = () => {
  modals.setModalVisibility("addUserTemplateModal", false);
};

const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    createTemplate();
  }
};

const selectFile = async () => {
  if (!trayStates.userPin) {
    await trayStates.togglePin()
  }
  // TODO this is broken
  const result = await DialogService.SelectFileDialog("Select Image File", "*");
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
    .then((response) => {
      projectTemplateStore.reloadProjectTemplate()
    })
    .catch((error) => {
      notificationStore.errorNotification('Error creating template', error);
    });

  closeModal();
};


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

/* .general-container {
  gap: 10px;
  box-sizing: border-box;
  width: 300px;
  color: #fff;
  display: flex;
  align-items: center;
  flex-direction: column;
  align-items: center;
} */

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

