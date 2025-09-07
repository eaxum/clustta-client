<template>
  <div ref="collectionMenu" class="filter-menu-container">
    <div class="input-section">
      <div class="horizontal-flex">
        <input ref="searchUserInput" v-stop-propagation v-model="searchUserTerm" class="input-short" type="text"
          placeholder="Search User" />
      </div>
    </div>

    <!-- Users -->
    <div v-if="assignee && !multipleTasks" class="assignee-list-container current-assignee">
      <AssigneeItem 
        :name="assignee.name" 
        :assigneeId="assignee.id"
        :photo="assignee.photo" 
        :avatarColor="assignee.avatarColor"
      >
        <template #actions>
          <span v-stop-propagation class="single-action-button" @click="unassignTask()" v-tooltip="'Unassign'">
            <img class="small-icons" :src="getAppIcon('person-minus')">
          </span>
        </template>
      </AssigneeItem>
    </div>
    
    <span v-if="assignee && collaboratorsList.length && !multipleTasks" class="menu-divider"></span>

    <!-- collaborators -->
    <div v-if="collaboratorsList && collaboratorsList.length" class="assignee-list-container">
      <AssigneeItem 
        v-stop-propagation
        v-for="(collaborator, index) in collaboratorsList" 
        :key="index" 
        :assigneeId="collaborator.id"
        :name="collaborator.name" 
        :userPhoto="collaborator.photo" 
        :avatarColor="collaborator.avatarColor"
        @click="assignTask(collaborator.id)"
      />
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref, onMounted, onBeforeUnmount } from 'vue';
import utils from '@/services/utils';
import emitter from '@/lib/mitt';

// services
import { TaskService } from "@/../bindings/clustta/services";

// states/store imports
import { useMenu } from '@/stores/menu';
import { useStageStore } from '@/stores/stages';
import { useUserStore } from '@/stores/users';
import { useAssetStore } from '@/stores/assets';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';

// components
import AssigneeItem from '@/instances/common/components/AssigneeItem.vue'
import { useProjectStore } from '@/stores/projects';

// states/stores
const userStore = useUserStore();
const menu = useMenu();
const stage = useStageStore();
const notificationStore = useNotificationStore();
const assetStore = useAssetStore();
const projectStore = useProjectStore();
const iconStore = useIconStore();

// refs
const collectionMenu = ref(null);
const searchUserInput = ref(null);
const searchUserTerm = ref('');

// computed properties
const task = computed(() => { return assetStore.selectedTask });
const multipleTasks = computed(() => { return stage.markedItems.length > 1 });

const projectCollaborators = computed(() => {
  return userStore.getProjectCollaborators;
});

const assignee = computed(() => {
  if (!task.value) {
    return
  }

  if (!task.value.assignee_id) {
    return
  };

  const user = userStore.getUserData(task.value.assignee_id);
  const userData = {
    name: `${user.first_name} ${user.last_name}` || user,
    photo: user.photo || "",
    avatarColor: userStore.userProfileColor(user.id),
    id: user.id,
  }

  return userData;
});

const collaboratorsList = computed(() => {

  const allCollaborators = projectCollaborators.value;
  if (multipleTasks.value) {
    const availableCollaborators = allCollaborators.filter((item) => item.username.toLowerCase().includes(searchUserTerm.value))
    return utils.sortAlphabetically(formatCollaborators(availableCollaborators))
  } else if (!task.value.assignee_id) {
    const availableCollaborators = allCollaborators.filter((item) => item.username.toLowerCase().includes(searchUserTerm.value))
    return utils.sortAlphabetically(formatCollaborators(availableCollaborators))
  };

  const filteredCollaborators = allCollaborators.filter((item) => item.id !== assignee.value.id && item.username.toLowerCase().includes(searchUserTerm.value));
  const result = formatCollaborators(filteredCollaborators);
  return utils.sortAlphabetically(result);
});

const getAppIcon = (iconName) => {
	const icon = iconStore.getAppIcon(iconName);
	return icon
};

const emitTaskUpdates = (taskId, updates) => {
  const updateData = { itemId: taskId, updates };
  
  // Emit to both Browser and VirtuaItem components
  emitter.emit('update-root-data', updateData);
  emitter.emit('update-children', updateData);
};

// methods
const formatCollaborators = (arr) => {
  return arr.map((user, index) => ({
    name: `${user.first_name} ${user.last_name}` || user,
    photo: user.photo || "",
    avatarColor: userStore.userProfileColor(user.id),
    id: user.id,
    index: index.toString(),
  }));
};

const assignTask = (assigneeId) => {
  if (!multipleTasks.value) {
    assignSingleTask(assigneeId);
  } else {
    assignMultipleTasks(assigneeId)
  }
};

