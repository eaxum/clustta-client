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
import { computed, ref, onMounted, onBeforeUnmount } from 'vue';
import utils from '@/services/utils';
import emitter from '@/lib/mitt';

// services
import { AssetService, ProjectService } from "@/services";

// states/store imports
import { useMenu } from '@/stores/menu';
import { useStageStore } from '@/stores/stages';
import { useUserStore } from '@/stores/users';
import { useAssetStore } from '@/stores/assets';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useStudioStore } from '@/stores/studio';

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
const studioStore = useStudioStore();

// refs
const collectionMenu = ref(null);
const loadingUserIds = ref([]);
const searchUserInput = ref(null);
const searchUserTerm = ref('');

// computed properties
const task = computed(() => { return assetStore.selectedAsset });
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

// Returns studio users who are not in the current project, filtered by search term.
const filteredStudioUsers = computed(() => {
  if (!searchUserTerm.value) {
    return [];
  }
  
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
  await AssetService.AssignAsset(projectStore.activeProject.uri, taskId, userId)
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
  await AssetService.UnassignAsset(projectStore.activeProject.uri, taskId)
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
    await AssetService.AssignAsset(projectStore.activeProject.uri, taskId, userId)
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
    await AssetService.UnassignAsset(projectStore.activeProject.uri, taskId)
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

// Adds a studio user to the project and then assigns the task to them.
const assignStudioUser = async (user) => {
  if (loadingUserIds.value.includes(user.id)) return;
  
  loadingUserIds.value.push(user.id);
  
  try {
    // Get the default role (Artist or first available)
    const roles = userStore.getRolesNames || [];
    const defaultRole = roles.find(role => role.toLowerCase() === 'artist') || roles[0];
    
    if (!defaultRole) {
      notificationStore.errorNotification("Error", "No roles available");
      return;
    }
    
    // First, add the user to the project
    await ProjectService.AddUser(projectStore.activeProject.uri, user.email, defaultRole);
    
    // Refresh to get updated user list
    await userStore.reloadUsers();
    console.log(user.id)
    
    // Now assign the task to this user
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

.assignee-scroll-container {
  /* display: flex; */
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




