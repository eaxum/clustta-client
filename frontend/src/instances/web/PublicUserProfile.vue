<template>
  <div class="public-profile-root">
    <div class="public-profile-header">
      <!-- <ClusttaLogo :boldText="true" :showText="true" :colored="true" size="small" @click="goHome" class="header-logo" /> -->
      <TitleBar />
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>Loading profile...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="error-container">
      <img :src="getAppIcon('alert-circle')" alt="Error" class="error-icon" />
      <h2>{{ errorTitle }}</h2>
      <p>{{ errorMessage }}</p>
      <ActionButton 
        label="Go Home"
        :icon="getAppIcon('home')"
        :iconAfter="false"
        :useBackground="true"
        @click="goHome"
      />
    </div>

    <!-- Profile Content -->
    <div v-else class="public-profile-body">
      <div class="public-profile-container">
        
        <!-- Header Card -->
        <ProfileCard title="Profile Information">
          <div class="header-layout">
            <ProfileAvatar
              :userPhoto="profileData.photo"
              :avatarColor="profileColor"
              :isEditing="false"
              :readonly="true"
            />
            
            <div class="header-info">
              <div class="display-mode-fields">
                <div class="profile-name-row">
                  <div class="profile-name">{{ fullName }}</div>
                  <ActionButton :icon="getAppIcon('copy')" v-tooltip="'Copy profile link'" @click="copyProfileLink" />
                </div>
                <div v-if="profileData.bio" class="profile-title">{{ profileData.bio }}</div>
                
                <div class="meta-info">
                  <div v-if="profileData.country" class="info-item">
                    <img class="info-icon small-icons" :src="getAppIcon('map-pin')" alt="">
                    <span>{{ profileData.country }}</span>
                  </div>
                  
                  <div 
                    v-if="profileData.availability" 
                    class="availability-badge"
                    :style="{ backgroundColor: profileData.availability === 'available' ? '#24811E' : 'rgba(255, 255, 255, 0.1)' }"
                  >
                    <img class="info-icon small-icons" :src="getAppIcon('check-circle')" alt="">
                    <span>{{ capitalizeStr(profileData.availability) }}</span>
                  </div>
                </div>
                
                <!-- Professional Links in Display Mode -->
                <div class="social-links">
                  <LinksManager
                    :links="profileData.links"
                    :isEditing="false"
                    :readonly="true"
                  />
                </div>
              </div>
            </div>
          </div>
        </ProfileCard>

        <!-- Studios Card -->
        <ProfileCard v-if="profileData.studios && profileData.studios.length" title="Studios">
          <div class="studios-container">
            <div
              v-for="studio in profileData.studios"
              :key="studio.id"
              class="studio-item"
              :title="studio.name"
            >
              <img 
                v-if="studio.logo" 
                :src="studio.logo" 
                :alt="studio.name" 
                class="studio-logo"
                @error="handleStudioLogoError"
              />
              <img 
                v-else 
                :src="getAppIcon('stall')" 
                alt="Studio" 
                class="studio-logo studio-icon-default"
              />
              <span class="studio-name">{{ studio.name }}</span>
            </div>
          </div>
        </ProfileCard>

        <!-- Skills Card -->
        <ProfileCard v-if="profileData.skills && profileData.skills.length" title="Skills">
          <SkillsManager
            :skills="profileData.skills"
            :allSkills="[]"
            :isEditing="false"
            :readonly="true"
          />
        </ProfileCard>

        <!-- Tools & Software Card -->
        <ProfileCard v-if="profileData.tools && profileData.tools.length" title="Tools & Software">
          <ToolsManager
            :tools="profileData.tools"
            :allTools="[]"
            :isEditing="false"
            :readonly="true"
          />
        </ProfileCard>

        <!-- Activity Card -->
        <ProfileCard title="Activity [Coming soon]">
          <ContributionGraph :readonly="true" />
        </ProfileCard>

      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useIconStore } from '@/stores/icons';
