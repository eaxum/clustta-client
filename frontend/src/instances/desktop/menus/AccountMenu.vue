<template>
  <div ref="accountMenu" class="filter-menu-container">
    <!-- Current Account -->
    <div class="current-account-section">
      <div class="account-item-container">
        <div class="account-item current-account">
          <div class="account-avatar">
            <div class="profile-picture" :style="{ backgroundColor: profileColor(currentAccount?.id) }">
              <img v-if="currentAccount?.photo" class="profile-img" :src="currentAccount.photo">
              <img v-else class="profile-img" :src="generateAvatar(currentAccount?.id)">
            </div>
          </div>
          <div class="account-info">
            <div class="account-name">{{ currentAccount?.first_name }} {{ currentAccount?.last_name }}</div>
            <div v-if="!isOfflineMode" class="account-email">{{ currentAccount?.email }}</div>
          </div>
          <div class="account-status">
            <span v-if="isOfflineMode" class="status-indicator offline" v-tooltip="$t('menus.offlineModeTooltip')">●</span>
            <span v-else class="status-indicator active">●</span>
          </div>
        </div>
        <div v-if="!isOfflineMode" class="account-remove">
          <ActionButton :icon="getAppIcon('cog')" :showLabel="false" :fullWidth="false" v-tooltip="$t('menus.accountSettings')" :buttonFunction="openAccountSettings" />
          <ActionButton :icon="getAppIcon('logout')" :showLabel="false" :fullWidth="false" v-tooltip="$t('common.signOut')" :buttonFunction="signOutCurrentAccount" />
        </div>
      </div>

      <ActionButton v-if="isOfflineMode" :icon="getAppIcon('login')" :showLabel="true" :fullWidth="true" :label="$t('common.signIn')" :buttonFunction="signInFromOffline" />
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
              <img v-else class="profile-img" :src="generateAvatar(account.id)">
            </div>
          </div>
          <div class="account-info">
            <div class="account-name">{{ account.first_name }} {{ account.last_name }}</div>
            <div class="account-email">{{ account.email }}</div>
          </div>
        </div>
        <div class="account-remove">
          <ActionButton :icon="getAppIcon('logout')" :showLabel="false" :fullWidth="false" v-tooltip="$t('menus.removeAccount')" :buttonFunction="() => removeAccountFromList(account.id)" />
        </div>
      </div>
    </div>

    <span v-if="additionalAccounts.length > 0" class="menu-divider"></span>

    <!-- Actions -->
    <div class="account-actions">
      <ActionButton v-if="!isOfflineMode" :icon="getAppIcon('person-plus')" :showLabel="true" :fullWidth="true" :label="$t('menus.addAccount')" :buttonFunction="addAccount" />
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { resetStoreInitialization } from '@/router';
import { generateAvatar } from '@/lib/avatar';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// services
import { AccountService, AuthService } from "@/services";

// stores
import { useAccountStore } from '@/stores/accounts';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';
import { useThemeStore } from '@/stores/theme';
import { useTrayStates } from '@/stores/TrayStates';
import { useUserStore } from '@/stores/users';

const { t } = useI18n();
const accountStore = useAccountStore();
const iconStore = useIconStore();
const menu = useMenu();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const router = useRouter();
const stage = useStageStore();
const themeStore = useThemeStore();
const trayStates = useTrayStates();
const userStore = useUserStore();

// refs
const accountMenu = ref(null);

// computed
// Returns the list of additional accounts.
const additionalAccounts = computed(() => accountStore.additionalAccounts);

// Returns the current active account.
const currentAccount = computed(() => accountStore.currentAccount?.user || userStore.user);

// Checks if the app is in offline mode.
const isOfflineMode = computed(() => accountStore.isOfflineMode);

// methods
// Opens the add account modal.
const addAccount = () => {
  try {
    accountStore.isAdditionalAccount = true;
    modals.setModalVisibility('loginModal', true);
    menu.hideContextMenu();
  } catch (error) {
    notificationStore.errorNotification(t('notifications.addAccountFailed'), error);
  }
};

