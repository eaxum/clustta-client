<template>

  <div class="modal-container" ref="modalContainer" v-stop-propagation v-esc="closeModal">
    <HeaderArea :title="title" :icon="'person-plus'" :showSearch="showSearch" />

    <div class="general-container">

      <div class="horizontal-flex">
        <CollaboratorSuggestions :allowMultipleEntries="true" :placeholder="placeholder" :selectedItems="selectedUsers" :allItems="isR2Remote ? [] : studioUsers"
          @tagAdded="addUser" @tagRemoved="removeUser" />
      </div>

      <div v-if="!isR2Remote" class="horizontal-flex">
        <DropDownBox :items="userStore.getRolesNames" :onSelect="selectRole"
          :selectedItem="collaboratorRole" :placeHolder="$t('common.none')" :fullWidth="true" />
      </div>

      <div v-if="isR2Remote" class="horizontal-flex">
        <DropDownBox :items="personalRemoteRoles" :onSelect="selectRole"
          :selectedItem="collaboratorRole" :placeHolder="$t('common.none')" :fullWidth="true" />
      </div>

      <!-- Notification section for non-studio users -->
      <div class="notification-area">
      <div v-if="!isR2Remote && nonStudioUsers.length > 0" class="horizontal-flex">
        <NotificationBox 
          type="warning"
          :icon="getAppIcon('alert')"
          :iconAlt="$t('common.alert')"
          :title="$t('modals.studioAdditionRequired')"
          :message="$t('modals.studioAdditionMessage', nonStudioUsers.length)"
        />
      </div>

      <!-- Notification section for new users -->
      <div v-if="newUsers.length > 0" class="horizontal-flex">
        <NotificationBox 
          type="invitation"
          :icon="getAppIcon('mail')"
          :iconAlt="$t('common.invitation')"
          :title="$t('modals.invitationRequired')"
          :message="$t('modals.invitationMessage', newUsers.length)"
        />
      </div>
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="$t('common.add')" :buttonFunction="addCollaborators" :isActive="!!selectedUsers.length"
          :loading="isAwaitingResponse" />
      </div>


    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import CollaboratorSuggestions from '@/instances/common/components/CollaboratorSuggestions.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import NotificationBox from '@/instances/common/components/NotificationBox.vue';

// services
import { AuthService, CollaboratorService, ProjectService, StudioService } from "@/services";

// stores
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const studioStore = useStudioStore();
const trayStates = useTrayStates();
const userStore = useUserStore();

const { t } = useI18n();

import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStudioStore } from '@/stores/studio';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';

// emits
const emit = defineEmits(['addedCollaborator']);

// constants
const personalRemoteRoles = ['admin', 'artist'];
const placeholder = t('placeholders.nameOrEmailSeparated');
const showSearch = false;
const title = t('modals.manageCollaborators');

// refs
const allProjectCollaborators = ref([]);
const collaboratorRole = ref(
  projectStore.isR2Remote
    ? 'artist'
    : userStore.getRolesNames[userStore.getRolesNames.length - 1]
);
const isAwaitingResponse = ref(false);
const modalContainer = ref(null);
const selectedUserEmails = ref([]);
const unregisteredUserEmails = ref([]);

// computed
// Whether the active project is a personal remote project.
const isR2Remote = computed(() => projectStore.isR2Remote);

const newUsers = computed(() => {
  return selectedUsers.value.filter(user => user.userType === 'new');
});

const nonStudioUsers = computed(() => {
  return selectedUsers.value.filter(user => user.userType === 'user');
});

