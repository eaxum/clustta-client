<template>
  <div ref="selectorRoot" class="item-selector" v-esc="escape">
    <div class="combo">
      <div @click="focusInput()" class="selector-container tint combo-linear" ref="comboBoxRoot">
        <div class="search-input-container" ref="searchInputContainer">
          <input 
            v-focus 
            ref="inputField" 
            :placeholder="placeholder" 
            v-model="searchQuery" 
            autocomplete="off"
            @keydown="handleKeyDown" 
            @blur="handleInputBlur" 
            @focus="handleInputFocus"
            @input="handleInput" 
            class="input-field" 
          />
        </div>
      </div>
      
      <Teleport to="#app">
        <div 
          v-if="showSuggestions" 
          ref="suggestionsParent" 
          class="suggestions-parent" 
          v-esc="hideSuggestions" 
          v-stop-propagation
          :style="dropdownStyles"
        >
          <div 
            v-for="item in filteredItems" 
            :key="item.id"
            class="item-suggestion" 
            @click="addItem(item)"
          >
              <img :class="['large-icons', { 'no-filter': !iconFilter }]" :src="getItemIcon(item)">
            <div class="item-meta">
              <div class="item-suggestion-name">{{ item.name }}</div>
              <div v-if="item.category" class="item-suggestion-category">{{ item.category }}</div>
            </div>
          </div>
          
          <div v-if="filteredItems.length === 0 && searchQuery" class="no-results">
            {{ $t('components.itemSelector.noMatchingItems', { type: itemType }) }}
          </div>
        </div>
      </Teleport>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue';
import { useIconStore } from '@/stores/icons';
import { getToolLogo, getSkillIcon } from '@/utils/iconMappers';

const iconStore = useIconStore();

// Emits
const emit = defineEmits([
  'itemAdded',
  'input', 
  'on-focus', 
  'on-blur'
]);

// Props
const props = defineProps({
  placeholder: { 
    type: String, 
    default: 'Search or add...' 
  },
  selectedItems: { 
    type: Array, 
    default: () => [] 
  },
  allItems: { 
    type: Array, 
    default: () => [] 
  },
  itemType: {
    type: String,
    default: 'item' // 'skill' or 'tool'
  },
  allowMultiple: {
    type: Boolean,
    default: true
  },
  iconFilter: {
    type: Boolean,
    default: true
  }
});

// Element refs
const selectorRoot = ref(null);
const comboBoxRoot = ref(null);
const inputField = ref(null);
const suggestionsParent = ref(null);

// Refs
const searchQuery = ref('');
const showDropdown = ref(false);
const isInputActive = ref(false);

// Dropdown positioning
const dropdownTop = ref(0);
const dropdownLeft = ref(0);
const dropdownWidth = ref(0);
const dropdownMaxHeight = ref(300);

// Computed styles for dropdown
const dropdownStyles = computed(() => ({
  position: 'fixed',
  top: `${dropdownTop.value}px`,
  left: `${dropdownLeft.value}px`,
  width: `${dropdownWidth.value}px`,
  maxHeight: `${dropdownMaxHeight.value}px`
}));

// Computed

const showSuggestions = computed(() => {
  return showDropdown.value && filteredItems.value.length > 0;
});

const filteredItems = computed(() => {
  // Get items that are not already selected
  const availableItems = props.allItems.filter(item => 
    !props.selectedItems.some(selected => selected.id === item.id)
  );
  
  if (!searchQuery.value) {
    return availableItems;
  }
  
  const lowerSearchTerm = searchQuery.value.toLowerCase();
  return availableItems.filter(item => {
    const searchRange = `${item.name} ${item.category || ''}`;
    return searchRange.toLowerCase().includes(lowerSearchTerm);
  });
});

// Methods
const getAppIcon = (iconName) => {
  return iconStore.resolveIcon(iconName);
};

// Get icon/logo for an item based on itemType
const getItemIcon = (item) => {
  if (props.itemType === 'tool') {
    // For tools, get the file icon logo
    const toolName = item.tool_name || item.ToolName || item.name || '';
    return getToolLogo(toolName);
  } else if (props.itemType === 'skill') {
    // For skills, get the thematic icon from iconStore
    const skillName = item.skill_name || item.SkillName || item.name || '';
    const category = item.skill_category || item.SkillCategory || item.category || '';
    const iconName = getSkillIcon(skillName, category);
    return iconStore.resolveIcon(iconName);
  }
  // Fallback for other item types
  return item.icon ? iconStore.resolveIcon(item.icon) : null;
};

const addItem = (item) => {
  emit('itemAdded', item);
  inputField.value.focus();
  searchQuery.value = '';
};

const hideSuggestions = () => {
  showDropdown.value = false;
  searchQuery.value = '';
};

const escape = () => {
  hideSuggestions();
};

const handleKeyDown = (event) => {
  if (event.key === 'Escape') {
    hideSuggestions();
  } else if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
    event.preventDefault();
    // TODO: Add keyboard navigation through suggestions
  }
};

const handleInput = (event) => {
  emit('input', event.target.value);
  if (!showDropdown.value && event.target.value) {
    showDropdown.value = true;
    updateDropdownPosition();
  }
};

