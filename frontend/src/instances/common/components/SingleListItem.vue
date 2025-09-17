<template>
  <div  class="scroll-list-container">
      <div class="scroll-list-item">
      <img v-if="useIcons" class="small-icons" :src="item.icon">
      <div class="scroll-list-item-name"> {{ item.name }}</div>
      <div v-if="useMeta" class="scroll-list-item-meta"> {{ item.meta }}</div>
      <div class="scroll-list-item-actions">
          <span v-if="editItems" class="single-action-button" @click="editListItem(index)" v-tooltip="'Edit'">
            <img class="small-icons" src="/icons/edit.svg">
          </span>
          <span v-if="deleteItems" class="single-action-button" :class="{'item-inactive' : buttonInactive}" @click="deleteListItem(index)"  v-tooltip="'Delete'">
            <img class="small-icons" src="/icons/delete.svg">
          </span>
          
      </div>
      </div>
  </div>
</template>


<script setup>
const props = defineProps({
    item: Object,
    useIcons: {
        type:Boolean, 
        default: false
    }, 
    useMeta: {
        type:Boolean, 
        default: false
    }, 
    editItems: {
        type:Boolean, 
        default: false
    }, 
    deleteItems: {
        type:Boolean, 
        default: false
    },  
    buttonInactive: {
        type:Boolean, 
        default: false
    },  
    editListItem: Function,   
    deleteListItem: Function,   
});
</script>

<style scoped>

/* @import "@/assets/tray.css"; */


.scroll-list-item-name{
    font-family: 'Inter', sans-serif;
    font-weight: 200;
    color: white;
    font-size: 16px;
    display: flex;
    flex: 1;
    height: 100%;
    align-items: center;
    justify-content: flex-start;
}

.scroll-list-item-meta {
    color: rgb(219, 219, 219);
    background-color: rgba(0, 0, 0, 0.216);
    padding: .3rem;
    border-radius: 5px;
    font-size: 12px;
}
.scroll-list-container {
  box-sizing: border-box;
  padding: 0px .5rem;
  align-items: center;
  flex-direction: column;
  gap: .2rem;
  background-color: rgba(0, 0, 0, 0.144);
  /* flex: 1; */
  min-height: max-content;
  overflow: hidden;
  overflow-y: scroll;
  width: 100%;
  border-radius: 10px;
}

.scroll-list-container::-webkit-scrollbar {
  width: 4px;
}

.scroll-list-container::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: rgba(255, 255, 255, 0.295);
  /* background-color: rgb(236, 0, 0); */
}

.scroll-list-container::-webkit-scrollbar-track {
  border-radius: 10px;
}

.scroll-list-item {
  color: white;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .5rem;
  width: 100%;
  border-bottom: 1px solid rgba(255, 255, 255, 0.096);
  height: 40px;
  overflow: hidden;
  padding: .2rem;
}
.scroll-list-item:last-child{
  border-bottom: 0px;
  /* background-color: firebrick; */

}
.scroll-list-item-actions {
  display: flex;
}
</style>

