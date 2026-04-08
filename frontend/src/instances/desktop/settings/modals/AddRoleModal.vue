<template>

  <div class="modal-container" v-esc="escape" @keydown.enter="handleEnterKey">
    <HeaderArea :title="title" :icon="'scale'" :showSearch="showSearch" />

    <div class="general-container">

      <div class="input-section">
        <input v-model="roleParameters.name" class="input-short" type="text" :placeholder="$t('placeholders.roleName')" v-focus />
      </div>

      <div ref="collectionMenu" class="role-config" v-stop-propagation>

        <div v-for="(permissions, groupName) in groupedPermissions" :key="groupName" class="role-config-group">

          <div class="role-config-group-meta">
            <div class="role-config-group-name">
              {{ formatLabel(groupName) }}
            </div>
            <span class="active-count">
              {{ $t('modals.permissionCount', { count: activePermissionsCount[groupName] }) }}
            </span>
          </div>
          <div class="menu-divider"></div>

          <span v-for="permission in permissions" :key="permission" class="role-item" @click="toggleField(permission)">
            <!-- <img class="small-icons" :src="getStatusIcon(status)"> -->
            <div class="horizontal-flex">
              <div> {{ formatLabel(permission) }}</div>
              <ToggleSwitch :switchValueProp="roleParameters[permission]" />
            </div>
          </span>

        </div>

      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="$t('common.addRole')" :fullWidth="true" @click="addRole" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>

    </div>
  </div>
</template>


<script setup>
// imports
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';

// services
import { UserService } from '@/services';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useUserStore } from '@/stores/users';

const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const userStore = useUserStore();

const { t } = useI18n();

// refs
const collectionMenu = ref(null);
const isAwaitingResponse = ref(false);

// constants
const defaultRole = {
  "id": "9c3403b7-aa5b-4958-afbb-2cf192009f84",
  "mtime": 1738949861,
  "name": "artist",
  "synced": true,
  "view_collection": false,
  "create_collection": false,
  "update_collection": false,
  "delete_collection": false,
  "view_asset": false,
  "create_asset": false,
  "update_asset": false,
  "delete_asset": false,
  "view_template": false,
  "create_template": false,
  "update_template": false,
  "delete_template": false,
  "view_checkpoint": true,
  "create_checkpoint": true,
  "delete_checkpoint": false,
  "pull_chunk": true,
  "assign_asset": false,
  "unassign_asset": false,
  "add_user": false,
  "remove_user": false,
  "change_role": false,
  "change_status": true,
  "set_done_asset": false,
  "set_retake_asset": false,
  "view_done_asset": false,
  "manage_dependencies": false,
  "manage_share_links": false,
};

const initialSettings = ref({ ...defaultRole });

const permissionGroups = {
  assets: ['view_asset', 'create_asset', 'update_asset', 'delete_asset', 'manage_dependencies'],
  assignation: ['assign_asset', 'unassign_asset'],
  collections: ['view_collection', 'create_collection', 'update_collection', 'delete_collection'],
  users: ['add_user', 'remove_user', 'change_role'],
  status: ['view_done_asset', 'change_status', 'set_done_asset', 'set_retake_asset'],
  templates: ['view_template', 'create_template', 'update_template', 'delete_template'],
  checkpoints: ['view_checkpoint', 'create_checkpoint', 'delete_checkpoint', 'pull_chunk'],
  sharing: ['manage_share_links'],
};

const roleParameters = ref({ ...defaultRole });
const showSearch = false;
const title = t('modals.addNewRole');

// computed
// Computes active permissions count per group.
const activePermissionsCount = computed(() => {
  const counts = {};
  Object.entries(permissionGroups).forEach(([groupName, permissions]) => {
    counts[groupName] = permissions.reduce((count, permission) => {
      return roleParameters.value[permission] === true ? count + 1 : count;
    }, 0);
  });
  return counts;
});

// Groups permissions by category.
const groupedPermissions = computed(() => {
  const groups = {};
  Object.entries(permissionGroups).forEach(([groupName, permissions]) => {
    groups[groupName] = permissions.filter(permission =>
      typeof roleParameters.value[permission] === 'boolean' && permission !== 'synced'
    );
  });
  return groups;
});

// Checks if any changes have been made.
const isValueChanged = computed(() => {
  return Object.keys(roleParameters.value).some(key => {
    return roleParameters.value[key] !== initialSettings.value[key];
  });
});

// methods
// Adds a new role to the project.
const addRole = async () => {
  let parameters = roleParameters.value;
  await UserService.AddRole(projectStore.activeProject.uri, parameters.name, parameters)
    .then((response) => {
      notificationStore.addNotification(t('notifications.roleCreated'), "", "success");
      const index = userStore.roles.findIndex(role => role.id === parameters.id);
      userStore.roles[index] = response;
      closeModal();
    })
    .catch((error) => {
      console.log(error);
      notificationStore.errorNotification(t('notifications.errorCreatingRole'), error);
    });
};

// Closes the modal.
const closeModal = () => {
  modals.setModalVisibility("addRoleModal", false);
};

// Closes the modal on escape key.
const escape = () => {
  modals.setModalVisibility('addRoleModal', false);
};

// Formats a permission key to display label.
const formatLabel = (key) => {
  return key.replace(/_/g, ' ')
    .split(' ')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
};

// Handles enter key press to trigger add role.
const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    addRole();
  }
};

// Toggles a permission field.
const toggleField = (key) => {
  roleParameters.value[key] = !roleParameters.value[key];
};
</script>

<style scoped>
@import "@/assets/desktop.css";

.general-container{
  min-width: 500px;
}

.horizontal-flex {
  font-weight: 400;
}

.input-short {
  flex: 1;
  width: 100%;
}

.role-config {
  display: flex;
  flex-direction: column;
  color: var(--white);
  align-items: center;
  gap: .3rem;
  padding: .6rem;
  box-sizing: border-box;
  width: 100%;
  height: max-content;
  max-height: 50vh;
  border-radius: var(--normal-radius);
  overflow: hidden;
  overflow-y: scroll;
}

.role-config::-webkit-scrollbar {
  width: 4px;
}

.role-config::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--light-steel);
}

.role-config::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.role-config-group {
  display: flex;
  flex-direction: column;
  color: var(--white);
  align-items: center;
  gap: .3rem;
  padding: .6rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  border-radius: var(--normal-radius);
  background-color: var(--dark-steel);
}

.role-config-group-meta {
  display: flex;
  color: var(--white);
  align-items: center;
  gap: .5rem;
  padding: .6rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: space-between;
}

.role-config-group-name {
  font-size: large;
  font-weight: 300;
  display: flex;
  color: var(--white);
  align-items: flex-start;
  gap: .3rem;
  box-sizing: border-box;
  width: max-content;
  height: min-content;
}

.role-item {
  overflow: hidden;
  background-color: transparent;
  text-align: center;
  font-size: 14px;
  line-height: 14px;
  color: var(--white);
  position: relative;
  border-radius: var(--small-radius);
  box-sizing: border-box;
  cursor: pointer;
  display: flex;
  gap: 10px;
  align-items: center;
  padding: .3rem;
  min-width: max-content;
  min-height: max-content;
  width: 100%;
  transition: all 0.3s ease;
}

.role-item:hover {
  background-color: rgba(255, 255, 255, 0.05);
}

[data-theme="dark"] .horizontal-flex {
  font-weight: 200;
}
</style>

