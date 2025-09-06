<template>
<div class="data-container" >
  <div v-if="!enterEditData" class="data-item">
    <label class="data-name">{{ newData }}</label>
    <div class="data-item-actions">
      <span v-if="originalData !== newData" class="single-action-button" @click="revertData()"><img class="small-icons" src="/icons/revert.svg"></span>
      <span v-if="isEditable" class="single-action-button" @click="enterEditMode()"><img class="small-icons" src="/icons/edit.svg"></span>
    </div>
  </div>

  <div v-else class="data-item">
    <input ref="dataNameInput" spellcheck="false" v-model="mutableData" class="data-name-input" type="text" 
    :placeholder="mutableData" v-focus v-return="handleEnterKey" v-esc="cancelChange"/>
    <div class="data-item-actions">
      <span class="single-action-button" @click="confirmChange()"><img class="small-icons" src="/icons/confirm.svg"></span>
      <span class="single-action-button" @click="cancelChange()"><img class="small-icons" src="/icons/cancel.svg"></span>
    </div>
  </div>
</div>
</template>

<script setup>
import { ref, computed, onMounted, watch, nextTick, onBeforeUnmount } from 'vue';
import { useTrayStates } from '@/stores/TrayStates';


const trayStates = useTrayStates();


const emit = defineEmits(['mutatedData', 'queryTagRemoved', 'isInputActive']);
const props = defineProps({
  tags: { type: Array, default: () => [] },
  originalData: { type: String, default: '' },
  newData: { type: String, default: '' },
  title: { type: String, default: '' },
  isEditable: { type: Boolean, default: false}
});

const handleEnterKey = (event) => {
    confirmChange();
};

const dataNameInput = ref();
const mutableData = ref('');
const enterEditData = ref(false);

const confirmChange = () => {
  emit('mutatedData', mutableData.value);
  enterEditMode();
};

const revertData = () => {
  emit('mutatedData', props.originalData)
  mutableData.value = props.originalData;
};

const enterEditMode = () => {
  if(mutableData.value === ''){
    mutableData.value = props.originalData;
  }
  mutableData.value = props.newData;
  enterEditData.value = !enterEditData.value;
  if(enterEditData.value){
    emit('isInputActive', true);
  } else{
    emit('isInputActive', false);
  }
};

const cancelChange = () =>{
  enterEditData.value = false;
  // console.log('canceled');
}



const toggleSearch = () => {
  emit('toggleSearch')
}

onMounted(() => {
  mutableData.value = props.originalData;
  // console.log(props.originalData)
});
</script>

<style scoped>

/* @import "@/assets/tray.css"; */

.data-container{
  z-index: 20;
  display: flex;
 /* background-color: red; */
 width: 100%;
 overflow: hidden;

}
.data-name{
  /* background-color: goldenrod; */
  width: 100%;
  /* display: flex; */
  /* flex: 1; */
}
.data-item {
  color: white;
  /* margin-top: 1rem; */
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .5rem;
  width: 100%;
  border-bottom: 1px solid rgba(255, 255, 255, 0.096);
  height: max-content;
  height: 35px;
  padding: .2rem;

  /* background-color: greenyellow; */
}

.data-item-actions {

  /* background-color: red; */
  display: flex;
  flex: 1;

}

.data-name-input{
  z-index: 26;
  /* z-index: initial; */
  font-family: 'Inter', sans-serif;
  font-style: italic;
  box-sizing: border-box;
  border-radius: 8px;
  font-size: 16px;
  outline: none;
  /* box-shadow: 0px 1px 2px 0px rgba(16, 24, 40, 0.05000000074505806); */
  padding: 10px;
  height: 100%;
  /* border-color: rgba(234, 236, 240, 1); */
  border-width: 0px;
  /* flex: 1; */
  width: 100%;
  /* border-style: solid; */
  /* background-color: rgba(255, 255, 255, 1); */
  background-color: transparent;
  /* background-color: rgba(255, 255, 255, 0.094); */
  background-color: #3c3c3c;

  color: rgb(255, 255, 255);
  outline: none;
  outline: solid 1px white;
  outline-offset: -1px;

}
</style>
