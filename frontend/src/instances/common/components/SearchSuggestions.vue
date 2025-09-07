<template>
  <div ref="searchTagsRoot" class="search-tags" :class="{ 'search-tags-closed': !displayTagContainer }" v-esc="escape">
    <div class="combo">

      <div @click="focusNewChip()" class="search-container tint combo-linear" ref="comboBoxRoot"
        :class="{ 'combo--focus': isInputActive, 'combo--error': isError }">


        <input ref="listBoxParent" :placeholder="placeholder" v-model="trayStates.tagSearchQuery"
          v-return="handleEnterKey" @keydown.delete.stop="removeLastChip" @keydown="addNew" @blur="handleInputBlur"
          @focus="handleInputFocus" @input="makeItNormal" class="combo-new-chip" />

        <!--  for tag suggestions -->
        <Teleport to="#app">
          <div v-if="showTagSuggestions && !props.noWrap" ref="tagParent" class="tag-parent" v-esc="hideSuggestions"
            :style="{ top: listItemsAnchor + 'px', width: listItemsWidth + 'px', maxHeight: listItemMaxHeight + 'px', left: listItemsLeft + 'px' }">
            <div ref="tagItem" v-for="(tag) in filteredTags" :key="tag" @click="addTag(tag)" v-stop-propagation
              class="tag-item chip">
              {{ tag.toLowerCase().replace(/_/g, " ") }}
              <img class="small-icons add-chip-icon" src="/icons/add.svg">
            </div>
          </div>

          <div v-if="isList" class="list-tags" v-esc="hideSuggestions"
            :style="{ top: listItemsAnchor + 'px', width: listItemsWidth + 'px', maxHeight: listItemMaxHeight + 'px', left: listItemsLeft + 'px' }">
            <div ref="tagItem" v-for="(tag) in filteredTags" :key="tag" @click="addTag(tag)" class="combo-list-item">
              {{ tag.toLowerCase().replace(/_/g, " ") }}
              <!-- <img class="small-icons" src="/icons/add.svg"> -->
            </div>
          </div>

        </Teleport>
      </div>

      <div class="combo-actions">
        <span class="single-action-button" ref="suggestionsButton"
          :class="{ 'single-action-button-pressed': searchTags }" @click="toggleSearchTags"
          v-tooltip="trayStates.pinSuggestions ? 'Hide Suggestion' : 'Show Suggestions'">
          <img class="small-icons" src="/icons/suggestions.svg">
        </span>
        <span v-if="showSearchToggle" class="single-action-button" @click="toggleSearch" v-tooltip="'Hide Search'"><img
            class="small-icons" src="/icons/search.svg">
        </span>
      </div>



    </div>
    <div class="tag-container">
      <TagContainer :tags="tags" :queryTag="queryTag" @tagRemoved="removeTag" @queryTagRemoved="removeQueryTag" />

    </div>


  </div>

</template>

<script setup>
import { ref, computed, onMounted, watch, nextTick, onBeforeUnmount } from 'vue';
import { useTrayStates } from '@/stores/TrayStates';
import { useCommonStore } from '@/stores/common';
import TagContainer from '@/instances/common/components/TagContainer.vue';


const trayStates = useTrayStates();
const commonStore = useCommonStore();

// const emptyTags = computed(() => {return props.tags.length === 0 ? true : false || !props.queryTag });
const emptyTags = computed(() => { return props.tags.length !== 0 });
const queryTagAvailable = computed(() => { return props.queryTag });
const displayTagContainer = computed(() => { return emptyTags.value || queryTagAvailable.value });
// const showTags = computed(() => {return !emptyTags.value});

const emit = defineEmits(['input', 'update:modelValue', 'on-limit', 'on-chips-changed', 'on-remove', 'on-error', 'on-focus', 'on-blur',
  'toggleSearch', 'tagRemoved', 'tagAdded', 'queryTagAdded', 'queryTagRemoved', 'changePlaceholder', 'suggestionsVisibility']);
