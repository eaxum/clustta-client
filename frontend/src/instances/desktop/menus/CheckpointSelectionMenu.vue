<template>
  <div ref="menuRoot" class="checkpoint-selection-menu" v-stop-propagation>
    <div class="checkpoint-selection-search" @click.stop @keydown.stop>
      <SearchBar v-model="searchTerm" :placeholder="menuData.searchPlaceholder"
        :isLoading="menuData.searchLoading" />
    </div>
    <div v-if="menuData.loading" class="checkpoint-selection-state">Loading options...</div>
    <div v-else-if="!filteredOptions.length" class="checkpoint-selection-state">
      {{ searchTerm ? 'No results' : 'No options available' }}
    </div>
    <div v-else class="checkpoint-selection-options">
      <button v-for="option in filteredOptions" :key="option.id" class="checkpoint-selection-option"
        :class="{ 'checkpoint-selection-option-selected': option.id === menuData.selectedId,
          'checkpoint-selection-checkpoint': option.description,
          'checkpoint-selection-divider': option.showDivider }"
        :disabled="option.disabled || isSaving" type="button" @click="selectOption(option)">
        <img v-if="option.icon" class="small-icons checkpoint-selection-icon" :src="option.icon">
        <span v-if="option.description" class="checkpoint-option" :title="option.label">
          <span class="checkpoint-option-message">{{ option.label }}</span>
          <span class="checkpoint-option-date">{{ option.description }}</span>
        </span>
        <span v-else class="checkpoint-selection-label">{{ option.label }}</span>
        <img v-if="option.id === menuData.selectedId" class="small-icons checkpoint-selection-pin"
          :src="getAppIcon('pin')">
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue';
import SearchBar from '@/instances/desktop/components/SearchBar.vue';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';

const iconStore = useIconStore();
const menu = useMenu();
const isSaving = ref(false);
const menuRoot = ref(null);
const searchTerm = ref('');

const menuData = computed(() => menu.checkpointSelectionMenuData);
const filteredOptions = computed(() => {
  const query = searchTerm.value.trim().toLowerCase();
  const options = menuData.value.options.filter((option) => {
    if (!query) return true;
    const searchableText = [option.label, option.searchText, ...(option.searchTerms || [])]
      .filter(Boolean)
      .join(' ')
      .toLowerCase();
    return searchableText.includes(query);
  });
  const selectedIndex = options.findIndex(option => option.id === menuData.value.selectedId);
  if (selectedIndex > 0) {
    const selectedOption = options[selectedIndex];
    const groupStart = options.findIndex(option => option.group === selectedOption.group);
    if (selectedIndex > groupStart) {
      options.splice(groupStart, 0, ...options.splice(selectedIndex, 1));
    }
  }
  const firstCheckpointIndex = options.findIndex(option => option.group === 'checkpoints');
  return options.map((option, index) => ({
    ...option,
    showDivider: firstCheckpointIndex > 0 && index === firstCheckpointIndex,
  }));
});

const getAppIcon = iconName => iconStore.getAppIcon(iconName);

const selectOption = async (option) => {
  if (option.disabled || isSaving.value || !menuData.value.onSelect) return;
  if (option.id === menuData.value.selectedId) {
    menu.hideContextMenu();
    return;
  }
  isSaving.value = true;
  try {
    const shouldClose = await menuData.value.onSelect(option);
    if (shouldClose !== false) menu.hideContextMenu();
  } finally {
    isSaving.value = false;
  }
};
</script>

<style scoped>
.checkpoint-selection-menu {
  display: flex;
  width: 260px;
  max-height: min(420px, 70vh);
  flex-direction: column;
  gap: .15rem;
  padding: .4rem;
  box-sizing: border-box;
  border: 1px solid var(--border-color);
  border-radius: var(--large-radius);
  background: color-mix(in srgb, var(--surface-2) 82%, transparent);
  backdrop-filter: blur(35px);
  box-shadow: 0 10px 28px rgba(0, 0, 0, .3);
  color: var(--text);
}

.checkpoint-selection-search {
  box-sizing: border-box;
  width: 100%;
  position: relative;
  background: var(--surface-2);
  z-index: 1;
  flex: 0 0 auto;
}

.checkpoint-selection-options {
  display: flex;
  min-height: 0;
  overflow-y: auto;
  flex-direction: column;
  gap: .15rem;
}

.checkpoint-selection-state {
  padding: .75rem;
  color: var(--text-muted);
  font-size: 13px;
  text-align: center;
}

.checkpoint-selection-option {
  display: flex;
  align-items: center;
  gap: .5rem;
  width: 100%;
  min-height: 36px;
  padding: .35rem .5rem;
  border: 0;
  border-radius: var(--normal-radius);
  background: transparent;
  color: var(--text);
  cursor: pointer;
  text-align: left;
}

.checkpoint-selection-option:hover,
.checkpoint-selection-option-selected {
  background: var(--surface-4);
}

.checkpoint-selection-option:disabled {
  cursor: default;
  opacity: .55;
}

.checkpoint-selection-checkpoint {
  padding: 0;
}

.checkpoint-selection-divider {
  margin-top: .35rem;
  border-top: var(--transparent-line);
  border-radius: 0 0 var(--normal-radius) var(--normal-radius);
}

.checkpoint-selection-checkpoint .checkpoint-selection-pin {
  margin-right: .5rem;
}

.checkpoint-selection-label {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
  font-weight: 400;
}

.checkpoint-option {
  display: flex;
  flex-direction: column;
  min-width: 0;
  width: 100%;
  gap: .25rem;
  padding: .5rem;
  box-sizing: border-box;
}

.checkpoint-option-message,
.checkpoint-option-date {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.checkpoint-option-message {
  color: var(--text-muted);
  font-family: 'Inter', sans-serif;
  font-size: 14px;
  font-weight: 400;
}

.checkpoint-option-date {
  color: var(--text-muted);
  font-family: 'Inter', sans-serif;
  font-size: 12px;
  font-weight: 400;
}

.checkpoint-selection-icon,
.checkpoint-selection-pin {
  flex: 0 0 auto;
}

.checkpoint-selection-options::-webkit-scrollbar {
  width: 4px;
}

.checkpoint-selection-options::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--surface-4);
}

.checkpoint-selection-options::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}
</style>
