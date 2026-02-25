<template>
  <div class="user-list-card" @click="$emit('click')">
    <div class="user-card-content">
      <!-- Avatar -->
      <div class="user-avatar" :style="{ backgroundColor: avatarColor }">
        <img v-if="user.photo" class="avatar-img" :src="photoUrl" alt="Profile Photo">
        <img v-else class="avatar-img" :src="generateAvatar(user.id)" alt="Default Avatar">
      </div>

      <!-- Main Info -->
      <div class="user-main-info">
        <div class="user-identity">
          <span class="user-name">{{ fullName }}</span>
          <span v-if="user.job_title" class="user-title">• {{ user.job_title }}</span>
          <div 
            v-if="user.availability" 
            class="availability-badge"
            :class="availabilityClass"
          >
            <span class="availability-dot"></span>
            <span>{{ availabilityLabel }}</span>
          </div>
        </div>

        <div v-if="user.bio" class="user-bio">{{ truncatedBio }}</div>

        <div class="user-location">
          <ActionButton
            :icon="getAppIcon('map-pin')"
            :isInactive="true"
            :showIcon="true"
            :showLabel="false"
            :isMini="true"
          />
          <span class="location-text" :class="{ 'location-not-set': !location }">{{ location || 'Location not Set' }}</span>
        </div>

        <!-- Divider -->
        <div v-if="hasExtendedInfo && showDetails" class="section-divider">
          <div class="menu-divider"></div>
        </div>

        <!-- Extended Info Section (Skills, Tools, Studios) -->
        <div v-if="hasExtendedInfo && showDetails" class="extended-info">
          <!-- Skills -->
          <div v-if="user.skills && user.skills.length" class="user-tags-row">
            <SkillsManager
              :skills="displayedSkills"
              :allSkills="[]"
              :isEditing="false"
              :readonly="true"
            />
            <span v-if="remainingSkills > 0" class="more-tag">+{{ remainingSkills }} more</span>
          </div>

          <!-- Tools -->
          <div v-if="user.tools && user.tools.length" class="user-tags-row">
            <ToolsManager
              :tools="displayedTools"
              :allTools="[]"
              :isEditing="false"
              :readonly="true"
            />
            <span v-if="remainingTools > 0" class="more-tag">+{{ remainingTools }} more</span>
          </div>

          <!-- Studios -->
          <div v-if="user.studios && user.studios.length" class="user-studios-row">
            <span class="tags-label">Studios:</span>
            <div class="studios-container">
              <div 
                v-for="studio in user.studios.slice(0, 4)" 
                :key="studio.id" 
                class="studio-badge"
                :title="studio.name"
              >
                <img 
                  v-if="studio.logo" 
                  :src="studio.logo" 
                  :alt="studio.name" 
                  class="studio-logo"
                />
                <span v-else class="studio-initial">{{ studio.name?.charAt(0) || 'S' }}</span>
              </div>
              <span v-if="user.studios.length > 4" class="more-tag">+{{ user.studios.length - 4 }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Contact Section -->
      <div class="contact-section">
        <ActionButton
          :icon="getAppIcon('mail')"
          label="Get in touch"
          :showIcon="true"
          :showLabel="true"
          :useOutline="true"
          :buttonFunction="handleContact"
        />
        <div v-if="hasLinks" class="user-links-row">
          <LinksManager
            :links="displayedLinks"
            :isEditing="false"
            :readonly="true"
            :showLabels="false"
          />
          <span v-if="remainingLinks > 0" class="more-links">+{{ remainingLinks }} links</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useIconStore } from '@/stores/icons';
import { generateAvatar } from '@/lib/avatar';
import SkillsManager from '@/instances/desktop/components/SkillsManager.vue';
import ToolsManager from '@/instances/desktop/components/ToolsManager.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import LinksManager from '@/instances/desktop/components/LinksManager.vue';
import utils from '@/services/utils';

const props = defineProps({
  user: {
    type: Object,
    required: true
  },
  maxSkills: {
    type: Number,
    default: 4
  },
  maxTools: {
    type: Number,
    default: 4
  },
  maxLinks: {
    type: Number,
    default: 2
  },
  showDetails: {
    type: Boolean,
    default: false
  }
});

defineEmits(['click']);

