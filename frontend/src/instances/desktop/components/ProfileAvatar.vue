<template>
  <div class="profile-avatar-container" :class="{ 'editable': isEditing }">
    <div class="avatar-wrapper">
      <div class="profile-avatar" :style="{ backgroundColor: avatarColor }">
        <img v-if="displayPhoto" class="avatar-img" :src="displayPhoto" alt="Profile Photo">
        <img v-else class="avatar-img" :src="fallbackAvatar" alt="Default Avatar">
      </div>
      
      <div v-if="isEditing && !readonly" class="avatar-overlay">
        <div class="avatar-actions">
          <button 
            class="avatar-action-button" 
            @click="selectPhoto"
            v-tooltip="$t('components.profileAvatar.changePhoto')"
          >
            <CiCamera :size="20" class="action-icon" />
          </button>
          <button 
            v-if="photoPreview || userPhoto" 
            class="avatar-action-button" 
            @click="removePhoto"
            v-tooltip="$t('components.profileAvatar.removePhoto')"
          >
            <CiClose :size="20" class="action-icon" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { CiCamera, CiClose } from '@clustta/icons-vue';
import { resolveIcon } from '@/lib/icon-map';
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useIconStore } from '@/stores/icons';
import { DialogService, FSService } from '@/services';
import { generateAvatar } from '@/lib/avatar';
import utils from '@/services/utils';

const { t } = useI18n();

const iconStore = useIconStore();

const props = defineProps({
  userPhoto: {
    type: String,
    default: ''
  },
  userId: {
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
  },
  readonly: {
    type: Boolean,
    default: false
  }
});

const emit = defineEmits(['photoChanged', 'photoRemoved']);

const photoPreview = ref(null);

// Generates a DiceBear avatar based on userId as fallback.
const fallbackAvatar = computed(() => {
  return generateAvatar(props.userId);
});

const displayPhoto = computed(() => {
  return photoPreview.value || props.userPhoto;
});

const selectPhoto = async () => {
  try {
    const result = await DialogService.SelectFileDialog(
      t('components.profileAvatar.selectProfilePicture'), 
      "*.png;*.jpg;*.jpeg;*.gif;*.bmp;*.webp"
    );
    
    if (result) {
      // Convert file path to base64 for preview
      const base64String = await utils.base64FromFile(result);
      photoPreview.value = base64String;
      
      // Emit the file path and preview
      emit('photoChanged', result, base64String);
    }
  } catch (error) {
    console.error('Error selecting photo:', error);
  }
};

const removePhoto = () => {
  photoPreview.value = null;
  emit('photoRemoved');
};

const getAppIcon = (iconName) => {
  return iconStore.resolveIcon(iconName);
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
  /* border: 2px solid rgba(255, 255, 255, 0.1); */
  outline: var(--transparent-line);
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