const focusInput = () => {
  inputField.value?.focus();
};

const updateDropdownPosition = () => {
  if (!comboBoxRoot.value) return;
  
  const rect = comboBoxRoot.value.getBoundingClientRect();
  const viewportHeight = window.innerHeight;
  const spaceBelow = viewportHeight - rect.bottom;
  const spaceAbove = rect.top;
  
  dropdownTop.value = rect.bottom + 8;
  dropdownLeft.value = rect.left;
  dropdownWidth.value = rect.width;
  
  // Calculate max height based on available space
  // Prefer showing below, but if not enough space, check above
  if (spaceBelow < 200 && spaceAbove > spaceBelow) {
    dropdownTop.value = rect.top - Math.min(300, spaceAbove - 8);
    dropdownMaxHeight.value = Math.min(300, spaceAbove - 16);
  } else {
    dropdownMaxHeight.value = Math.min(300, spaceBelow - 16);
  }
};

const handleInputFocus = (event) => {
  isInputActive.value = true;
  showDropdown.value = true;
  updateDropdownPosition();
  emit('on-focus', event);
};

const handleInputBlur = (event) => {
  isInputActive.value = false;
  // Delay hiding to allow click on suggestions
  setTimeout(() => {
    if (!isInputActive.value) {
      showDropdown.value = false;
    }
  }, 200);
  emit('on-blur', event);
};

const handleClickOutside = (event) => {
  if (selectorRoot.value && !selectorRoot.value.contains(event.target)) {
    if (suggestionsParent.value && !suggestionsParent.value.contains(event.target)) {
      hideSuggestions();
    }
  }
};

onMounted(() => {
  document.addEventListener('click', handleClickOutside);
  window.addEventListener('scroll', updateDropdownPosition, true);
  window.addEventListener('resize', updateDropdownPosition);
});

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside);
  window.removeEventListener('scroll', updateDropdownPosition, true);
  window.removeEventListener('resize', updateDropdownPosition);
});
</script>

<style scoped>
.item-selector {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  width: 100%;
  height: min-content;
}

.combo {
  width: 100%;
  line-height: 1.4;
  overflow: hidden;
  text-align: left;
  display: flex;
  gap: 0.3rem;
  height: min-content;
}

.selector-container {
  width: 100%;
  min-height: 40px;
  display: flex;
  padding: 0.3rem;
  border-radius: var(--large-radius);
  background-color: var(--steel);
  cursor: text;
}

.selector-container:hover {
  outline: var(--transparent-line);
  outline-offset: -1px;
}

.search-input-container {
  position: relative;
  width: 100%;
  height: max-content;
  display: flex;
  align-items: center;
  box-sizing: border-box;
}

.input-field {
  flex: 1;
  min-width: 60px;
  height: 40px;
  font-family: Inter, sans-serif;
  font-size: 14px;
  font-weight: 300;
  color: var(--white);
  background: transparent;
  border: 0;
  outline: none;
  white-space: nowrap;
  width: 100%;
  box-sizing: border-box;
  /* background-color: crimson; */
  min-height: 40px;
}

.input-field::placeholder {
  color: var(--white);
  font-style: italic;
}

/* Suggestions Dropdown */
.suggestions-parent {
  z-index: 10000;
  min-height: 32px;
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  padding: 0.5rem;
  box-sizing: border-box;
  background-color: var(--black);
  border-radius: var(--large-radius);
  outline: var(--transparent-line);
  outline-offset: -1px;
  overflow-y: auto;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.suggestions-parent::-webkit-scrollbar {
  width: 6px;
  border-radius: 6px;
}

.suggestions-parent::-webkit-scrollbar-thumb {
  background-color: var(--light-steel);
  border-radius: 3px;
}

.suggestions-parent::-webkit-scrollbar-track {
  background-color: var(--dark-steel);
  border-radius: 3px;
}

.item-suggestion {
  color: var(--white);
  display: flex;
  align-items: center;
  box-sizing: border-box;
  gap: 0.5rem;
  font-size: 14px;
  width: 100%;
  height: min-content;
  padding: 0.5rem 0.75rem;
  border-radius: var(--normal-radius);
  cursor: pointer;
  transition: background-color 0.2s;
}

.item-suggestion:hover {
  background-color: var(--steel);
}

.item-meta {
  width: 100%;
  height: min-content;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.2rem;
}

.item-suggestion-name {
  font-weight: 400;
  width: 100%;
  display: flex;
  align-items: center;
  text-wrap: nowrap;
}

.item-suggestion-category {
  font-size: 12px;
  opacity: 0.6;
  width: 100%;
  display: flex;
  align-items: center;
  text-wrap: nowrap;
}

.no-results {
  color: rgba(255, 255, 255, 0.5);
  text-align: center;
  padding: 1rem;
  font-size: 0.875rem;
}

/* Animations */
.list-move,
.list-enter-active,
.list-leave-active {
  transition: all 0.2s ease;
}

.list-enter-from,
.list-leave-to {
  opacity: 0;
  transform: translateX(-10px);
}
</style>
