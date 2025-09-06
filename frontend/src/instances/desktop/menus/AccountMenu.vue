<template>
  <div ref="accountMenu" class="filter-menu-container">
    <!-- Current Account -->
    <div class="current-account-section">
      <div class="account-item current-account">
        <div class="account-avatar">
          <div class="profile-picture" :style="{ backgroundColor: profileColor(currentAccount?.id) }">
            <img v-if="currentAccount?.photo" class="profile-img" :src="currentAccount.photo">
            <img v-else class="profile-img" :src="getAppIcon('person')">
          </div>
        </div>
        <div class="account-info">
          <div class="account-name">{{ currentAccount?.first_name }} {{ currentAccount?.last_name }}</div>
          <div class="account-email">{{ currentAccount?.email }}</div>
        </div>
        <div class="account-status">
          <span class="status-indicator active">●</span>
        </div>
      </div>
    </div>

    <span class="menu-divider"></span>

    <!-- Additional Accounts -->
    <div v-if="additionalAccounts.length > 0" class="additional-accounts-section">
      <div 
        v-for="account in additionalAccounts" 
        :key="account.id" 
        class="account-item-container"
      >
        <div 
          class="account-item clickable-account"
          @click="switchToAccount(account.id)"
        >
          <div class="account-avatar">
            <div class="profile-picture" :style="{ backgroundColor: profileColor(account.id) }">
              <img v-if="account.photo" class="profile-img" :src="account.photo">
              <img v-else class="profile-img" :src="getAppIcon('person')">
            </div>
          </div>
          <div class="account-info">
            <div class="account-name">{{ account.first_name }} {{ account.last_name }}</div>
            <div class="account-email">{{ account.email }}</div>
          </div>
          <div class="account-status">
            <img class="small-icons" :src="getAppIcon('switch')">
          </div>
        </div>
        <div class="account-remove">
          <ActionButton 
            :icon="getAppIcon('trash')" 
            :showLabel="false" 
            :fullWidth="false" 
            :buttonFunction="() => removeAccountFromList(account.id)"
            class="remove-account-btn"
          />
        </div>
      </div>
    </div>

    <span v-if="additionalAccounts.length > 0" class="menu-divider"></span>

    <!-- Actions -->
    <div class="account-actions">
      <ActionButton 
        :icon="getAppIcon('person-plus')" 
        :showLabel="true" 
        :fullWidth="true" 
        label="Add Account"
        :buttonFunction="addAccount" 
      />
      
      <ActionButton 
        :icon="getAppIcon('cog')" 
        :showLabel="true" 
        :fullWidth="true" 
        label="Account Settings"
        :buttonFunction="openAccountSettings" 
      />
      
      <ActionButton 
        :icon="getAppIcon('logout')" 
        :showLabel="true" 
        :fullWidth="true" 
        label="Sign Out"
        :buttonFunction="signOutCurrentAccount" 
      />
    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted } from 'vue';
import { useUserStore } from '@/stores/users';
import { useProjectStore } from '@/stores/projects';
import { useThemeStore } from '@/stores/theme';
import { useTrayStates } from '@/stores/TrayStates';
import { useMenu } from '@/stores/menu';
import { useStageStore } from '@/stores/stages';
import { useIconStore } from '@/stores/icons';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useAccountStore } from '@/stores/accounts';
import { AccountService, AuthService } from "@/../bindings/clustta/services";

// Components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// Stores
const userStore = useUserStore();
const projectStore = useProjectStore();
const themeStore = useThemeStore();
const trayStates = useTrayStates();
const menu = useMenu();
const stage = useStageStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const accountStore = useAccountStore();

// Refs
const accountMenu = ref(null);

// Computed
const currentAccount = computed(() => accountStore.currentAccount?.user || userStore.user);

// Use account store data for additional accounts
const additionalAccounts = computed(() => accountStore.additionalAccounts);

// Methods
const getAppIcon = (iconName) => {
  const icon = iconStore.getAppIcon(iconName);
  return icon;
};

const profileColor = (uuid) => {
  if (!uuid) return '#000000';
  const parts = uuid.split('-');
  return '#' + (parts[0] || '000000').substring(0, 6);
};

const switchToAccount = async (accountId) => {
  try {
    stage.operationActive = true;
    
    // Switch account using the store method with UI updates
    await accountStore.switchToAccount(accountId, {
      userStore,
      projectStore,
      trayStates,
      themeStore,
      notificationStore,
      stageStore: stage
    });
    
    menu.hideContextMenu();
  } catch (error) {
    // Error handling is done in the store method
  } finally {
    stage.operationActive = false;
  }
};

