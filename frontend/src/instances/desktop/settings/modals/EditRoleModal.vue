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
        <GeneralButton :label="$t('common.update')" :fullWidth="true" @click="updateRole" :isActive="isValueChanged"
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
const initialSettings = ref({ ...userStore.selectedRole });

const permissionGroups = {
  assets: ['view_asset', 'create_asset', 'update_asset', 'delete_asset', 'manage_dependencies'],
  assignation: ['assign_asset', 'unassign_asset'],
  collections: ['view_collection', 'create_collection', 'update_collection', 'delete_collection'],
  users: ['add_user', 'remove_user', 'change_role'],
  status: ['view_done_asset', 'change_status', 'set_done_asset', 'set_retake_asset'],
  templates: ['view_template', 'create_template', 'update_template', 'delete_template'],
  checkpoints: ['view_checkpoint', 'create_checkpoint', 'delete_checkpoint', 'pull_chunk'],
};

const roleParameters = ref({ ...userStore.selectedRole });
const showSearch = false;
const title = t('modals.editRole');

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
// Closes the modal.
const closeModal = () => {
  modals.setModalVisibility("editRoleModal", false);
};

// Closes the modal on escape key.
const escape = () => {
  modals.setModalVisibility('editRoleModal', false);
};

// Formats a permission key to display label.
const formatLabel = (key) => {
  return key.replace(/_/g, ' ')
    .split(' ')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
};

// Handles enter key press to trigger update role.
const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    updateRole();
  }
};

// Toggles a permission field.
const toggleField = (key) => {
  roleParameters.value[key] = !roleParameters.value[key];
};

// Updates the role with current parameters.
const updateRole = async () => {
  let parameters = roleParameters.value;
  await UserService.UpdateRole(projectStore.activeProject.uri, parameters.id, parameters.name, parameters)
    .then((response) => {
      notificationStore.addNotification(t('notifications.roleUpdated'), "", "success");
      const index = userStore.roles.findIndex(role => role.id === parameters.id);
      userStore.roles[index] = response;
      closeModal();
    })
    .catch((error) => {
      console.log(error);
      notificationStore.errorNotification(t('notifications.errorUpdatingRole'), error);
    });
};
</script>

<style scoped>
@import "@/assets/desktop.css";

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
  padding-left: .3rem;
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

