<template>
  <div class="skills-manager">
    <!-- Display selected skills -->
    <div v-if="skills.length > 0" class="skills-container">
      <span
        v-for="skill in skills"
        :key="skill.id"
        class="skill-badge"
      >
        <img v-if="skill.icon" class="small-icons" :src="getAppIcon(skill.icon)" alt="">
        <span class="skill-name">{{ skill.name }}</span>
        <button
          v-if="isEditing"
          @click="removeSkill(skill)"
          class="skill-remove-button"
          title="Remove"
        >
          <img class="remove-icon" :src="getAppIcon('close')" alt="Remove">
        </button>
      </span>
    </div>
    
    <!-- ItemSelector for adding new skills -->
    <ItemSelector
      v-if="isEditing"
      :selectedItems="skills"
      :allItems="allSkills"
      :placeholder="'Search and add skills...'"
      :itemType="'skill'"
      @itemAdded="addSkill"
    />
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useIconStore } from '@/stores/icons';
import ItemSelector from './ItemSelector.vue';

const iconStore = useIconStore();

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
    default: () => [
      { id: '1', name: '3D Modeling', category: 'Technical', icon: 'box' },
      { id: '2', name: 'Character Animation', category: 'Artistic', icon: 'man-running' },
      { id: '3', name: 'Texturing', category: 'Artistic', icon: 'palette' },
      { id: '4', name: 'Rigging', category: 'Technical', icon: 'network' },
      { id: '5', name: 'VFX', category: 'Artistic', icon: 'sparkles' },
      { id: '6', name: 'Lighting', category: 'Technical', icon: 'bulb' },
      { id: '7', name: 'Rendering', category: 'Technical', icon: 'bulb' },
      { id: '8', name: 'Sculpting', category: 'Artistic', icon: 'palette' },
      { id: '9', name: 'UV Mapping', category: 'Technical', icon: 'box' },
      { id: '10', name: 'Motion Graphics', category: 'Artistic', icon: 'sparkles' },
      { id: '11', name: 'Compositing', category: 'Technical', icon: 'layers' },
      { id: '12', name: 'Concept Art', category: 'Artistic', icon: 'palette' },
      { id: '13', name: 'Game Design', category: 'Management', icon: 'gamepad' },
      { id: '14', name: 'Project Management', category: 'Management', icon: 'check-circle' },
      { id: '15', name: 'Storyboarding', category: 'Artistic', icon: 'file' }
    ]
  }
});

const emit = defineEmits(['skillAdded', 'skillRemoved']);

const addSkill = (skill) => {
  emit('skillAdded', skill);
};

const removeSkill = (skill) => {
  emit('skillRemoved', skill.name);
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
  padding: 0.5rem 0.75rem;
  background-color: var(--steel);
  border-radius: var(--normal-radius);
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--white);
  position: relative;
  transition: background-color 0.2s;
}

.skill-badge:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.skill-remove-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  margin-left: 0.25rem;
  opacity: 0.7;
  transition: opacity 0.2s;
}

.skill-remove-button:hover {
  opacity: 1;
}

.remove-icon {
  width: 12px;
  height: 12px;
  filter: brightness(0) invert(1);
}

.skill-icon {
  width: 14px;
  height: 14px;
  filter: brightness(0) invert(1);
}

.skill-name {
  user-select: none;
}
</style>