import { useNotificationStore } from '@/stores/notifications';
import { ProfileService, ClipboardService } from '@/services';

// Components
import TitleBar from '@/instances/desktop/components/TitleBar.vue';
import ClusttaLogo from '@/instances/common/components/ClusttaLogo.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import ProfileCard from '@/instances/desktop/components/ProfileCard.vue';
import ProfileAvatar from '@/instances/desktop/components/ProfileAvatar.vue';
import SkillsManager from '@/instances/desktop/components/SkillsManager.vue';
import ToolsManager from '@/instances/desktop/components/ToolsManager.vue';
import LinksManager from '@/instances/desktop/components/LinksManager.vue';
import ContributionGraph from '@/instances/desktop/components/ContributionGraph.vue';

const route = useRoute();
const router = useRouter();
const iconStore = useIconStore();
const notificationStore = useNotificationStore();

// State
const loading = ref(true);
const error = ref(false);
const errorTitle = ref('');
const errorMessage = ref('');
const profileData = ref({
  first_name: '',
  last_name: '',
  username: '',
  bio: '',
  country: '',
  availability: '',
  photo: '',
  links: {},
  studios: [],
  skills: [],
  tools: [],
  contributions: null
});

// Computed
const fullName = computed(() => {
  return `${profileData.value.first_name} ${profileData.value.last_name}`.trim() || 'User';
});

const profileColor = computed(() => {
  if (profileData.value.id) {
    const parts = profileData.value.id.split('-');
    return '#' + parts[0];
  }
  return '#666';
});

// Methods
const capitalizeStr = (str) => {
  if (!str) return '';
  // Replace underscores with spaces and capitalize each word
  return str
    .split('_')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ');
};

const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

const handleStudioLogoError = (event) => {
  event.target.src = getAppIcon('stall');
  event.target.classList.add('studio-icon-default');
};

// Convert photo data to displayable URL
const getPhotoUrl = (photo) => {
  if (!photo) return '';
  
  // Already a data URL or http URL
  if (typeof photo === 'string') {
    if (photo.startsWith('data:') || photo.startsWith('http')) {
      return photo;
    }
    // Assume it's base64 encoded
    if (photo.length > 0) {
      return `data:image/png;base64,${photo}`;
    }
  }
  
  // Array of bytes (Uint8Array or regular array)
  if (Array.isArray(photo) || photo instanceof Uint8Array) {
    if (photo.length === 0) return '';
    const base64 = btoa(String.fromCharCode(...new Uint8Array(photo)));
    return `data:image/png;base64,${base64}`;
  }
  
  return '';
};

const goHome = () => {
  router.push('/');
};

const copyProfileLink = async () => {
  const profileUrl = `https://app.clustta.com/user/${profileData.value.username}`;
  try {
    await ClipboardService.WriteText(profileUrl);
    notificationStore.addNotification('Profile Link Copied', 'Profile link copied to clipboard', 'success');
  } catch (err) {
    console.error('Failed to copy profile link:', err);
    notificationStore.errorNotification('Copy Failed', 'Failed to copy profile link to clipboard');
  }
};

