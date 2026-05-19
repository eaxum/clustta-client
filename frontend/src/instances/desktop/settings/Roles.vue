<template>
  <div class="settings-component-root">
    <div class="settings-component-container">

      <ActionBar :itemType="$t('settings.addRole')" :addFunction="addRole" />

      <div v-if="projectRoles.length" class="roles-list-wrapper">
        <div class="roles-list">
          <RoleItem v-for="role in projectRoles" :key="role.id" :role="role" :canEdit="role.can_edit" :canDelete="role.can_delete" :onEdit="prepEditRole" :onDelete="deleteRole" />
        </div>
      </div>

      <PageState v-else :message="message()" :illustration="illustration()" :secondaryIcon="getAppIcon('plus-circle')" :secondaryActionMessage="secondaryActionMessage()" :secondaryActionFunction="secondaryActionFunction" />

    </div>
  </div>
</template>

<script setup>
// imports
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import utils from '@/services/utils';

// components
import ActionBar from '@/instances/desktop/components/ActionBar.vue';
import PageState from '@/instances/common/components/PageState.vue';
import RoleItem from '@/instances/desktop/components/RoleItem.vue';

// services
import { UserService } from '@/services';

// stores
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const userStore = useUserStore();

import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useUserStore } from '@/stores/users';

const { t } = useI18n();

// computed
const projectRoles = computed(() => {
  let projectRoles = userStore.getProjectRoles;
  let projectUsers = userStore.getProjectCollaborators;

  let usedProjectRoleIds = [];
  for (const user of projectUsers) {
    if (!usedProjectRoleIds.includes(user.role_id)) {
      usedProjectRoleIds.push(user.role_id);
    }
  }

  return projectRoles.map(role => ({
    ...role,
    name: utils.capitalizeStr(role.name),
    can_delete: !usedProjectRoleIds.includes(role.id),
    can_edit: role.name !== 'admin',
  }));
});

// methods
// Opens the add role modal.
const addRole = () => {
  modals.setModalVisibility('addRoleModal', true);
};

// Deletes a role from the project.
const deleteRole = async (roleId) => {
  UserService.DeleteRole(projectStore.activeProject.uri, roleId)
    .then(() => {
      notificationStore.addNotification(t('notifications.roleDeleted'), "", "success");
      const index = userStore.roles.findIndex(role => role.id === roleId);
      userStore.roles.splice(index, 1);
    })
    .catch((error) => {
      notificationStore.errorNotification(t('notifications.errorDeletingRole'), error);
    });
};

// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Returns the empty state illustration path.
const illustration = () => {
  return '/page-states/resources.png';
};

// Returns the empty state message.
const message = () => {
  return t('settings.noUserRoles');
};

// Opens the edit role modal for the given role.
const prepEditRole = (roleId) => {
  const allRoles = userStore.getProjectRoles;
  const selectedRole = allRoles.find((item) => item.id === roleId);
  userStore.selectedRole = selectedRole;
  modals.setModalVisibility('editRoleModal', true);
};

// Returns the empty state secondary action label.
const secondaryActionFunction = () => {
  addRole();
};

// Returns the empty state secondary action label.
const secondaryActionMessage = () => {
  return t('settings.addRole');
};
</script>


<style scoped>
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
  width: 96%;
  gap: .5rem;
  align-items: center;
  color: white;
  justify-content: space-between;
  padding: 1rem;
  background-color: var(--surface-1);
  border-radius: var(--very-large-radius);
}

.roles-list-wrapper {
  width: 100%;
  min-height: 0;
  flex: 1;
  overflow-y: auto;
}

.roles-list-wrapper::-webkit-scrollbar {
  width: 4px;
}

.roles-list-wrapper::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--surface-4);
}

.roles-list-wrapper::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.roles-list {
  display: flex;
  flex-direction: column;
  gap: .5rem;
  width: 100%;
  box-sizing: border-box;
}
</style>

