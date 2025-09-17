<template>
  <div class="settings-component-root">
    <div class="settings-component-container">
      <ActionBar v-if="userStore.canDo('add_user')" :itemType="'Add collaborator'" :addFunction="addCollaborator" />

      <ScrollList v-if="projectCollaborators.length" :items="projectCollaborators" :useAvatar="true" :useItemId="true"
        :useMeta="true" :editItems="true" :editListItem="prepEditCollaborator" :deleteItems="true"
        :deleteListItem="deleteCollaborator" :forCollab="true" />

    </div>
  </div>
</template>

<script setup>

// imports
import { computed } from 'vue';
import { ProjectService } from "@/../bindings/clustta/services";
import utils from '@/services/utils';

// store imports
import { useAssetStore } from '@/stores/assets';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';

// components
import ScrollList from '@/instances/desktop/components/ScrollList.vue';
import ActionBar from '@/instances/desktop/components/ActionBar.vue'
import { useUserStore } from '@/stores/users';
import { useDesktopModalStore } from '@/stores/desktopModals';


// states
const assetStore = useAssetStore();
const userStore = useUserStore();
const projectStore = useProjectStore();

const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();

const addCollaborator = () => {
  modals.setModalVisibility('manageCollaboratorModal', true);
};

// refs

const isLastAdmin = computed(() => {
  let projectUsers = userStore.getProjectCollaborators;
  const projectRoles = projectUsers.map((user) => user.role.name);
  const isLastAdmin = projectRoles.filter(roleName => roleName === 'admin').length < 2;
  return isLastAdmin
});

const activeUserId = computed(() => {
  return userStore.user?.id;
});

const canRemoveUser = computed(() => { return userStore.canDo('remove_user') });
const canChangeRole = computed(() => { return userStore.canDo('change_role') });

const projectCollaborators = computed(() => {

  let projectUsers = userStore.getProjectCollaborators;
  let assignedUserIds = [];
  let tasks = assetStore.assets;

  for (const task of tasks) {
    let taskAssigneeId = task.assignee_id;
    if (!assignedUserIds.includes(taskAssigneeId)) {
      assignedUserIds.push(taskAssigneeId)
    }
  }

  const users = projectUsers.map(user => (
    {
      name: `${user.first_name} ${user.last_name}` || user,
      profile: user.photo || "",
      meta: user.role.name || "",
      id: user.id,
      avatarColor: userStore.userProfileColor(user.id),
      can_edit: user.id !== activeUserId.value && canChangeRole.value,
      can_delete: !assignedUserIds.includes(user.id) && user.id !== activeUserId.value && (user.role.name !== 'admin' || !isLastAdmin.value) && canRemoveUser.value,
    }
  ));
  return utils.sortAlphabetically(users);
});


// computed props
const prepEditCollaborator = (userId) => {
  let allCollaborators = userStore.getProjectCollaborators;
  let collaborator = allCollaborators.find(item => item.id === userId);
  userStore.selectedUser = collaborator
  modals.setModalVisibility('editCollaboratorModal', true);
};


const deleteCollaborator = (userId) => {
  let allCollaborators = userStore.getProjectCollaborators;
  let collaborator = allCollaborators.find(item => item.id === userId);
  // let collaborator = userStore.getProjectCollaborators[index];
  ProjectService.RemoveUser(projectStore.activeProject.uri, collaborator.id)
    .then(async (data) => {
      let users = userStore.users;
      let userIndex = users.indexOf(collaborator)
      userStore.users.splice(userIndex, 1)
      notificationStore.addNotification("User Removed Successfully.", "", "success")
    })
    .catch((error) => {
      notificationStore.errorNotification("Error Removing User", error);
    })
};

</script>

<style scoped>
.input-short {
  flex: 1;
  width: 100%;
}

.settings-component-root {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 5px;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
}

.settings-component-container {
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  height: 100%;
  overflow: hidden;
  box-sizing: border-box;
  width: 96%;
  gap: .5rem;
  align-items: center;
  color: white;
  justify-content: space-between;
  border-radius: var(--large-radius);
  padding: 1rem;
  background-color: crimson;
  background-color: var(--black-steel);
}
</style>


