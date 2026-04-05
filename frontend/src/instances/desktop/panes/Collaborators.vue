<template>
  <div class="general-pane-header">
    <SearchBar v-model="searchQuery" :placeholder="$t('panes.searchByNameOrEmail')" @clear="clearSearch" />
  </div>

  <div class="general-pane-root">
    <div v-if="hasResults" class="collaborators-scroll-container">
      <div class="collaborators-list">
        <CollaboratorItem v-for="(collaborator, index) in filteredProjectCollaborators" :key="collaborator.id"
          :collaborator="collaborator" :index="index" :compact="true" :roles="availableRoles"
          :onRoleChange="changeCollaboratorRole" :onDelete="deleteCollaborator"
          :canEdit="collaborator.can_edit" :canDelete="collaborator.can_delete"
          :isLoading="loadingCollaboratorIds.includes(collaborator.id)" />

        <div v-if="searchQuery && filteredStudioCollaborators.length && filteredProjectCollaborators.length" class="collaborators-divider">
          <span class="divider-text">{{ $t('panes.studioMembers') }}</span>
        </div>

        <CollaboratorItem v-for="(collaborator, index) in filteredStudioCollaborators" :key="'studio-' + collaborator.id"
          :collaborator="collaborator" :index="index" :compact="true"
          :onAdd="addCollaboratorToProject" :canEdit="canAddUser" :canDelete="false" :isProjectMember="false"
          :isLoading="loadingCollaboratorIds.includes(collaborator.id)" />
      </div>
    </div>

    <PageState v-else :message="message()" :illustration="illustration()" />
  </div>
</template>

<script setup>
// imports
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import utils from '@/services/utils';

// components
import CollaboratorItem from '@/instances/desktop/components/CollaboratorItem.vue';
import PageState from '@/instances/common/components/PageState.vue';
import SearchBar from '@/instances/desktop/components/SearchBar.vue';

// services
import { CollaboratorService, ProjectService } from "@/services";

// stores
import { useAssetStore } from '@/stores/assets';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStudioStore } from '@/stores/studio';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';

const assetStore = useAssetStore();
const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const studioStore = useStudioStore();
const trayStates = useTrayStates();
const userStore = useUserStore();

const { t } = useI18n();

// refs
const loadingCollaboratorIds = ref([]);
const searchBar = ref(null);
const searchQuery = ref('');

// computed props
const activeUserId = computed(() => {
  return userStore.user?.id;
});

const availableRoles = computed(() => {
  return userStore.getRolesNames;
});

const canAddUser = computed(() => {
  return userStore.canDo('add_user');
});

const canChangeRole = computed(() => {
  return userStore.canDo('change_role');
});

const canRemoveUser = computed(() => {
  return userStore.canDo('remove_user');
});

// Returns filtered project collaborators based on search query.
const filteredProjectCollaborators = computed(() => {
  if (!searchQuery.value) {
    return projectCollaborators.value;
  }
  const query = searchQuery.value.toLowerCase();
  return projectCollaborators.value.filter(collaborator => {
    return collaborator.name.toLowerCase().includes(query) ||
      collaborator.email?.toLowerCase().includes(query);
  });
});

// Returns filtered studio collaborators (not in project) based on search query.
const filteredStudioCollaborators = computed(() => {
  if (!searchQuery.value) {
    return [];
  }
  const query = searchQuery.value.toLowerCase();
  return studioCollaborators.value.filter(collaborator => {
    return collaborator.name.toLowerCase().includes(query) ||
      collaborator.email?.toLowerCase().includes(query);
  });
});

// Returns true if there are any results to display.
const hasResults = computed(() => {
  return filteredProjectCollaborators.value.length > 0 || filteredStudioCollaborators.value.length > 0;
});

const isLastAdmin = computed(() => {
  const projectUsers = userStore.getProjectCollaborators;
  const projectRoles = projectUsers.map((user) => user.role.name);
  return projectRoles.filter(roleName => roleName === 'admin').length < 2;
});

// Returns the list of project collaborators with computed permissions.
const projectCollaborators = computed(() => {
  const projectUsers = userStore.getProjectCollaborators;
  const assignedUserIds = [];
  const assets = assetStore.assets;

  for (const asset of assets) {
    const assetAssigneeId = asset.assignee_id;
    if (!assignedUserIds.includes(assetAssigneeId)) {
      assignedUserIds.push(assetAssigneeId);
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
      can_delete: !assignedUserIds.includes(user.id) && user.id !== activeUserId.value && (user.role?.name !== 'admin' || !isLastAdmin.value) && canRemoveUser.value,
    };
  });
  return utils.sortAlphabetically(users);
});

