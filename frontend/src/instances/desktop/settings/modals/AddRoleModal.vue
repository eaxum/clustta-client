<template>

  <div class="modal-container" v-esc="escape" @keydown.enter="handleEnterKey">
    <HeaderArea :title="title" :icon="'scale'" :showSearch="showSearch" />

    <div class="general-container">

      <div class="input-section">
        <input v-model="roleParameters.name" class="input-short" type="text" placeholder="Role Name" v-focus />
      </div>

      <div ref="collectionMenu" class="role-config" v-stop-propagation>

        <div v-for="(permissions, groupName) in groupedPermissions" :key="groupName" class="role-config-group">

          <div class="role-config-group-meta">
            <div class="role-config-group-name">
              {{ formatLabel(groupName) }}
            </div>
            <span class="active-count">
              {{ activePermissionsCount[groupName] }} permissions
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
        <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Add Role'" :fullWidth="true" @click="addRole" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>

    </div>
  </div>
</template>


<script setup>
import { ref, onMounted, computed, watch } from 'vue';

// stores
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useUserStore } from '@/stores/users';

//   components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';
import { UserService } from '@/../bindings/clustta/services/index';
import { useProjectStore } from '@/stores/projects';

// header
let title = 'Add new role';
let showSearch = false;

// stores
const notificationStore = useNotificationStore();
const modals = useDesktopModalStore();
const userStore = useUserStore();
const projectStore = useProjectStore();

// refs
const isAwaitingResponse = ref(false);

const defaultRole = {
  "id": "9c3403b7-aa5b-4958-afbb-2cf192009f84",
  "mtime": 1738949861,
  "name": "artist",
  "synced": true,
  "view_entity": false,
  "create_entity": false,
  "update_entity": false,
  "delete_entity": false,
  "view_task": false,
  "create_task": false,
  "update_task": false,
  "delete_task": false,
  "view_template": false,
  "create_template": false,
  "update_template": false,
  "delete_template": false,
  "view_checkpoint": true,
  "create_checkpoint": true,
  "delete_checkpoint": false,
  "pull_chunk": true,
  "assign_task": false,
  "unassign_task": false,
  "add_user": false,
  "remove_user": false,
  "change_role": false,
  "change_status": true,
  "set_done_task": false,
  "set_retake_task": false,
  "view_done_task": false,
  "manage_dependencies": false,
}
const roleParameters = ref({ ...defaultRole })
const initialSettings = ref({ ...defaultRole });

const handleEnterKey = (event) => {
  if (event.key === 'Enter') {
    addRole();
  }
};



const escape = () => {
  modals.setModalVisibility('addRoleModal', false);
};

const closeModal = () => {
  modals.setModalVisibility("addRoleModal", false);
};


const addRole = async () => {
  let parameters = roleParameters.value
  await UserService.AddRole(projectStore.activeProject.uri, parameters.name, parameters)
    .then((response) => {
      notificationStore.addNotification("Role Created", "", "success");
      const index = userStore.roles.findIndex(role => role.id === parameters.id);
      userStore.roles[index] = response
      closeModal()
    })
    .catch((error) => {
      console.log(error)
      notificationStore.errorNotification("Error Creating Role", error);
    });
};


const permissionGroups = {

  // general: ['change_status'],
  tasks: ['view_task', 'create_task', 'update_task', 'delete_task', 'manage_dependencies'],
  assignation: ['assign_task', 'unassign_task'],
  entities: ['view_entity', 'create_entity', 'update_entity', 'delete_entity'],
  users: ['add_user', 'remove_user', 'change_role'],
  status: ['view_done_task', 'change_status', 'set_done_task', 'set_retake_task',],
  templates: ['view_template', 'create_template', 'update_template', 'delete_template'],
  checkpoints: ['view_checkpoint', 'create_checkpoint', 'delete_checkpoint', 'pull_chunk'],
}

const groupedPermissions = computed(() => {
  const groups = {}
  Object.entries(permissionGroups).forEach(([groupName, permissions]) => {
    groups[groupName] = permissions.filter(permission =>
      typeof roleParameters.value[permission] === 'boolean' && permission !== 'synced'
    )
  })
  return groups
});

// Compute if any changes have been made
const isValueChanged = computed(() => {
  return Object.keys(roleParameters.value).some(key => {
    return roleParameters.value[key] !== initialSettings.value[key]
  })
})

// Compute active permissions count per group
const activePermissionsCount = computed(() => {
  const counts = {}
  Object.entries(permissionGroups).forEach(([groupName, permissions]) => {
    counts[groupName] = permissions.reduce((count, permission) => {
      return roleParameters.value[permission] === true ? count + 1 : count
    }, 0)
  })
  return counts
})

const formatLabel = (key) => {
  return key.replace(/_/g, ' ')
    .split(' ')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

const toggleField = (key) => {
  roleParameters.value[key] = !roleParameters.value[key]
}

onMounted(() => {
});


</script>

<style scoped>
@import "@/assets/desktop.css";

.role-config {
  display: flex;
  flex-direction: column;
  color: var(--white);
  align-items: center;
  gap: .3rem;
  padding: .6rem;
  box-sizing: border-box;
  width: max-content;
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
  border-radius: 10px;
  background-color: var(--light-steel);
}

.role-config::-webkit-scrollbar-track {
  border-radius: 10px;
}

.role-config-group {
  display: flex;
  flex-direction: column;
  color: var(--white);
  align-items: center;
  gap: .3rem;
  padding: .6rem;
  box-sizing: border-box;
  width: max-content;
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
  width: max-content;
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
  background-color: transparent;
  color: var(--white);
  position: relative;
  border-radius: 8px;
  border-radius: var(--small-radius);
  box-sizing: border-box;
  cursor: pointer;
  display: flex;
  gap: 10px;
  align-items: center;
  padding-left: .3rem;
  height: max-content;
  width: max-content;
  min-width: max-content;
  min-height: max-content;
  width: 100%;
  transition: all 0.3s ease;
}

.role-item:hover {
  background-color: rgba(255, 255, 255, 0.05);
}

.entity-item-menu-visible {
  opacity: 1;
  visibility: visible;
}

.input-short {
  flex: 1;
  width: 100%;
}

.listbox-short {

  flex: 1;
  width: 130px;
}

.input-label {

  font-family: Inter, sans-serif;
  color: var(--white);
  font-size: 16px;
  white-space: nowrap;
  flex: 1;

}

.category-area {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  color: var(--white);
  width: 98%;
}

.category-list {
  box-sizing: border-box;
  display: flex;
  padding: .5rem;
  align-items: center;
  flex-direction: column;
  gap: .2rem;
  background-color: rgba(0, 0, 0, 0.144);
  height: 290px;
  overflow: hidden;
  overflow-y: scroll;
  width: 100%;
  border-radius: 10px;
}

.category-list::-webkit-scrollbar {
  width: 4px;
}

.category-list::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: rgba(255, 255, 255, 0.295);
}

.category-list::-webkit-scrollbar-track {
  border-radius: 10px;
}

.category-item {
  color: var(--white);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .5rem;
  width: 100%;
  border-bottom: 1px solid rgba(255, 255, 255, 0.096);
  height: max-content;
  padding: .2rem;
}

.category-item-actions {
  display: flex;

}

.horizontal-flex{
  font-weight: 400;
}

[data-theme="dark"] .horizontal-flex{
  font-weight: 200;
}
</style>

