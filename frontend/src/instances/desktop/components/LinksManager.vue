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
        v-if="links.behance"
        :icon="getAppIcon('brand-behance')"
        label="Behance"
        :iconAfter="false"
        :useOutline="true"
        @click="openLink(links.behance)"
      />
      <ActionButton
        v-if="links.artstation"
        :icon="getAppIcon('brand-artstation')"
        label="ArtStation"
        :iconAfter="false"
        :useOutline="true"
        @click="openLink(links.artstation)"
      />
      <ActionButton
        v-if="links.portfolio"
        :icon="getAppIcon('link')"
        label="Portfolio"
        :iconAfter="false"
        :useOutline="true"
        @click="openLink(links.portfolio)"
      />
      <ActionButton
        v-if="links.linkedin"
        :icon="getAppIcon('brand-linkedin')"
        label="LinkedIn"
        :iconAfter="false"
        :useOutline="true"
        @click="openLink(links.linkedin)"
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

const localLinks = ref({ ...props.links });

const hasAnyLinks = computed(() => {
  return Object.values(props.links).some(link => link && link.trim());
});

// Watch for changes in edit mode and emit updates
watch(localLinks, (newLinks) => {
  if (props.isEditing) {
    emit('update:links', { ...newLinks });
  }
}, { deep: true });

// Sync local links when props change (e.g., when canceling edit)
watch(() => props.links, (newLinks) => {
  localLinks.value = { ...newLinks };
}, { deep: true });

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