const props = defineProps({
  readOnly: { type: Boolean, default: false },
  noWrap: { type: Boolean, default: false },
  parentClose: { type: Boolean, default: false },
  modelValue: { type: String, default: '', },
  validate: { type: [String, Function, Object], default: "" },
  addChipOnKeys: {
    type: Array,
    default: function () {
      return [
        13, // Enter
        188, // Comma ','
        // 32, // Space
      ];
    }
  },

  placeholder: { type: String, default: '' },
  chips: { type: Array, default: () => [] },
  limit: { type: Number, default: -1 },
  chipLength: { type: Number, default: -1 },
  allowDuplicates: { type: Boolean, default: false },
  showSearchToggle: { type: Boolean, default: false },
  addChipOnBlur: { type: Boolean, default: false },
  inputChips: { type: Array, default: () => [] },
  suggestions: { type: Array, default: () => [] },
  outerSearch: { type: String, default: '' },
  pinSuggestions: { type: Boolean, default: false },
  displaySuggestions: { type: Boolean, default: true },
  forSearch: { type: Boolean, default: false },
  tags: { type: Array, default: () => [] },
  projectTags: { type: Array, default: () => [] },
  queryTag: { type: String, default: '' },
});

// Define reactive data properties
const listBoxParent = ref(null);
const searchTagsRoot = ref(null);
const searchTags = ref(false);
const searchPlaceholder = ref('Search Tasks')
const outerChipWrapper = ref(null);
const comboBoxRoot = ref(null);
const tagItem = ref(null);
const tagParent = ref(null);
const suggestionsButton = ref(null);
const isInputActive = ref(false);
const isError = ref(false);
const newChip = ref('');
const tagSearchQuery = computed(() => { return trayStates.tagSearchQuery });
const innerChips = ref([]);
const selectedCategory = ref('');
const selectedCategoryIndex = ref(0);
const useInputActivity = computed(() => { return trayStates.tagSearchQuery && props.displaySuggestions && filteredTags.value.length !== 0 });
const showTagSuggestions = computed(() => { return searchTags.value && filteredTags.value.length !== 0 || useInputActivity.value });
const isList = computed(() => { return showTagSuggestions.value && props.noWrap });
const observer = ref(null);

const filteredTags = computed(() => {

  const originalTags = props.projectTags;

  if (!tagSearchQuery.value) {
    return originalTags;
  }

  const lowerSearchTerm = tagSearchQuery.value.toLowerCase();
  let matches = 0;

  return originalTags.filter(tag => {
    if (tag.toLowerCase().includes(lowerSearchTerm) && matches < 10) {
      matches++;
      return tag;
    }
  });
});

const toggleSearch = () => {
  // if(searchTags.value){
  //   searchTags.value = false
  // }
  emit('toggleSearch', commonStore.viewSearchQuery);
  // searchTags.value = false;
  // //console.log('jjjj')

};


const addTag = (tag) => {
  emit('tagAdded', tag);
  listBoxParent.value.focus();
  trayStates.tagSearchQuery = '';
}

const removeTag = (tag) => {
  emit('tagRemoved', tag);
}

const removeQueryTag = () => {
  emit('queryTagRemoved');
}

const addQueryTag = (tag) => {
  emit('queryTagAdded');
}

const hideSuggestions = () => {
  updateAnchor();
  emit('suggestionsVisibility', searchTags.value);
  trayStates.tagSearchQuery = '';
  searchTags.value = false;
}

const handleEnterKey = (event) => {
  // confirmChange();
};


const escape = () => {
  if (trayStates.showTraySearch) {
    trayStates.showTraySearch = false;
  }
};

