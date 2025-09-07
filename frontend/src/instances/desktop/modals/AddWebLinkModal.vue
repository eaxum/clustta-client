<template>
  <div ref="modalContainer" class="modal-container" v-stop-propagation>
    <HeaderArea :title="title" :icon="getAppIcon('website-link')" :showSearch="showSearch" :showPin="true" />
    <div class="general-container">

      <div class="input-section">
        <div class="compound-input-section">
          <input v-model="taskName" class="input-short" type="text" placeholder="Task Name" v-focus
            @keydown.enter="handleEnterKey" />
        </div>
      </div>

      <div class="input-section">
        <div class="horizontal-flex">
          <input v-model="taskWebLink" class="input-short" type="text" placeholder="Web link" ref="taskWebLinkInput" />
          <span @click="pasteWebLink" class="single-action-button" v-tooltip="'Paste link'"><img class="small-icons"
              :src="getAppIcon('clipboard')"></span>
        </div>
      </div>

      <div class="input-section drop-down-box-section">
        <DropDownBox :items="itemTypes" :selectedItem="itemType" :onSelect="changeItemType" />
        <DropDownBox :items="taskTypeNames" :selectedItem="taskType" :onSelect="selectTaskType" />
      </div>
      <div class="input-section">
        <SearchSuggestions :placeholder="placeholder" :showTags="true" :tags="tags" :projectTags="projectTags"
          :forSearch="false" @tagAdded="addTag" @tagRemoved="removeTag" />
      </div>
      <div class="pop-up-actions">
        <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Create'" :fullWidth="true" @click="createTask(false)" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { ref, computed, watchEffect, onMounted } from 'vue';
import { isValidWeblink } from '@/lib/pointer';
import utils from '@/services/utils';
import emitter from '@/lib/mitt';

import { ClipboardService, AssetService, FSService } from "@/../bindings/clustta/services";

// state imports
import { useTrayStates } from '@/stores/TrayStates';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useStageStore } from '@/stores/stages';
import { useAssetStore } from '@/stores/assets';
import { useProjectStore } from '@/stores/projects';
import { useIconStore } from '@/stores/icons';

// components
import Apps from '@/instances/common/components/Apps.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import SearchSuggestions from '@/instances/common/components/SearchSuggestions.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';

// vars
let placeholder = 'Add Tags, use commas to confirm'

// states
const iconStore = useIconStore();
const trayStates = useTrayStates();
const assetStore = useAssetStore();
const projectStore = useProjectStore();

// stores
const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();
const stageStore = useStageStore();
const menu = useMenu();

// refs
const tags = ref([]);
const taskName = ref('');
const taskWebLink = ref('');
const taskWebLinkInput = ref(null);
const showSearch = false;
const exposeParams = ref(false);
const modalContainer = ref(null);
const showTaskOptions = ref(true);
const isAwaitingResponse = ref(false);
const isResource = ref(false);
const taskType = ref(assetStore.getAssetTypesNames[0]);

// computed properties
const title = 'Add web link';
const icon = computed(() => trayStates.popUpModalIcon);
const isValueChanged = computed(() => {
  return taskName.value !== '' && isValidWeblink(taskWebLink.value);
});
const projectTags = computed(() => {
  const allTags = assetStore.projectTags;
  return allTags.filter(item => !tags.value.includes(item));
});

const taskTypeNames = computed(() => {
  return assetStore.getAssetTypesNames;
});

const itemType = ref('Task');

const itemTypes = computed(() => {
  const allItemTypes = ['Task', 'Resource'];
  return allItemTypes.filter((item) => item !== itemType.value?.toLowerCase());
});

// methods
const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const pasteWebLink = async () => {
  ClipboardService.ReadText()
    .then(link => {
      if (isValidWeblink(link)) {
        taskWebLink.value = link;
      }
    }).catch(err => {
      console.error("Failed to paste from clipboard:", err);
    });
};
const removeTag = (tag) => {
  tags.value = tags.value.filter(t => t !== tag);
};

