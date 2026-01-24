<template>
  <div ref="collectionMenu" class="filter-menu-container">
    <div class="input-section">
      <div class="horizontal-flex">
        <input ref="searchUserInput" v-stop-propagation v-model="searchUserTerm" class="input-short" type="text"
          placeholder="Search User" />
      </div>
    </div>

    <div class="assignee-scroll-container">
      <!-- Current Assignee -->
      <div v-if="assignee && !multipleTasks" class="current-assignee-section">
        <div class="section-label">Assigned</div>
        <div class="assignee-list-container current-assignee">
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
      </div>

      <!-- Project Collaborators -->
      <div v-if="collaboratorsList && collaboratorsList.length" class="assignee-list-container">
        <AssigneeItem 
          v-stop-propagation
          v-for="(collaborator, index) in collaboratorsList" 
          :key="index" 
          :assigneeId="collaborator.id"
          :name="collaborator.name" 
          :userPhoto="collaborator.photo" 
          :avatarColor="collaborator.avatarColor"
          :isLoading="loadingUserIds.includes(collaborator.id)"
          @click="assignTask(collaborator.id)"
        />
      </div>

      <!-- Studio Users Divider -->
      <div v-if="searchUserTerm && filteredStudioUsers.length && collaboratorsList.length" class="studio-users-divider">
        <span class="divider-text">Studio Members</span>
      </div>

      <!-- Studio Users (not in project) -->
      <div v-if="searchUserTerm && filteredStudioUsers.length" class="assignee-list-container">
        <AssigneeItem 
          v-stop-propagation
          v-for="(user, index) in filteredStudioUsers" 
          :key="'studio-' + index" 
          :assigneeId="user.id"
          :name="user.name" 
          :userPhoto="user.photo" 
          :avatarColor="user.avatarColor"
          :isLoading="loadingUserIds.includes(user.id)"
          @click="assignStudioUser(user)"
        />
      </div>

      <!-- No Results -->
      <div v-if="searchUserTerm && !collaboratorsList.length && !filteredStudioUsers.length" class="no-results">
        No results
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import emitter from '@/lib/mitt';
import utils from '@/services/utils';

// components
import AssigneeItem from '@/instances/common/components/AssigneeItem.vue';

// services
import { AssetService, ProjectService } from "@/services";

// stores
import { useAssetStore } from '@/stores/assets';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useStudioStore } from '@/stores/studio';
import { useUserStore } from '@/stores/users';

const assetStore = useAssetStore();
const iconStore = useIconStore();
const menu = useMenu();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const studioStore = useStudioStore();
const userStore = useUserStore();

// refs
const collectionMenu = ref(null);
const loadingUserIds = ref([]);
const searchUserInput = ref(null);
const searchUserTerm = ref('');

// computed properties
// Returns the current assignee data formatted for display.
const assignee = computed(() => {
  if (!task.value || !task.value.assignee_id) return;

  const user = userStore.getUserData(task.value.assignee_id);
  return {
    name: `${user.first_name} ${user.last_name}` || user,
    photo: user.photo || "",
    avatarColor: userStore.userProfileColor(user.id),
    id: user.id,
  };
});

// Returns formatted list of project collaborators, excluding current assignee.
const collaboratorsList = computed(() => {
  const allCollaborators = projectCollaborators.value;
  if (multipleTasks.value) {
    const availableCollaborators = allCollaborators.filter((item) => item.username.toLowerCase().includes(searchUserTerm.value));
    return utils.sortAlphabetically(formatCollaborators(availableCollaborators));
  } else if (!task.value.assignee_id) {
    const availableCollaborators = allCollaborators.filter((item) => item.username.toLowerCase().includes(searchUserTerm.value));
    return utils.sortAlphabetically(formatCollaborators(availableCollaborators));
  }

  const filteredCollaborators = allCollaborators.filter((item) => item.id !== assignee.value.id && item.username.toLowerCase().includes(searchUserTerm.value));
  return utils.sortAlphabetically(formatCollaborators(filteredCollaborators));
});

// Returns studio users who are not in the current project, filtered by search term.
const filteredStudioUsers = computed(() => {
  if (!searchUserTerm.value) return [];
  
  const projectUserIds = projectCollaborators.value.map(user => user.id);
  const studioUsers = studioStore.studioUsers || [];
  const query = searchUserTerm.value.toLowerCase();
  
  const users = studioUsers
    .filter(user => !projectUserIds.includes(user.id))
    .filter(user => {
      const fullName = `${user.first_name} ${user.last_name}`.toLowerCase();
      return fullName.includes(query) || 
             user.email?.toLowerCase().includes(query) ||
             user.username?.toLowerCase().includes(query);
    })
    .map(user => ({
      name: `${user.first_name} ${user.last_name}` || user,
      photo: user.photo || "",
      email: user.email,
      avatarColor: userStore.userProfileColor(user.id),
      id: user.id,
    }));
  
  return utils.sortAlphabetically(users);
});

// Checks if multiple tasks are selected.
const multipleTasks = computed(() => stage.markedItems.length > 1);

// Returns the list of project collaborators.
const projectCollaborators = computed(() => userStore.getProjectCollaborators);

// Returns the currently selected task.
const task = computed(() => assetStore.selectedAsset);

// methods
// Assigns a task to a user, handling single or multiple selection.
const assignTask = (assigneeId) => {
  if (!multipleTasks.value) {
    assignSingleTask(assigneeId);
  } else {
    assignMultipleTasks(assigneeId);
  }
};