const addAccount = () => {
  try {
    console.log('Adding new account');
    
    // Set flag to indicate we're adding an additional account
    accountStore.isAdditionalAccount = true;
    
    modals.setModalVisibility('loginModal', true);
    menu.hideContextMenu();
  } catch (error) {
    notificationStore.errorNotification("Add Account Failed", error);
  }
};

const removeAccountFromList = async (accountId) => {
  try {
    console.log('Removing account:', accountId);
    
    // Remove the account using the store
    await accountStore.removeAccount(accountId);
    
    notificationStore.addNotification("Account Removed", "Account has been successfully removed", "success");
  } catch (error) {
    console.error('Remove account error:', error);
    notificationStore.errorNotification("Remove Failed", error.message || 'Unable to remove account');
  }
};

const openAccountSettings = () => {
  try {
    console.log('Opening account settings');
    stage.setStageVisibility('account', true);
    menu.hideContextMenu();
  } catch (error) {
    notificationStore.errorNotification("Settings Failed", error);
  }
};

const signOutCurrentAccount = async () => {
  try {
    // Get current user ID before signing out
    const currentUserId = userStore.user?.id;
    menu.hideContextMenu();
    
    if (!currentUserId) {
      throw new Error('No active user to sign out');
    }

    // Remove the current account using the store
    await accountStore.removeAccount(currentUserId);
    
    // Reset stores after successful signout
    userStore.$reset();
    projectStore.$reset();
    trayStates.$reset();
    
    // Check if there are other accounts available
    const accountCount = await accountStore.getAccountCount();
    
    if (accountCount > 0) {
      // The store will automatically update after removal, get the new active account
      const activeAccount = accountStore.activeAccount;
      if (activeAccount && activeAccount.user) {
        userStore.user = activeAccount.user;
        userStore.isUserAuthenticated = true;
        
        // Refresh application data for the new active user
        await themeStore.initializeTheme();
        await projectStore.loadStudios();
        await projectStore.loadProjects();
        trayStates.refreshData();
        
        notificationStore.addNotification("Account Switched", `Switched to ${activeAccount.user.first_name} ${activeAccount.user.last_name}`);
      }
    } else {
      // No accounts left, reset to unauthenticated state
      userStore.user = null;
      userStore.isUserAuthenticated = false;
      // modals.setModalVisibility('loginModal', true);
      notificationStore.addNotification("Signed Out", "All accounts signed out");
    }
    
    console.log('Signed out current account');
    menu.hideContextMenu();
  } catch (error) {
    console.error('Sign out error:', error);
    notificationStore.errorNotification("Sign Out Failed", error.message || 'Unable to sign out');
  }
};

onMounted(() => {
  // Ensure accounts are loaded when component mounts
  if (!accountStore.isLoaded) {
    accountStore.initialize();
  }
  console.log('Account menu mounted');
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.filter-menu-container {
  min-width: 280px;
  padding: 0.5rem;
}

.current-account-section {
  margin-bottom: 0.5rem;
}

.additional-accounts-section {
  margin-bottom: 0.5rem;
}

.account-actions {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.account-item-container {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.25rem;
}

.account-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem;
  border-radius: var(--normal-radius);
  transition: background-color 0.2s ease;
  flex: 1;
}

.account-remove {
  flex-shrink: 0;
}

.remove-account-btn {
  padding: 0.25rem !important;
  min-width: 32px !important;
  height: 32px !important;
  background-color: var(--red) !important;
  border: none !important;
}

.remove-account-btn:hover {
  background-color: var(--dark-red) !important;
}

.current-account {
  background-color: var(--dark-glass);
  border: 1px solid var(--steel);
}

.clickable-account {
  cursor: pointer;
}

.clickable-account:hover {
  background-color: var(--dark-glass);
  outline: 1px solid var(--steel);
}

.account-avatar {
  flex-shrink: 0;
}

.profile-picture {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.profile-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.account-info {
  flex: 1;
  min-width: 0;
}

.account-name {
  font-weight: 500;
  font-size: 0.875rem;
  color: var(--white);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.account-email {
  font-size: 0.75rem;
  color: var(--white);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.account-status {
  flex-shrink: 0;
}

.status-indicator {
  font-size: 0.75rem;
}

.status-indicator.active {
  color: #22c55e; /* Green for active */
}

.status-indicator.inactive {
  color: var(--light-grey); /* Grey for inactive */
}

.menu-divider {
  display: block;
  height: 1px;
  background-color: var(--steel);
  margin: 0.5rem 0;
  border: none;
}
</style>
