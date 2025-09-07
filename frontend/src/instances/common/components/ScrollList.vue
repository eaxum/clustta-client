<template>
  <div  class="scroll-list-container">
      <div v-for="(item, index) in items" :key="index" :index="index" class="scroll-list-item">
      <div v-if="useAvatar" class="profile-picture" :style="{ backgroundColor: item.avatarColor}">
        <img class="profile-img"  :src=" item.profile ? item.profile : '/icons/default_profile_picture.svg'">
      </div>
      <img v-else-if="useIcons" class="small-icons" :src="item.icon">
      <div class="scroll-list-item-name"> {{ item.name }}</div>
      <div v-if="item.meta" class="scroll-list-item-meta"> {{ item.meta }}</div>

      <div v-if="useItemId" class="scroll-list-item-actions">
          <span v-if="editItems" v-stop-propagation class="single-action-button" @click="editListItem(item.id)" v-tooltip="'Edit'">
            <img class="small-icons" src="/icons/edit.svg">
          </span>
          <span v-if="deleteItems" v-stop-propagation class="single-action-button" :class="{'item-inactive' : buttonInactive(item.id)}" @click="deleteListItem(item.id)"  v-tooltip="'Delete'">
            <img class="small-icons" src="/icons/delete.svg">
          </span>
      </div>
      <div v-else class="scroll-list-item-actions">
          <span v-if="editItems" v-stop-propagation class="single-action-button" @click="editListItem(index)" v-tooltip="'Edit'">
            <img class="small-icons" src="/icons/edit.svg">
          </span>
          <span v-if="deleteItems" v-stop-propagation class="single-action-button" :class="{'item-inactive' : buttonInactive(index)}" @click="deleteListItem(index)"  v-tooltip="'Delete'">
            <img class="small-icons" src="/icons/delete.svg">
          </span>
      </div>

      </div>
  </div>
</template>


<script setup>
const props = defineProps({
    items: Array,
    useAvatar: {
        type:Boolean, 
        default: false
    }, 
    avatarColor: {
        type:String, 
        default: '#7FFF00'
    }, 
    useIcons: {
        type:Boolean, 
        default: false
    }, 
    useItemId: {
        type:Boolean, 
        default: false
    }, 
    wrapItems: {
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
        type:Function, 
        default: () => {
          return false
        }
    },  
    editListItem: Function,   
    deleteListItem: Function,   
});
</script>

<style scoped>

.profile-picture{
    background-color: red;
    height: 24px;
    min-width: 24px;
    overflow: hidden;
    display: flex;
    align-items: center;
    border-radius: 24px;
    /* padding: 5px; */
}
.profile-img{
    width: 100%;
    height: 100%;
}
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
    color: white;
    background-color: rgba(0, 0, 0, 0.216);
    padding: .3rem;
    border-radius: 5px;
    font-size: 12px;
}
.scroll-list-container {
  box-sizing: border-box;
  padding: .5rem;
  align-items: center;
  flex-direction: column;
  gap: .2rem;
  background-color: rgba(0, 0, 0, 0.144);
  /* flex: 1; */
  /* min-height: max-content; */
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
  background-color: white;
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
.scroll-list-item:hover>*:last-child {
  opacity: 1;
  visibility: visible;
  transition: opacity 0.2s ease-in-out;
  display: flex;
}

.scroll-list-item-actions {
  display: none;
  opacity: 0;
  visibility: hidden;
}
</style>

