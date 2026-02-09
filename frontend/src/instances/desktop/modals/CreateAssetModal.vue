<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>
    <HeaderArea :title="title" :icon="icon" :customIcon="customIcon" :showSearch="showSearch" :showPin="true" />
    <div class="general-container">

      <!-- Asset Creation Context -->
      <template v-if="!displayTypeCreator">
        <div class="input-section">
          <div class="compound-input-section">
            <input v-model="taskName" class="input-short" type="text" placeholder="Task Name" v-focus @keydown.enter="handleEnterKey" />
            <ActionButton :icon="getAppIcon('switches')" :buttonFunction="toggleOptions" :isActive="exposeParams" v-tooltip="'Show Options'" />
          </div>
        </div>

        <div class="input-section drop-down-box-section">
          <div class="horizontal-flex">
            <div class="dropdown-wrapper">
              <DropDownBox :items="taskTypeNames" :selectedItem="taskType" :onSelect="selectTaskType" />
            </div>
            <span @click="toggleTypeCreator" class="single-action-button" v-tooltip="'Add New Asset Type'">
              <img class="small-icons" :src="getAppIcon('plus-circle')">
            </span>
          </div>
        </div>

        <div class="task-options-container" :class="{ 'task-options-container-closed': showTaskOptions === true }">
          <div class="input-section">
            <Apps />
          </div>
        </div>

        <div class="pop-up-actions">
          <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
          <GeneralButton :label="'Create'" :fullWidth="true" @click="createTask(false)" :isActive="isValueChanged" :loading="isAwaitingResponse" />
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
import { computed, onMounted, onUnmounted, ref, watchEffect } from 'vue';
import emitter from '@/lib/mitt';
import utils from '@/services/utils';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import Apps from '@/instances/common/components/Apps.vue';
import AssetTypeForm from '@/instances/common/components/AssetTypeForm.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// services
import { AssetService, FSService } from '@/services';

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useTemplateStore } from '@/stores/template';
import { useTrayStates } from '@/stores/TrayStates';

const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stageStore = useStageStore();
const templateStore = useTemplateStore();
const trayStates = useTrayStates();

// refs
const displayTypeCreator = ref(false);
const exposeParams = ref(false);
const isAwaitingResponse = ref(false);
const isResource = ref(true);
const modalContainer = ref(null);
const newTypeIcon = ref('generic');
const selectedTemplate = ref('');
const showTaskOptions = ref(true);
const tags = ref([]);
const taskName = ref('');
const taskType = ref(assetStore.getAssetTypesNames[0]);
const typeFormRef = ref(null);

// constants
const showSearch = false;

// computed
// Returns the custom icon path from tray states (used in asset creation context).
const customIcon = computed(() => {
  if (displayTypeCreator.value) {
    return null;
  }
  return trayStates.popUpModalIcon;
});

// Returns the type icon name (used in type creation context).
const icon = computed(() => {
  if (displayTypeCreator.value) {
    return newTypeIcon.value || 'generic';
  }
  return null;
});

// Returns whether the task name is not empty.
const isValueChanged = computed(() => {
  return taskName.value !== '';
});

// Returns the list of asset type names.
const taskTypeNames = computed(() => {
  return assetStore.getAssetTypesNames;
});

// Returns the modal title from tray states or type creator title.
const title = computed(() => {
  if (displayTypeCreator.value) {
    return 'Add Asset Type';
  }
  return trayStates.popUpModalTitle;
});

// methods
// Closes the modal.
const closeModal = () => {
  trayStates.searchTags = false;
  modals.setModalVisibility('createAssetModal', false);
};