// Returns studio users who are not in the current project.
const studioCollaborators = computed(() => {
  const projectUserIds = userStore.getProjectCollaborators.map(user => user.id);
  const studioUsers = studioStore.studioUsers || [];

  const users = studioUsers
    .filter(user => !projectUserIds.includes(user.id))
    .map(user => {
      return {
        ...user,
        name: `${user.first_name} ${user.last_name}` || user,
        profile: user.photo || "",
        role_name: user.role_name || 'user',
        id: user.id,
        avatarColor: userStore.userProfileColor(user.id),
        can_edit: true,
        can_delete: false,
      };
    });
  return utils.sortAlphabetically(users);
});

// computed
// Whether the active project is a studio project (cloud or private).
const isStudioProject = computed(() => {
  return projectStore.selectedStudio && projectStore.selectedStudio.name !== 'Personal';
});

// methods
// Adds a studio collaborator to the project with 'Artist' role (or first available).
const addCollaboratorToProject = async (userId) => {
  const collaborator = studioCollaborators.value.find(user => user.id === userId);
  if (!collaborator) {
    notificationStore.errorNotification(t('notifications.errorAddingUserToProject'), "User not found");
    return;
  }

  // Use 'Artist' role if available, otherwise fall back to first available role
  const roles = availableRoles.value || [];
  const defaultRole = roles.find(role => role.toLowerCase() === 'artist') || roles[0];
  
  if (!defaultRole) {
    notificationStore.errorNotification(t('notifications.errorAddingUserToProject'), t('notifications.noRolesAvailable'));
    return;
  }

  loadingCollaboratorIds.value.push(userId);

  try {
    if (isStudioProject.value) {
      const remoteUrl = projectStore.getActiveProjectUrl;
      await CollaboratorService.AddCollaboratorsWithRole(remoteUrl, [userId], defaultRole);
      await ProjectService.AddUserSynced(projectStore.activeProject.uri, collaborator.email, defaultRole);
    } else {
      await ProjectService.AddUser(projectStore.activeProject.uri, collaborator.email, defaultRole);
    }
    notificationStore.addNotification(t('notifications.userAddedToProject'), "", "success");
    await trayStates.refreshData();
  } catch (error) {
    notificationStore.errorNotification(t('notifications.errorAddingUserToProject'), error);
  } finally {
    loadingCollaboratorIds.value = loadingCollaboratorIds.value.filter(id => id !== userId);
  }
};

// Updates the collaborator's role in the project.
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

// Clears the search query.
const clearSearch = () => {
  searchQuery.value = '';
};

// Removes a collaborator from the project.
const deleteCollaborator = async (userId) => {
  const allCollaborators = userStore.getProjectCollaborators;
  const collaborator = allCollaborators.find(item => item.id === userId);
  
  loadingCollaboratorIds.value.push(userId);
  
  try {
    if (isStudioProject.value || projectStore.isR2Remote) {
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
  } finally {
    loadingCollaboratorIds.value = loadingCollaboratorIds.value.filter(id => id !== userId);
  }
};

// Returns the icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Returns the illustration path for the empty state.
const illustration = () => {
  return '/page-states/resources.png';
};

// Returns the message for the empty state.
const message = () => {
  if (searchQuery.value) {
    return t('notifications.noCollaboratorsMatch');
  }
  return t('notifications.noCollaboratorsOnProject');
};

// Handles search input updates.
const updateSearch = () => {
  // Search is handled reactively via filteredCollaborators computed
};

</script>

<style scoped>
@import "@/assets/desktop.css";

.collaborators-scroll-container {
  width: 100%;
  height: 100%;
  overflow-y: auto;
  overflow-x: hidden;
  box-sizing: border-box;
  padding-right: 5px;
  padding-bottom: 1rem;
}

.collaborators-scroll-container::-webkit-scrollbar {
  width: 4px;
}

.collaborators-scroll-container::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--light-steel);
}

.collaborators-scroll-container::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.collaborators-list {
  display: flex;
  flex-direction: column;
  gap: .5rem;
  width: 100%;
  box-sizing: border-box;
}

.collaborators-divider {
  display: flex;
  align-items: center;
  width: 100%;
  padding: .5rem 0;
  gap: .5rem;
}

.collaborators-divider::before,
.collaborators-divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background-color: var(--light-steel);
  opacity: 0.3;
}

.divider-text {
  font-size: 12px;
  color: var(--white);
  opacity: 0.5;
  white-space: nowrap;
}
</style>
