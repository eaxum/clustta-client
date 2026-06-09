<template>
  <div class="language-switcher" ref="switcherRoot">
    <div class="language-trigger" @click="toggleDropdown">
      <img :src="getAppIcon('translation')" class="language-icon small-icons" />
      <span class="language-label">{{ currentLanguageName }}</span>
      <img :src="getAppIcon('chevron-down')" class="chevron-icon small-icons" :class="{ 'chevron-open': isOpen }" />
    </div>

    <div v-if="isOpen" class="language-dropdown">
      <div v-for="(name, code) in languages" :key="code" class="language-option" :class="{ 'language-active': code === currentLocale }" @click="selectLanguage(code)">
        {{ name }}
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { SUPPORTED_LANGUAGES, setLocale } from '@/i18n';

// store imports
import { useIconStore } from '@/stores/icons';

// stores
const iconStore = useIconStore();

const { locale } = useI18n();

// refs
const isOpen = ref(false);
const switcherRoot = ref(null);

// computed properties
const currentLocale = computed(() => locale.value);

const currentLanguageName = computed(() => {
  return languages[locale.value] || 'English';
});

const languages = SUPPORTED_LANGUAGES;

// methods/functions

// Closes the dropdown if the click is outside the component.
const handleClickOutside = (event) => {
  if (switcherRoot.value && !switcherRoot.value.contains(event.target)) {
    isOpen.value = false;
  }
};

// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Selects a language and persists it to localStorage.
const selectLanguage = (code) => {
  setLocale(code);
  isOpen.value = false;
};

// Toggles the dropdown open/closed.
const toggleDropdown = () => {
  isOpen.value = !isOpen.value;
};

// lifecycle hooks
onMounted(() => {
  document.addEventListener('click', handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside);
});
</script>

<style scoped>
.language-switcher {
    /* background-color: hsl(var(--destructive)); */
    top: 0px;
    right: .5rem;
    position: absolute;
    z-index: 10;
}

.language-trigger {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.35rem 0.6rem;
  border-radius: var(--small-radius);
  cursor: pointer;
  opacity: 0.5;
  transition: opacity 0.2s;
}

.language-trigger:hover {
  opacity: 1;
}

.language-icon {
  width: 16px;
  height: 16px;
}

.language-label {
  font-size: 0.8rem;
  font-weight: 400;
  color: hsl(var(--foreground));
}

.chevron-icon {
  width: 12px;
  height: 12px;
  transition: transform 0.2s;
}

.chevron-open {
  transform: rotate(180deg);
}

.language-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  min-width: 160px;
  max-height: 300px;
  overflow-y: auto;
  background-color: hsl(var(--card));
  border-radius: var(--normal-radius);
  border: 1px solid hsl(var(--border));
  
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  z-index: 100;
}

.language-dropdown::-webkit-scrollbar {
  width: 4px;
}

.language-dropdown::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: hsl(var(--border));
}

.language-dropdown::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.language-option {
  padding: 0.5rem 0.75rem;
  font-size: 0.8rem;
  font-weight: 300;
  color: hsl(var(--foreground));
  cursor: pointer;
  transition: background-color 0.15s;
}

.language-option:first-child {
  border-radius: var(--normal-radius) var(--normal-radius) 0 0;
}

.language-option:last-child {
  border-radius: 0 0 var(--normal-radius) var(--normal-radius);
}

.language-option:hover {
  background-color: hsl(var(--muted));
}

.language-active {
  color: hsl(var(--primary));
  font-weight: 500;
}
</style>
