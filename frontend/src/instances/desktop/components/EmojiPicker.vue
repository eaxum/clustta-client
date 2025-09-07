<template>
  <div class="apps-container-full" v-stop-propagation>
    

  <div class="category-buttons">
    <button  v-for="(category, key) in emojiData"  :key="key"
      @click="selectCategory(key)" :class="{ active: selectedCategory === key }"
      class="category-button" >
      <span class="emoji">{{ category.icon }}</span>
    </button>
  </div>
  
  <div class="menu-divider"></div>
    <div class="apps-container">

      <div ref="scrollableElement" class="apps-grid">
          <div v-for="emoji in filteredEmojis" @click="selectEmoji(emoji)" class="apps-grid-item" :class="{ 'apps-grid-item-selected' : selectedEmoji === emoji}" >
            <span class="emoji">{{ emoji }}</span>
          </div>
      </div>

    </div>
  </div>



</template>

<script setup>
import { ref, computed } from 'vue';
import emojiData from "@/data/emojiData.json";

const selectedCategory = ref(Object.keys(emojiData)[0]);
const selectedEmoji = ref('');
const selectedUnicode = ref('');
const selectedEmojiEntity = ref('');
const searchTerm = ref('');

const selectCategory = (category) => {
selectedCategory.value = category;
};

const getEmojiUnicode = (emoji) => {
return [...emoji].map(char => 
'U+' + char.codePointAt(0).toString(16).toUpperCase().padStart(4, '0')
).join(' ');
};

const unicodeToNumericEntity = (unicode) =>  {
// Remove the 'U+' prefix and convert the hexadecimal to decimal
const hex = unicode.replace('U+', '');
const decimal = parseInt(hex, 16);

// Return the numeric entity
return `&#${decimal}`;
}

const selectEmoji = (emoji) => {
selectedEmoji.value = emoji;
selectedUnicode.value = getEmojiUnicode(emoji);
selectedEmojiEntity.value = unicodeToNumericEntity(selectedUnicode.value);
emit('select', {
emoji,
unicode: selectedUnicode.value,
entity: unicodeToNumericEntity(selectedUnicode.value)
});
};

const filteredEmojis = computed(() => {
if (searchTerm.value) {
const allEmojis = Object.values(emojiData)
.flatMap(category => category.emojis);
return allEmojis.filter(emoji => 
emoji.toLowerCase().includes(searchTerm.value.toLowerCase())
);
}

return emojiData[selectedCategory.value].emojis;
});

const emit = defineEmits(['select']);
</script>

<style scoped>

@import "@/assets/desktop.css";

.emoji-container {
max-width: 400px;
margin: 0 auto;
padding: 1rem;
border: 1px solid #ccc;
border-radius: 8px;
}

.search-input {
width: 100%;
padding: 8px;
margin-bottom: 1rem;
border: 1px solid #ddd;
border-radius: 4px;
font-size: 1rem;
box-sizing: border-box;
}

.category-buttons {
  justify-content: space-between;
display: flex;
gap: 0.5rem;
/* margin-bottom: 1rem; */
overflow: hidden;
overflow-x: auto;
padding-bottom: 0.5rem;
min-height: 30px;
/* height: 100px; */
width: 100%;
/* flex : 1; */
/* background-color: crimson; */
}

.category-buttons::-webkit-scrollbar {
  height: 4px;
}

.category-buttons::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: rgba(255, 255, 255, 0.295);
}

.category-buttons::-webkit-scrollbar-track {
  border-radius: 10px;
  /* background-color: rgba(0, 0, 0, 0.295); */
}

.category-button {
/* padding: 0.5rem; */
/* border: 1px solid #ddd; */
border-radius: 4px;
background: none;
cursor: pointer;
transition: background-color 0.2s;
height: 100px;
width: 35px;
height: 35px;
display: flex;
overflow: hidden;
align-items: center;
justify-content: center;
box-sizing: border-box;
padding: .2rem;


  border-radius: 16px;
  border-width: 0px;
  cursor: pointer;

}

.category-button.active {
background-color: #e6e6e6;
/* background-color: var(--); */
}

.emoji-grid {
display: grid;
grid-template-columns: repeat(auto-fill, minmax(40px, 1fr));
gap: 0.5rem;
max-height: 300px;
overflow-y: auto;
padding-right: 0.5rem;
}

.emoji-button {
padding: 0.5rem;
border: none;
background: none;
cursor: pointer;
border-radius: 4px;
transition: background-color 0.2s;
}

.emoji-button:hover {
background-color: #f0f0f0;
}

.emoji {
  font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI Emoji", 
  "Segoe UI", "Apple Color Emoji", sans-serif;
  font-size: 2rem;
}

.selected-emoji {
padding: 0.5rem;
border-top: 1px solid #ddd;
margin-top: 1rem;
}

.unicode {
font-family: monospace;
font-size: 0.9rem;
color: #666;
margin-top: 0.5rem;
}

/* ******************************************************* */
.icon-picker-icon{
  box-sizing: border-box;
  padding: 2px;
  height: 32px;
  width: 32px;
  overflow: hidden;
  display: flex;
  align-items: center;
}

.apps-container-full {
  box-sizing: border-box;
  flex-direction: column;
  gap: .5rem;
  width: 100%;
  display: flex;
  position: relative;
  display: flex;
  justify-content: space-between;
  align-items: center;
  /* max-width: 300px; */
  max-height: 260px;
  background-color: crimson;
  background-color: var(--black-steel);
  border-radius: var(--large-radius);
  padding: .5rem;
  outline: var(--transparent-line);
  outline-offset: -1px;
  overflow: hidden;
}

.menu-divider{
  color: red;
}
.apps-container {
  box-sizing: border-box;
  overflow: hidden;
  /* background-color: chartreuse; */
  width: 100%;
  padding: .2rem;
  flex: 1;
  display: flex;
  position: relative;
  display: flex;
  justify-content: space-between;
  align-items: center;
  align-items: baseline;
  overflow: hidden;
  overflow-y: scroll;
  /* max-height: 400px; */
}

.apps-container::-webkit-scrollbar {
  width: 4px;
}

.apps-container::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: rgba(255, 255, 255, 0.295);
}

.apps-container::-webkit-scrollbar-track {
  border-radius: 10px;
  /* background-color: rgba(0, 0, 0, 0.295); */
}

.apps-grid {
  /* background-color: deeppink; */
  width: 100%;
  display: grid;
  gap: 10px;
  grid-template-columns: auto auto auto auto auto auto auto;
	grid-template-columns: repeat(auto-fill, minmax(35px, 1fr));
  box-sizing: border-box;
}


.apps-grid-item-container {
  box-sizing: border-box;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  background-color: rebeccapurple;
  overflow: hidden;
  /* gap: 10px; */
  /* padding: .1rem; */
}


.apps-grid-item{
  box-sizing: border-box;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  /* background-color: rebeccapurple; */
  overflow: hidden;
  border-radius: 12px;
  /* padding: .2rem; */
  /* background-color: crimson; */
  
}
.apps-grid-item:hover {
  background-color: rgb(121, 121, 121);
  background-color: #ffffff15;
  outline: var(--transparent-line);
  outline-offset: -1px;
}

.apps-grid-item:active {
  background-color: rgb(70, 70, 70);
  background-color: #00000013;
}

.apps-grid-item-selected {
  box-sizing: border-box;
  background-color: rgba(0, 0, 0, 0.216);
  outline: solid 1px white;
  outline-offset: -1px;

}

.apps-grid-item-selected:hover {
  background-color: rgba(0, 0, 0, 0.216);

}
</style>

