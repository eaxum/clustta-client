<template>

  <div ref="modalContainer" class="modal-container" v-esc="closeModal" v-stop-propagation>
    <HeaderArea :title="title" :icon="typeIcon" :customIcon="customIcon" />

    <div class="general-container">

      <!-- Asset Edit Context -->
      <template v-if="!displayTypeCreator">
        <div class="input-section">
          <input v-model="taskName" class="input-short" type="text" placeholder="Task Name" v-focus />
        </div>

        <div v-if="task.is_link" class="input-section">
          <div class="horizontal-flex">
            <input v-model="taskWebLink" class="input-short" type="text" placeholder="Web link" ref="taskWebLinkInput" />
            <span @click="pasteWebLink" class="single-action-button" v-tooltip="'Paste link'">
              <img class="small-icons" :src="getAppIcon('clipboard')">
            </span>
          </div>
        </div>

        <div v-if="!task.is_link" class="input-section drop-down-box-section">
          <div class="horizontal-flex">
            <div class="dropdown-wrapper">
              <DropDownBox :items="taskTypeNames" :selectedItem="taskType" :onSelect="selectTaskType" />
            </div>
            <span @click="toggleTypeCreator" class="single-action-button" v-tooltip="'Add New Asset Type'">
              <img class="small-icons" :src="getAppIcon('plus-circle')">
            </span>
          </div>
        </div>

        <div class="pop-up-actions">
          <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
          <GeneralButton :label="'Confirm'" :fullWidth="true" @click="updateTask()" :isActive="isValueChanged" :loading="isAwaitingResponse" />
        </div>
      </template>

      <!-- Asset Type Creation Context -->
      <template v-else>
        <AssetTypeForm ref="typeFormRef" mode="create" @created="handleTypeCreated" @cancel="toggleTypeCreator" @iconChange="handleTypeIconChange" />
      </template>

    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref, watchEffect } from 'vue';
import { isValidWeblink } from '@/lib/pointer';
import emitter from '@/lib/mitt';
import utils from '@/services/utils';

// components
import AssetTypeForm from '@/instances/common/components/AssetTypeForm.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// services
import { AssetService, ClipboardService } from '@/services';

// stores
import { useAssetStore } from '@/stores/assets';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';

const assetStore = useAssetStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const trayStates = useTrayStates();
const userStore = useUserStore();

// refs
const displayTypeCreator = ref(false);
const isAwaitingResponse = ref(false);
const isResource = ref(false);
const modalContainer = ref(null);
const newTypeIcon = ref('generic');
const oldTags = ref([]);
const oldTaskName = ref('');
const oldTaskWebLink = ref('');
const tags = ref([]);
const taskName = ref('');
const taskType = ref('');
const taskTypeId = ref('');
const taskWebLink = ref('');
const typeFormRef = ref(null);

// computed
// Returns the custom icon path from the task (used in edit context).
const customIcon = computed(() => {
  if (displayTypeCreator.value) {
    return null;
  }
  return task.value.icon;
});

// Returns whether form values have changed.
const isValueChanged = computed(() => {
  const currentTask = assetStore.selectedAsset;
  if (!currentTask) {
    return false;
  }
  const restrictedEntries = [oldTaskName.value, ''];
  const isNameChanged = !restrictedEntries.includes(taskName.value);
  const isPointerChanged = isValidWeblink(taskWebLink.value) && (taskWebLink.value !== oldTaskWebLink.value) && !!taskWebLink.value.length;
  const isTaskTypeChanged = currentTask.task_type_id !== taskTypeId.value;
  const isTagsUpdated = tags.value.length === oldTags.value.length &&
    tags.value.every(tag => oldTags.value.includes(tag));
  return isNameChanged || isTaskTypeChanged || !isTagsUpdated || isPointerChanged;
});

// Returns the currently selected task.
const task = computed(() => {
  return assetStore.selectedAsset;
});

// Returns the list of asset type names.
const taskTypeNames = computed(() => {
  return assetStore.getAssetTypesNames;
});

