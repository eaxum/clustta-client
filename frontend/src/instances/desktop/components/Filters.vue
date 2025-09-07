<template>
    <div class="filter-container" @wheel="handleWheel">
      <div ref="scrollableElement" class="filter-scroller">
            <div v-for="(item, index) in items" :key="index" :index="index" @click="selectItem(item)" 
                class="filter-item" :style="{ backgroundColor: item.backgroundColor, color: item.textColor}">
                <div v-if="useAvatar" class="profile-picture" :style="{ backgroundColor: item.avatarColor}">
                    <img class="profile-img"  :src=" item.profile ? item.profile : '/icons/default_profile_picture.svg'">
                </div>
                <div v-else-if="useIcons && item.icon" class="task-item-icon-container">
                    <img class="small-icons" :src="item.icon">
                </div>

                <div class="task-item-content" >
                    <div  class="task-item-details">
                        {{ item.name.toUpperCase().replace(/_/g, " ") }}
                    </div>
                </div>
                <!-- <div  class="task-item-actions">
                    <span v-if="editItems" class="filter-action-button">
                        <img class="small-icons" src="/icons/close.svg">
                    </span>
                </div> -->
            </div>
        </div>
    </div>
</template>
  
<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue';
const scrollableElement = ref(null);

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'

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
    backgroundColor: {
        type:String, 
        default: 'red'
    },  
    useIcons: {
        type:Boolean, 
        default: false
    }, 
    editItems: {
        type:Boolean, 
        default: false
    }, 
    selectItem: {
        type:Function, 
        default: (item) => {
          console.log(item);
        }
    },     
});

const handleWheel = (event) => {
  const element = scrollableElement.value;
  const delta = Math.sign(event.deltaY);
  element.scrollLeft += delta * 50; // Adjust scroll speed as needed
  event.preventDefault(); // Prevent the default vertical scrolling behavior
}
 
</script>
  
<style scoped>

@import "@/assets/desktop.css";

/* action button */
.filter-action-button{
  overflow: hidden;
  background-color: transparent;
  text-align: center;
  font-size: 14px;
  line-height: 14px;
  background-color: transparent;
  color: white;
  position: relative;
  /* border-radius: 8px; */
  border-radius: 20px;
  box-sizing: border-box;
  cursor: pointer;
  display: flex;
  /* flex: 1; */
  gap: 10px;
  align-items: center;
  /* padding: .3rem; */
  height: max-content;
  width: max-content;
  min-width: max-content;
  min-height: max-content;
  /* aspect-ratio: 1/1; */
  transition: all 0.3s ease;
  /* background-color: red; */

}

.filter-action-button:hover{
  background-color: rgb(121, 121, 121);
  background-color: #ffffff15;
}
.filter-action-button:active{
  background-color: rgb(70, 70, 70);
  background-color: #00000013;
}

.filter-action-button-pressed {
  box-sizing: border-box;
  background-color: rgba(0, 0, 0, 0.216);
  outline: solid 1px white;
  outline-offset: -1px;
}

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

.filter-item:hover>*:last-child {
  opacity: 1;
  transition: opacity 0.2s ease-in-out;
}

.filter-root{
	/* z-index: 5; */
	position: relative;
	display: flex;
  flex-direction: column;
	padding-right: .4rem;
	overflow: hidden;
	overflow-y: scroll;
	height: 100%;
  /* height: min-content; */
	width: 100%;
	box-sizing: border-box;
  align-items: flex-start;
	/* max-width: 600px; */
	min-width: 300px;
  /* background-color: red; */
}
.filter-root-single{
  height: min-content;
}

.filter-root::-webkit-scrollbar {
  width: 8px;
}

.filter-root::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: var(--dark-steel);
}

.filter-root::-webkit-scrollbar-track {
  border-radius: 10px;
}

.filter-item-meta {
    color: rgb(219, 219, 219);
    color: white;
    background-color: rgba(0, 0, 0, 0.216);
    padding: .3rem;
    border-radius: 5px;
    font-size: 12px;
}

.filter-container {
  box-sizing: border-box;
  display: flex;
  height: min-content;
  height: 50px;
  flex-direction: row;
  /* padding: .5rem; */
  /* gap: .3rem; */
  overflow: hidden;
  align-items: center;
  /* overflow-x: scroll; */
}

.filter-scroller {
  box-sizing: border-box;
  display: flex;
  height: min-content;
  height: 50px;
  flex-direction: row;
  padding: .5rem;
  gap: .3rem;
  overflow: hidden;
  align-items: center;
  /* background-color: green; */
}

.filter-container::-webkit-scrollbar {
  height: 8px;
}

