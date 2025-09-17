import { defineStore } from 'pinia';
import { AccountService, SettingsService } from "@/../bindings/clustta/services";
import { useUserStore } from '@/stores/users';
import { useProjectStore } from '@/stores/projects';
import { useTrayStates } from '@/stores/TrayStates';
import { useThemeStore } from '@/stores/theme';
import { useNotificationStore } from '@/stores/notifications';
import { useStageStore } from '@/stores/stages';
import { useDesktopModalStore } from '@/stores/desktopModals';

export const useAccountStore = defineStore('accounts', {
  state: () => ({
    accounts: [],           // All accounts (active + additional)
    activeAccount: null,    // Currently active account
    isLoaded: false,       // Whether accounts have been initially loaded
    isLoading: false,      // Loading state for operations
    lastUpdated: null,     // Timestamp for cache invalidation
    isAdditionalAccount: false // Flag to indicate when adding additional account vs first login
  }),

  getters: {
    // Get all accounts except the active one
    additionalAccounts: (state) => {
      const activeAccountId = state.activeAccount?.user?.id;
      return state.accounts.filter(account => account.id !== activeAccountId);
    },

    // Get the current active account
    currentAccount: (state) => {
      const activeAccountId = state.activeAccount?.user?.id;
      return state.accounts.find(account => account.id === activeAccountId) || state.activeAccount;
    },

    // Get total account count
    accountCount: (state) => {
      return state.accounts.length;
    },

    // Check if we have multiple accounts
    hasMultipleAccounts: (state) => {
      return state.accounts.length > 1;
    }
  },

  actions: {
    // Initial load of all accounts on app startup
    async loadAccounts() {
      if (this.isLoading) return; // Prevent multiple simultaneous loads
      
      this.isLoading = true;
      try {
        
        // Get all stored accounts and active account
        const [accountsData, activeAccountData] = await Promise.all([
          AccountService.GetAllAccounts(),
          AccountService.GetActiveAccount()
        ]);


        // Convert to array format expected by the UI
        const accountArray = Object.values(accountsData).map(token => ({
          id: token.user.id,
          first_name: token.user.first_name,
          last_name: token.user.last_name,
          email: token.user.email,
          username: token.user.username,
          photo: token.user.photo,
          session_id: token.session_id,
          last_used: new Date(), // Could be enhanced to track actual last used time
          expires_at: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000), // 30 days from now
          is_active: token.user.id === activeAccountData?.user?.id
        }));

        this.accounts = accountArray;
        this.activeAccount = activeAccountData;
        this.isLoaded = true;
        this.lastUpdated = new Date();

      } catch (error) {
        this.accounts = [];
        this.activeAccount = null;
      } finally {
        this.isLoading = false;
      }
    },

    // Add new account after login
    async addAccount(token) {
      try {
        await AccountService.AddAccount(token);
        
        // Refresh accounts to get updated list
        await this.refreshAccounts();
        
        return true;
      } catch (error) {
        return false;
      }
    },

    // Remove account
    async removeAccount(userId) {
      try {
        await AccountService.RemoveAccount(userId);
        
        // Refresh accounts to get updated list
        await this.refreshAccounts();
        
        return true;
      } catch (error) {
        return false;
      }
    },

    // Switch active account
    async switchAccount(userId) {
      try {
        
        // Check if account exists in our store
        const targetAccount = this.accounts.find(account => account.id === userId);
        if (!targetAccount) {
          throw new Error(`Account with ID ${userId} not found in stored accounts`);
        }

        // Switch in backend
        await AccountService.SwitchAccount(userId);
        
        // Update local state immediately for better UX
        this.accounts.forEach(account => {
          account.is_active = account.id === userId;
        });

        // Get fresh active account data
        const activeAccountData = await AccountService.GetActiveAccount();
        this.activeAccount = activeAccountData;

        return activeAccountData;
      } catch (error) {
        // Refresh accounts to ensure consistency if switch failed
        await this.refreshAccounts();
        throw error;
      }
    },

    // Complete account switch with UI updates (call from components)
    async switchToAccount(userId) {
      try {
        const userStore = useUserStore();
        const projectStore = useProjectStore();
        const trayStates = useTrayStates();
        const themeStore = useThemeStore();
        const notificationStore = useNotificationStore();
        const stageStore = useStageStore();
        const modals = useDesktopModalStore();
        
        // Change stage to projects before account switch for better UX
        if (stageStore) {
          stageStore.setStageVisibility('projects', true);
        }
        
        // Switch account using the core method
        const activeAccount = await this.switchAccount(userId);
        
        // Reset stores after successful switch
        userStore.$reset();
        projectStore.$reset();
        trayStates.$reset();
        
        if (activeAccount && activeAccount.user) {
          // Update user store with the switched account
          userStore.user = activeAccount.user;
          userStore.isUserAuthenticated = true;
          
          // Refresh application data for the new user
          await themeStore.initializeTheme();
          await projectStore.loadStudios();
          
          // Check if project directory exists before loading projects
          const projectDirectoryExists = await SettingsService.GetProjectDirectory();
          
          if (projectDirectoryExists) {
            console.log('exists')
            // Load projects if directory exists
            await projectStore.loadProjects();
            trayStates.refreshData();
          } else {
            console.log('not-exists')
            // Trigger directory onboarding if directory doesn't exist
              modals.setModalVisibility('dirOnboardModal', true);
          }
          
          notificationStore.addNotification("Account Switched", `Switched to ${activeAccount.user.first_name} ${activeAccount.user.last_name}`, "success");
        } else {
          throw new Error('Failed to get active account data');
        }
        
        // Reset the additional account flag after successful switch
        this.isAdditionalAccount = false;
        
        return activeAccount;
      } catch (error) {
        console.error('Account switch error:', error);
        const notificationStore = useNotificationStore();
        notificationStore.errorNotification("Switch Failed", error.message || 'Unable to switch accounts');
        throw error;
      }
    },

    // Get account count
    async getAccountCount() {
      try {
        const count = await AccountService.GetAccountCount();
        return count;
      } catch (error) {
        console.error('❌ Failed to get account count:', error);
        return this.accounts.length; // Fallback to local count
      }
    },

    // Force refresh from backend
    async refreshAccounts() {
      this.isLoaded = false;
      await this.loadAccounts();
    },

    // Initialize store (call on app startup)
    async initialize() {
      if (!this.isLoaded) {
        await this.loadAccounts();
      }
    }
  }
});