const loadPublicProfile = async () => {
  const username = route.params.username;
  
  if (!username) {
    error.value = true;
    errorTitle.value = 'Invalid Profile';
    errorMessage.value = 'No username provided.';
    loading.value = false;
    return;
  }

  try {
    const profile = await ProfileService.GetPublicProfile(username);
    
    if (!profile) {
      error.value = true;
      errorTitle.value = 'Profile Not Found';
      errorMessage.value = 'This profile does not exist or is set to private.';
      loading.value = false;
      return;
    }

    profileData.value = {
      ...profile,
      // Transform backend link fields into links object
      links: {
        behance: profile.behance_link || profile.BehanceLink || '',
        artstation: profile.artstation_link || profile.ArtstationLink || '',
        portfolio: profile.portfolio_link || profile.PortfolioLink || '',
        linkedin: profile.linkedin_link || profile.LinkedInLink || '',
        instagram: profile.instagram_link || profile.InstagramLink || '',
      },
      // Transform country from location field
      country: profile.country_name || profile.location || profile.Location || '',
      // Handle photo - use empty string if no photo to trigger placeholder
      photo: getPhotoUrl(profile.photo),
      studios: profile.studios || [],
      skills: profile.skills || [],
      tools: profile.tools || []
    };
    
    loading.value = false;
  } catch (err) {
    console.error('Error loading public profile:', err);
    error.value = true;
    
    const errMsg = err.message || '';
    
    if (errMsg.includes('not found') || errMsg.includes('404')) {
      errorTitle.value = 'Profile Not Found';
      errorMessage.value = 'This profile does not exist or is set to private.';
    } else if (errMsg.includes('private') || errMsg.includes('403')) {
      errorTitle.value = 'Profile Private';
      errorMessage.value = 'This user has set their profile to private.';
    } else {
      errorTitle.value = 'Error Loading Profile';
      errorMessage.value = 'An error occurred while loading this profile. Please try again later.';
    }
    
    loading.value = false;
  }
};

onMounted(() => {
  loadPublicProfile();
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.public-profile-root {
  width: 100%;
  min-height: 100vh;
  max-height: 100vh;
  background-color: var(--night);
  display: flex;
  flex-direction: column;
  align-items: center;
  overflow-y: auto;
}

.public-profile-header {
  width: 100%;
  display: flex;
  align-items: center;
  background-color: var(--black);
  border-bottom: var(--transparent-line);
}

.header-logo {
  padding-left: 0.5rem;
  flex: unset !important;
}

.public-profile-root::-webkit-scrollbar {
  width: 8px;
}

.public-profile-root::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--light-steel);
}

.public-profile-root::-webkit-scrollbar-track {
  border-radius: 10px;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  color: var(--white);
  gap: 1rem;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid rgba(255, 255, 255, 0.2);
  border-top-color: var(--white);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  color: var(--white);
  gap: 1rem;
  text-align: center;
  padding: 2rem;
}

.error-icon {
  width: 64px;
  height: 64px;
  filter: brightness(0) invert(1);
  opacity: 0.5;
}

.error-container h2 {
  margin: 0;
  font-size: 1.5rem;
}

.error-container p {
  margin: 0;
  opacity: 0.7;
  max-width: 400px;
}

.public-profile-body {
  width: 100%;
  max-width: 800px;
  padding: 2rem;
  box-sizing: border-box;
}

.public-profile-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

/* Header Layout */
.header-layout {
  display: flex;
  gap: 1.5rem;
  align-items: flex-start;
}

.header-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.display-mode-fields {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.profile-name-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.profile-name {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--white);
}

.profile-title {
  font-size: 1rem;
  color: var(--white);
  opacity: 0.7;
}

.meta-info {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-top: 0.5rem;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  color: var(--white);
  opacity: 0.7;
  font-size: 0.875rem;
}

.info-icon {
  width: 16px;
  height: 16px;
  filter: brightness(0) invert(1);
}

.availability-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.25rem 0.75rem;
  border-radius: 1rem;
  font-size: 0.875rem;
  color: var(--white);
}

.social-links {
  margin-top: 1rem;
}

/* Studios */
.studios-container {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
}

.studio-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background-color: var(--steel);
  border-radius: var(--normal-radius);
}

.studio-logo {
  width: 24px;
  height: 24px;
  border-radius: 4px;
  object-fit: cover;
}

.studio-icon-default {
  filter: brightness(0) invert(1);
  opacity: 0.7;
}

.studio-name {
  color: var(--white);
  font-size: 0.875rem;
}

/* Mobile responsiveness */
@media (max-width: 600px) {
  .header-layout {
    flex-direction: column;
    align-items: center;
    text-align: center;
  }

  .meta-info {
    justify-content: center;
  }

  .social-links {
    display: flex;
    justify-content: center;
  }

  .public-profile-body {
    padding: 1rem;
  }
}
</style>
