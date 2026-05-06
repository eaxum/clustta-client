<template>
  <div class="searchbar-container" :class="{ 'searchbar-large': size === 'large' }" v-esc="handleClear">
    <input ref="inputRef" v-model="model" class="searchbar-input" type="text" :placeholder="placeholder"
      @input="$emit('input', $event)" spellcheck="false" />
    <ActionButton v-if="model.length" :isLoading="isLoading" :icon="getAppIcon( isLoading? 'loading' : 'close')" :allowDeactivate="true"
      v-tooltip="isLoading ? $t('components.searchBar.loading') : $t('components.searchBar.clearSearch')" :buttonFunction="handleClear" />
  </div>
</template>

<script setup>
// imports
import { ref } from 'vue';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// stores
import { useIconStore } from '@/stores/icons';

const iconStore = useIconStore();

// props
const props = defineProps({
  placeholder: { type: String, default: 'Search' },
  isLoading: { type: Boolean, default: false },
  size: { type: String, default: 'normal' } // 'normal' | 'large'
});

// emits
const emit = defineEmits(['input', 'clear']);

// model
const model = defineModel({ type: String, default: '' });

// refs
const inputRef = ref(null);

// methods

// Returns the app icon path for the given icon name.
const getAppIcon = (iconName) => iconStore.resolveIcon(iconName);

// Clears the search input and emits clear event.
const handleClear = () => {
  model.value = '';
  emit('clear');
};

// Focuses the input element.
const focus = () => {
  inputRef.value?.focus();
};

// Expose focus method to parent
defineExpose({ focus });
</script>

<style scoped>
@import "@/assets/desktop.css";

.searchbar-container {
  display: flex;
  align-items: center;
  width: 100%;
  padding-right: .2rem;
  box-sizing: border-box;
  outline: none;
  background-color: var(--midnight-steel);
  border-radius: var(--large-radius);
  height: 40px;
  min-height: 40px;
  outline: var(--transparent-line);
}

.searchbar-container:hover {
  outline: var(--transparent-line);
  outline-offset: 1px;
}

.searchbar-input {
  font-family: 'Inter', sans-serif;
  font-weight: 400;
  font-size: 14px;
  width: 100%;
  padding: 10px;
  box-sizing: border-box;
  border: 0px;
  outline: none;
  background-color: transparent;
  color: var(--white);
  border-radius: var(--large-radius);
  transition: width 0.2s ease-out;
}

/* Large variant */
.searchbar-container.searchbar-large {
  border-radius: var(--very-large-radius);
  height: 50px;
  min-height: 50px;
}

.searchbar-container.searchbar-large .searchbar-input {
  font-size: 18px;
}
</style>
