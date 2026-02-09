<template>
  <div ref="modalContainer" class="modal-container">

    <div class="general-pane-header">
      <HeaderArea :title="title" :icon="getAppIcon('briefcase-plus')" :showSearch="false" />
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
        <GeneralButton :label="'Create'" :fullWidth="true" @click="createProject" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref, watchEffect } from 'vue';

// components
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// services
import { FSService, ProjectService } from "@/services";

// stores
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectTemplateStore = useProjectTemplateStore();

import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectTemplateStore } from '@/stores/project_template';

// constants
const title = 'Add project template';

// refs
const isAwaitingResponse = ref(false);
const modalContainer = ref(null);
const projectTemplateIsCreated = ref(false);
const projectTemplateName = ref('');

// computed
const isValueChanged = computed(() => {
  return !projectTemplateNameEmpty.value && !projectTemplateNameInUse.value;
});

const projectTemplateNameEmpty = computed(() => {
  return projectTemplateName.value === '';
});

const projectTemplateNameInUse = computed(() => {
  return restrictedNames.value.includes(projectTemplateName.value.toLowerCase());
});

const restrictedNames = computed(() => {
  const allProjectTemplateNames = projectTemplateStore.projectTemplateNames;
  let names = [];
  for (let i = 0; i < allProjectTemplateNames?.length; i++) {
    names.push(allProjectTemplateNames[i].toLowerCase());
  }
  return names;
});

// methods

// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Creates a new project template.
const createProject = async () => {
  let projectTemplatesFolder = await FSService.UserProjectTemplatesPath();

  isAwaitingResponse.value = true;
  let name = projectTemplateName.value;
  let path = projectTemplatesFolder + '/' + name + ".clst";
  path = path.replace(/\\/g, '/');
  
  ProjectService.CreateProject(path, 'Personal', "", "").then(async () => {
    await projectTemplateStore.loadProjectTemplates();
    closeModal();
    isAwaitingResponse.value = false;
  }).catch((error) => {
    console.log(error);
    notificationStore.errorNotification('Error creating project', error);
  });
};

// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles enter key press to create project.
const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    createProject();
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

.general-container {
  gap: 1rem;
}

.input-short {
  flex: 1;
  width: 100%;
}
</style>


