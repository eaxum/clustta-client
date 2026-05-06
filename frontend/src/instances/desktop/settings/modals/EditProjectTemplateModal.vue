<template>
  <div ref="modalContainer" class="modal-container">

      <HeaderArea :title="title" :icon="CiEdit" :showSearch="false" />

    <div class="general-container">

      <div class="input-section">
        <div class="horizontal-flex">
          <input v-model="projectTemplateName" class="input-short" type="text" :placeholder="$t('placeholders.templateName')"
            @keydown.enter="handleEnterKey" v-focus />
        </div>
        <div v-if="!projectTemplateIsCreated && projectTemplateNameInUse" class="horizontal-flex input-alert">
          {{ $t('modals.templateNameExists') }}
        </div>
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="$t('common.update')" :fullWidth="true" @click="renameProjectTemplate" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import { CiEdit } from '@clustta/icons-vue';
import { resolveIcon } from '@/lib/icon-map';

// components
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// services
import { FSService } from "@/services";

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
const title = 'Edit project template';

// refs
const isAwaitingResponse = ref(false);
const modalContainer = ref(null);
const projectTemplateIsCreated = ref(false);
const projectTemplateName = ref(projectTemplateStore.activeProjectTemplate.name);

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

// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.resolveIcon(iconName);
};

// Handles enter key press to rename project.
const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    renameProjectTemplate();
  }
};

// Renames the project template.
const renameProjectTemplate = async () => {
  isAwaitingResponse.value = true;
  let projectTemplatesFolder = await FSService.UserProjectTemplatesPath();

  let oldName = projectTemplateStore.activeProjectTemplate.name + ".clst";
  let newName = projectTemplateName.value + ".clst";

  let oldTemplatePath = await FSService.JoinPath(projectTemplatesFolder, oldName);
  let newTemplatePath = await FSService.JoinPath(projectTemplatesFolder, newName);
  
  await FSService.Rename(oldTemplatePath, newTemplatePath).then(async () => {
    await projectTemplateStore.loadProjectTemplates();
    projectTemplateStore.selectActiveProjectTemplate(projectTemplateName.value);
    closeModal();
    isAwaitingResponse.value = false;
  }).catch((error) => {
    isAwaitingResponse.value = false;
    console.log(error);
    notificationStore.errorNotification(t('notifications.errorRenamingProjectTemplate'), error);
  });
  isAwaitingResponse.value = false;
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