const toggleSearchTags = () => {
  updateAnchor();

  searchTags.value = !searchTags.value;
  emit('suggestionsVisibility', searchTags.value);

  if (searchTags.value) {
    searchPlaceholder.value = "Search Tags";
    emit('changePlaceholder', searchPlaceholder.value);
  } else {
    searchPlaceholder.value = "Search Tasks";
    emit('changePlaceholder', searchPlaceholder.value);
  }


  listBoxParent.value.focus();
  emit('input', trayStates.tagSearchQuery);

  if (searchTags.value && commonStore.viewSearchQuery !== '') {
    emit('queryTagAdded', commonStore.viewSearchQuery);
    //console.log('queryTagAdded');
    trayStates.tagSearchQuery = '';
  } else if (!searchTags.value) {
    trayStates.tagSearchQuery = props.queryTag;
    commonStore.viewSearchQuery = props.queryTag;
    removeQueryTag();
    //console.log('queryTagRemoved');
  }

}
// Define watchers
watch(() => props.inputChips, (value) => {
  innerChips.value = value;
  //console.log(innerChips.value)
});

// Define computed properties
const isLimit = computed(() => {
  let isLimit = props.limit > 0 && Number(props.limit) === innerChips.value.length;
  if (isLimit) {
    emit('on-limit');
  }
  return isLimit;
});


const handleWheel = (event) => {
  const element = outerChipWrapper.value;
  const delta = Math.sign(event.deltaY);
  element.scrollLeft += delta * 50; // Adjust scroll speed as needed
  event.preventDefault(); // Prevent the default vertical scrolling behavior
};

const scrollElement = (direction) => {
  if (outerChipWrapper.value) {
    const elementWidth = outerChipWrapper.value.clientWidth;
    // canScrollRef();
    outerChipWrapper.value.scrollBy({
      left: elementWidth * direction,
      behavior: 'smooth',
    });
  }
};



const addNewTag = (event) => {
  let newTag = event.target.value
  emit('tagAdded', newTag);
  listBoxParent.value.focus();
  // newChip.value = '';
}

const handleKeyDown = (event) => {
  if (event) {
    if (event.keyCode === 188 || event.keyCode === 13) {
      // if (event.keyCode === 188 || event.keyCode === 32) {
      if (searchTags.value || !props.forSearch) {
        if (event.target.value.trim() !== "") {
          addNewTag(event);
          trayStates.tagSearchQuery = '';
        }
      }
    } else if (event.keyCode === 38 || event.keyCode === 40) {
      event.preventDefault();
      navigate(event.keyCode === 38 ? -1 : 1);
    } else if (event.keyCode === 13) {
      //console.log(selectedCategory.value);
      newChip.value = selectedCategory.value;
      addNew();
      listBoxParent.value.focus();
      //displaySuggestions.value = true;
      newChip.value = '';
    }
    // //console.log('not prevented');
  }
};

const navigate = (direction) => {
  if (filteredTags.value.length && !props.pinSuggestions) {
    selectedCategoryIndex.value = (selectedCategoryIndex.value + direction + filteredTags.value.length) % filteredTags.value.length;
    selectedCategory.value = filteredTags.value[selectedCategoryIndex.value];
    const element = document.querySelectorAll('.combo-list-item')[selectedCategoryIndex.value];
    element.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }
};

const makeItNormal = (event) => {
  emit('input', event.target.value);

  handleKeyDown(event);
  //displaySuggestions.value = true;
  selectedCategory.value = filteredTags.value[0];
  selectedCategoryIndex.value = 0;
  if (!searchTags.value) {
    emit('input', event.target.value);
  }
};

const resetData = () => {
  innerChips.value = [];
};

const focusNewChip = () => {
  updateAnchor();
  if (props.readOnly || !document.querySelector(".combo-new-chip")) {
    return;
  }
};

const listItemsBoundary = computed(() => trayStates.listItemsBoundary);
const listItemsAnchor = ref(null);
const listItemsWidth = ref(null);
const listItemMaxHeight = ref(0);
const listItemsLeft = ref(0);
const trackedPin = ref(null);

// watch(trackedPin, (newValue, oldValue) => {
//   if (newValue !== oldValue) {
//     //console.log('trackedPin changed:', newValue);
//     updateAnchor();
//   }
// });

