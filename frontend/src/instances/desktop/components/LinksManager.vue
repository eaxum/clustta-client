<template>
  <div class="links-manager">
    <div v-if="isEditing" class="links-edit-mode">
      <FormInput
        v-model="localLinks.portfolio"
        placeholder="https://youtube.com/watch?v=... or https://vimeo.com/..."
        label="Portfolio URL (YouTube/Vimeo)"
        type="url"
        :error="linkErrors.portfolio"
        :showValidation="!!localLinks.portfolio"
        :valid="!linkErrors.portfolio && !!localLinks.portfolio"
      />
      <FormInput
        v-model="localLinks.artstation"
        placeholder="https://artstation.com/yourprofile"
        label="ArtStation URL"
        type="url"
        :error="linkErrors.artstation"
        :showValidation="!!localLinks.artstation"
        :valid="!linkErrors.artstation && !!localLinks.artstation"
      />
      <FormInput
        v-model="localLinks.behance"
        placeholder="https://behance.net/yourprofile"
        label="Behance URL"
        type="url"
        :error="linkErrors.behance"
        :showValidation="!!localLinks.behance"
        :valid="!linkErrors.behance && !!localLinks.behance"
      />
      <FormInput
        v-model="localLinks.linkedin"
        placeholder="https://linkedin.com/in/yourprofile"
        label="LinkedIn URL"
        type="url"
        :error="linkErrors.linkedin"
        :showValidation="!!localLinks.linkedin"
        :valid="!linkErrors.linkedin && !!localLinks.linkedin"
      />
      <FormInput
        v-model="localLinks.instagram"
        placeholder="https://instagram.com/yourprofile"
        label="Instagram URL"
        type="url"
        :error="linkErrors.instagram"
        :showValidation="!!localLinks.instagram"
        :valid="!linkErrors.instagram && !!localLinks.instagram"
      />
    </div>
    
    <div v-else class="links-display-mode">
      <ActionButton
        v-if="safeLinks.portfolio"
        :icon="getAppIcon('video-camera')"
        label="Portfolio"
        :iconAfter="false"
        :useOutline="true"
        @click="openLink(safeLinks.portfolio)"
      />
      <ActionButton
        v-if="safeLinks.artstation"
        :icon="getAppIcon('brand-artstation')"
        label="ArtStation"
        :iconAfter="false"
        :useOutline="true"
        @click="openLink(safeLinks.artstation)"
      />
      <ActionButton
        v-if="safeLinks.behance"
        :icon="getAppIcon('brand-behance')"
        label="Behance"
        :iconAfter="false"
        :useOutline="true"
        @click="openLink(safeLinks.behance)"
      />
      <ActionButton
        v-if="safeLinks.linkedin"
        :icon="getAppIcon('brand-linkedin')"
        label="LinkedIn"
        :iconAfter="false"
        :useOutline="true"
        @click="openLink(safeLinks.linkedin)"
      />
      <ActionButton
        v-if="safeLinks.instagram"
        :icon="getAppIcon('brand-instagram')"
        label="Instagram"
        :iconAfter="false"
        :useOutline="true"
        @click="openLink(safeLinks.instagram)"
      />
      <p v-if="!hasAnyLinks" class="no-links-message">
        No professional links added yet
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue';
import { useIconStore } from '@/stores/icons';
import { Browser } from "@wailsio/runtime";
import FormInput from './FormInput.vue';
import ActionButton from './ActionButton.vue';

const iconStore = useIconStore();

const props = defineProps({
  links: {
    type: Object,
    default: () => ({
      behance: '',
      artstation: '',
      portfolio: '',
      linkedin: '',
      instagram: ''
    })
  },
  isEditing: {
    type: Boolean,
    default: false
  }
});

const emit = defineEmits(['update:links', 'update:linksValid']);

// Ensure links is always an object, even if null/undefined is passed
const safeLinks = computed(() => props.links || {
  behance: '',
  artstation: '',
  portfolio: '',
  linkedin: '',
  instagram: ''
});

const localLinks = ref({ ...safeLinks.value });
const isInternalUpdate = ref(false);

// Validation errors for each link
const linkErrors = ref({
  behance: '',
  artstation: '',
  portfolio: '',
  linkedin: '',
  instagram: ''
});

const hasAnyLinks = computed(() => {
  if (!props.links) return false;
  return Object.values(props.links).some(link => link && link.trim());
});

// Check if all filled links are valid
const linksValid = computed(() => {
  return Object.keys(localLinks.value).every(platform => {
    const url = localLinks.value[platform];
    // Empty links are valid (optional)
    if (!url || url.trim() === '') return true;
    // Non-empty links must pass validation
    return !linkErrors.value[platform];
  });
});

// Watch linksValid and emit to parent
watch(linksValid, (isValid) => {
  emit('update:linksValid', isValid);
}, { immediate: true });

// Validation patterns for each platform
const validateLink = (platform, url) => {
  if (!url || url.trim() === '') {
    linkErrors.value[platform] = '';
    return true;
  }

  const patterns = {
    behance: /^https:\/\/(www\.)?behance\.net\/.+/i,
    artstation: /^https:\/\/(www\.)?artstation\.com\/.+/i,
    linkedin: /^https:\/\/(www\.)?linkedin\.com\/(in)\/.+/i,
    portfolio: /^https:\/\/(www\.)?(youtube\.com|youtu\.be|vimeo\.com)\/.+/i, // YouTube or Vimeo only
    instagram: /^https:\/\/(www\.)?instagram\.com\/.+/i
  };

  const isValid = patterns[platform]?.test(url.trim());
  
  if (!isValid) {
    const messages = {
      behance: 'Must start with https://behance.net/',
      artstation: 'Must start with https://artstation.com/',
      linkedin: 'Must start with https://linkedin.com/in/',
      portfolio: 'Must be a YouTube or Vimeo link',
      instagram: 'Must start with https://instagram.com/'
    };
    linkErrors.value[platform] = messages[platform];
    return false;
  }
  
  linkErrors.value[platform] = '';
  return true;
};

// Watch for changes in edit mode, validate, and emit updates
watch(localLinks, (newLinks) => {
  if (props.isEditing && !isInternalUpdate.value) {
    // Validate each link
    Object.keys(newLinks).forEach(platform => {
      validateLink(platform, newLinks[platform]);
    });
    
    emit('update:links', { ...newLinks });
  }
}, { deep: true });

// Sync local links when props change (e.g., when canceling edit or switching to edit mode)
watch(() => props.links, (newLinks) => {
  // Only sync if we're not editing or if the links are significantly different
  if (!props.isEditing) {
    isInternalUpdate.value = true;
    localLinks.value = { ...(newLinks || {
      behance: '',
      artstation: '',
      portfolio: '',
      linkedin: '',
      instagram: ''
    }) };
    // Use nextTick to ensure the update completes before resetting the flag
    setTimeout(() => {
      isInternalUpdate.value = false;
    }, 0);
  }
}, { deep: true });

// Initialize localLinks when entering edit mode
watch(() => props.isEditing, (isEditing) => {
  if (isEditing) {
    isInternalUpdate.value = true;
    localLinks.value = { ...safeLinks.value };
    setTimeout(() => {
      isInternalUpdate.value = false;
    }, 0);
  }
});

const openLink = (url) => {
  if (url) {
    Browser.OpenURL(url);
  }
};

const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};
</script>

<style scoped>
.links-manager {
  width: 100%;
}

.links-edit-mode {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.links-display-mode {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.no-links-message {
  color: var(--white);
  font-size: 0.875rem;
  opacity: .5;
  margin: 0;
  font-style: italic;
}
</style>