const addTag = (tag) => {
  if (tags.value.includes(tag)) {
    return
  }
  else {
    tags.value.push(tag);
  }
};

const selectTaskType = (taskTypeName) => {
  taskType.value = taskTypeName;

  const allTaskTypeNames = taskTypeNames.value;
  const currentTaskName = taskName.value.toLowerCase();

  if (allTaskTypeNames.includes(currentTaskName)) {
    taskName.value = utils.capitalizeStr(taskTypeName);
  }

};

const changeItemType = (newItemTypeName) => {

  const itemTypeName = newItemTypeName.toLowerCase() + 's';
  isResource.value = itemTypeName !== 'tasks';
  itemType.value = newItemTypeName;

};

const toggleOptions = () => {
  showTaskOptions.value = !showTaskOptions.value;
  exposeParams.value = !exposeParams.value;
  scrollAppIntoView();
}

const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    createTask(false);
  }
};

const scrollAppIntoView = () => {
  const selectedIcon = document.querySelector(`.apps-flex-item-selected`);
  const appsCenter = trayStates.appsContainer.offsetWidth / 2;
  const iconCenter = selectedIcon.offsetWidth / 2;
  const scrollPosition = selectedIcon.offsetLeft - (appsCenter - iconCenter);
  trayStates.appsContainer.scrollTo({
    left: scrollPosition,
    behavior: 'smooth'
  });
};

const closeModal = () => {
  trayStates.searchTags = false;
  modals.setModalVisibility("addWebLinkModal", false);
};
const createTask = async (launch = false, comment = "new file") => {
  isAwaitingResponse.value = true;
  let selectedTaskType = assetStore.assetTypes.find(item => item.name === taskType.value);
  let entities = stageStore.markedEntities
  if (entities.length <= 1) {
    let entityId = ""
    if (entities.length > 0) {
      entityId = entities[0]
    }
    await AssetService.CreateAsset(
      projectStore.activeProject.uri,
      taskName.value,
      "",
      selectedTaskType.id,
      entityId,
      isResource.value,
      "",
      "",
      taskWebLink.value,
      true,
      tags.value,
      "",
      comment
    )
      .then(async (data) => {
        let successMessage = 'Creating ' + taskName.value + '...'
        notificationStore.addNotification(successMessage, "", "success");
        if (!trayStates.keepModalOpen) {
          closeModal();
        } else {
          taskName.value = "";
          tags.value = [];
        }
        isAwaitingResponse.value = false;
        successMessage = 'Created ' + taskName.value + ' successfully.'
        notificationStore.addNotification(successMessage, "", "success")
        if (launch) {
          FSService.LaunchFile(data.file_path)
        }
        emitter.emit('refresh-browser');
      })
      .catch((error) => {
        console.log(error)
        notificationStore.errorNotification("Error creating task", error)
      });
  } else {

  }

};

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

// onMounted hook
onMounted(() => {
  menu.clickOutsideMask = null;
  taskName.value = utils.capitalizeStr(assetStore.getAssetTypesNames[0]);
  trayStates.listItemsBoundary = modalContainer.value;
  trayStates.tagSearchQuery = '';
  trayStates.itemTags = [];

});


</script>


<style scoped>
@import "@/assets/desktop.css";

.general-container {
  gap: 1rem;
}

.modal-container {
  background-color: red;
}

.task-options-container {
  position: relative;
  box-sizing: border-box;
  width: 100%;
  height: max-content;
  height: 60px;
  transition: all .2s ease-in-out;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  margin: 0;
  /* background-color: chocolate; */
}

.task-options-container-closed {
  height: 0px;
  padding: 0;
  margin-bottom: -1.5rem;
}

.input-short {
  width: 100%;
}

.listbox-short {
  width: 130px;
}

.input-label {
  font-family: Inter, sans-serif;
  color: white;
  font-size: 16px;
  white-space: nowrap;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 1rem;

}

.compound-input-section {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: .4rem;
}

.pop-up-prompt {
  gap: 10px;
  align-items: center;
  justify-content: space-between;
}


.pop-up-actions {
  padding: 0px;
  margin-top: 0;
}
</style>





