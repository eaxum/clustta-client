<template>
  <div ref="modalContainer" class="modal-container">

    <div class="general-pane-header">
      <HeaderArea :title="title" :icon="getAppIcon('edit')" :showSearch="false" />
    </div>

    <div class="general-container">

      <div class="input-section">
        <div class="horizontal-flex">
          <input v-model="projectTemplateName" class="input-short" type="text" placeholder="Template Name"
            @keydown.enter="handleEnterKey" v-focus />
        </div>
        <div v-if="!projectTemplateIsCreated && projectTemplateNameInUse" class="horizontal-flex input-alert">
          A template with this name already exists.
        </div>
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Update'" :fullWidth="true" @click="renameProjectTemplate" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { ref, onMounted, computed, watchEffect } from 'vue';

//stores
import { useMenu } from '@/stores/menu';
import { useIconStore } from '@/stores/icons';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useSettingsStore } from '@/stores/settings';

//components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import { useProjectTemplateStore } from '@/stores/project_template';

import { ProjectService, FSService } from "@/../bindings/clustta/services";
import { useNotificationStore } from '@/stores/notifications';

//header vars
let title = 'Edit project template';

// stores/states
const modals = useDesktopModalStore();
const menu = useMenu();
const settings = useSettingsStore();
const iconStore = useIconStore();
const projectTemplateStore = useProjectTemplateStore();
const notificationStore = useNotificationStore();

//refs
const projectTemplateName = ref(projectTemplateStore.activeProjectTemplate.name);
const isAwaitingResponse = ref(false);
const projectTemplateIsCreated = ref(false);
const modalContainer = ref(null);

const restrictedNames = computed(() => {
  const allProjectTemplateNames = projectTemplateStore.projectTemplateNames;
  let restrictedNames = [];
  for (let i = 0; i < allProjectTemplateNames?.length; i++) {
    restrictedNames.push(allProjectTemplateNames[i].toLowerCase())
  }
  return restrictedNames;
});

const projectTemplateNameInUse = computed(() => {
  return restrictedNames.value.includes(projectTemplateName.value.toLowerCase());
});

const projectTemplateNameEmpty = computed(() => {
  return projectTemplateName.value === ''
});

const isValueChanged = computed(() => {
  return !projectTemplateNameEmpty.value && !projectTemplateNameInUse.value
});

// methods
const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const closeModal = () => {
  modals.disableAllModals();
};

const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    renameProjectTemplate();
  }
};

const renameProjectTemplate = async () => {
  isAwaitingResponse.value = true
  let projectTemplatesFolder = await FSService.UserProjectTemplatesPath()

  let oldName = projectTemplateStore.activeProjectTemplate.name + ".clst"
  let newName = projectTemplateName.value + ".clst"

  let oldTemplatePath = await FSService.JoinPath(projectTemplatesFolder, oldName)
  let newTemplatePath = await FSService.JoinPath(projectTemplatesFolder, newName)
  await FSService.Rename(oldTemplatePath, newTemplatePath).then(async (project) => {
    await projectTemplateStore.loadProjectTemplates()
    projectTemplateStore.selectActiveProjectTemplate(projectTemplateName.value)
    closeModal();
    isAwaitingResponse.value = false;

  }).catch((error) => {
    isAwaitingResponse.value = false
    console.log(error)
    notificationStore.errorNotification('Error creating project', error);
  });
  isAwaitingResponse.value = false


};

watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

</script>

<style scoped>
@import "@/assets/desktop.css";
@import "@/assets/modals.css";

.general-container {
  gap: 1rem;
}

.modal-info {
  display: flex;
  flex-direction: column;
  max-width: 100%;
  justify-content: flex-start;
  align-self: stretch;
  width: 464px;
  align-items: flex-start;
  box-sizing: border-box;

}

.modal-text-container {
  display: flex;
  flex-direction: column;
  max-width: 100%;
  justify-content: flex-start;
  align-self: stretch;
  width: 464px;
  align-items: flex-start;
  /* margin-top: 20px; */
}

.modal-title {
  max-width: 100%;
  align-self: stretch;
  width: 464px;
  color: rgba(16, 24, 40, 1);
  color: white;
  font-size: 18px;
  line-height: 28px;
  letter-spacing: 0%;
  text-align: left;
}

.input-header {
  /* background-color: lightblue; */
  width: 100%;
  display: flex;
  align-items: center;
  margin: 10px 0px;
}

.input-count {
  background-color: none;
  font-size: 14px;
  color: white;
}

.modal-subtitle {
  /* background-color: beige; */
  /* max-width: 100%; */
  align-self: stretch;
  width: 464px;
  color: rgba(16, 24, 40, 1);
  color: white;
  font-size: 14px;
  /* line-height: 28px; */
  letter-spacing: 0%;
  text-align: left;
}



.modal-body {
  box-sizing: border-box;
  max-width: 100%;
  align-self: stretch;
  width: 464px;
  margin: 8px 0px;
  font-size: 14px;
  color: rgba(16, 24, 40, 1);
  line-height: 20px;
  letter-spacing: 0%;
  text-align: left;
}

.modal-actions {
  box-sizing: border-box;
  padding: 1rem 2rem;
  gap: 2rem;
  display: flex;
  flex-direction: row;
  max-width: 100%;
  align-self: stretch;
  align-items: center;
  justify-content: space-evenly;
  width: 464px;
  margin-top: 32px;
}

.div-10 {
  display: flex;
}

.task-options-container {
  position: relative;
  box-sizing: border-box;
  width: 100%;
  height: max-content;
  height: 40px;
  transition: all .2s ease-in-out;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  margin: 0;
}

.task-options-container-closed {
  height: 0px;
  padding: 0;
  margin-bottom: -1rem;
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

.pop-up-prompt {
  gap: 10px;
  align-items: center;
  justify-content: center;
}
</style>
