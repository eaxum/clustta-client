<template>
  <div class="skills-manager">
    <!-- Display selected skills -->
    <div v-if="skills.length > 0" class="skills-container">
      <span
        v-for="skill in skills"
        :key="skill.id"
        class="skill-badge"
      >
        <ActionButton
          :icon="getSkillIconPath(skill)"
          :isInactive="true"
          :showIcon="true"
          :showLabel="false"
        />
        <span class="skill-name">{{ skill.skill_name }}</span>
        <ActionButton
          :icon="getAppIcon('close')"
          :buttonFunction="() => removeSkill(skill)"
          :showIcon="true"
          :showLabel="false"
          :noFilter="false"
        />
      </span>
    </div>
    
    <!-- Empty state -->
    <div v-else-if="!isEditing" class="empty-state">
      No skills to display
    </div>
    
    <!-- ItemSelector for adding new skills -->
    <div v-if="isEditing">
      <ItemSelector
        v-if="skills.length < 5"
        :selectedItems="skills"
        :allItems="normalizedAllSkills"
        :placeholder="'Search and add skills...'"
        :itemType="'skill'"
        @itemAdded="addSkill"
      />
      <div v-else class="limit-message">
        <img :src="getAppIcon('info')" alt="Info" class="limit-icon" />
        <span>Maximum of 5 skills reached. Remove a skill to add another.</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import { useIconStore } from '@/stores/icons';
import { useUserStore } from '@/stores/users';
import { useProfileStore } from '@/stores/profile';
import { useNotificationStore } from '@/stores/notifications';
import { ProfileService } from "@/services";
import ItemSelector from './ItemSelector.vue';
import ActionButton from './ActionButton.vue';
import { getSkillIcon } from '@/utils/iconMappers';

const iconStore = useIconStore();
const userStore = useUserStore();
const profileStore = useProfileStore();
const notificationStore = useNotificationStore();

const props = defineProps({
  skills: {
    type: Array,
    default: () => []
  },
  isEditing: {
    type: Boolean,
    default: false
  },
  allSkills: {
    type: Array,
    default: () => []
  }
});

// Normalize allSkills to have consistent 'name' field for ItemSelector
const normalizedAllSkills = computed(() => {
  // Filter out skills that are already selected
  const selectedSkillIds = props.skills.map(s => s.id);
  return props.allSkills
    .filter(skill => !selectedSkillIds.includes(skill.id))
    .map(skill => ({
      ...skill,
      name: skill.skill_name || skill.name,
      category: skill.skill_category || skill.category
    }));
});

const addSkill = (skill) => {
  ProfileService.AddUserSkill(userStore.user.id, {
    skill_id: skill.id,
    proficiency_level: skill.proficiency_level || 'intermediate'
  })
    .then(() => {
      // Transform the skill to match the expected structure
      const transformedSkill = {
        ...skill,
        skill_id: skill.id,
        skill_name: skill.skill_name || skill.name,
        skill_category: skill.skill_category || skill.category
      };
      profileStore.addSkill(transformedSkill);
      notificationStore.addNotification("Skill added", "Skill added successfully.", "success", false);
    })
    .catch((err) => {
      notificationStore.errorNotification("Failed to add skill", err?.message || err);
    });
};

const removeSkill = (skill) => {
  console.log(skill)
  ProfileService.RemoveUserSkill(userStore.user.id, skill.skill_id)
    .then(() => {
      profileStore.removeSkill(skill.id);
      notificationStore.addNotification("Skill removed", "Skill removed successfully.", "success", false);
    })
    .catch((err) => {
      notificationStore.errorNotification("Failed to remove skill", err?.message || err);
    });
};

// Get skill icon path dynamically at render time
const getSkillIconPath = (skill) => {
  // Use the skill name and category to get the appropriate icon
  const skillName = skill.skill_name;
  const category = skill.skill_category;
  const iconName = getSkillIcon(skillName, category);
  return iconStore.getAppIcon(iconName);
};

const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};
</script>

<style scoped>
.skills-manager {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  width: 100%;
}

.skills-container {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.skill-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.3rem;
  background-color: var(--steel);
  border-radius: var(--large-radius);
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--white);
  position: relative;
  transition: background-color 0.2s;
}

.skill-badge:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.skill-icon {
  width: 14px;
  height: 14px;
  filter: brightness(0) invert(1);
}

.skill-name {
  user-select: none;
}

.limit-message {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  background-color: rgba(255, 193, 7, 0.1);
  border: 1px solid rgba(255, 193, 7, 0.3);
  border-radius: var(--normal-radius);
  color: rgba(255, 193, 7, 0.9);
  font-size: 0.875rem;
}

.limit-icon {
  width: 16px;
  height: 16px;
  filter: invert(82%) sepia(89%) saturate(548%) hue-rotate(359deg) brightness(103%) contrast(98%);
  flex-shrink: 0;
}

.empty-state {
  padding: .5rem;
  text-align: center;
  color: var(--white);
  opacity: .5;
  font-style: italic;
  font-size: 0.875rem;
}
</style>