// Assigns multiple tasks to a user.
const assignMultipleTasks = async (assigneeId) => {
  let taskIds = stage.markedItems;

  for (const taskId of taskIds) {
    await AssetService.AssignAsset(projectStore.activeProject.uri, taskId, assigneeId)
      .then(async () => {
        emitTaskUpdates(taskId, [{ property: 'assignee_id', value: assigneeId }]);
        menu.disableAllMenus();
      })
      .catch((error) => {
        console.log(error);
        notificationStore.errorNotification("Error Assigning Task", error);
      });
  }
  notificationStore.addNotification("Tasks Assigned Successfully.", "", "success");
};

// Assigns a single task to a user.
const assignSingleTask = async (assigneeId) => {
  let selectedTask = task.value;
  let taskId = selectedTask.id;
  let user = collaboratorsList.value.find((item) => item.id === assigneeId);
  let userId = user ? user.id : "";
  
  await AssetService.AssignAsset(projectStore.activeProject.uri, taskId, userId)
    .then(async () => {
      selectedTask.assignee_id = userId;
      emitTaskUpdates(taskId, [{ property: 'assignee_id', value: userId }]);
      menu.disableAllMenus();
      notificationStore.addNotification("Task Assigned Successfully.", "", "success");
    })
    .catch((error) => {
      console.log(error);
      notificationStore.errorNotification("Error Assigning Task", error);
    });
};

// Adds a studio user to the project and assigns the task to them.
const assignStudioUser = async (user) => {
  if (loadingUserIds.value.includes(user.id)) return;
  
  loadingUserIds.value.push(user.id);
  
  try {
    const roles = userStore.getRolesNames || [];
    const defaultRole = roles.find(role => role.toLowerCase() === 'artist') || roles[0];
    
    if (!defaultRole) {
      notificationStore.errorNotification("Error", "No roles available");
      return;
    }
    
    await ProjectService.AddUser(projectStore.activeProject.uri, user.email, defaultRole);
    await userStore.reloadUsers();
    
    if (!multipleTasks.value) {
      await assignSingleTask(user.id);
    } else {
      await assignMultipleTasks(user.id);
    }
    
    notificationStore.addNotification("User added to project and task assigned.", "", "success");
  } catch (error) {
    console.error(error);
    notificationStore.errorNotification("Error adding user to project", error);
  } finally {
    loadingUserIds.value = loadingUserIds.value.filter(id => id !== user.id);
  }
};

// Emits task update events to Browser and VirtualItem components.
const emitTaskUpdates = (taskId, updates) => {
  const updateData = { itemId: taskId, updates };
  emitter.emit('update-root-data', updateData);
  emitter.emit('update-children', updateData);
};

// Formats collaborator data for display.
const formatCollaborators = (arr) => {
  return arr.map((user, index) => ({
    name: `${user.first_name} ${user.last_name}` || user,
    photo: user.photo || "",
    avatarColor: userStore.userProfileColor(user.id),
    id: user.id,
    index: index.toString(),
  }));
};

// Returns the icon path for a given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Unassigns a task, handling single or multiple selection.
const unassignTask = () => {
  if (!multipleTasks.value) {
    unassignSingleTask();
  } else {
    unassignMultipleTasks();
  }
};

// Unassigns multiple tasks.
const unassignMultipleTasks = async () => {
  let taskIds = stage.markedItems;

  for (const taskId of taskIds) {
    await AssetService.UnassignAsset(projectStore.activeProject.uri, taskId)
      .then(async () => {
        let task = assetStore.findAsset(taskId);
        task.assignee_id = null;
        emitTaskUpdates(taskId, [{ property: 'assignee_id', value: null }]);
        menu.disableAllMenus();
      })
      .catch((error) => {
        console.log(error);
        notificationStore.errorNotification("Error Assigning Task", error);
      });
  }
  notificationStore.addNotification("Tasks Unassigned Successfully.", "", "success");
};

// Unassigns a single task.
const unassignSingleTask = async () => {
  let selectedTask = task.value;
  let taskId = selectedTask.id;
  
  await AssetService.UnassignAsset(projectStore.activeProject.uri, taskId)
    .then(async () => {
      selectedTask.assignee_id = null;
      emitTaskUpdates(taskId, [{ property: 'assignee_id', value: null }]);
      notificationStore.addNotification("Task Unassigned Successfully.", "", "success");
      menu.disableAllMenus();
    })
    .catch((error) => {
      console.log(error);
      notificationStore.errorNotification("Error Unassigning Task", error);
    });
};

// lifecycle hooks
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

.assignee-scroll-container {
  flex-direction: column;
  gap: .3rem;
  max-height: 50vh;
  overflow: hidden;
  overflow-y: auto;
  width: 100%;
}

.assignee-scroll-container::-webkit-scrollbar {
  width: 4px;
}

.assignee-scroll-container::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--light-steel);
}

.assignee-scroll-container::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.assignee-list-container {
  box-sizing: border-box;
  align-items: center;
  flex-direction: column;
  gap: .2rem;
  overflow: hidden;
  width: 100%;
  border-radius: 10px;
}

.current-assignee {
  overflow: hidden;
  min-height: min-content;
}

.current-assignee-section {
  display: flex;
  flex-direction: column;
  gap: .3rem;
  padding-bottom: .4rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  margin-bottom: .2rem;
}

.section-label {
  font-size: 11px;
  color: var(--white);
  opacity: 0.5;
  padding-left: .2rem;
}

.no-results {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 1rem;
  font-size: 12px;
  color: var(--white);
  opacity: 0.5;
}

.studio-users-divider {
  display: flex;
  align-items: center;
  width: 100%;
  padding: .3rem 0;
  gap: .5rem;
}

.studio-users-divider::before,
.studio-users-divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background-color: var(--white);
  opacity: 0.2;
}

.divider-text {
  font-size: 11px;
  color: var(--white);
  opacity: 0.5;
  white-space: nowrap;
}
</style>




