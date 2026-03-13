<template>
  <div class="settings-component-root">
    <div class="settings-component-container">
      <ActionBar v-if="userStore.canDo('add_user')" :itemType="$t('settings.addCollaborator').toLowerCase()" :addFunction="addCollaborator" />

      <div v-if="projectCollaborators.length" class="collaborators-list-wrapper">
        <div class="collaborators-list">
          <CollaboratorItem 
            v-for="(collaborator, index) in projectCollaborators" 
            :key="collaborator.id"
            :collaborator="collaborator"
            :index="index"
            :roles="availableRoles"
            :onRoleChange="changeCollaboratorRole"
            :onDelete="deleteCollaborator"
            :canEdit="collaborator.can_edit"
            :canDelete="collaborator.can_delete"
          />
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>

// imports
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { ProjectService } from "@/services";
import utils from '@/services/utils';

// store imports
import { useAssetStore } from '@/stores/assets';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';

// components
import CollaboratorItem from '@/instances/desktop/components/CollaboratorItem.vue';
import ActionBar from '@/instances/desktop/components/ActionBar.vue'
import { useUserStore } from '@/stores/users';
import { useDesktopModalStore } from '@/stores/desktopModals';


// states
const assetStore = useAssetStore();
const userStore = useUserStore();
const projectStore = useProjectStore();
const trayStates = useTrayStates();

const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();
const { t } = useI18n();

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

const availableRoles = computed(() => {
  return userStore.getRolesNames;
});

const projectCollaborators = computed(() => {

  let projectUsers = userStore.getProjectCollaborators;
  let assignedUserIds = [];
  let assets = assetStore.assets;

  for (const asset of assets) {
    let assetAssigneeId = asset.assignee_id;
    if (!assignedUserIds.includes(assetAssigneeId)) {
      assignedUserIds.push(assetAssigneeId)
    }
  }

  const users = projectUsers.map(user => {
    // Find the matching role name from availableRoles (case-insensitive match)
    const userRoleName = user.role?.name || "";
    const matchedRole = userStore.getRolesNames.find(
      role => role.toLowerCase() === userRoleName.toLowerCase()
    ) || userRoleName;

    return {
      ...user,
      name: `${user.first_name} ${user.last_name}` || user,
      profile: user.photo || "",
      role_name: matchedRole,
      id: user.id,
      avatarColor: userStore.userProfileColor(user.id),
      can_edit: user.id !== activeUserId.value && canChangeRole.value,
      can_delete: !assignedUserIds.includes(user.id) && user.id !== activeUserId.value && (user.role?.name !== 'admin' || !isLastAdmin.value) && canRemoveUser.value,
    };
  });
  return utils.sortAlphabetically(users);
});

const changeCollaboratorRole = async (userId, newRole) => {
  await ProjectService.ChangeRole(projectStore.activeProject.uri, userId, newRole)
    .then(async () => {
      notificationStore.addNotification(t('notifications.userUpdated'), "", "success");
      await trayStates.refreshData();
    })
    .catch((error) => {
      notificationStore.errorNotification(t('notifications.errorUpdatingUser'), error);
    });
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
      notificationStore.addNotification(t('notifications.userRemoved'), "", "success")
    })
    .catch((error) => {
      notificationStore.errorNotification(t('notifications.errorRemovingUser'), error);
    })
};

</script>

<style scoped>
.input-short {
  flex: 1;
  width: 100%;
}

.collaborators-list-wrapper {
  width: 100%;
  min-height: 0;
  flex: 1;
  overflow-y: auto;
}

.collaborators-list-wrapper::-webkit-scrollbar {
  width: 4px;
}

.collaborators-list-wrapper::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--light-steel);
}

.collaborators-list-wrapper::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.collaborators-list {
  display: flex;
  flex-direction: column;
  gap: .5rem;
  width: 100%;
  box-sizing: border-box;
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
  border-radius: var(--very-large-radius);
}
</style>


