<template>
  <div class="modal-container" ref="modalContainer" v-stop-propagation v-esc="closeModal">
    <HeaderArea :title="title" :icon="'file-plus'" :showSearch="showSearch" />

    <div class="general-container">
      <div v-if="selectedFiles.length === 0" class="category-item">
        <ActionButton :icon="getAppIcon('plus-circle')" :label="'Select files'" :showLabel="true" :buttonFunction="selectFiles" />
      </div>

      <div v-else class="category-area">
        <div class="category-list">
          <div v-for="(file, index) in selectedFiles" :key="index" class="file-item-wrapper">
            <div class="category-item">
              <img v-if="file.icon" :src="file.icon" class="file-icon small-icons no-filter " />
              <input v-model="file.name" class="input-short" type="text" placeholder="Template Name" />
              <span v-if="file.extension" class="extension-badge">{{ file.extension }}</span>
              <div class="category-item-actions">
                <ActionButton :icon="getAppIcon('trash')" :useDanger="true" :noFilter="true" :buttonFunction="() => removeFile(index)" />
              </div>
            </div>
            <InputAlert 
              :show="duplicateNameIndices.includes(index)" 
              :message="'Duplicate template name'" 
            />
          </div>
        </div>
        <ActionButton :icon="getAppIcon('plus-circle')" :label="'Add more files'" :showLabel="true" :buttonFunction="selectFiles" />
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
import { useStageStore } from '@/stores/stages';
import { useIconStore } from '@/stores/icons';
import { DialogService } from '@/../bindings/clustta/services/index';

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import InputAlert from '@/instances/common/components/InputAlert.vue';

// stores
const modals = useDesktopModalStore();
const projectStore = useProjectStore();
const trayStates = useTrayStates();
const notificationStore = useNotificationStore();
const stage = useStageStore();
const iconStore = useIconStore();

// header
let title = 'Add Templates';
let showSearch = false;

// refs
const selectedFiles = ref([]);
const isAwaitingResponse = ref(false);

const hasDuplicateNames = computed(() => {
  const names = selectedFiles.value.map(file => file.name.trim().toLowerCase());
  return names.some((name, index) => names.indexOf(name) !== index);
});

const duplicateNameIndices = computed(() => {
  const names = selectedFiles.value.map(file => file.name.trim().toLowerCase());
  const indices = [];
  
  names.forEach((name, index) => {
    if (name && names.indexOf(name) !== index) {
      indices.push(index);
    }
  });
  
  // Also add the first occurrence of duplicates
  names.forEach((name, index) => {
    if (name && names.lastIndexOf(name) !== index && !indices.includes(index)) {
      indices.push(index);
    }
  });
  
  return indices;
});

const isValueChanged = computed(() => {
  return selectedFiles.value.length > 0 && 
         selectedFiles.value.every(file => file.name.trim() !== '') &&
         !hasDuplicateNames.value;
});

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

// methods
const closeModal = () => {
  modals.setModalVisibility("addTemplateModal", false);
};

const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    createTemplate();
  }
};

const processTemplateFileIcons = async (files) => {
  if (!files || !Array.isArray(files)) {
    return files;
  }

  for (let i = 0; i < files.length; i++) {
    let file = files[i];
    let extension = "";
    
    // Get extension from file name
    const extensionMatch = file.fullName.match(/\.([^.]+)$/);
    if (extensionMatch) {
      extension = extensionMatch[1].toLowerCase();
    }
    
    // Get icon from icon store
    let iconPath = (await iconStore.getIcon(extension)) || "";
    file.icon = iconPath;
    file.extension = extension ? `.${extension}` : '';
  }
  
  return files;
};

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
        icon: '', // Will be set by processTemplateFileIcons
        extension: '' // Will be set by processTemplateFileIcons
      });
    });
    
    // Process icons for the new files
    await processTemplateFileIcons(newFiles);
    
    // Add to selected files
    selectedFiles.value.push(...newFiles);
  }
};

const removeFile = (index) => {
  selectedFiles.value.splice(index, 1);
};

const createTemplate = async () => {
  if (selectedFiles.value.length === 0) {
    notificationStore.addNotification('No Files Selected', 'Please select at least one template file', "error");
    return;
  }

  // Check if all files have names
  const filesWithoutNames = selectedFiles.value.filter(file => !file.name.trim());
  if (filesWithoutNames.length > 0) {
    notificationStore.addNotification('Template Names Required', 'All templates must have a name', "error");
    return;
  }

  // Check for duplicate names
  if (hasDuplicateNames.value) {
    notificationStore.addNotification('Duplicate Names Found', 'All templates must have unique names', "error");
    return;
  }

  try {
    // Show operation in progress
    stage.operationActive = true;
    isAwaitingResponse.value = true;

    // Create all templates
    for (const file of selectedFiles.value) {
      await TemplateService.CreateTemplate(projectStore.activeProject.uri, file.name, file.path);
    }
    
    trayStates.refreshData();
    notificationStore.addNotification('Templates Created', `Successfully created ${selectedFiles.value.length} template(s)`, "success");
    closeModal();
  } catch (error) {
    notificationStore.errorNotification('Error creating templates', error);
  } finally {
    stage.operationActive = false;
    isAwaitingResponse.value = false;
  }
};

onMounted(() => {
  selectFiles();
})

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
  font-size: 14px;
}

.listbox-short {

  flex: 1;
  width: 130px;
}

.input-label {

  font-family: Inter, sans-serif;
  color: var(--white);
  /* font-size: 16px; */
  white-space: nowrap;
  flex: 1;

}

.category-area {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  color: var(--white);
  width: 98%;
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
}

.file-item-wrapper {
  width: 100%;
  display: flex;
  flex-direction: column;
  border-bottom: 1px solid rgba(255, 255, 255, 0.096);
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

.file-icon {
  flex-shrink: 0;
  min-width: 24px;
  max-width: 24px;
  height: 24px;
  object-fit: contain;
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
</style>

