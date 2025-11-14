<template>
  <div class="links-manager">
    <div v-if="isEditing" class="links-edit-mode">
      <FormInput
        v-model="localLinks.behance"
        placeholder="https://behance.net/yourprofile"
        label="Behance URL"
        type="url"
      />
      <FormInput
        v-model="localLinks.artstation"
        placeholder="https://artstation.com/yourprofile"
        label="ArtStation URL"
        type="url"
      />
      <FormInput
        v-model="localLinks.portfolio"
        placeholder="https://yourportfolio.com"
        label="Portfolio URL"
        type="url"
      />
      <FormInput
        v-model="localLinks.linkedin"
        placeholder="https://linkedin.com/in/yourprofile"
        label="LinkedIn URL"
        type="url"
      />
    </div>
    
    <div v-else class="links-display-mode">
      <ActionButton
        v-if="safeLinks.behance"
        :icon="getAppIcon('brand-behance')"
        label="Behance"
        :iconAfter="false"
        :useOutline="true"
        @click="openLink(safeLinks.behance)"
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
        v-if="safeLinks.portfolio"
        :icon="getAppIcon('link')"
        label="Portfolio"
        :iconAfter="false"
        :useOutline="true"
        @click="openLink(safeLinks.portfolio)"
      />
      <ActionButton
        v-if="safeLinks.linkedin"
        :icon="getAppIcon('brand-linkedin')"
        label="LinkedIn"
        :iconAfter="false"
        :useOutline="true"
        @click="openLink(safeLinks.linkedin)"
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
      linkedin: ''
    })
  },
  isEditing: {
    type: Boolean,
    default: false
  }
});

const emit = defineEmits(['update:links']);

// Ensure links is always an object, even if null/undefined is passed
const safeLinks = computed(() => props.links || {
  behance: '',
  artstation: '',
  portfolio: '',
  linkedin: ''
});

const localLinks = ref({ ...safeLinks.value });
const isInternalUpdate = ref(false);

const hasAnyLinks = computed(() => {
  if (!props.links) return false;
  return Object.values(props.links).some(link => link && link.trim());
});

// Watch for changes in edit mode and emit updates
watch(localLinks, (newLinks) => {
  if (props.isEditing && !isInternalUpdate.value) {
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
      linkedin: ''
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