const selectedUsers = computed(() => {
  if (isR2Remote.value) {
    const registeredUsers = selectedUserEmails.value.map(email => ({
      id: email,
      email: email,
      full_name: email,
      avatarColor: generateAvatarColor(email),
      userType: 'user'
    }));

    const unregisteredUsers = unregisteredUserEmails.value.map(email => ({
      id: email,
      email: email,
      full_name: email,
      avatarColor: generateAvatarColor(email),
      userType: 'new'
    }));

    return [...registeredUsers, ...unregisteredUsers];
  }

  const studioUsersList = studioStore.studioUsers
    .map(user => ({
      ...user,
      full_name: `${user.first_name} ${user.last_name}`,
      avatarColor: userStore.userProfileColor(user.id),
      userType: 'studio'
    }))
    .filter((user) => selectedUserEmails.value.includes(user.email));

  const registeredUsers = selectedUserEmails.value
    .filter(id => !studioStore.studioUsers.some(user => (user.email) === id))
    .map(email => ({
      id: email,
      email: email,
      full_name: email,
      avatarColor: generateAvatarColor(email),
      userType: 'user'
    }));

  const unregisteredUsers = unregisteredUserEmails.value
    .filter(id => !studioStore.studioUsers.some(user => (user.email) === id))
    .map(email => ({
      id: email,
      email: email,
      full_name: email,
      avatarColor: generateAvatarColor(email),
      userType: 'new'
    }));

  return [...studioUsersList, ...registeredUsers, ...unregisteredUsers];
});

const studioUsers = computed(() => {
  const projectUserEmails = userStore.getProjectCollaborators.map((user) => user.email);
  const availableUsers = studioStore.studioUsers
    .map(user => ({
      ...user,
      full_name: `${user.first_name} ${user.last_name}`,
      avatarColor: userStore.userProfileColor(user.id)
    }))
    .filter((user) => !selectedUserEmails.value.includes(user.email) && !projectUserEmails.includes(user.email));
  return availableUsers;
});

// methods

// Adds collaborators to the project, handling studio users, global users, and new invites.
const addCollaborators = async () => {
  isAwaitingResponse.value = true;

  try {
    if (isR2Remote.value) {
      await addPersonalRemoteCollaborators();
    } else {
      await addStudioCollaborators();
    }
  } catch (error) {
    console.error('Error in addCollaborators:', error);
    notificationStore.errorNotification(t('notifications.errorAddingUsers'), error);
  } finally {
    if (!isR2Remote.value) {
      await studioStore.getStudioUsers();
    }
    isAwaitingResponse.value = false;
    closeModal();
    await trayStates.refreshData();
  }
};

// Adds collaborators to a personal remote project via the Clustta server.
const addPersonalRemoteCollaborators = async () => {
  const remoteUrl = projectStore.activeProject.remote;
  const projectUri = projectStore.activeProject.uri;
  const registeredEmails = [];
  const resolvedUserIds = [];
  const newUsersList = [];

  for (const user of selectedUsers.value) {
    try {
      const userData = await CollaboratorService.FetchUserByEmail(user.email);
      if (userData?.id) {
        resolvedUserIds.push(userData.id);
        registeredEmails.push(user.email);
      } else {
        newUsersList.push(user);
      }
    } catch (error) {
      newUsersList.push(user);
    }
  }

  if (resolvedUserIds.length > 0) {
    await CollaboratorService.AddCollaborators(remoteUrl, resolvedUserIds);

    for (const email of registeredEmails) {
      try {
        await ProjectService.AddUser(projectUri, email, collaboratorRole.value).then((result)=>{
          console.log(result)
        });
      } catch (error) {
        console.error('Error adding user to project:', error);
      }
    }

    notificationStore.addNotification(t('notifications.usersAddedSuccessfully', { count: resolvedUserIds.length }), "", "success");
  }

  for (const user of newUsersList) {
    try {
      await AuthService.SendInvitationEmail(user.email, 'Clustta', projectStore.activeProject.name || 'Project');
    } catch (error) {
      console.error('Error sending invitation:', error);
      notificationStore.errorNotification(t('notifications.errorSendingInvitation'), error);
    }
  }

  if (newUsersList.length > 0) {
    notificationStore.addNotification(t('notifications.invitationsSent', { count: newUsersList.length }), "", "info");
  }

  await userStore.reloadUsers();
};