// Creates a new task/asset in the project.
const createTask = async (launch = false, comment = 'Asset created') => {
  isAwaitingResponse.value = true;
  const selectedTaskType = assetStore.assetTypes.find(item => item.name === taskType.value);
  const entities = stageStore.markedEntities;
  const template = templateStore.templates.find(template => template.name === templateStore.selectedTemplateName);
  templateStore.lastUsedTemplate = template.name;
  const isNested = commonStore.navigatorMode && !!collectionStore.navigatedCollection;
  if (entities.length <= 1) {
    let entityId = '';
    if (isNested) {
      entityId = collectionStore.navigatedCollection.id;
    } else if (entities.length > 0) {
      entityId = entities[0];
    }
    await AssetService.CreateAsset(
      projectStore.activeProject.uri,
      taskName.value,
      '',
      selectedTaskType.id,
      entityId,
      isResource.value,
      template.id,
      '',
      '',
      false,
      tags.value,
      '',
      comment
    )
      .then(async (data) => {
        const newAsset = data;
        notificationStore.addNotification('Creating ' + taskName.value + '...', '', 'success');
        if (!trayStates.keepModalOpen) {
          closeModal();
        } else {
          taskName.value = '';
          tags.value = [];
        }
        isAwaitingResponse.value = false;
        stageStore.selectedItem = newAsset;
        assetStore.selectedAsset = newAsset;
        stageStore.firstSelectedItemId = newAsset.id;
        stageStore.markedItems = [newAsset.id];
        notificationStore.addNotification('Created ' + taskName.value + ' successfully.', '', 'success');
        emitter.emit('refresh-browser');
        if (launch) {
          FSService.LaunchFile(newAsset.file_path);
        }
      })
      .catch((error) => {
        console.log(error);
        notificationStore.errorNotification('Error creating task', error);
      });
  }
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles enter key press to submit form.
const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    createTask(false);
  }
};

// Handles successful type creation from the form.
const handleTypeCreated = (response) => {
  taskType.value = response.name;
  displayTypeCreator.value = false;
};

// Handles icon change from the type form.
const handleTypeIconChange = (icon) => {
  newTypeIcon.value = icon;
};

// Scrolls the selected app icon into view.
const scrollAppIntoView = () => {
  const selectedIcon = document.querySelector('.apps-flex-item-selected');
  const appsCenter = trayStates.appsContainer.offsetWidth / 2;
  const iconCenter = selectedIcon.offsetWidth / 2;
  const scrollPosition = selectedIcon.offsetLeft - (appsCenter - iconCenter);
  trayStates.appsContainer.scrollTo({
    left: scrollPosition,
    behavior: 'smooth'
  });
};

// Selects a task type from the dropdown.
const selectTaskType = (taskTypeName) => {
  taskType.value = taskTypeName;
  const allTaskTypeNames = taskTypeNames.value;
  const currentTaskName = taskName.value.toLowerCase();
  if (allTaskTypeNames.includes(currentTaskName)) {
    taskName.value = utils.capitalizeStr(taskTypeName);
  }
};

// Toggles the visibility of task options.
const toggleOptions = () => {
  showTaskOptions.value = !showTaskOptions.value;
  exposeParams.value = !exposeParams.value;
  scrollAppIntoView();
};

// Toggles the type creator context.
const toggleTypeCreator = () => {
  displayTypeCreator.value = !displayTypeCreator.value;
  if (!displayTypeCreator.value) {
    newTypeIcon.value = 'generic';
  }
};

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

// lifecycle hooks
onMounted(() => {
  menu.clickOutsideMask = null;
  taskName.value = utils.capitalizeStr(assetStore.getAssetTypesNames[0]);
  trayStates.listItemsBoundary = modalContainer.value;
  trayStates.tagSearchQuery = '';
  trayStates.itemTags = [];
  if (templateStore.lastUsedTemplate) {
    selectedTemplate.value = templateStore.lastUsedTemplate;
  } else {
    selectedTemplate.value = templateStore.templates[0].name;
  }
});

onUnmounted(() => {
  stageStore.markedEntities = [];
});
</script>


<style scoped>
@import "@/assets/desktop.css";

.compound-input-section {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: .4rem;
}

.dropdown-wrapper {
  flex: 1;
  width: 100%;
}

.general-container {
  gap: 20px;
}

.input-short {
  width: 100%;
}

.pop-up-actions {
  padding: 0px;
  margin-top: 0;
}

.task-options-container {
  position: relative;
  box-sizing: border-box;
  width: 100%;
  height: 60px;
  transition: all .2s ease-in-out;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  margin: 0;
}

.task-options-container-closed {
  height: 0px;
  padding: 0;
  margin-bottom: -1.5rem;
}
</style>








