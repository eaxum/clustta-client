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
        <div class="new-user-notification">
          <div class="notification-icon">
            <img class="small-icons no-filter" :src="getAppIcon('mail')" alt="Invitation" />
          </div>
          <div class="notification-content">
            <div class="notification-title">Invitation Required</div>
            <div class="notification-message">
              {{ newUsers.length }} user{{ newUsers.length > 1 ? 's' : '' }} 
              {{ newUsers.length > 1 ? 'are' : 'is' }}n't on Clustta yet. 
              {{ newUsers.length > 1 ? 'They' : 'This user' }} will be sent an invite to signup.
            </div>
          </div>
        </div>
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
import { ref, onMounted, computed, watchEffect } from 'vue';

// services
import { StudioService, AuthService } from "@/../bindings/clustta/services";

// store/state imports
import { useTrayStates } from '@/stores/TrayStates';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useUserStore } from '@/stores/users';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useMenu } from '@/stores/menu';
import { useStudioStore } from '@/stores/studio';
import { useIconStore } from '@/stores/icons';

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import CollaboratorSuggestions from '@/instances/common/components/CollaboratorSuggestions.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';

// stores
const trayStates = useTrayStates();
const studioStore = useStudioStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const userStore = useUserStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const menu = useMenu();

// vars
let placeholder = 'Enter names or emails to add to studio';

const selectedUserEmails = ref([]);
const unregisteredUserEmails = ref([]);
const modalContainer = ref(null);
const studioRoles = ref(["admin", "user"]);


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

  // return studioUsers
  return [...studioUsers, ...registeredUsers, ...unregisteredUsers];
});

// Computed property to track completely new users who need invitations
const newUsers = computed(() => {
  return selectedUsers.value.filter(user => user.userType === 'new');
});

const removeUser = (user) => {
    const userEmail = user.email;
    if(user.userType !== 'new'){
      selectedUserEmails.value = selectedUserEmails.value.filter(t => t !== userEmail);
    } else {
      unregisteredUserEmails.value = unregisteredUserEmails.value.filter(t => t !== userEmail);
    }
};

const addUser = (user) => {
  const userEmail = user.email.toLowerCase();
  const studioUserEmails = studioStore.studioUsers.map((user) => user.email);
  if (selectedUserEmails.value.includes(userEmail)) {
    return
  }
  if (studioUserEmails.includes(userEmail)) {
    notificationStore.addNotification(`User is already in the studio.`, "", "success");
    return
  }
  else {
    if(!user.userType){
      selectedUserEmails.value.push(userEmail);
    }
    
    if(!userStore.userCanCreateProject) return 

    if(user.userType !== 'new'){
      selectedUserEmails.value.push(userEmail);
    } else {
      unregisteredUserEmails.value.push(userEmail);
    }
  }
};

// Generate a color for avatar based on email
const generateAvatarColor = (email) => {
  const colors = ['#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4', '#FFEAA7', '#DDA0DD', '#98D8C8', '#F7DC6F'];
  let hash = 0;
  for (let i = 0; i < email.length; i++) {
    hash = email.charCodeAt(i) + ((hash << 5) - hash);
  }
  return colors[Math.abs(hash) % colors.length];
};

const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon;
};

// refs
const studioCollaboratorRole = ref(studioRoles.value[0]);
const isAwaitingResponse = ref(false);

// methods
const closeModal = () => {
  modals.setModalVisibility('addCollaboratorModal', false);
};

const selectRole = (role) => {
  studioCollaboratorRole.value = role;
};

