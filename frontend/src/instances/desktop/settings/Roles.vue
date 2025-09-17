<template>
  <div class="settings-component-root">
    <div class="settings-component-container">

      <ActionBar :itemType="'Add Role'" :addFunction="addRole" />

      <ScrollList v-if="projectRoles.length" :items="projectRoles" :useIcons="true" :useItemId="true" :wrapItems="false"
        :editItems="true" :editListItem="prepEditRole" :deleteItems="true" :deleteListItem="deleteRole" />

      <PageState v-else :message="message()" :illustration="illustration()" :secondaryIcon="getAppIcon('plus-circle')"
        :secondaryActionMessage="secondaryActionMessage()" :secondaryActionFunction="secondaryActionFunction" />

    </div>
  </div>
</template>

<script setup>
import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
    const icon = iconStore.getAppIcon(iconName);
    return icon
};

// imports
import { onMounted, computed, ref } from 'vue';
import utils from '@/services/utils';

// state imports
import { useTrayStates } from '@/stores/TrayStates';

// store imports
import { useNotificationStore } from '@/stores/notifications';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useUserStore } from '@/stores/users';
import { useProjectStore } from '@/stores/projects';

// components
import ScrollList from '@/instances/desktop/components/ScrollList.vue';
import ActionBar from '@/instances/desktop/components/ActionBar.vue';
import PageState from '@/instances/common/components/PageState.vue';
import { UserService } from '@/../bindings/clustta/services/index';

// states
const userStore = useUserStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();

// computed props
const getRoleTypeIcon = (icon) => {
  return '/icons/person.svg'
}

const projectRoles = computed(() => {

  let projectRoles = userStore.getProjectRoles;
  let projectUsers = userStore.getProjectCollaborators;

  let usedProjectRoleIds = [];

  for (const user of projectUsers){
    if(!usedProjectRoleIds.includes(user.role_id)){
      usedProjectRoleIds.push(user.role_id)
    }
  }

  const roles = projectRoles.map(type => (
    {
      ...type,
      icon: getRoleTypeIcon(type.name),
      can_delete: !usedProjectRoleIds.includes(type.id),
      can_edit: type.name !== 'admin',
    }
  ));
  
  return roles
});

// methods
const message = () => {
  return 'You have no user roles';
};

const illustration = () => {
  return '/page-states/resources.png';
};

const secondaryActionMessage = () => {
  return 'Add Template'
};

const secondaryActionFunction = () => {
  addRole();
};

const addRole = () => {
  modals.setModalVisibility('addRoleModal', true);
};

const prepEditRole = (roleId) => {
  const allRoles = userStore.getProjectRoles;
  const selectedRole = allRoles.find((item) => item.id === roleId);
  userStore.selectedRole = selectedRole
  console.log(userStore.selectedRole);

  modals.setModalVisibility('editRoleModal', true);
};

const replaceSymbols = (name) => {
  return name.replace(/_/g, " ").toLowerCase().replace(/(^\w|\s\w)/g, match => match.toUpperCase());
};

const deleteRole = async (roleId) => {
  UserService.DeleteRole(projectStore.activeProject.uri, roleId)
    .then((response) => {
      notificationStore.addNotification("Role Deleted", "", "success");
      const index = userStore.roles.findIndex(role => role.id === roleId);
      userStore.roles.splice(index, 1);
    })
    .catch((error) => {
      notificationStore.errorNotification("Error Deleting Role", error);
    });
};

// onMounted hook
onMounted(async () => {

});
</script>


<style scoped>
.input-short {
  flex: 1;
  width: 100%;
}

.settings-component-root{
  width:100%;
  height:100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 5px;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
}

.settings-component-container{
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