const updateAnchor = () => {
  // Check if required elements exist before accessing their properties
  if (!listItemsBoundary.value || !searchTagsRoot.value || !listBoxParent.value) {
    return;
  }

  const trayRootHeight = listItemsBoundary.value.getBoundingClientRect().height;
  const searchTagsRootOffset = searchTagsRoot.value.offsetTop;
  const searchTagsRootHeight = searchTagsRoot.value.getBoundingClientRect().height;
  const searchTagsRootGlobalY = searchTagsRoot.value.getBoundingClientRect().top;
  const searchTagsRootLeft = searchTagsRoot.value.getBoundingClientRect().left;

  const listBoxParentBottom = listBoxParent.value.getBoundingClientRect().bottom;

  listItemsAnchor.value = searchTagsRootOffset + searchTagsRootHeight + 10;
  listItemsAnchor.value = searchTagsRootGlobalY + searchTagsRootHeight + 10;
  listItemsWidth.value = searchTagsRoot.value.getBoundingClientRect().width;
  listItemsLeft.value = searchTagsRootLeft;
  listItemMaxHeight.value = trayRootHeight - searchTagsRootHeight - searchTagsRootGlobalY;
};

const trackHeightChange = (entries) => {
  // const entry = entries[0];
  // for (let entry of entries) {
  //       //console.log('Height changed:', entry.contentRect.height);
  //     }
  // //console.log('Height changed:', entry.contentRect.height);
  updateAnchor();
};

onMounted(() => {
  observer.value = new ResizeObserver(trackHeightChange);
  observer.value.observe(searchTagsRoot.value);
  // observer.value.observe(listBoxParent.value);
  document.addEventListener('click', handleClickOutside);

});

onBeforeUnmount(() => {
  if (observer.value) {
    observer.value.disconnect();
  }
  document.removeEventListener('click', handleClickOutside);

});

const handleInputFocus = (event) => {
  //displaySuggestions.value = true;
  isInputActive.value = true;
  emit('on-focus', event);
};

const handleInputBlur = (event) => {
  isInputActive.value = false;
  addNew(event);
  emit('on-blur', event);
};

const handleClickOutside = (event) => {
  if (!suggestionsButton.value.contains(event.target)) {
    if (!listBoxParent.value.contains(event.target)) {
      if (tagParent.value && !tagParent.value.contains(event.target)) {
        if (!event.target.classList.contains('chip')) {
          hideSuggestions();
        }
      };
    };
  };
};

const addNew = (event) => {
  handleKeyDown(event);
  const keyShouldAddChip = event ? props.addChipOnKeys.indexOf(event.keyCode) !== -1 : true;
  const typeIsNotBlur = event && event.type !== "blur";
  if (
    (!keyShouldAddChip && (typeIsNotBlur || !props.addChipOnBlur)) ||
    isLimit.value
  ) {
    return;
  }
  if (
    newChip.value &&
    (props.allowDuplicates || innerChips.value.indexOf(newChip.value) === -1) &&
    validateIfNeeded(newChip.value) && (props.chipLength === -1 || newChip.value.length <= props.chipLength)
  ) {
    innerChips.value.push(newChip.value);
    newChip.value = "";
    emit('update:modelValue', '');
    chipChange();
    event && event.preventDefault();
  } else {
    if (validateIfNeeded(newChip.value)) {
      if (newChip.value && (props.chipLength === -1 || newChip.value.length <= props.chipLength)) {
        makeItError(true);
      } else {
        makeItError('maxLength');
      }
    } else {
      makeItError(false);
    }
    event && event.preventDefault();
  }

  nextTick(() => {
    comboBoxRoot.value.scrollTop = comboBoxRoot.value.scrollHeight;
    selectedCategoryIndex.value = 0;
  });
};

const makeItError = (isDuplicatedOrMaxLength) => {
  props.inputChips.className = 'combo-new-chip combo-new-chip--error';
  // props.inputChips.style.textDecoration = "underline";
  //console.log(isDuplicatedOrMaxLength);
  emit('on-error', isDuplicatedOrMaxLength);
};

