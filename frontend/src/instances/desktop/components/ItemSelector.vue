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
        
        <Teleport to="#app">
          <div 
            v-if="showSuggestions" 
            ref="suggestionsParent" 
            class="suggestions-parent" 
            v-esc="hideSuggestions" 
            v-stop-propagation
            :style="{ 
              top: dropdownTop + 'px', 
              width: dropdownWidth + 'px', 
              maxHeight: dropdownMaxHeight + 'px', 
              left: dropdownLeft + 'px' 
            }"
          >
            <div 
              v-for="item in filteredItems" 
              :key="item.id"
              class="item-suggestion" 
              @click="addItem(item)"
            >
              <img class="small-icons" :src="getItemIcon(item)" alt="">
              <div class="item-meta">
                <div class="item-suggestion-name">{{ item.name }}</div>
                <div v-if="item.category" class="item-suggestion-category">{{ item.category }}</div>
              </div>
            </div>
            
            <div v-if="filteredItems.length === 0 && searchQuery" class="no-results">
              No matching {{ itemType }}s found
            </div>
          </div>
        </Teleport>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, watchEffect } from 'vue';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';
import { getToolLogo, getSkillIcon } from '@/utils/iconMappers';

const iconStore = useIconStore();
const menu = useMenu();

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
  }
});

// Element refs
const observer = ref(null);
const selectorRoot = ref(null);
const comboBoxRoot = ref(null);
const inputField = ref(null);
const suggestionsParent = ref(null);
const searchInputContainer = ref(null);

// Refs
const searchQuery = ref('');
const showDropdown = ref(false);
const isInputActive = ref(false);

// Dropdown positioning
const dropdownTop = ref(0);
const dropdownLeft = ref(0);
const dropdownWidth = ref(0);
const dropdownMaxHeight = ref(0);

// Computed
const listItemsBoundary = computed(() => menu.contextMenuBounds);

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
  return iconStore.getAppIcon(iconName);
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
    return iconStore.getAppIcon(iconName);
  }
  // Fallback for other item types
  return item.icon ? iconStore.getAppIcon(item.icon) : null;
};

const addItem = (item) => {
  emit('itemAdded', item);
  inputField.value.focus();
  searchQuery.value = '';
};

const hideSuggestions = () => {
  showDropdown.value = false;
  searchQuery.value = '';
  updateDropdownPosition();
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
  }
  updateDropdownPosition();
};

const focusInput = () => {
  inputField.value?.focus();
  updateDropdownPosition();
};

const updateDropdownPosition = () => {
  if (!selectorRoot.value) return;
  
  const trayRootHeight = listItemsBoundary.value?.getBoundingClientRect().height || window.innerHeight;
  const rootRect = selectorRoot.value.getBoundingClientRect();
  const rootHeight = rootRect.height;
  const rootGlobalY = rootRect.top;
  const rootLeft = rootRect.left;

  dropdownTop.value = rootGlobalY + rootHeight + 10;
  dropdownWidth.value = rootRect.width;
  dropdownLeft.value = rootLeft;
  dropdownMaxHeight.value = trayRootHeight - rootHeight - rootGlobalY - 20;
};

const trackHeightChange = () => {
  updateDropdownPosition();
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

watchEffect(() => {
  if (showDropdown.value) {
    updateDropdownPosition();
  }
});

onMounted(() => {
  observer.value = new ResizeObserver(trackHeightChange);
  observer.value.observe(selectorRoot.value);
  document.addEventListener('click', handleClickOutside);
});

onBeforeUnmount(() => {
  if (observer.value) {
    observer.value.disconnect();
  }
  document.removeEventListener('click', handleClickOutside);
});
</script>

<style scoped>
.item-selector {
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
  border-radius: var(--normal-radius);
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
  color: rgba(255, 255, 255, 0.4);
}

/* Suggestions Dropdown */
.suggestions-parent {
  position: absolute;
  z-index: 10000;
  min-height: 32px;
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  padding: 0.5rem;
  box-sizing: border-box;
  background-color: var(--black);
  border-radius: var(--small-radius);
  outline: var(--transparent-line);
  outline-offset: -1px;
  overflow-y: auto;
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
  border-radius: var(--small-radius);
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
