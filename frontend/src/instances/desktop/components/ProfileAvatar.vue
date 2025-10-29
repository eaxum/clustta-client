<template>
  <div class="profile-avatar-container" :class="{ 'editable': isEditing }">
    <div class="avatar-wrapper">
      <div class="profile-avatar" :style="{ backgroundColor: avatarColor }">
        <img v-if="displayPhoto" class="avatar-img" :src="displayPhoto" alt="Profile Photo">
        <img v-else class="avatar-img" :src="getAppIcon('person')" alt="Default Avatar">
      </div>
      
      <div v-if="isEditing" class="avatar-overlay">
        <div class="avatar-actions">
          <label class="avatar-action-button" title="Change Photo">
            <img class="action-icon" :src="getAppIcon('camera')" alt="Change">
            <input 
              type="file" 
              @change="handlePhotoChange" 
              accept="image/*" 
              ref="photoInput" 
              class="photo-input"
            />
          </label>
          <button 
            v-if="photoPreview || userPhoto" 
            class="avatar-action-button" 
            @click="removePhoto"
            title="Remove Photo"
          >
            <img class="action-icon" :src="getAppIcon('close')" alt="Remove">
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import { useIconStore } from '@/stores/icons';

const iconStore = useIconStore();

const props = defineProps({
  userPhoto: {
    type: String,
    default: ''
  },
  avatarColor: {
    type: String,
    default: '#666'
  },
  isEditing: {
    type: Boolean,
    default: false
  },
  size: {
    type: String,
    default: 'large' // large, medium, small
  }
});

const emit = defineEmits(['photoChanged', 'photoRemoved']);

const photoInput = ref(null);
const photoPreview = ref(null);

const displayPhoto = computed(() => {
  return photoPreview.value || props.userPhoto;
});

const handlePhotoChange = (event) => {
  const file = event.target.files[0];
  if (file) {
    const reader = new FileReader();
    reader.onload = (e) => {
      photoPreview.value = e.target.result;
      emit('photoChanged', file, e.target.result);
    };
    reader.readAsDataURL(file);
  }
};

const removePhoto = () => {
  photoPreview.value = null;
  if (photoInput.value) {
    photoInput.value.value = '';
  }
  emit('photoRemoved');
};

const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};
</script>

<style scoped>
.profile-avatar-container {
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar-wrapper {
  position: relative;
  width: 96px;
  height: 96px;
}

.profile-avatar {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid rgba(255, 255, 255, 0.1);
  box-sizing: border-box;
}

.avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.2s;
}

.editable .avatar-wrapper:hover .avatar-overlay {
  opacity: 1;
}

.avatar-actions {
  display: flex;
  gap: 0.5rem;
}

.avatar-action-button {
  background-color: var(--white);
  border: none;
  border-radius: 50%;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: transform 0.2s;
  padding: 0;
}

.avatar-action-button:hover {
  transform: scale(1.1);
}

.action-icon {
  width: 20px;
  height: 20px;
  filter: brightness(0);
}

.photo-input {
  display: none;
}

/* Size variants */
.profile-avatar-container.medium .avatar-wrapper {
  width: 80px;
  height: 80px;
}

.profile-avatar-container.small .avatar-wrapper {
  width: 48px;
  height: 48px;
}
</style>