const unassignTask = () => {
  if (!multipleTasks.value) {
    unassignSingleTask();
  } else {
    unassignMultipleTasks()
  }
};

const assignSingleTask = async (assigneeId) => {
  let selectedTask = task.value;
  let taskId = selectedTask.id;
  let user = collaboratorsList.value.find((item) => item.id === assigneeId);
  let userId = user ? user.id : "";
  await TaskService.AssignTask(projectStore.activeProject.uri, taskId, userId)
    .then(async (data) => {
      // Update local task data
      selectedTask.assignee_id = userId;
      
      // Emit updates using helper function
      emitTaskUpdates(taskId, [
        { property: 'assignee_id', value: userId }
      ]);
      
      menu.disableAllMenus();
      notificationStore.addNotification("Task Assigned Successfully.", "", "success");
    })
    .catch((error) => {
      console.log(error)
      notificationStore.errorNotification("Error Assigning Task", error)
    });
};

const unassignSingleTask = async () => {
  let selectedTask = task.value;
  let taskId = selectedTask.id;
  await TaskService.UnassignTask(projectStore.activeProject.uri, taskId)
    .then(async (data) => {
      selectedTask.assignee_id = null;
      
      emitTaskUpdates(taskId, [
        { property: 'assignee_id', value: null }
      ]);
      
      notificationStore.addNotification("Task Unassigned Successfully.", "", "success");
      menu.disableAllMenus();
    })
    .catch((error) => {
      console.log(error)
      notificationStore.errorNotification("Error Unassigning Task", error)
    });
};

const assignMultipleTasks = async (assigneeId) => {
  let taskIds = stage.markedItems;

  for (const taskId of taskIds) {
    let userId = assigneeId;
    await TaskService.AssignTask(projectStore.activeProject.uri, taskId, userId)
      .then(async (data) => {
        
        emitTaskUpdates(taskId, [
          { property: 'assignee_id', value: userId }
        ]);
        
        menu.disableAllMenus();
      })
      .catch((error) => {
        console.log(error)
        notificationStore.errorNotification("Error Assigning Task", error)
      });

  }
  notificationStore.addNotification("Tasks Assigned Successfully.", "", "success");
};

const unassignMultipleTasks = async () => {
  let taskIds = stage.markedItems;

  for (const taskId of taskIds) {
    await TaskService.UnassignTask(projectStore.activeProject.uri, taskId)
      .then(async (data) => {
        // Update local task data
        let task = assetStore.findAsset(taskId);
        task.assignee_id = null;
        
        // Emit updates using helper function
        emitTaskUpdates(taskId, [
          { property: 'assignee_id', value: null }
        ]);
        
        menu.disableAllMenus();
      })
      .catch((error) => {
        console.log(error)
        notificationStore.errorNotification("Error Assigning Task", error)
      });
  }
  notificationStore.addNotification("Tasks Unssigned Successfully.", "", "success");
};

// onMounted hook
onMounted(() => {
  searchUserInput.value.focus();
  menu.assetMenuWidth = collectionMenu.value.getBoundingClientRect().width;
  menu.collectionMenu = collectionMenu.value;
});

onBeforeUnmount(() => {
  menu.assetMenuWidth = collectionMenu.value.getBoundingClientRect().width;
  menu.assetMenuHeight = collectionMenu.value.getBoundingClientRect().height;
});
</script>

<style scoped>
@import "@/assets/desktop.css";
@import "@/assets/menu.css";

.input-section {
  min-height: min-content;
}

.input-short {
  flex: 1;
  width: 100%;
  font-size: 14px;
}

.entity-item-menu-container {
  z-index: 10;
  display: flex;
  top: 0;
  left: 0;
  flex-direction: column;
  color: var(--white);
  align-items: center;
  gap: .3rem;
  padding: .6rem;
  box-sizing: border-box;
  width: max-content;
  width: 250px;
  height: max-content;
  border-radius: 16px;
  outline: var(--transparent-line);
  outline-offset: -1px;
  background-color: var(--light-steel);
}

.entity-item-menu-visible {
  opacity: 1;
  visibility: visible;
}

.assignee-list-container {
  box-sizing: border-box;
  /* padding: .5rem; */
  align-items: center;
  flex-direction: column;
  gap: .2rem;
  overflow: hidden;
  overflow-y: scroll;
  width: 100%;
  border-radius: 10px;
}

.assignee-list-container::-webkit-scrollbar {
  width: 4px;
}

.assignee-list-container::-webkit-scrollbar-thumb {
  border-radius: 8px;
  background-color: var(--midnight-steel);
  background-color: var(--white);
}

.assignee-list-container::-webkit-scrollbar-track {
  margin-top: 5px;
  border-radius: 10px;
}

.current-assignee {
  overflow: hidden;
  min-height: min-content;
}
</style>