const validateIfNeeded = (chipValue) => {
  if (props.validate === "" || props.validate === undefined) {
    return true;
  }
  if (typeof props.validate === "function") {
    return props.validate(chipValue);
  }
  if (
    typeof props.validate === "string" &&
    Object.keys(validators).indexOf(props.validate) > -1
  ) {
    return validators[props.validate].test(chipValue);
  }
  if (
    typeof props.validate === "object" &&
    props.validate.test !== undefined
  ) {
    return props.validate.test(chipValue);
  }
  return true;
};

const removeLastChip = () => {
  if (newChip.value) {
    return;
  }
  innerChips.value.pop();
  chipChange();
};

const remove = (index) => {
  innerChips.value.splice(index, 1);
  chipChange();
  emit("on-remove", index);
};

const chipChange = () => {
  emit("on-chips-changed", innerChips.value);
};
</script>

<style scoped>
@import "@/assets/desktop.css";

.invisible-mask {
  position: absolute;
  z-index: 1;
  width: 100%;
  height: 100%;
  background-color: red;
  /* opacity: .2; */
  border-radius: 10px;

}

.tag-container {
  position: relative;
  box-sizing: border-box;

  width: 100%;
  /* height: 40px; */
  overflow: hidden;
  height: max-content;
  transition: all .2s ease-in-out;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  align-items: center;
  /* background-color: aquamarine; */
}

.tag-container-closed {
  /* display: none; */
  padding: 0px;
  /* background-color: coral; */
  height: 0px;
  gap: 0px;
  box-sizing: border-box;
  margin: 0px;

}

.autoComp-main {
  background-color: #f9fafb;
  background-color: cornflowerblue;
  /* min-width: 100vw;
    min-height: 100vh; */
  /* height: 100%;
    width: 200px;
    display: flex;
    justify-content: center;
    align-items: center; */
  flex: 1;
  min-width: 60px;
  /* height: 27px; */
  box-sizing: border-box;
  overflow: hidden;
  display: flex;
  justify-content: center;
  align-items: center;
}

.autoComp-container {
  position: relative;
  outline: none;
  padding: 2px;
  height: 100%;
  box-sizing: border-box;
  /* height: 27px; */
  /* overflow: visible; */

  /* background: transparent;
    border: 0;
    font-weight: 400;
    margin: 3px;
    outline: none;
    padding: 0 4px;
    flex: 1;
    min-width: 60px;
    height: 27px;
    background-color: aquamarine; */
  width: 100%;
  box-sizing: border-box;
  overflow: hidden;
  display: flex;
  justify-content: center;
  align-items: center;
}

.autoComp-input {
  box-sizing: border-box;
  /* padding: 1rem; */
  /* margin-bottom: 0.5rem; */
  width: 100%;
  height: 100%;
  border: 1px solid #e0e0e0;
  background-color: red;
}

.tag-parent {
  z-index: 1000;
  box-sizing: border-box;
  position: absolute;
  border-radius: 8px;
  min-height: 32px;
  line-height: 1.4 !important;
  background-color: #e7e7e7;
  background-color: #2a2a2a;
  overflow: hidden;
  overflow-y: auto;
  /* text-align: left; */
  display: flex;
  flex-wrap: wrap;
  /* flex-direction: column; */
  gap: .2rem;
  padding: .3rem .3rem;
  outline: 1px solid white;
  outline-offset: -1px;
  /* right: 50%;
  transform: translateX(50%); */
}

.list-tags {
  z-index: 1000;
  box-sizing: border-box;
  position: absolute;
  border-radius: 8px;
  min-height: 32px;
  line-height: 1.4 !important;
  background-color: #e7e7e7;
  background-color: #2a2a2a;
  overflow: hidden;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  flex-wrap: nowrap;
  gap: .2rem;
  padding: .3rem .3rem;
  outline: 1px solid white;
  outline-offset: -1px;
  right: 50%;
  transform: translateX(50%);
  /* background-color: red; */
}


.tag-parent::-webkit-scrollbar {

  border-radius: 8px;
  width: 8px;
}