const iconStore = useIconStore();

const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Computed properties
const fullName = computed(() => {
  const name = `${props.user.first_name || ''} ${props.user.last_name || ''}`.trim() || 'User';
  return name.split(' ').map(part => utils.capitalizeStr(part)).join(' ');
});

const avatarColor = computed(() => {
  if (props.user.id) {
    const parts = props.user.id.split('-');
    return '#' + (parts[0] || '666666').substring(0, 6);
  }
  return '#666666';
});

const photoUrl = computed(() => {
  if (!props.user.photo) return null;
  if (props.user.photo.startsWith('data:') || props.user.photo.startsWith('http')) {
    return props.user.photo;
  }
  return 'data:image/png;base64,' + props.user.photo;
});

const location = computed(() => {
  const parts = [];
  if (props.user.location) parts.push(props.user.location);
  if (props.user.country?.name) parts.push(props.user.country.name);
  else if (props.user.country?.code) parts.push(props.user.country.code);
  return parts.join(', ');
});

const availabilityClass = computed(() => {
  switch (props.user.availability) {
    case 'available': return 'available';
    case 'busy': return 'busy';
    case 'not_looking': return 'not-looking';
    default: return '';
  }
});

const availabilityLabel = computed(() => {
  switch (props.user.availability) {
    case 'available': return 'Available';
    case 'busy': return 'Busy';
    case 'not_looking': return 'Not Looking';
    default: return props.user.availability || '';
  }
});

const truncatedBio = computed(() => {
  if (!props.user.bio) return '';
  return props.user.bio.length > 120 ? props.user.bio.substring(0, 120) + '...' : props.user.bio;
});

// User profile links
const userLinks = computed(() => {
  const user = props.user;
  return {
    portfolio: user.portfolio_link || user.PortfolioLink || user.links?.portfolio || '',
    artstation: user.artstation_link || user.ArtstationLink || user.links?.artstation || '',
    behance: user.behance_link || user.BehanceLink || user.links?.behance || '',
    linkedin: user.linkedin_link || user.LinkedInLink || user.links?.linkedin || '',
    instagram: user.instagram_link || user.InstagramLink || user.links?.instagram || ''
  };
});

const hasLinks = computed(() => {
  const links = userLinks.value;
  return links.portfolio || links.artstation || links.behance || links.linkedin || links.instagram;
});

// Get total link count and displayed/remaining links
const linkEntries = computed(() => {
  const links = userLinks.value;
  return Object.entries(links).filter(([_, value]) => !!value);
});

const displayedLinks = computed(() => {
  const entries = linkEntries.value.slice(0, props.maxLinks);
  return Object.fromEntries([
    ['portfolio', ''],
    ['artstation', ''],
    ['behance', ''],
    ['linkedin', ''],
    ['instagram', ''],
    ...entries
  ]);
});

const remainingLinks = computed(() => {
  return Math.max(0, linkEntries.value.length - props.maxLinks);
});

// Check if user has any extended info (skills, tools, studios)
const hasExtendedInfo = computed(() => {
  return (props.user.skills && props.user.skills.length) ||
         (props.user.tools && props.user.tools.length) ||
         (props.user.studios && props.user.studios.length);
});

// Handle contact button click
const handleContact = (event) => {
  event.stopPropagation();
  // Could open email or show contact modal in the future
  if (props.user.email) {
    window.open(`mailto:${props.user.email}`, '_blank');
  }
};

const displayedSkills = computed(() => {
  const skills = (props.user.skills || []).slice(0, props.maxSkills);
  // Normalize skill data to ensure skill_name and skill_category properties exist
  return skills.map(s => ({
    ...s,
    skill_name: s.skill_name || s.SkillName || s.name || '',
    skill_category: s.skill_category || s.SkillCategory || s.category || ''
  }));
});

const remainingSkills = computed(() => {
  return Math.max(0, (props.user.skills?.length || 0) - props.maxSkills);
});

const displayedTools = computed(() => {
  const tools = (props.user.tools || []).slice(0, props.maxTools);
  // Normalize tool data to ensure tool_name property exists
  return tools.map(t => ({
    ...t,
    tool_name: t.tool_name || t.ToolName || t.name || ''
  }));
});

