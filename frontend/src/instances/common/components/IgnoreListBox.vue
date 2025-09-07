<template>
  <div ref="ignoreListRoot" class="ignore-list-root"  v-esc="escape">
      <div @click="focusOnInput()" class="ignore-list-container tint combo-linear">
            <div class="added-users">
                <div v-if="selectedItems.length" class="ignored-item-wrapper">
                    <div v-for="item in selectedItems" 
                    class="ignored-item" :class="{ 'ignored-folder': isFileWithExtension(item) }">
                        <div class="ignored-item-name">
                            {{item}}
                        </div>
                        <span class="single-action-button"  @click="removeItem(item)"  v-tooltip="'Remove'">
                          <img class="small-icons" src="/icons/close.svg">
                        </span>
                    </div>
                </div>
                <input ref="listBoxParent" v-focus :placeholder="placeholder" v-model="searchQuery"
                    v-return="handleEnterKey" @keydown.delete.stop="removeLastItem" @blur="handleInputBlur"
                    @focus="handleInputFocus" @paste="handlePaste" @keydown="handleKeyDown" class="input-field new-user" />
            </div>
      </div>
  </div>
</template>
  
<script setup>

import { ref, computed, onMounted, onBeforeUnmount, onUnmounted, watchEffect } from 'vue';

// stores
import { useMenu } from '@/stores/menu';

// states
const menu = useMenu();

// emits
const emit = defineEmits([
  'input', 'on-remove', 
  'on-focus', 'on-blur', 
  'itemRemoved', 'itemAdded'
]);

// props
const props = defineProps({
  placeholder: { type: String, default: '' },
  chips: { type: Array, default: () => [] },
  suggestions: { type: Array, default: () => [] },
  displaySuggestions: { type: Boolean, default: true },
  selectedItems: { type: Array, default: () => [] },
  allItems: { type: Array, default: () => [] },
});
  
// element refs
const listBoxParent = ref(null);
const ignoreListRoot = ref(null);

// refs
const searchQuery = ref('');
const isInputActive = ref(false);
  
// methods
const isFileWithExtension = (item) => {
  return item && 
         typeof item === 'string' && 
         !item.includes('/') && 
         !item.includes('\\');
};

const trimLeadingPeriods = (str) => {
  return str.replace(/^\.+/, '');
};

const addItem = (event) => {
  let input = event.target.value.trim();
  emit('itemAdded', input);
  searchQuery.value = '';
  listBoxParent.value.focus();
}

const handlePaste = (event) => {
  const pastedText = event.clipboardData.getData('text');

  if (pastedText && /[,\s\n]/.test(pastedText)) {
    event.preventDefault();
    
    processInputEntries(pastedText);
  }
};

const processInputEntries = (input) => {

  const entries = input.split('\n').filter(entry => {
    const trimmed = entry.trim();

    return trimmed !== '' && 
           !trimmed.startsWith('#') &&
           !trimmed.startsWith('//') &&
           !trimmed.startsWith('/*');
  });
  
  // Add each valid entry
  for (const entry of entries) {
    const trimmedEntry = entry.trim();
    if (trimmedEntry && !props.selectedItems.includes(trimmedEntry)) {
      emit('itemAdded', trimmedEntry);
    }
  }
  
  searchQuery.value = '';
}

const removeItem = (item) => {
  emit('itemRemoved', item);
}
  
const handleEnterKey = (event) => {
};
  
  
const escape = () => {
};
  
const handleKeyDown = (event) => {
  if (event) {
      if (event.keyCode === 13) {
        if (event.target.value.trim() !== "") {
            addItem(event);
        }
    }
  }
};

const handleInputFocus = (event) => {
  isInputActive.value = true;
  emit('on-focus', event);
};

const focusOnInput = () => {
    listBoxParent.value.focus();
};
  
const handleInputBlur = (event) => {
  isInputActive.value = false;
  emit('on-blur', event);
};

const removeLastItem = () => {
  if(searchQuery.value) return
  let selectedItems = props.selectedItems;
  if(!selectedItems.length) return;
  const lastEntry = getLastEntry(selectedItems);
  removeItem(lastEntry);
};

const getLastEntry = (array) => {
  if (!array || array.length === 0) {
      return undefined;
  }
  return array[array.length - 1];
};


onMounted(() => {

});

onBeforeUnmount(() => {

});

onUnmounted(() => {

});

</script>
  
<style scoped>
@import "@/assets/desktop.css";

/* User Item Styles */
.ignored-item {
  color: var(--white);
  display: flex;
  align-items: center;
  gap: .5rem;
  font-size: medium;
  width: min-content;
  height: min-content;
  padding: .1rem .1rem .1rem .8rem;
  background-color: var(--steel);
  border-radius: var(--small-radius);
}

.ignored-folder {
    background-color: rgb(0, 161, 86);
}

.ignored-item-name {
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
}

.ignored-item-wrapper {
  width: 100%;
  height: min-content;
  display: flex;
  flex-wrap: wrap;
  border-radius: var(--small-radius);
  gap: .4rem;
}

.new-user {
  width: 100%;
  min-height: 35px;
  background-color: hotpink;
  border-radius: var(--small-radius);
}

/* Search Tags and Combo Styles */
.ignore-list-root {
  display: flex;
  flex-direction: column;
  gap: .5rem;
  width: 100%;
  height: 100%;
  /* background-color: cadetblue; */
}

.ignore-list-container {
  width: 100%;
  /* min-height: 32px; */
  height: 100%;
  display: flex;
  flex-wrap: wrap;
  gap: .1rem;
  padding: .1rem;
  border-radius: var(--small-radius);
  background-color: var(--transparent-black);
  overflow-y: auto;
}

.ignore-list-container:hover {
  outline: var(--transparent-line);
  outline-offset: -1px;
}

.ignore-list-container::-webkit-scrollbar {
  width: 4px;
}

.ignore-list-container::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--light-steel);
}

.ignore-list-container::-webkit-scrollbar-track {
  border-radius: 10px;
}

.input-field {
  flex: 1;
  min-width: 60px;
  height: 27px;
  font-family: Inter, sans-serif;
  font-size: 16px;
  color: var(--white);
  background: transparent;
  border: 0;
  outline: none;
  white-space: nowrap;
  width: 100%;
  box-sizing: border-box;
}


.desktop-input-long{
  font-family: 'Inter', sans-serif;
  font-weight: 300;
  box-sizing: border-box;
  font-size: 16px;
  border-radius: 8px;
  box-shadow: 0px 1px 2px 0px rgba(16, 24, 40, 0.05000000074505806);
  padding: 10px;
  border-color: rgba(234, 236, 240, 1);
  border-width: 0px;
  border-style: solid;
  background-color: rgba(255, 255, 255, 1);
  color: rgba(52, 64, 84, 1);
  width: 100%;
  min-height: 60px;
  height: 100px;
  max-height: 180px;
  margin: auto;
  margin-top: 20px;
  outline: none;
  resize: none;
  outline: none;
  background-color: var(--midnight-steel);
  color: var(--white);
}

.desktop-input-long:hover{
  outline: 1px solid var(--white);
  outline-offset: -1px;
}

.desktop-input-long:focus{
  outline: 1px solid var(--white);
  outline-offset: -1px;
}
</style>


