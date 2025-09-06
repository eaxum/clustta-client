<template>
  <div ref="searchTagsRoot" class="search-tags" v-esc="escape">
    <div class="combo">
      <div @click="focusNewChip()" class="search-container tint combo-linear" ref="comboBoxRoot">
        <div class="added-users" :class="{'multiple-entries' : allowMultipleEntries }" ref="addedUsersContainer">
          <div v-if="selectedItems.length" class="user-item-wrapper">
            <div v-for="item in selectedItems" class="user-item" :class="getUserItemClass(item)">
              <div class="profile-picture" :style="{ backgroundColor: item.avatarColor }">
                <img class="profile-img" :src="item.photo || '/icons/default_profile_picture.svg'">
              </div>
              <div class="user-item-name">
                {{ getUserDisplayName(item) }}
              </div>
              <div v-if="item.userType !== 'studio'" class="user-type-indicator" :class="getUserTypeClass(item.userType)">
                {{ getUserTypeLabel(item.userType) }}
              </div>
              <span class="single-action-button" @click="removeItem(item)" v-tooltip="'Remove'">
                <img class="small-icons" src="/icons/close.svg">
              </span>
            </div>
          </div>
          <input v-focus ref="listBoxParent" :placeholder="placeholder" v-model="searchQuery" v-return="handleEnterKey"
            autocomplete="off"
            @keydown.delete.stop="removeLastItem" @keydown="handleKeyDown" 
            @blur="handleInputBlur" @focus="handleInputFocus"
            @input="makeItNormal" class="input-field new-user" />
        </div>
        <Teleport to="#app">
          <div v-if="showTagSuggestions" ref="tagParent" class="tag-parent" v-esc="hideSuggestions" v-stop-propagation
            :style="{ top: listItemsAnchor + 'px', width: listItemsWidth + 'px', maxHeight: listItemMaxHeight + 'px', left: listItemsLeft + 'px' }">

            <div v-for="item in filteredItems" class="user-item-suggestion" @click="addItem(item)">
              <div class="profile-picture" :style="{ backgroundColor: item.avatarColor }">
                <img class="profile-img" :src="item.photo ? item.photo : '/icons/default_profile_picture.svg'">
              </div>
              <div class="user-item-meta">
                <div class="user-item-name" :class="{ 'user-item-name-compact': !displayEmail }">
                  {{ item.full_name }}
                </div>
                <div v-if="displayEmail" class="user-item-email">
                  {{ item.email }}
                </div>
              </div>
            </div>

          </div>
        </Teleport>

      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, onUnmounted, watchEffect } from 'vue';

// stores
import { useMenu } from '@/stores/menu';

// services (for email validation and API calls)
import { AuthService } from "@/../bindings/clustta/services";

// states
const menu = useMenu();

// emits
const emit = defineEmits([
  'input', 'on-remove',
  'on-focus', 'on-blur',
  'tagRemoved', 'tagAdded', 'suggestionsVisibility'
]);

// props
const props = defineProps({
  placeholder: { type: String, default: '' },
  chips: { type: Array, default: () => [] },
  suggestions: { type: Array, default: () => [] },
  displaySuggestions: { type: Boolean, default: true },
  allowMultipleEntries: { type: Boolean, default: false },
  displayEmail: { type: Boolean, default: true },
  selectedItems: { type: Array, default: () => [] },
  allItems: { type: Array, default: () => [] },
});

// element refs
const observer = ref(null);
const tagItem = ref(null);
const tagParent = ref(null);
const comboBoxRoot = ref(null);
const listBoxParent = ref(null);
const searchTagsRoot = ref(null);
const suggestionsButton = ref(null);
const listItemsAnchor = ref(null);
const listItemsWidth = ref(null);
const addedUsersContainer = ref(null);

// refs
const newChip = ref('');
const searchQuery = ref('');
const selectedCategory = ref('');
const pendingEmails = ref(new Set()); // Track emails being validated

const searchTags = ref(false);
const isInputActive = ref(false);

const listItemsLeft = ref(0);
const listItemMaxHeight = ref(0);
const selectedCategoryIndex = ref(0);

// Email validation regex
const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

// computed props
const listItemsBoundary = computed(() => menu.contextMenuBounds);
const useInputActivity = computed(() => { return searchQuery.value && props.displaySuggestions && filteredItems.value.length !== 0 });
const showTagSuggestions = computed(() => { return searchTags.value && filteredItems.value.length !== 0 || useInputActivity.value });