const remainingTools = computed(() => {
  return Math.max(0, (props.user.tools?.length || 0) - props.maxTools);
});
</script>

<style scoped>
@import "@/assets/desktop.css";
.user-list-card {
    background-color: var(--black-steel);
    border-radius: var(--very-large-radius);
    border-radius: var(--gigantic-radius);
    outline: var(--transparent-line);
    outline-offset: -1px;
    cursor: pointer;
    transition: all .2s ease-in-out;
    width: 100%;
  box-sizing: border-box;
  color: var(--white);
}

.user-list-card:hover {
  /* background-color: var(--dark-steel); */
  border-radius: var(--large-radius);
  /* box-shadow: 0 0px 10px rgba(0, 0, 0, 0.4); */
  box-shadow: 0 0px 8px rgba(0, 0, 0, 0.1);
}

.user-card-content {
  display: flex;
  gap: 1.25rem;
  padding: 1.25rem;
}

/* Avatar */
.user-avatar {
  flex-shrink: 0;
  width: 72px;
  height: 72px;
  border-radius: 50%;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  outline: var(--transparent-line);
}

.avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

/* Main Info */
.user-main-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.user-identity {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.user-name {
  font-size: 1.1rem;
  font-weight: 500;
  color: var(--white);
}

.user-title {
  font-size: 0.875rem;
  color: var(--text-secondary);
}

/* Contact Section */
.contact-section {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  align-self: flex-start;
  gap: 0.35rem;
  flex-shrink: 0;
}

.user-links-row {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.more-links {
  font-size: 0.7rem;
  color: var(--text-tertiary);
  padding: 0.2rem 0.35rem;
}

.user-bio {
  font-size: 0.875rem;
  color: var(--text-secondary);
  line-height: 1.4;
}

.user-location {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.location-text {
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.location-text.location-not-set {
  font-style: italic;
  opacity: 0.6;
}

/* Section Divider */
.section-divider {
  display: flex;
  align-items: center;
  margin: 0.5rem 0;
}

.divider-line {
  flex: 1;
  height: 1px;
  background-color: var(--transparent-line);
  background-color: crimson;
}

/* Extended Info */
.extended-info {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.availability-badge {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.2rem 0.55rem;
  border-radius: 1rem;
  font-size: 0.7rem;
  font-weight: 500;
  background-color: transparent;
  border: 1px solid rgba(255, 255, 255, 0.25);
  color: var(--white);
}

.availability-badge.available {
  border-color: #4ade80;
  background-color: rgba(74, 222, 128, 0.15);
  color:var(--white);
}

.availability-badge.available .availability-dot {
  background-color: #4ade80;
}

.availability-badge.busy {
  border-color: #fbbf24;
  background-color: rgba(251, 191, 36, 0.15);
  color:var(--white);
}

.availability-badge.busy .availability-dot {
  background-color: #fbbf24;
}

.availability-badge.not-looking {
  border-color:var(--white);
  background-color: rgba(255, 255, 255, 0.1);
  color:var(--white);
}

.availability-badge.not-looking .availability-dot {
  background-color:var(--white);
}

.availability-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

/* Tags */
.user-tags-row,
.user-studios-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.studios-container {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  flex-wrap: wrap;
}

.more-tag {
  font-size: 0.75rem;
  color: var(--text-tertiary);
  padding: 0.2rem 0.4rem;
}

/* Studios */
.studio-badge {
  width: 24px;
  height: 24px;
  border-radius: 4px;
  overflow: hidden;
  background-color: rgba(255, 255, 255, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
}

.studio-logo {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.studio-initial {
  font-size: 0.7rem;
  font-weight: 600;
  color: var(--text-secondary);
}

/* Responsive */
@media (max-width: 768px) {
  .user-card-content {
    flex-direction: column;
    align-items: center;
    text-align: center;
  }

  .user-avatar {
    width: 64px;
    height: 64px;
  }

  .user-identity {
    justify-content: center;
  }

  .contact-section {
    width: 100%;
    flex-direction: row;
    align-items: center;
    flex-wrap: wrap;
    justify-content: center;
  }

  .user-links-row {
    justify-content: center;
  }

  .user-location {
    justify-content: center;
  }

  .user-tags-row,
  .user-studios-row {
    justify-content: center;
  }
}
</style>
