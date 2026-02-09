<template>
  <div class="modal-container" ref="modalContainer" v-stop-propagation v-esc="closeModal">
    <HeaderArea :title="'Add Studio Collaborator'" :icon="getAppIcon('person-plus')" :showSearch="false" />
    <div class="general-container">

      <div class="horizontal-flex">
        <CollaboratorSuggestions :allowMultipleEntries="true" :placeholder="placeholder" :selectedItems="selectedUsers" 
          @tagAdded="addUser" @tagRemoved="removeUser"  />
      </div>

      <div class="horizontal-flex">
        <DropDownBox :items="studioRoles" :onSelect="selectRole"
          :selectedItem="studioCollaboratorRole" :placeHolder="'None'" :fullWidth="true" />
      </div>

      <!-- Notification section for non-studio users -->
      <div class="notification-area">

      <!-- Notification section for new users -->
      <div v-if="newUsers.length > 0" class="horizontal-flex">
        <NotificationBox 
          type="invitation"
          :icon="getAppIcon('mail')"
          iconAlt="Invitation"
          title="Invitation Required"
          :message="`${newUsers.length} user${newUsers.length > 1 ? 's' : ''} ${newUsers.length > 1 ? 'are' : 'is'}n't on Clustta yet. ${newUsers.length > 1 ? 'They' : 'This user'} will be sent an invite to signup.`"
        />
      </div>
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="'Cancel'" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Add'" :buttonFunction="addCollaborators" :isActive="!!selectedUsers.length"
          :loading="isAwaitingResponse" />
      </div>

    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref, watchEffect } from 'vue';

// components
import CollaboratorSuggestions from '@/instances/common/components/CollaboratorSuggestions.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import NotificationBox from '@/instances/common/components/NotificationBox.vue';

// services
import { AuthService, StudioService } from "@/services";

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStudioStore } from '@/stores/studio';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';

const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const studioStore = useStudioStore();
const trayStates = useTrayStates();
const userStore = useUserStore();

// refs
const isAwaitingResponse = ref(false);
const modalContainer = ref(null);
const selectedUserEmails = ref([]);
const studioCollaboratorRole = ref('admin');
const studioRoles = ref(["admin", "user"]);
const unregisteredUserEmails = ref([]);

// constants
const placeholder = 'Enter names or emails to add to studio';

// computed
// Tracks new users who need invitation emails.
const newUsers = computed(() => {
  return selectedUsers.value.filter(user => user.userType === 'new');
});

// Aggregates selected users from different sources.
const selectedUsers = computed(() => {
  const studioUsers = studioStore.studioUsers
    .map(user => ({
      ...user,
      full_name: `${user.first_name} ${user.last_name}`,
      avatarColor: userStore.userProfileColor(user.id),
      userType: 'studio'
    }))
    .filter((user) => selectedUserEmails.value.includes(user.email));

  const registeredUsers = selectedUserEmails.value
    .filter(id => !studioStore.studioUsers.some(user => (user.email) === id))
    .map(email => {
      return {
        id: email,
        email: email,
        full_name: email,
        avatarColor: generateAvatarColor(email),
        userType: 'user'
      };
    });

  const unregisteredUsers = unregisteredUserEmails.value
    .filter(id => !studioStore.studioUsers.some(user => (user.email) === id))
    .map(email => {
      return {
        id: email,
        email: email,
        full_name: email,
        avatarColor: generateAvatarColor(email),
        userType: 'new'
      };
    });

  return [...studioUsers, ...registeredUsers, ...unregisteredUsers];
});

// methods
// Adds all selected collaborators to the studio.
const addCollaborators = async () => {
  isAwaitingResponse.value = true;

  try {
    const globalUsers = [];
    const newUsersList = [];

    for (const user of selectedUsers.value) {
      if (user.userType === 'studio') {
        continue;
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

    for (const user of globalUsers) {
      try {
        await StudioService.AddCollaborator(user.email, projectStore.selectedStudio.id, studioCollaboratorRole.value);
      } catch (error) {
        console.error('Error adding global user to studio:', error);
        notificationStore.errorNotification("Error Adding User to Studio", error);
      }
    }

    for (const user of newUsersList) {
      try {
        await AuthService.SendInvitationEmail(
          user.email, 
          projectStore.selectedStudio.name || 'Clustta Studio',
          ''
        );
      } catch (error) {
        console.error('Error sending invitation:', error);
        notificationStore.errorNotification("Error Sending Invitation", error);
      }
    }

    const successCount = globalUsers.length;
    const invitationCount = newUsersList.length;

    if (successCount > 0) {
      notificationStore.addNotification(`${successCount} user(s) added to studio successfully.`, "", "success");
    }
    
    if (invitationCount > 0) {
      notificationStore.addNotification(`${invitationCount} invitation(s) sent.`, "", "info");
    }

  } catch (error) {
    console.error('Error in addCollaborators:', error);
    notificationStore.errorNotification("Error Adding Users", error);
  } finally {
    await studioStore.getStudioUsers();
    isAwaitingResponse.value = false;
    closeModal();
    await trayStates.refreshData();
  }
};

// Adds a user to the selected list.
const addUser = (user) => {
  const userEmail = user.email.toLowerCase();
  const studioUserEmails = studioStore.studioUsers.map((user) => user.email);
  if (selectedUserEmails.value.includes(userEmail)) {
    return;
  }
  if (studioUserEmails.includes(userEmail)) {
    notificationStore.addNotification(`User is already in the studio.`, "", "success");
    return;
  } else {
    if (!user.userType) {
      selectedUserEmails.value.push(userEmail);
    }
    
    if (!userStore.userCanCreateProject) return;

    if (user.userType !== 'new') {
      selectedUserEmails.value.push(userEmail);
    } else {
      unregisteredUserEmails.value.push(userEmail);
    }
  }
};

// Closes the modal.
const closeModal = () => {
  modals.setModalVisibility('addCollaboratorModal', false);
};

// Generates a color for avatar based on email.
const generateAvatarColor = (email) => {
  const colors = ['#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4', '#FFEAA7', '#DDA0DD', '#98D8C8', '#F7DC6F'];
  let hash = 0;
  for (let i = 0; i < email.length; i++) {
    hash = email.charCodeAt(i) + ((hash << 5) - hash);
  }
  return colors[Math.abs(hash) % colors.length];
};

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Removes a user from the selected list.
const removeUser = (user) => {
  const userEmail = user.email;
  if (user.userType !== 'new') {
    selectedUserEmails.value = selectedUserEmails.value.filter(t => t !== userEmail);
  } else {
    unregisteredUserEmails.value = unregisteredUserEmails.value.filter(t => t !== userEmail);
  }
};

// Selects a role from dropdown.
const selectRole = (role) => {
  studioCollaboratorRole.value = role;
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