// Returns the icon path for a given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Opens the account settings view.
const openAccountSettings = () => {
  try {
    stage.setStageVisibility('account', true);
    menu.hideContextMenu();
  } catch (error) {
    notificationStore.errorNotification(t('notifications.settingsFailed'), error);
  }
};

// Generates a profile color based on UUID.
const profileColor = (uuid) => {
  if (!uuid) return '#000000';
  const parts = uuid.split('-');
  return '#' + (parts[0] || '000000').substring(0, 6);
};

// Removes an account from the list.
const removeAccountFromList = async (accountId) => {
  try {
    await accountStore.removeAccount(accountId);
    notificationStore.addNotification(t('notifications.accountRemoved'), t('notifications.accountRemovedDesc'), "success");
  } catch (error) {
    console.error('Remove account error:', error);
    notificationStore.errorNotification(t('notifications.removeFailed'), error.message || t('notifications.unableToRemoveAccount'));
  }
};

// Signs in from offline mode.
const signInFromOffline = async () => {
  try {
    const offlineUserId = 'offline-user';
    await accountStore.removeAccount(offlineUserId);
    
    userStore.$reset();
    projectStore.$reset();
    trayStates.$reset();
    resetStoreInitialization();
    
    menu.hideContextMenu();
    router.push('/auth/login');
  } catch (error) {
    console.error('Sign in from offline error:', error);
    menu.hideContextMenu();
    router.push('/auth/login');
  }
};

// Signs out the current account.
const signOutCurrentAccount = async () => {
  try {
    const currentUserId = userStore.user?.id;
    menu.hideContextMenu();
    
    if (!currentUserId) {
      throw new Error('No active user to sign out');
    }

    await accountStore.removeAccount(currentUserId);
    
    userStore.$reset();
    projectStore.$reset();
    trayStates.$reset();
    
    const accountCount = await accountStore.getAccountCount();
    
    if (accountCount > 0) {
      const activeAccount = accountStore.activeAccount;
      if (activeAccount && activeAccount.user) {
        await accountStore.switchToAccount(activeAccount.user.id);
      }
    } else {
      userStore.user = null;
      userStore.isUserAuthenticated = false;
      notificationStore.addNotification(t('notifications.signedOut'), t('notifications.allAccountsSignedOut'), "info");
      resetStoreInitialization();
      router.push('/auth/login');
    }
    
    menu.hideContextMenu();
  } catch (error) {
    console.error('Sign out error:', error);
    notificationStore.errorNotification(t('notifications.signOutFailed'), error.message || t('notifications.unableToSignOut'));
  }
};

// Switches to a different account.
const switchToAccount = async (accountId) => {
  try {
    stage.operationActive = true;
    
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
    console.error('Switch account error:', error);
  } finally {
    stage.operationActive = false;
  }
};

// lifecycle hooks
onMounted(() => {
  if (!accountStore.isLoaded) {
    accountStore.initialize();
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.filter-menu-container {
  min-width: 250px;
  max-width: 250px;
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
  min-width: 0;
  overflow: hidden;
}

.account-remove {
  display: flex;
  gap: 0.25rem;
  flex-shrink: 0;
  max-width: 0;
  opacity: 0;
  overflow: hidden;
  transition: max-width 0.2s ease, opacity 0.2s ease;
}

.account-item-container:hover .account-remove {
  max-width: 80px;
  opacity: 1;
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
  border: 1px solid var(--surface-3);
}

.clickable-account {
  cursor: pointer;
}

.clickable-account:hover {
  background-color: var(--dark-glass);
  outline: 1px solid var(--surface-3);
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
  color: var(--text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.account-email {
  font-size: 0.75rem;
  color: var(--text);
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

.status-indicator.offline {
  color: #f59e0b; /* Amber/orange for offline */
}

.offline-banner {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  margin-top: 0.5rem;
  background: rgba(245, 158, 11, 0.15);
  border: 1px solid rgba(245, 158, 11, 0.3);
  border-radius: 6px;
  font-size: 0.75rem;
  color: #fbbf24;
}

.offline-banner .small-icons {
  width: 14px;
  height: 14px;
  filter: none;
  opacity: 0.9;
}

.menu-divider {
  display: block;
  height: 1px;
  background-color: var(--surface-3);
  margin: 0.5rem 0;
  border: none;
}
</style>