const addCollaborators = async () => {
  isAwaitingResponse.value = true;

  try {
    // Categorize users by type
    const existingUsers = [];
    const globalUsers = [];
    const newUsers = [];

    for (const user of selectedUsers.value) {
      if (user.userType === 'studio') {
        // User is already in studio, skip
        continue;
      } else {
        // For email-based users, check if they exist globally
        try {
          const emailExists = await AuthService.CheckEmailExists(user.email);
          if (emailExists) {
            globalUsers.push(user);
          } else {
            newUsers.push(user);
          }
        } catch (error) {
          console.error('Error checking email:', error);
          newUsers.push(user); // Treat as new user if API fails
        }
      }
    }

    // Process global users (add to studio only)
    for (const user of globalUsers) {
      try {
        await StudioService.AddCollaborator(user.email, projectStore.selectedStudio.id, studioCollaboratorRole.value);
      } catch (error) {
        console.error('Error adding global user to studio:', error);
        notificationStore.errorNotification("Error Adding User to Studio", error);
      }
    }

    // Process new users (send invitation emails)
    for (const user of newUsers) {
      try {
        await AuthService.SendInvitationEmail(
          user.email, 
          projectStore.selectedStudio.name || 'Clustta Studio',
          '' // No project name since we're only adding to studio
        );
      } catch (error) {
        console.error('Error sending invitation:', error);
        notificationStore.errorNotification("Error Sending Invitation", error);
      }
    }

    const successCount = globalUsers.length;
    const invitationCount = newUsers.length;

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

watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

onMounted(() => {
  trayStates.tagSearchQuery = '';
})

</script>

<style scoped>
@import "@/assets/desktop.css";

.horizontal-flex {
  /* background-color: goldenrod; */
  padding: 0 .4rem;
}

.notification-area{
  width: 100%;
  display: flex;
  flex-direction: column;
  /* background-color: crimson; */
  overflow: hidden;
  gap: .5rem;
}

.add-category {

  display: flex;
  gap: .5rem;
  flex-direction: row;
  /* background-color: chocolate; */
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
  /* margin-top: 1rem; */
  display: flex;
  /* align-items: center; */
  flex-direction: column;
  /* justify-content: space-between; */
  gap: 1rem;
  /* background-color: darkkhaki; */
  color: var(--white);
  width: 98%;
}

.category-list {
  box-sizing: border-box;
  /* margin-top: 1rem; */
  display: flex;
  padding: .5rem;
  align-items: center;
  flex-direction: column;
  /* justify-content: space-between; */
  gap: .2rem;
  /* background-color: rgb(57, 122, 108); */

  background-color: rgba(0, 0, 0, 0.144);
  height: 290px;
  /* height: 90%; */
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
  /* background-color: rgba(0, 0, 0, 0.295); */
}

.category-item {
  color: var(--white);
  /* margin-top: 1rem; */
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .5rem;
  width: 100%;
  border-bottom: 1px solid rgba(255, 255, 255, 0.096);
  height: max-content;
  padding: .2rem;

  /* background-color: greenyellow; */
}

.category-item-actions {

  /* background-color: red; */
  display: flex;

}

/* Studio notification styles */
.studio-notification {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  width: 100%;
  padding: 0.75rem;
  background-color: rgba(255, 165, 0, 0.1);
  border: 1px solid #FFA500;
  border-radius: var(--small-radius);
  /* margin: 0.5rem 0; */
}

/* New user notification styles */
.new-user-notification {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  width: 100%;
  padding: 0.75rem;
  background-color: rgba(45, 156, 219, 0.1);
  border: 1px solid #2D9CDB;
  border-radius: var(--small-radius);
  /* margin: 0.5rem 0; */
}

.notification-icon {
  flex-shrink: 0;
  margin-top: 0.1rem;
}

.notification-icon img {
  width: 20px;
  height: 20px;
}

.studio-notification .notification-icon img {
  filter: brightness(0) saturate(100%) invert(69%) sepia(100%) saturate(2582%) hue-rotate(14deg) brightness(101%) contrast(101%);
}

.new-user-notification .notification-icon img {
  filter: brightness(0) saturate(100%) invert(47%) sepia(74%) saturate(6614%) hue-rotate(201deg) brightness(95%) contrast(91%);
}

.notification-content {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  width: 100%;
}

.notification-title {
  font-weight: 400;
  color: var(--white);
  font-size: 14px;
}

.notification-message {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.8);
  line-height: 1.4;
}
</style>