// Returns the modal title based on task type or type creator.
const title = computed(() => {
  if (displayTypeCreator.value) {
    return 'Add Asset Type';
  }
  return task.value.is_link ? 'Edit link' : 'Edit task';
});

// Returns the type icon name (used in type creation context).
const typeIcon = computed(() => {
  if (displayTypeCreator.value) {
    return newTypeIcon.value || 'generic';
  }
  return null;
});

// methods
// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles successful type creation from the form.
const handleTypeCreated = (response) => {
  taskType.value = response.name;
  taskTypeId.value = response.id;
  displayTypeCreator.value = false;
};

// Handles icon change from the type form.
const handleTypeIconChange = (icon) => {
  newTypeIcon.value = icon;
};

// Pastes a web link from clipboard if valid.
const pasteWebLink = async () => {
  ClipboardService.ReadText()
    .then(link => {
      if (isValidWeblink(link)) {
        taskWebLink.value = link;
      }
    })
    .catch(err => {
      console.error('Failed to paste from clipboard:', err);
    });
};

// Selects a task type from the dropdown.
const selectTaskType = (taskTypeName) => {
  const taskTypes = assetStore.getAssetTypes;
  const newTaskType = taskTypes.find((item) => item.name === taskTypeName);
  taskType.value = taskTypeName;
  taskTypeId.value = newTaskType.id;
  const allTaskTypeNames = taskTypeNames.value;
  const currentTaskName = taskName.value.toLowerCase();
  if (allTaskTypeNames.includes(currentTaskName)) {
    taskName.value = utils.capitalizeStr(taskTypeName);
  }
};

// Toggles the type creator context.
const toggleTypeCreator = () => {
  displayTypeCreator.value = !displayTypeCreator.value;
  if (!displayTypeCreator.value) {
    newTypeIcon.value = 'generic';
  }
};

// Updates the task with the new values.
const updateTask = async () => {
  isAwaitingResponse.value = true;
  const taskId = assetStore.selectedAsset.id;
  const currentTask = assetStore.selectedAsset;
  const newTaskTags = tags.value;
  const taskTypes = assetStore.getAssetTypes;
  const newTaskType = taskTypes.find((item) => item.id === taskTypeId.value);
  if (taskName.value === '') {
    notificationStore.addNotification('Task name cant be empty', 'Task name cant be empty', 'error');
    return;
  }
  await AssetService.UpdateAsset(projectStore.activeProject.uri, taskId, taskName.value, taskTypeId.value, isResource.value, taskWebLink.value, newTaskTags)
    .then(() => {
      currentTask.name = taskName.value;
      currentTask.pointer = taskWebLink.value;
      currentTask.is_resource = isResource.value;
      currentTask.tags = newTaskTags;
      currentTask.task_type_name = newTaskType.name;
      currentTask.task_type_icon = newTaskType.icon;
      currentTask.task_type_id = newTaskType.id;
      emitter.emit('refresh-browser');
      isAwaitingResponse.value = false;
    })
    .catch((error) => {
      isAwaitingResponse.value = false;
      console.error('Error:', error);
    });
  closeModal();
};

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

// lifecycle hooks
onMounted(() => {
  trayStates.tagSearchQuery = '';
  const currentTask = assetStore.selectedAsset;
  taskName.value = currentTask.name;
  taskWebLink.value = currentTask.pointer;
  isResource.value = currentTask.is_resource;
  taskType.value = currentTask.task_type_name;
  taskTypeId.value = currentTask.task_type_id;
  oldTaskName.value = currentTask.name;
  oldTaskWebLink.value = currentTask.pointer;
  tags.value = Array.from(currentTask.tags);
  oldTags.value = Array.from(currentTask.tags);
});
</script>


<style scoped>
@import "@/assets/desktop.css";

.dropdown-wrapper {
  flex: 1;
  width: 100%;
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
  color: var(--white);
  font-size: 14px;
  white-space: nowrap;
  flex: 1;
}
</style>