.filter-container::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: white;
  /* background-color: rgb(236, 0, 0); */
}

.filter-container::-webkit-scrollbar-track {
  border-radius: 10px;
}


.filter-item{
  display: flex;
  /* gap: .2rem; */
  color: white;
  align-items: center;
  padding-left: .3rem;
  padding: .3rem .5rem ;
  box-sizing: border-box;
  width: 100%;
  width: min-content;
  background-color: var(--light-steel);
  border-radius: var(--large-radius);
  /* overflow: hidden; */
  height: min-content;
  cursor: pointer;
  
}
.filter-item-wrap{
  width: 100%;
}


.filter-item:hover{
  outline: var(--transparent-line);
  outline-offset: -1px;
}

.task-item-container{
  display: flex;
  gap: .5rem;
  color: white;
  align-items: center;
  padding: .4rem;
  box-sizing: border-box;
  width: 100%;
  height: 50px;
  justify-content: space-between;
}

.task-item-container-cards{
  height: 100%;
  flex-direction: column;
}

.task-item-container-selected{
  outline: 1px solid rgb(255, 255, 255);
  outline-offset: -1px;
}

.task-item-container-selected:hover{
  outline: 1px solid rgb(255, 255, 255);
  outline-offset: -1px;
}
.selected-item{
  outline: 1px solid white;
  outline-offset: -1px;
}

.selected-item:hover{
  outline: 1px solid white;
  outline-offset: -1px;
}
.task-spacer{
  display: flex;
  align-items: center;
  justify-content: center;
  max-width: 40px;
  min-width: 40px;
  aspect-ratio: 1/1;
  height: 100%;
  overflow: hidden;
  /* background-color: purple; */
  /* flex: 1; */
}

.task-spacer-empty{
  background-color: moccasin;
}

.checkboxes{
  width: 18px;
  height: 18px;
  border-radius: 4px;
  border: 2px solid yellow;
  background: #FFF;
  padding: 10px;
}

.action-column {
  text-align: center;
  padding: 2px;
  box-sizing: border-box;
  transition: all .3s ease-in;
}

.checkbox-container{
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  padding: .1rem;
  overflow: hidden;
  min-width: min-content;
  height: 100%;
  /* background-color: royalblue; */
}


.task-item-icon-container{
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  min-width: min-content;
  padding: .1rem;
  overflow: hidden;
  height: 100%;
  /* background-color: firebrick; */
}

.task-item-content{
  gap: .4rem;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  /* background-color: indigo; */
  width: 100%;
  overflow: hidden;
}

.task-item-content-cards{
  height: max-content;
}

.task-item-details{
    /* display: flex; */
    padding: .2rem;
    flex-wrap: nowrap;
    overflow: hidden;
    box-sizing: border-box;
    align-items: center;
    justify-content: space-between;
    /* width: 50%; */
    height: 100%;
    height: min-content;
    white-space: nowrap;
    text-overflow: ellipsis;
    font-size: 12px;
    font-weight: 700;
    /* color: black; */
    /* background-color: forestgreen; */
}

.task-item-meta{
  display: flex;
  padding: .2rem;
  box-sizing: border-box;
  align-items: center;
  gap: .4rem;
  width: 100%;
  height: 100%;
  overflow: hidden;
  /* background-color: rosybrown; */
}
.task-item-tag{
  display: flex;
  box-sizing: border-box;
  overflow: hidden;
  padding: .1rem .4rem;
  font-size: 12px;
  background-color: black;
  border-radius: 20px;
}


.task-item-status-container{
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  padding: .2rem;
  height: 100%;
  /* background-color: darkorange; */
  /* flex: 1; */
}

.task-item-container-footer{
  /* background-color: royalblue; */
  align-items: center;
  display: flex;
  width: min-content;
  box-sizing: border-box;
  padding: .2rem;
  /* opacity: 0; */
}

.task-item-container-footer-cards{
  width: 100%;
  justify-content: space-between;

}
.task-item-status{
  display: flex;
  border-radius: var(--normal-radius);
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  width: 80px;
  padding: .4rem .4rem;
  height: max-content;
  /* background-color: firebrick; */
  font-size: 12px;
  text-transform: uppercase;
  font-weight: 700;
  color: black;
}

.task-item-actions{
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: max-content;
  min-width: max-content;
  padding: .2rem;
  /* opacity: 0; */
  /* gap: .7rem; */
  /* height: 100%; */
  /* background-color: darkcyan; */
  /* flex: 1; */
}

.task-item-assignee{
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  width: min-content;
  min-width: max-content;
  gap: .7rem;
  height: 100%;
  /* background-color: darkcyan; */
  /* flex: 1; */
}
</style>
  
  
  