// Adds collaborators to a studio project via the studio server.
const addStudioCollaborators = async () => {
  const studioUsersList = [];
  const globalUsers = [];
  const newUsersList = [];

  for (const user of selectedUsers.value) {
    if (user.userType === 'studio') {
      studioUsersList.push(user);
    } else {
      try {
        const emailExists = await AuthService.CheckEmailExists(user.email);
        if (emailExists) {
          globalUsers.push(user);
        } else {
          newUsersList.push(user);
        }
      } catch (error) {
        console.error('Error checking email:', error);
        newUsersList.push(user);
      }
    }
  }

  for (const user of studioUsersList) {
    try {
      await ProjectService.AddUser(projectStore.activeProject.uri, user.email, collaboratorRole.value);
    } catch (error) {
      console.error('Error adding studio user:', error);
      notificationStore.errorNotification(t('notifications.errorAddingStudioUser'), error);
    }
  }

  for (const user of globalUsers) {
    try {
      await StudioService.AddCollaborator(user.email, projectStore.selectedStudio.id, 'user');
      await ProjectService.AddUser(projectStore.activeProject.uri, user.email, collaboratorRole.value);
    } catch (error) {
      console.error('Error adding global user:', error);
      notificationStore.errorNotification(t('notifications.errorAddingGlobalUser'), error);
    }
  }

  for (const user of newUsersList) {
    try {
      await AuthService.SendInvitationEmail(
        user.email, 
        projectStore.selectedStudio.name || 'Clustta Studio',
        projectStore.activeProject.name || 'Project'
      );
    } catch (error) {
      console.error('Error sending invitation:', error);
      notificationStore.errorNotification(t('notifications.errorSendingInvitation'), error);
    }
  }

  const successCount = studioUsersList.length + globalUsers.length;
  const invitationCount = newUsersList.length;

  if (successCount > 0) {
    notificationStore.addNotification(t('notifications.usersAddedSuccessfully', { count: successCount }), "", "success");
  }
  
  if (invitationCount > 0) {
    notificationStore.addNotification(t('notifications.invitationsSent', { count: invitationCount }), "", "info");
  }
};

// Adds a user to the selection list for collaboration.
const addUser = async (user) => {
  const userEmail = user.email.toLowerCase();
  const projectUserEmails = userStore.getProjectCollaborators.map((user) => user.email);
  
  if (selectedUserEmails.value.includes(userEmail) || unregisteredUserEmails.value.includes(userEmail)) {
    return;
  }
  
  if (projectUserEmails.includes(userEmail)) {
    notificationStore.addNotification(t('notifications.userAlreadyInProject'), "", "success");
    return;
  }

  if (isR2Remote.value) {
    try {
      const emailExists = await AuthService.CheckEmailExists(userEmail);
      if (emailExists) {
        selectedUserEmails.value.push(userEmail);
      } else {
        unregisteredUserEmails.value.push(userEmail);
      }
    } catch (error) {
      unregisteredUserEmails.value.push(userEmail);
    }
    return;
  }

  if (!user.userType) {
    selectedUserEmails.value.push(userEmail);
  }
  
  if (!userStore.userCanCreateProject) return;

  if (user.userType !== 'new') {
    selectedUserEmails.value.push(userEmail);
  } else {
    unregisteredUserEmails.value.push(userEmail);
  }
};

// Closes the manage collaborator modal.
const closeModal = () => {
  modals.setModalVisibility('manageCollaboratorModal', false);
};

// Generates a consistent avatar color based on email hash.
const generateAvatarColor = (email) => {
  const colors = ['#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4', '#FFEAA7', '#DDA0DD', '#98D8C8', '#F7DC6F'];
  let hash = 0;
  for (let i = 0; i < email.length; i++) {
    hash = email.charCodeAt(i) + ((hash << 5) - hash);
  }
  return colors[Math.abs(hash) % colors.length];
};

// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Removes a user from the selection list.
const removeUser = (user) => {
  const userEmail = user.email;
  if (user.userType !== 'new') {
    selectedUserEmails.value = selectedUserEmails.value.filter(t => t !== userEmail);
  } else {
    unregisteredUserEmails.value = unregisteredUserEmails.value.filter(t => t !== userEmail);
  }
};

// Selects a role for the collaborators being added.
const selectRole = (role) => {
  collaboratorRole.value = role;
};

// watchers
watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

// lifecycle
onMounted(() => {
  trayStates.tagSearchQuery = '';
  allProjectCollaborators.value = userStore.getProjectCollaborators.map(person => person.email);
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.horizontal-flex {
  padding: 0 .4rem;
}

.notification-area {
  width: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  gap: .5rem;
}
</style>