const filteredItems = computed(() => {
  const allItems = props.allItems;
  if (!searchQuery.value) {
    return allItems;
  }
  const lowerSearchTerm = searchQuery.value.toLowerCase();
  let matches = 0;

  return allItems.filter(item => {
    const searchRange = item.full_name + item.email
    if (searchRange.toLowerCase().includes(lowerSearchTerm) && matches < 10) {
      matches++;
      return item;
    }
  });
});

// methods
const addItem = (item) => {
  // console.log(item)
  emit('tagAdded', item);
  listBoxParent.value.focus();
  searchQuery.value = '';
  
  // Scroll to bottom to ensure input field remains visible
  if (addedUsersContainer.value) {
    setTimeout(() => {
      addedUsersContainer.value.scrollTop = addedUsersContainer.value.scrollHeight;
    }, 0);
  }
}

const removeItem = (item) => {
  emit('tagRemoved', item);
}

// Handle space or comma key press for email input
const handleSpaceOrComma = async (event) => {
  if (event.key === ' ' || event.key === ',') {
    event.preventDefault();
    const email = searchQuery.value.trim();
    
    if (email && emailRegex.test(email)) {
      await processEmailInput(email);
      searchQuery.value = '';
    }
  }
};

// Process email input and determine user type
const processEmailInput = async (email) => {
  if (pendingEmails.value.has(email)) return; // Avoid duplicate calls
  
  // Check if email is already selected
  const isAlreadySelected = props.selectedItems.some(item => item.email === email);
  if (isAlreadySelected) return;
  
  // Check if it's a studio user first
  const studioUser = props.allItems.find(item => item.email === email);
  if (studioUser) {
    // It's already a studio user, add normally
    addItem(studioUser);
    return;
  }
  
  pendingEmails.value.add(email);
  
  try {
    // Check if user exists globally
    const emailExists = await AuthService.CheckEmailExists(email);
    console.log('Email check result for', email, ':', emailExists);
    
    const newUser = {
      email: email,
      full_name: email, // Fallback name
      id: email, // Use email as ID for non-studio users
      avatarColor: generateAvatarColor(email),
      userType: emailExists ? 'user' : 'new',
      photo: null
    };
    
    addItem(newUser);
  } catch (error) {
    console.error('Error checking email:', error);
    // Treat as new user if API call fails
    const newUser = {
      email: email,
      full_name: email,
      id: email,
      avatarColor: generateAvatarColor(email),
      userType: 'new',
      photo: null
    };
    addItem(newUser);
  } finally {
    pendingEmails.value.delete(email);
  }
};

// Generate a color for avatar based on email
const generateAvatarColor = (email) => {
  const colors = ['#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4', '#FFEAA7', '#DDA0DD', '#98D8C8', '#F7DC6F'];
  let hash = 0;
  for (let i = 0; i < email.length; i++) {
    hash = email.charCodeAt(i) + ((hash << 5) - hash);
  }
  return colors[Math.abs(hash) % colors.length];
};

// Get CSS class for user item based on type
const getUserItemClass = (item) => {
  if (!item.userType) return '';
  return `user-item-${item.userType}`;
};

// Get display name for user
const getUserDisplayName = (item) => {
  return item.full_name || item.email;
};

// Get CSS class for user type indicator
const getUserTypeClass = (userType) => {
  return `user-type-${userType}`;
};

// Get label for user type
const getUserTypeLabel = (userType) => {
  switch (userType) {
    case 'user': return 'User';
    case 'new': return 'New';
    case 'studio': return ''; // Studio users don't need a label
    default: return '';
  }
};

const hideSuggestions = () => {
  updateAnchor();
  emit('suggestionsVisibility', searchTags.value);
  searchQuery.value = '';
  searchTags.value = false;
}

const handleEnterKey = (event) => {
};


const escape = () => {
};

const addNewTag = (event) => {
  let newTag = event.target.value
  emit('tagAdded', newTag);
  listBoxParent.value.focus();
}

const handleKeyDown = async (event) => {
  if (event) {
    // Handle space and comma for email input
    if (event.key === ' ' || event.key === ',') {
      await handleSpaceOrComma(event);
      return;
    }
    
    if (event.keyCode === 188 || event.keyCode === 13) {
      if (searchTags.value) {
        if (event.target.value.trim() !== "") {
          addNewTag(event);
          searchQuery.value = '';
        }
      }
    } else if (event.keyCode === 38 || event.keyCode === 40) {
      event.preventDefault();
    } else if (event.keyCode === 13) {
      newChip.value = selectedCategory.value;
      listBoxParent.value.focus();
      newChip.value = '';
    }
  }
};