.tag-parent::-webkit-scrollbar-thumb {
  border-radius: 8px;
  background-color: rgba(255, 255, 255, 0.295);
}

.tag-parent::-webkit-scrollbar-track {
  border-radius: 8px;
  background-color: rgba(0, 0, 0, 0.295);
}

.tag-item {
  /* transition: background-color 0.2s ease-in-out;
  padding: 0.3rem .4rem;
  border-radius: 8px;
  width: max-content;
  color: black;
  box-sizing: border-box; */

  background-color: rgb(65, 65, 65);
  color: white;
  outline: solid 1px #2e2e2e;
}

.tag-item:hover,
.tag-item-highlighted {

  background-color: red;
  background-color: rgb(110, 163, 241);
  background-color: #5f5f5f;

}



.combo-list-item {
  font-family: 'Inter', sans-serif;
  font-size: 18px;
  font-weight: 200;
  color: white;
  box-sizing: border-box;
  list-style: none;
  cursor: pointer;
  background-color: transparent;
  transition: background-color 0.2s ease-in-out;
  padding: 0.2rem .4rem;
  /* background-color: #858585; */
  border-radius: 8px;
  width: max-content;
  width: 100%;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  /* background-color: rgb(65, 129, 224); */
}

.combo-list-item:hover {
  background-color: #595959;
}

.combo-list-item-closed {
  opacity: .5;
}

.outer-chip-wrapper-parent {
  box-sizing: border-box;
  position: relative;
  /* border-radius: 8px; */
  min-height: 40px;
  /* line-height: 1.4 !important; */
  /* background-color: #fff; */
  /* overflow: unset; */
  overflow: hidden;
  overflow-x: hidden;
  /* text-align: left; */
  height: max-content;
  display: flex;
  /* flex-wrap: wrap; */
  flex-direction: row;
  gap: .2rem;
  border-right: 1px solid rgb(255, 255, 255);
  width: 100%;
  /* padding: .2rem ; */
  /* background: linear-gradient(to right, rgba(255, 0, 0, 0.5), rgba(255, 0, 0, 0) 90%); */
  /* box-shadow: inset -5px 0 5px -5px rgba(0, 0, 0, 0.3); */

}

/* .outer-chip-wrapper-parent::after {
  content: "";
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  width: 20px; 
  background: linear-gradient(to left, rgba(33, 33, 33, 0.5), rgba(255, 0, 0, 0) 90%);

} */

.outer-chip-wrapper {
  box-sizing: border-box;
  /* position: absolute; */
  /* border-radius: 8px; */
  min-height: 32px;
  line-height: 1.4 !important;
  /* background-color: #fff; */
  /* overflow: unset; */
  overflow-x: scroll;
  text-align: left;
  height: max-content;
  display: flex;
  /* flex-wrap: wrap; */
  flex-direction: row;
  gap: .2rem;
  padding: .2rem;
}

.outer-chip-wrapper::-webkit-scrollbar {
  display: none;
}

.outer-chip {
  background-color: #2e2e2e;
  color: rgb(205, 205, 205);
  outline: solid 1px white;
}

.outer-chip:hover {
  background-color: red;
  background-color: rgb(110, 163, 241);
  background-color: #ffffff17;
}

.arrow-buttons {
  position: relative;
  border-radius: 8px;
  box-sizing: border-box;
  cursor: pointer;
  display: flex;
  gap: 10px;
  align-items: center;
  padding: 5px 5px;
  width: 20px;
  overflow: hidden;
  align-items: center;
  justify-content: center;
}

.arrow-buttons-disabled {
  /* background-color: red; */
  opacity: .2;
  cursor: not-allowed;
}

.selected-display {
  font-size: 1.125rem;
  padding-top: 0.5rem;
  position: absolute;
  top: 300px;
  left: 0;
}

.search-tags {

  z-index: 1000000;
  height: 0px;
  display: flex;
  flex-direction: column;
  gap: .5rem;
  width: 100%;
  height: min-content;
  /* background-color: deeppink; */
  /* height: 100px; */
}

