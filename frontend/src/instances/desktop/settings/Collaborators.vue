<template>
  <div class="settings-component-root">
    <div class="settings-component-container">
      <div class="asset-header">
        <div class="create-menu">
          <ActionButton v-if="projectStore.isCloudHosted || userStore.canDo('add_user')" :icon="getAppIcon('person-plus')" :label="$t('settings.addCollaborator')" :showLabel="true"
            @click="addCollaborator" v-tooltip="$t('settings.addCollaborator')" />
          <ActionButton :icon="getAppIcon('refresh')" :label="$t('common.refresh')" v-tooltip="$t('common.refresh')"
            :buttonFunction="refresh" />
        </div>
      </div>

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
import { CollaboratorService, ProjectService } from "@/services";
import utils from '@/services/utils';

// store imports
import { useAssetStore } from '@/stores/assets';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import CollaboratorItem from '@/instances/desktop/components/CollaboratorItem.vue';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useUserStore } from '@/stores/users';


// stores
const assetStore = useAssetStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const trayStates = useTrayStates();
const userStore = useUserStore();

const { t } = useI18n();

const addCollaborator = () => {
  modals.setModalVisibility('manageCollaboratorModal', true);
};

// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Refreshes the project collaborators list.
const refresh = async () => {
  await userStore.reloadUsers();
};

// refs

// Whether the active project is a studio project (cloud or private).
const isStudioProject = computed(() => {
  return projectStore.selectedStudio && projectStore.selectedStudio.name !== 'Personal';
});

const isLastAdmin = computed(() => {
  if (projectStore.isCloudHosted) return false;
  let projectUsers = userStore.getProjectCollaborators;
  const projectRoles = projectUsers.map((user) => user.role.name);
  const isLastAdmin = projectRoles.filter(roleName => roleName === 'admin').length < 2;
  return isLastAdmin
});

const activeUserId = computed(() => {
  return userStore.user?.id;
});

const canRemoveUser = computed(() => { return projectStore.isCloudHosted || userStore.canDo('remove_user') });
const canChangeRole = computed(() => { return projectStore.isCloudHosted || userStore.canDo('change_role') });

const availableRoles = computed(() => {
  if (projectStore.isCloudHosted) return ['admin', 'artist'];
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
      can_delete: !assignedUserIds.includes(user.id) && user.id !== activeUserId.value && (projectStore.isCloudHosted || (user.role?.name !== 'admin' || !isLastAdmin.value)) && canRemoveUser.value,
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


const deleteCollaborator = async (userId) => {
  let allCollaborators = userStore.getProjectCollaborators;
  let collaborator = allCollaborators.find(item => item.id === userId);

  try {
    if (isStudioProject.value || projectStore.isCloudHosted) {
      const remoteUrl = projectStore.getActiveProjectUrl;
      await CollaboratorService.RemoveCollaborator(remoteUrl, collaborator.id);
      await ProjectService.RemoveUserSynced(projectStore.activeProject.uri, collaborator.id);
    } else {
      await ProjectService.RemoveUser(projectStore.activeProject.uri, collaborator.id);
    }
    await userStore.reloadUsers();
    notificationStore.addNotification(t('notifications.userRemoved'), "", "success");
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorRemovingUser'), error);
  }
};

</script>

<style scoped>
.input-short {
  flex: 1;
  width: 100%;
}

.asset-header {
  position: relative;
  display: flex;
  width: 100%;
  align-items: center;
  height: max-content;
  gap: 1rem;
  justify-content: space-between;
  padding: .2rem;
  box-sizing: border-box;
  min-width: max-content;
}

.create-menu {
  position: relative;
  display: flex;
  align-items: center;
  gap: .4rem;
  width: max-content;
  height: max-content;
  padding: .2rem;
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
  background-color: hsl(var(--border));
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
  color: hsl(var(--foreground));
  justify-content: space-between;
  
  padding: 1rem;
  background-color: hsl(var(--destructive));
  background-color: hsl(var(--background));
  
}
</style>


