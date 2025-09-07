<template>

  <div ref="modalContainer" class="modal-container" v-esc="closeModal" v-stop-propagation>
    <HeaderArea :title="title" :icon="icon" />

    <div class="general-container">
      <div class="input-section">
        <input v-model="taskName" class="input-short" type="text" placeholder="Task Name" v-focus />
      </div>

      <div v-if="task.is_link" class="input-section">
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
        <div v-if="!userStore.canDo('update_task')" class="input-label">Tags</div>

        <SearchSuggestions v-if="userStore.canDo('update_task')" :placeholder="placeholder" :tags="tags"
          :projectTags="projectTags" :showTags="true" :forSearch="false" @tagAdded="addTag" @tagRemoved="removeTag" />

        <TagContainer v-else :tags="tags" />
      </div>
      <div class="pop-up-actions">
        <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Confirm'" :fullWidth="true" @click="updateTask()" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>


    </div>
  </div>
</template>

<script setup>

// imports
import { ref, watchEffect, computed, onMounted } from 'vue';
import { isValidWeblink } from '@/lib/pointer';
import utils from '@/services/utils';
import emitter from '@/lib/mitt';

// services
import { TaskService, ClipboardService } from "@/../bindings/clustta/services";

// state imports
import { useUserStore } from '@/stores/users';
import { useTrayStates } from '@/stores/TrayStates';
import { useAssetStore } from '@/stores/assets';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useProjectStore } from '@/stores/projects';
import { useMenu } from '@/stores/menu';
import { useIconStore } from '@/stores/icons';

// components
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import SearchSuggestions from '@/instances/common/components/SearchSuggestions.vue';
import TagContainer from '@/instances/common/components/TagContainer.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';

// stores
const menu = useMenu();
const trayStates = useTrayStates();
const assetStore = useAssetStore();
const projectStore = useProjectStore();
const userStore = useUserStore();
const modals = useDesktopModalStore();
const iconStore = useIconStore();

// vars
let placeholder = 'Add Tags'

// refs
const modalContainer = ref(null);
const tags = ref([]);
const oldTags = ref([]);
const taskName = ref('');
const taskWebLink = ref('');
const taskType = ref('');
const taskTypeId = ref('');
const oldTaskName = ref('');
const oldTaskWebLink = ref('');
const isAwaitingResponse = ref(false);
const isResource = ref(false);

// computed properties
const projectTags = computed(() => {
  const allTags = assetStore.projectTags;
  return allTags.filter(item => !tags.value.includes(item));
});

const task = computed(() => {
  return assetStore.selectedTask;
});

const title = computed(() => {
  return task.value.is_link ? 'Edit link' : 'Edit task';
})

const icon = computed(() => {
  return task.value.is_link ? 'website-link' : task.value.icon;
})

const taskTypeNames = computed(() => {
  return assetStore.getTaskTypesNames;
});

const itemType = ref('');

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

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon
};

const itemTypes = computed(() => {
  const allItemTypes = ['Task', 'Resource'];
  return allItemTypes.filter((item) => item !== itemType.value?.toLowerCase());
});

const isValueChanged = computed(() => {
  const task = assetStore.selectedTask;
  if (!task) {
    return false
  }
  const restrictedEntries = [oldTaskName.value, ''];
  const isNameChanged = !restrictedEntries.includes(taskName.value);
  const isPointerChanged = isValidWeblink(taskWebLink.value) && (taskWebLink.value !== oldTaskWebLink.value) && !!taskWebLink.value.length;
  const isTypeChanged = task.is_resource !== isResource.value;
  const isTaskTypeChanged = task.task_type_id !== taskTypeId.value;
  const isTagsUpdated = tags.value.length === oldTags.value.length &&
    tags.value.every(tag => oldTags.value.includes(tag));
  return isNameChanged || isTypeChanged || isTaskTypeChanged || !isTagsUpdated || isPointerChanged
});

// methods
const selectTaskType = (taskTypeName) => {

  let newTaskType;
  const taskTypes = assetStore.getTaskTypes;
  newTaskType = taskTypes.find((item) => item.name === taskTypeName);

  taskType.value = taskTypeName;
  taskTypeId.value = newTaskType.id;

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

const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    if (trayStates.tagSearchQuery === '') {
      updateTask();
    }
  }
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
  console.log(tags.value)
};

const closeModal = (all) => {
  console.log('loud');
  modals.disableAllModals();
};

const updateTask = async () => {
  isResource.value;
  taskType.value;
  taskTypeId.value;

  isAwaitingResponse.value = true;
  let taskId = assetStore.selectedTask.id;
  let task = assetStore.selectedTask;
  let data = {};
  let newTaskTags = tags.value;

  let newTaskType;
  const taskTypes = assetStore.getTaskTypes;
  newTaskType = taskTypes.find((item) => item.id === taskTypeId.value);

  if (taskName.value === "") {
    notificationStore.addNotification("Task name cant be empty", "Task name cant be empty", "error")
    return
  }
  if (taskName.value !== task.name) {
    data.name = taskName.value;
  }
  if (newTaskTags !== task.tags) {
    data.tags = newTaskTags;
  }
  if (Object.keys(data).length === 0) {
    closeModal(true);
    return;
  }

  await TaskService.UpdateTask(projectStore.activeProject.uri, taskId, taskName.value, taskTypeId.value, isResource.value, taskWebLink.value, newTaskTags)
    .then((data) => {
      task.name = taskName.value;
      task.pointer = taskWebLink.value;
      task.is_resource = isResource.value;
      task.tags = newTaskTags;
      task.task_type_name = newTaskType.name;
      task.task_type_icon = newTaskType.icon;
      task.task_type_id = newTaskType.id;
      emitter.emit('refresh-browser');
      isAwaitingResponse.value = false;
    })
    .catch((error) => {
      isAwaitingResponse.value = false;
      console.error('Error:', error);
    });
  closeModal();
}

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

// onMounted
onMounted(() => {
  trayStates.tagSearchQuery = '';
  let task = assetStore.selectedTask;
  console.log(task)
  taskName.value = task.name;
  taskWebLink.value = task.pointer;
  itemType.value = !task.is_resource ? 'Task' : 'Resource';
  taskType.value = task.task_type_name;
  taskTypeId.value = task.task_type_id;
  oldTaskName.value = task.name;
  oldTaskWebLink.value = task.pointer;
  tags.value = Array.from(task.tags);
  oldTags.value = Array.from(task.tags);
});


</script>


<style scoped>
@import "@/assets/desktop.css";

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
</style>