const navigate = (direction) => {
  if (filteredItems.value.length) {
    selectedCategoryIndex.value = (selectedCategoryIndex.value + direction + filteredItems.value.length) % filteredItems.value.length;
    selectedCategory.value = filteredItems.value[selectedCategoryIndex.value];
    const element = document.querySelectorAll('.combo-list-item')[selectedCategoryIndex.value];
    element.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }
};

const makeItNormal = (event) => {
  emit('input', event.target.value);

  handleKeyDown(event);
  selectedCategory.value = filteredItems.value[0];
  selectedCategoryIndex.value = 0;
  if (!searchTags.value) {
    emit('input', event.target.value);
  }
};

const focusNewChip = () => {
  updateAnchor();
};

const updateAnchor = () => {
  const trayRootHeight = listItemsBoundary.value.getBoundingClientRect().height;
  const searchTagsRootOffset = searchTagsRoot.value.offsetTop;
  const searchTagsRootHeight = searchTagsRoot.value.getBoundingClientRect().height;
  const searchTagsRootGlobalY = searchTagsRoot.value.getBoundingClientRect().top;
  const searchTagsRootLeft = searchTagsRoot.value.getBoundingClientRect().left;

  listItemsAnchor.value = searchTagsRootOffset + searchTagsRootHeight + 10;
  listItemsAnchor.value = searchTagsRootGlobalY + searchTagsRootHeight + 10;
  listItemsWidth.value = searchTagsRoot.value.getBoundingClientRect().width;
  listItemsLeft.value = searchTagsRootLeft;
  listItemMaxHeight.value = trayRootHeight - searchTagsRootHeight - searchTagsRootGlobalY;
};

const trackHeightChange = (entries) => {
  updateAnchor();
};



const handleInputFocus = (event) => {
  isInputActive.value = true;
  emit('on-focus', event);
};

const handleInputBlur = (event) => {
  isInputActive.value = false;
  emit('on-blur', event);
};

const handleClickOutside = (event) => {
  if (!suggestionsButton.value?.contains(event.target)) {
    if (!listBoxParent.value.contains(event.target)) {
      if (tagParent.value && !tagParent.value.contains(event.target)) {

      };
    };
  };
};

const removeLastItem = () => {
  if (searchQuery.value) return
  let selectedItems = props.selectedItems;
  if (!selectedItems.length) return;
  const lastEntry = getLastEntry(selectedItems);
  removeItem(lastEntry);
};

const getLastEntry = (array) => {
  if (!array || array.length === 0) {
    return undefined;
  }
  return array[array.length - 1];
};

const hideListContent = (event) => {
  if (searchTags.value && (event.target !== listBoxParent.value)) {
    searchTags.value = false;
  }
};

watchEffect(() => {
  if (menu.clickOutsideMask) {
    menu.clickOutsideMask.addEventListener('click', hideListContent);
  }
});

onMounted(() => {
  observer.value = new ResizeObserver(trackHeightChange);
  observer.value.observe(searchTagsRoot.value);
  document.addEventListener('click', handleClickOutside);

});

onBeforeUnmount(() => {
  if (observer.value) {
    observer.value.disconnect();
  }
  document.removeEventListener('click', handleClickOutside);

});

onUnmounted(() => {
  if (menu.clickOutsideMask) {
    menu.clickOutsideMask.removeEventListener('click', hideListContent);
  }
});

</script>

<style scoped>
@import "@/assets/desktop.css";

/* Profile and User Item Styles */
.profile-picture {
  background-color: black;
  height: 24px;
  min-width: 24px;
  overflow: hidden;
  display: flex;
  align-items: center;
  border-radius: 24px;
}

.profile-img {
  width: 100%;
  height: 100%;
}

.user-item {
  display: flex;
  align-items: center;
  gap: .5rem;
  font-size: medium;
  font-size: 14px;
  width: min-content;
  height: min-content;
  padding: .2rem .2rem .2rem .4rem;
  background-color: var(--steel);
  border-radius: var(--small-radius);
}

.user-item-suggestion {
  color: var(--white);
  display: flex;
  align-items: center;
  box-sizing: border-box;
  gap: .5rem;
  font-size: medium;
  width: 100%;
  height: min-content;
  padding: .4rem .8rem .4rem .4rem;
  /* background-color: var(--steel); */
  border-radius: var(--small-radius);
}

.user-item-suggestion:hover {
  background-color: var(--black-steel);
  /* outline: var(--transparent-line);
  outline-offset: -1px; */
}

.user-item-meta {
  width: min-content;
  height: min-content;
  display: flex;
  /* background-color: hotpink; */
  width: 100%;
  justify-items: flex-start;
  flex-direction: column;
  align-items: flex-start;
  gap: .2rem;
}

