<template>
  <div class="skills-manager">
    <!-- Display selected skills -->
    <div v-if="skills.length > 0" class="skills-container">
      <Chip
        v-for="skill in skills"
        :key="skill.id"
        :icon="getSkillIconPath(skill)"
        :label="skill.skill_name"
        :onRemove="() => removeSkill(skill)"
        :readonly="readonly"
      />
    </div>
    
    <!-- Empty state -->
    <div v-else-if="!isEditing" class="empty-state">
      {{ $t('components.skillsManager.noSkills') }}
    </div>
    
    <!-- ItemSelector for adding new skills -->
    <div v-if="isEditing">
      <ItemSelector
        v-if="skills.length < 5"
        :selectedItems="skills"
        :allItems="normalizedAllSkills"
        :placeholder="$t('components.skillsManager.searchPlaceholder')"
        :itemType="'skill'"
        @itemAdded="addSkill"
      />
      <div v-else class="limit-message">
        <img :src="getAppIcon('info')" alt="Info" class="limit-icon" />
        <span>{{ $t('components.skillsManager.maxReached') }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useIconStore } from '@/stores/icons';
import { useUserStore } from '@/stores/users';
import { useProfileStore } from '@/stores/profile';
import { useNotificationStore } from '@/stores/notifications';
import { ProfileService } from "@/services";
import ItemSelector from './ItemSelector.vue';
import Chip from '@/instances/common/components/Chip.vue';
import { getSkillIcon } from '@/utils/iconMappers';

const iconStore = useIconStore();
const userStore = useUserStore();
const profileStore = useProfileStore();
const notificationStore = useNotificationStore();

const { t } = useI18n();

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
  },
  readonly: {
    type: Boolean,
    default: false
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
      notificationStore.addNotification(t('components.skillsManager.skillAdded'), t('components.skillsManager.skillAddedMessage'), "success", false);
    })
    .catch((err) => {
      notificationStore.errorNotification(t('components.skillsManager.failedToAddSkill'), err?.message || err);
    });
};

const removeSkill = (skill) => {
  console.log(skill)
  ProfileService.RemoveUserSkill(userStore.user.id, skill.skill_id)
    .then(() => {
      profileStore.removeSkill(skill.id);
      notificationStore.addNotification(t('components.skillsManager.skillRemoved'), t('components.skillsManager.skillRemovedMessage'), "success", false);
    })
    .catch((err) => {
      notificationStore.errorNotification(t('components.skillsManager.failedToRemoveSkill'), err?.message || err);
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
