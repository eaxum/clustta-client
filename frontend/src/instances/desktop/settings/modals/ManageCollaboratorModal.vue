<template>

  <div class="modal-container" ref="modalContainer" v-stop-propagation v-esc="closeModal">
    <HeaderArea :title="title" :icon="'person-plus'" :showSearch="showSearch" />

    <div class="general-container">

      <div class="horizontal-flex">
        <CollaboratorSuggestions :allowMultipleEntries="true" :placeholder="placeholder" :selectedItems="selectedUsers" :allItems="studioUsers"
          @tagAdded="addUser" @tagRemoved="removeUser" />
      </div>

      <div class="horizontal-flex">
        <DropDownBox :items="userStore.getRolesNames" :onSelect="selectRole"
          :selectedItem="collaboratorRole" :placeHolder="'None'" :fullWidth="true" />
      </div>

      <!-- Notification section for non-studio users -->
      <div class="notification-area">
      <div v-if="nonStudioUsers.length > 0" class="horizontal-flex">
        <NotificationBox 
          type="warning"
          :icon="getAppIcon('alert')"
          iconAlt="Alert"
          title="Studio Addition Required"
          :message="`${nonStudioUsers.length} user${nonStudioUsers.length > 1 ? 's' : ''} ${nonStudioUsers.length > 1 ? 'are' : 'is'} not currently in this studio. ${nonStudioUsers.length > 1 ? 'They' : 'This user'} will be added to the studio first, then to the project.`"
        />
      </div>

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
import { ref, onMounted, computed, watchEffect } from 'vue';

// services
import { ProjectService, StudioService, AuthService } from "@/services";

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
import NotificationBox from '@/instances/common/components/NotificationBox.vue';

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
let title = 'Manage Collaborators';
let showSearch = false;
let placeholder = 'Name or Email, separated by commas';

const selectedUserEmails = ref([]);
const unregisteredUserEmails = ref([]);
const modalContainer = ref(null);

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

const selectedUsers = computed(() => {
  const selectedStudio = projectStore.selectedStudio;
  // const studioUsers = selectedStudio.Users
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

// Computed property to track users that will be added to studio first
const nonStudioUsers = computed(() => {
  return selectedUsers.value.filter(user => user.userType === 'user');
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
  const projectUserEmails = userStore.getProjectCollaborators.map((user) => user.email);
  if (selectedUserEmails.value.includes(userEmail)) {
    return
  }
  if (projectUserEmails.includes(userEmail)) {
    notificationStore.addNotification(`User is already in the project.`, "", "success");
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
const newCollaboratorEmail = ref('');
const newCollaborators = ref([]);
const collaboratorRole = ref(userStore.getRolesNames[userStore.getRolesNames.length - 1]);
const allProjectCollaborators = ref([]);
const isAwaitingResponse = ref(false);

// computed props
const emit = defineEmits(['addedCollaborator']);
// const emailInvalid = computed(() => { return !emailRegex.test(newCollaboratorEmail.value) });
const emailInvalid = computed(() => { return newCollaboratorEmail.value === "" });
const projectCollaborators = computed(() => {
  const allCollaborators = allProjectCollaborators.value;
  return allCollaborators.filter(item => !newCollaborators.value.includes(item));
});

// methods
const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

const closeModal = () => {
  modals.setModalVisibility('manageCollaboratorModal', false);
};

const selectRole = (role) => {
  collaboratorRole.value = role;
};

const addCollaborators = async () => {
  isAwaitingResponse.value = true;

  try {
    // Categorize users by type
    const studioUsers = [];
    const globalUsers = [];
    const newUsers = [];

    for (const user of selectedUsers.value) {
      if (user.userType === 'studio') {
        studioUsers.push(user);
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

    // Process studio users (add directly to project)
    for (const user of studioUsers) {
      try {
        await ProjectService.AddUser(projectStore.activeProject.uri, user.email, collaboratorRole.value);
      } catch (error) {
        console.error('Error adding studio user:', error);
        notificationStore.errorNotification("Error Adding Studio User", error);
      }
    }

    // Process global users (add to studio first, then to project)
    for (const user of globalUsers) {
      try {
        // Add to studio first - use 'user' role instead of 'member'
        await StudioService.AddCollaborator(user.email, projectStore.selectedStudio.id, 'user');
        // Then add to project
        await ProjectService.AddUser(projectStore.activeProject.uri, user.email, collaboratorRole.value);
      } catch (error) {
        console.error('Error adding global user:', error);
        notificationStore.errorNotification("Error Adding Global User", error);
      }
    }

    // Process new users (send invitation emails)
    for (const user of newUsers) {
      try {
        await AuthService.SendInvitationEmail(
          user.email, 
          projectStore.selectedStudio.name || 'Clustta Studio',
          projectStore.activeProject.name || 'Project'
        );
      } catch (error) {
        console.error('Error sending invitation:', error);
        notificationStore.errorNotification("Error Sending Invitation", error);
      }
    }

    const successCount = studioUsers.length + globalUsers.length;
    const invitationCount = newUsers.length;

    if (successCount > 0) {
      notificationStore.addNotification(`${successCount} user(s) added successfully.`, "", "success");
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

const handleEnterKey = (event) => {

};

watchEffect(() => {
  if (modalContainer.value) {
    menu.clickOutsideMask = modalContainer.value;
  }
});

onMounted(() => {
  trayStates.tagSearchQuery = '';
  allProjectCollaborators.value = userStore.getProjectCollaborators.map(person => person.email)
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
</style>