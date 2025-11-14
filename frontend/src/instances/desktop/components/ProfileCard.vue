<template>
  <div class="profile-card">
    <div v-if="title" class="profile-card-header">
      <h2 class="profile-card-title">{{ title }}</h2>
      <button 
        v-if="showEditButton" 
        @click="$emit('toggleEdit')"
        class="card-edit-button"
        :title="isEditing ? 'Cancel editing' : 'Edit section'"
      >
        <img 
          :src="isEditing ? '/icons/close.svg' : '/icons/edit.svg'" 
          alt="Edit" 
          class="edit-icon"
        />
      </button>
    </div>
    <div class="profile-card-content">
      <slot></slot>
    </div>
  </div>
</template>

<script setup>
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
  background-color: var(--black-steel);
  border-radius: var(--large-radius);
  border-radius: 24px;
  padding: 1.5rem;
  box-sizing: border-box;
  width: 100%;
  outline: var(--transparent-line);
  outline-offset: -1px;
}

.profile-card-header {
  margin-bottom: 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  border-bottom: var(--transparent-line);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.profile-card-title {
  font-size: 1rem;
  font-weight: 300;
  color: var(--white);
  margin: 0;
  flex: 1;
}

.card-edit-button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  background-color: var(--steel);
  border-radius: var(--small-radius);
  cursor: pointer;
  transition: all 0.2s;
  padding: 0;
}

.card-edit-button:hover {
  background-color: rgba(255, 255, 255, 0.15);
  transform: scale(1.05);
}

.card-edit-button:active {
  transform: scale(0.95);
}

.edit-icon {
  width: 16px;
  height: 16px;
  filter: brightness(0) invert(1);
}

.profile-card-content {
  color: var(--white);
}
</style>
