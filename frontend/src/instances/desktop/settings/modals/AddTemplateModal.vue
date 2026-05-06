<template>
  <div class="modal-container" ref="modalContainer" v-stop-propagation v-esc="closeModal">
    <HeaderArea :title="title" :icon="'file-plus'" :showSearch="showSearch" />

    <div class="general-container">
      <div v-if="selectedFiles.length === 0" class="category-item">
        <ActionButton :icon="CiPlusCircle" :label="$t('modals.selectFiles')" :showLabel="true" :buttonFunction="selectFiles" />
      </div>

      <div v-else class="category-area">
        <div class="category-list">
          <div v-for="(file, index) in selectedFiles" :key="index" class="file-item-wrapper">
            <div class="category-item">
              <img v-if="file.icon" :src="file.icon" class="file-icon small-icons no-filter " />
              <input v-model="file.name" class="input-short" type="text" :placeholder="$t('placeholders.templateName')" />
              <span v-if="file.extension" class="extension-badge">{{ file.extension }}</span>
              <div class="category-item-actions">
                <ActionButton :icon="CiTrash" :useDanger="true" :noFilter="true" :buttonFunction="() => removeFile(index)" />
              </div>
            </div>
            <InputAlert 
              :show="duplicateNameIndices.includes(index)" 
              :message="$t('modals.duplicateTemplateName')" 
            />
          </div>
        </div>
        <ActionButton :icon="CiPlusCircle" :label="$t('modals.addMoreFiles')" :showLabel="true" :buttonFunction="selectFiles" />
      </div>


      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="$t('common.create')" :fullWidth="true" @click="createTemplate" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>


    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { CiPlusCircle, CiTrash } from '@clustta/icons-vue';
import { resolveIcon } from '@/lib/icon-map';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import InputAlert from '@/instances/common/components/InputAlert.vue';

// services
import { DialogService, TemplateService } from "@/services";

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useTrayStates } from '@/stores/TrayStates';

const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const trayStates = useTrayStates();

const { t } = useI18n();

// refs
const isAwaitingResponse = ref(false);
const modalContainer = ref(null);
const selectedFiles = ref([]);

// constants
const showSearch = false;
const title = t('modals.addTemplates');

// computed
// Returns indices of files with duplicate names.
const duplicateNameIndices = computed(() => {
  const names = selectedFiles.value.map(file => file.name.trim().toLowerCase());
  const indices = [];
  
  names.forEach((name, index) => {
    if (name && names.indexOf(name) !== index) {
      indices.push(index);
    }
  });
  
  names.forEach((name, index) => {
    if (name && names.lastIndexOf(name) !== index && !indices.includes(index)) {
      indices.push(index);
    }
  });
  
  return indices;
});

// Checks if there are duplicate template names.
const hasDuplicateNames = computed(() => {
  const names = selectedFiles.value.map(file => file.name.trim().toLowerCase());
  return names.some((name, index) => names.indexOf(name) !== index);
});

// Checks if form is valid for submission.
const isValueChanged = computed(() => {
  return selectedFiles.value.length > 0 && 
         selectedFiles.value.every(file => file.name.trim() !== '') &&
         !hasDuplicateNames.value;
});

// methods
// Closes the modal.
const closeModal = () => {
  modals.setModalVisibility("addTemplateModal", false);
};

// Creates templates from selected files.
const createTemplate = async () => {
  if (selectedFiles.value.length === 0) {
    notificationStore.addNotification(t('notifications.noFilesSelected'), t('notifications.selectAtLeastOneFile'), "error");
    return;
  }

  const filesWithoutNames = selectedFiles.value.filter(file => !file.name.trim());
  if (filesWithoutNames.length > 0) {
    notificationStore.addNotification(t('notifications.templateNamesRequired'), t('notifications.allTemplatesMustHaveName'), "error");
    return;
  }

  if (hasDuplicateNames.value) {
    notificationStore.addNotification(t('notifications.duplicateNamesFound'), t('notifications.allTemplatesMustBeUnique'), "error");
    return;
  }

  try {
    stage.operationActive = true;
    isAwaitingResponse.value = true;

    for (const file of selectedFiles.value) {
      await TemplateService.CreateTemplate(projectStore.activeProject.uri, file.name, file.path);
    }
    
    trayStates.refreshData();
    notificationStore.addNotification(t('notifications.templatesCreated'), t('notifications.successfullyCreatedTemplates', { count: selectedFiles.value.length }), "success");
    closeModal();
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorCreatingTemplates'), error);
  } finally {
    stage.operationActive = false;
    isAwaitingResponse.value = false;
  }
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.resolveIcon(iconName);
};

// Processes icons for template files.
const processTemplateFileIcons = async (files) => {
  if (!files || !Array.isArray(files)) {
    return files;
  }

  for (let i = 0; i < files.length; i++) {
    let file = files[i];
    let extension = "";
    
    const extensionMatch = file.fullName.match(/\.([^.]+)$/);
    if (extensionMatch) {
      extension = extensionMatch[1].toLowerCase();
    }
    
    let iconPath = (await iconStore.getIcon(extension)) || "";
    file.icon = iconPath;
    file.extension = extension ? `.${extension}` : '';
  }
  
  return files;
};

// Removes a file from the selected files list.
const removeFile = (index) => {
  selectedFiles.value.splice(index, 1);
};

// Opens dialog to select template files.
const selectFiles = async () => {
  const result = await DialogService.SelectFilesDialog("Select Template Files", "");
  if (result && result.length > 0) {
    const newFiles = [];
    
    result.forEach(filePath => {
      let normalizedPath = filePath.replace(/\\/g, '/');
      let fileName = normalizedPath.split('/').pop();
      let templateName = fileName.split('.').slice(0, -1).join('.');
      
      newFiles.push({
        path: normalizedPath,
        fullName: fileName,
        name: templateName,
        icon: '',
        extension: ''
      });
    });
    
    await processTemplateFileIcons(newFiles);
    selectedFiles.value.push(...newFiles);
  }
};

// lifecycle
onMounted(() => {
  selectFiles();
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.category-area {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  color: var(--white);
  width: 98%;
}

.category-item {
  color: var(--white);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .5rem;
  width: 100%;
  height: max-content;
  padding: .2rem;
}

.category-item-actions {
  display: flex;
}

.category-list {
  box-sizing: border-box;
  display: flex;
  padding: .5rem;
  align-items: center;
  flex-direction: column;
  gap: .2rem;
  background-color: var(--dark-steel);
  height: 290px;
  overflow: hidden;
  overflow-y: scroll;
  width: 100%;
  border-radius: var(--normal-radius);
}

.category-list::-webkit-scrollbar {
  width: 4px;
}

.category-list::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--light-steel);
}

.category-list::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.extension-badge {
  font-family: Inter, sans-serif;
  color: var(--white);
  font-size: 14px;
  font-weight: 400;
  background-color: var(--steel);
  padding: 0.2rem 0.5rem;
  border-radius: 6px;
  white-space: nowrap;
  flex-shrink: 0;
}

.file-icon {
  flex-shrink: 0;
  min-width: 24px;
  max-width: 24px;
  height: 24px;
  object-fit: contain;
}

.file-item-wrapper {
  width: 100%;
  display: flex;
  flex-direction: column;
  border-bottom: 1px solid rgba(255, 255, 255, 0.096);
}

.input-short {
  flex: 1;
  width: 100%;
  font-size: 14px;
}
</style>

