<template>
  <div class="profile-card">
    <div v-if="title" class="profile-card-header">
      <h2 class="profile-card-title">{{ title }}</h2>
      <ActionButton 
        v-if="showEditButton"
        :icon="isEditing ? '/icons/close.svg' : '/icons/edit.svg'"
        :buttonFunction="() => $emit('toggleEdit')"
        v-tooltip="isEditing ? $t('components.profileCard.cancelEditing') : $t('components.profileCard.editSection')"
      />
    </div>
    <div class="profile-card-content">
      <slot></slot>
    </div>
  </div>
</template>

<script setup>
import ActionButton from './ActionButton.vue';

const props = defineProps({
  title: {
    type: String,
    default: ''
  },
  showEditButton: {
    type: Boolean,
    default: false
  },
  isEditing: {
    type: Boolean,
    default: false
  }
});

defineEmits(['toggleEdit']);
</script>

<style scoped>
.profile-card {
  background-color: hsl(var(--background));
  overflow: hidden;
  background-color: hsl(var(--background));
  border-radius: var(--very-large-radius);
  box-sizing: border-box;
  width: 100%;
  border: 1px solid hsl(var(--border));
  
}

.profile-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-sizing: border-box;
  gap: 1rem;
  padding: 1rem 1.5rem;
  background-color: hsl(var(--card));
  border-radius: var(--normal-radius);
  border: 1px solid hsl(var(--border));
  
}

.profile-card-title {
  font-size: 1rem;
  font-weight: 300;
  color: hsl(var(--foreground));
  margin: 0;
  flex: 1;
}

.profile-card-content {
  padding: 1.5rem;
  color: hsl(var(--foreground));
}
</style>