.user-item-name {
  /* background-color: forestgreen; */
  font-weight: 250;
  width: min-content;
  height: min-content;
  display: flex;
  align-items: center;
  text-wrap: nowrap;
}

.user-item-name-compact {
  font-size: 14px;
}

.user-item-email {
  font-size: 14px;
  /* background-color: forestgreen; */
  opacity: .6;
  width: min-content;
  height: min-content;
  display: flex;
  align-items: center;
  text-wrap: nowrap;
}

/* Container Styles */
.added-users {
  position: relative;
  width: 100%;
  height: max-content;
  display: flex;
  flex-direction: column;
  gap: .2rem;
  align-items: center;
  padding: .2rem;
  box-sizing: border-box;
  overflow: hidden;
  transition: all .2s ease-in-out;
  max-height: 200px;
  overflow: hidden;
  /* background-color: crimson; */
}

.multiple-entries{
  overflow-y: scroll;
}

.added-users::-webkit-scrollbar {
  width: 6px;
  border-radius: 6px;
}

.added-users::-webkit-scrollbar-thumb {
  background-color: rgba(255, 255, 255, 0.295);
  background-color: var(--steel);
  border-radius: 6px;
}

.added-users::-webkit-scrollbar-track {
  border-radius: 6px;
}

.user-item-wrapper {
  width: 100%;
  height: min-content;
  display: flex;
  flex-wrap: wrap;
  /* padding: .2rem .4rem; */
  border-radius: var(--small-radius);
  gap: .4rem;
}

.new-user {
  width: 100%;
  min-height: 35px;
  background-color: hotpink;
  /* padding: .2rem .4rem; */
  border-radius: var(--small-radius);
}

/* Tag Styles */
.tag-parent {
  position: absolute;
  /* z-index: 1000; */
  min-height: 32px;
  display: flex;
  flex-direction: column;
  gap: .5rem;
  padding: .3rem;
  box-sizing: border-box;
  background-color: var(--dark-steel);
  border-radius: var(--small-radius);
  outline: var(--transparent-line);
  outline-offset: -1px;
  overflow-y: auto;
}

.tag-parent::-webkit-scrollbar {
  width: 8px;
  border-radius: 8px;
}

.tag-parent::-webkit-scrollbar-thumb {
  background-color: rgba(255, 255, 255, 0.295);
  border-radius: 8px;
}

.tag-parent::-webkit-scrollbar-track {
  background-color: rgba(0, 0, 0, 0.295);
  border-radius: 8px;
}

.tag-item {
  width: 100%;
  padding: 10px;
  color: var(--white);
  outline: solid 1px #2e2e2e;
  border-radius: var(--small-radius);
  box-sizing: border-box;
}

.tag-item:hover {
  background-color: #5f5f5f;
}

/* Search Tags and Combo Styles */
.search-tags {
  /* z-index: 1000000; */
  display: flex;
  flex-direction: column;
  gap: .5rem;
  width: 100%;
  height: min-content;
}

.combo {
  width: 100%;
  line-height: 1.4;
  overflow: hidden;
  text-align: left;
  display: flex;
  gap: .3rem;
  height: min-content;
}

.search-container {
  width: 100%;
  min-height: 32px;
  display: flex;
  flex-wrap: wrap;
  gap: .1rem;
  padding: .1rem;
  max-height: 50vh;
  border-radius: var(--normal-radius);
  background-color: var(--light-steel);
  overflow-y: auto;
}

[data-theme="dark"] .search-container{
  background-color: var(--midnight-steel);
}

.search-container:hover {
  outline: var(--transparent-line);
  outline-offset: -1px;
}

.input-field {
  flex: 1;
  min-width: 60px;
  height: 27px;
  /* margin: 3px; */
  /* padding: 0 4px; */
  font-family: Inter, sans-serif;
  font-size: 14px;
  font-weight: 300;
  color: var(--white);
  background: transparent;
  border: 0;
  outline: none;
  white-space: nowrap;
  /* background-color: hotpink; */
  width: 100%;
  box-sizing: border-box;
}

/* User Type Styles */
.user-item-user {
  border: 1px solid #FFA500;
  background-color: rgba(255, 165, 0, 0.1);
}

.user-item-new {
  border: 1px solid #2D9CDB;
  background-color: rgba(45, 156, 219, 0.1);
}

.user-type-indicator {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 10px;
  text-transform: uppercase;
  font-weight: 500;
}

.user-type-user {
  background-color: #FFA500;
  color: white;
}

.user-type-new {
  background-color: #2D9CDB;
  color: white;
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

.list-leave-active {
  position: absolute;
  left: var(--x);
  top: var(--y);
}
</style>