.search-tags-closed {
  gap: 0;
}

.combo {
  /* border-radius: 8px;
  min-height: 32px; */
  width: 100%;
  line-height: 1.4 !important;
  /* background-color: #fff; */
  overflow: hidden;
  /* overflow-y: auto; */
  /* max-height: 100px; */
  /* height: 180px; */
  /* cursor: text; */
  text-align: left;
  display: flex;
  gap: .3rem;
  /* flex-direction: column; */
  /* flex-wrap: wrap; */
  /* background-color: goldenrod; */
  height: min-content;
}

/* .combo::-webkit-scrollbar {
  border-radius: 8px;
  width: 8px;
}

.combo::-webkit-scrollbar-thumb {
  border-radius: 8px;
  background-color: rgba(255, 255, 255, 0.295);
}

.combo::-webkit-scrollbar-track {
  border-radius: 8px;
  background-color: rgba(0, 0, 0, 0.295);
} */

.combo--focus {
  /* outline: 0;
  border-color: #000000;
  box-shadow: 0 0 0 1px #000000; */
  outline: solid 1px white;
  outline-offset: -1px;
}

.combo--error {
  border-color: #f56c6c;
}



.search-container {
  border-radius: 8px;
  min-height: 32px;
  background-color: #f56c6c;
  background-color: rgba(0, 0, 0, 0.103);
  background-color: transparent;
  width: 100%;
  display: flex;
  flex-wrap: wrap;
  gap: .1rem;
  line-height: 1.4 !important;
  padding: .1rem;

  overflow: hidden;
  overflow-y: auto;
  max-height: 100px;
  /* height: 180px; */
  /* cursor: text; */
  text-align: left;
  display: flex;
  flex-wrap: wrap;
  /* color: white; */
  background-color: var(--transparent-black);

}

.search-container:hover {

  outline: solid 1px white;
  outline-offset: -1px;
}

.combo-linear {
  /* background-color: #f56c6c; */
  flex-wrap: nowrap;

}

.combo-content::-webkit-scrollbar {
  border-radius: 8px;
  width: 8px;
}

.combo-content::-webkit-scrollbar-thumb {
  border-radius: 8px;
  background-color: rgba(255, 255, 255, 0.295);
}

.combo-content::-webkit-scrollbar-track {
  border-radius: 8px;
  background-color: rgba(0, 0, 0, 0.295);
}

.combo-actions {
  display: flex;
  flex-direction: row;
  gap: .3rem;
  align-items: center;
}

.inner-chip {
  background-color: rgb(65, 129, 224);
  outline: solid 1px white;
}

.combo-remove-chip {
  color: #1b1b1b;
  transition: opacity 0.3s ease;
  opacity: 0.5;
  cursor: pointer;
  /* padding: 0 5px 0 7px; */
}

.combo-remove-chip::before {
  /* content: "x"; */
}

.combo-remove-chip:hover {
  opacity: 1;
}

.combo .combo-new-chip {


  box-sizing: border-box;
  font-family: Inter, sans-serif;
  font-size: 16px;
  white-space: nowrap;
  color: rgba(255, 255, 255, 0.8);
  background: transparent;
  border: 0;
  font-weight: 400;
  margin: 3px;
  outline: none;
  padding: 0 4px;
  flex: 1;
  min-width: 60px;
  height: 27px;
  /* background-color: crimson; */
}


.combo .combo-new-chip--error {
  color: #f56c6c;
}

/* Apply transition to moving elements */
.list-move,
.list-enter-active,
.list-leave-active {
  transition: all 0.2s ease;
}

/* Don't animate input element to avoid jittery movement when user starts typing just after creating a chip. */
#input {
  transition: none;
}

.list-enter-from,
.list-leave-to {
  opacity: 0;
  transform: translateX(-10px);
}

/* Ensure leaving items are taken out of layout flow so that moving animations can be calculated correctly. */
.list-leave-active {
  position: absolute;
  left: var(--x);
  top: var(--y);
}
</style>


