<template>
  <div class="navigation-bar">
    <div class="nav-left">
      <ClusttaLogo 
        :boldText="true" 
        :showText="true" 
        :colored="true" 
        size="medium" 
        @click="goHome" 
        v-stop-propagation 
      />
    </div>

    <!-- <div class="nav-center">
      <router-link to="/discover" class="nav-link" :class="{ active: isActive('/discover') }">
        <img :src="getAppIcon('person')" class="nav-icon" />
        <span>Discover</span>
      </router-link>
    </div> -->

    <div class="nav-right">
      <!-- Auth buttons when not logged in -->
      <template v-if="!userStore.isUserAuthenticated">
        <ActionButton 
          :icon="getAppIcon('launch')" 
          :label="isWideScreen ? 'Sign Up' : ''" 
          color="var(--grape)" 
          forceIconColor="light" 
          :buttonFunction="goToSignUp" 
          v-tooltip="!isWideScreen ? 'Sign Up' : ''" 
        />
        <ActionButton 
          :icon="getAppIcon('login')" 
          :label="isWideScreen ? 'Login' : ''" 
          :useOutline="true" 
          :buttonFunction="goToLogin" 
          v-tooltip="!isWideScreen ? 'Login' : ''" 
        />
      </template>

      <!-- User menu when logged in -->
      <template v-else>
        <div class="user-menu" @click="toggleUserMenu" v-stop-propagation>
          <div class="user-avatar" :style="{ backgroundColor: avatarColor }">
            <img v-if="userStore.user?.photo" class="avatar-img" :src="photoUrl" alt="Profile">
            <img v-else class="avatar-img" :src="generateAvatar(userStore.user?.id)" alt="Profile">
          </div>
        </div>

        <!-- User dropdown -->
        <Teleport to="body">
          <div v-if="showUserMenu" class="user-dropdown" ref="dropdownRef">
            <div class="dropdown-header">
              <div class="account-info">
                <div class="account-name">{{ fullName }}</div>
                <div class="account-email">{{ userStore.user?.email }}</div>
              </div>
            </div>
            <div class="dropdown-divider"></div>
            <div class="dropdown-actions">
              <ActionButton v-if="canDiscoverTalent"
                :icon="getAppIcon('person-search')" 
                :showLabel="true" 
                :fullWidth="true" 
                label="Discover"
                :buttonFunction="goToDiscover" 
              />
              <ActionButton 
                :icon="getAppIcon('person')" 
                :showLabel="true" 
                :fullWidth="true" 
                label="My Profile"
                :buttonFunction="goToProfile" 
              />
              <ActionButton 
                :icon="getAppIcon('cog')" 
                :showLabel="true" 
                :fullWidth="true" 
                label="Settings"
                :buttonFunction="goToSettings" 
              />
            </div>
            <div class="dropdown-divider"></div>
            <div class="dropdown-actions">
              <ActionButton 
                :icon="getAppIcon('logout')" 
                :showLabel="true" 
                :fullWidth="true" 
                label="Logout"
                :buttonFunction="handleLogout"
                :useDanger="true"
              />
            </div>
          </div>
        </Teleport>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useIconStore } from '@/stores/icons';
import { useUserStore } from '@/stores/users';
import { useEntitlementStore } from '@/stores/entitlements';
import { generateAvatar } from '@/lib/avatar';

import ClusttaLogo from '@/instances/common/components/ClusttaLogo.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

const router = useRouter();
const route = useRoute();
const iconStore = useIconStore();
const userStore = useUserStore();
const entitlementStore = useEntitlementStore();

const showUserMenu = ref(false);
const dropdownRef = ref(null);
const screenWidth = ref(window.innerWidth);

const isWideScreen = computed(() => screenWidth.value >= 500);

const canDiscoverTalent = computed(() => entitlementStore.hasFeature('talent_discovery'));

const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

const isActive = (path) => {
  return route.path.startsWith(path);
};

const fullName = computed(() => {
  const user = userStore.user;
  if (!user) return 'User';
  return `${user.first_name || ''} ${user.last_name || ''}`.trim() || 'User';
});

const avatarColor = computed(() => {
  if (userStore.user?.id) {
    const parts = userStore.user.id.split('-');
    return '#' + (parts[0] || '666666').substring(0, 6);
  }
  return '#666666';
});

const photoUrl = computed(() => {
  if (!userStore.user?.photo) return null;
  if (userStore.user.photo.startsWith('data:') || userStore.user.photo.startsWith('http')) {
    return userStore.user.photo;
  }
  return 'data:image/png;base64,' + userStore.user.photo;
});

const goHome = () => {
  router.push('/discover');
};

const goToLogin = () => {
  router.push('/auth/login');
};

const goToSignUp = () => {
  router.push('/auth/signup');
};

const goToDiscover = () => {
  showUserMenu.value = false;
  router.push('/discover');
};

const goToProfile = () => {
  showUserMenu.value = false;
  router.push('/profile');
};

const goToSettings = () => {
  showUserMenu.value = false;
  router.push('/settings');
};

const handleLogout = async () => {
  showUserMenu.value = false;
  await userStore.logout();
  router.push('/auth/login');
};

const toggleUserMenu = () => {
  showUserMenu.value = !showUserMenu.value;
};

const handleClickOutside = (event) => {
  if (showUserMenu.value && dropdownRef.value && !dropdownRef.value.contains(event.target)) {
    showUserMenu.value = false;
  }
};

const updateScreenWidth = () => {
  screenWidth.value = window.innerWidth;
};

onMounted(() => {
  document.addEventListener('click', handleClickOutside);
  window.addEventListener('resize', updateScreenWidth);
});

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside);
  window.removeEventListener('resize', updateScreenWidth);
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.navigation-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  min-height: 56px;
  padding: 0 1.5rem;
  background-color:transparent;
  box-sizing: border-box;
}

.nav-left,
.nav-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.nav-center {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 0.9rem;
  border-radius: var(--normal-radius);
  transition: all 0.2s;
}

.nav-link:hover {
  color: var(--text);
  background-color: var(--hover);
}

.nav-link.active {
  color: var(--text);
  background-color: rgba(255, 255, 255, 0.1);
}

.nav-icon {
  width: 18px;
  height: 18px;
  opacity: 0.8;
}

/* User Menu */
.user-menu {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.25rem;
  border-radius: var(--normal-radius);
  border-radius: 50%;
  cursor: pointer;
  transition: background-color 0.2s;
}

.user-menu:hover {
  background-color: var(--hover);
  outline: var(--transparent-line);
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.chevron-icon {
  width: 14px;
  height: 14px;
  opacity: 0.6;
}

/* User Dropdown */
.user-dropdown {
  position: fixed;
  top: 60px;
  right: 1.5rem;
  min-width: 200px;
  background-color: var(--surface-1);
  border-radius: var(--large-radius);
  outline: var(--transparent-line);
  outline-offset: -1px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
  z-index: 10000;
  overflow: hidden;
}

.dropdown-header {
  padding: 0.75rem;
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
  opacity: 0.7;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.dropdown-divider {
  height: 1px;
  background-color: var(--surface-3);
  margin: 0.25rem 0;
}

.dropdown-actions {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding: 0.25rem;
}

/* Responsive */
@media (max-width: 600px) {
  .navigation-bar {
    padding: 0 1rem;
  }

  .nav-link span {
    display: none;
  }

  .nav-link {
    padding: 0.5rem;
  }
}
</style>
