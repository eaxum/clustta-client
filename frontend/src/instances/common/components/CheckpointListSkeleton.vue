<template>
<div class="checkpoint-list-container" >

    <div class="checkpoint-item-skeleton" v-for="(skeleton, index) in skeletonArray" :key="index"  :style="{ animationDelay : `${(skeleton - 1) * 0.2}s` }">
        <div class="checkpoint-item-thumb-skeleton"  >
            <!-- <span ><img class="thumbs-skeleton"></span> -->
        </div>
        <div class="checkpoint-item-content-skeleton">
            <div class="checkpoint-item-meta-skeleton" >
            </div>
            <div class="checkpoint-item-meta-skeleton"> 
            </div>
            <div class="checkpoint-item-meta-skeleton"> 
            </div>
        </div>
        <span class="icon-skeleton" ></span>
    </div>
</div>
</template>

<script setup>

import { ref, onMounted  } from 'vue';

import { useTrayStates } from '@/stores/TrayStates';

const trayStates = useTrayStates();
const skeletonArray = ref([1,2,3,4,5]);

onMounted(async () => {

});

</script>
  



<style scoped>
  @import "@/assets/tray.css";


.checkpoint-item-skeleton{
box-sizing: border-box;
display: flex;
flex-direction: row;
align-items: center;
justify-content: center;
width: 100%;
padding:5px;
border-bottom: solid 1px rgba(0, 0, 0, 0.295);
border-bottom: 1px solid rgba(255, 255, 255, 0.096);
flex-wrap: wrap;
opacity: 0;
animation: fadeInFadeOut infinite  4s ease-in-out;

}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.checkpoint-item-thumb-skeleton{
box-sizing: border-box;
position: relative;
background-color: white;
opacity: .1;
width: 96px;
height: 54px;
overflow: hidden;
transition: all 0.2s ease-in;
border-radius: 8px;
overflow: hidden;
}



.thumbs-skeleton{
    box-sizing: border-box;
    /* padding: 2px; */
    height: 100%;
    /* width: 20px; */
    overflow: hidden;
    display: flex;
    align-items: center;
  }


.checkpoint-item-content-skeleton{
box-sizing: border-box;
height: 100%;
overflow: hidden;
display: flex;
flex-direction: column;
overflow: hidden;
align-items: center;
padding: .2rem .2rem;
padding-left: .4rem;
flex: 1;
height: max-content;
gap: .2rem;
}


.checkpoint-item-meta-skeleton{
box-sizing: border-box;
border-radius: 5px;
overflow: hidden;
display: flex;
height: 100%;
width: 100%;
white-space: nowrap;
background-color: white;
opacity: .1;
align-items: center;
height: 18px;
text-overflow: ellipsis;
}

.checkpoint-item-meta-active{

/* background-color: darkorange; */
width: 100%;
display: flex;
white-space: pre-wrap;
height: min-content;
margin-bottom: 10px;
border-bottom: 1px solid rgba(255, 255, 255, 0.096);
transition: all 0.2s ease-in;

}

.checkpoint-item-label-text{
font-family: 'Inter', sans-serif;
font-size: 14px;
color: white;
/* background-color: rebeccapurple; */
max-width: 100%;

}

.large{

    font-size: 14px;

}

.checkpoint-item-actions{
display: flex;
flex-direction: row;
max-width: 100%;
align-self: stretch;
align-items: center;
justify-content: space-evenly;
gap: 5%;
}




.icon-skeleton{
padding: 2px;
height: 20px;
width: 20px;
overflow: hidden;
/* display: flex; */
align-items: center;
background-color: white;
opacity: .1;
border-radius: 10px;
}

.checkpoint-list-container{

  width: 98%;
  height: 100%;
  /* background-color: red; */
  display: flex;
  flex-direction: column;
  gap: 1.5px;
  /* overflow-y: scroll; */
  padding: 0% 1%;
  /* margin: 3% 0%; */
}


@keyframes fadeInFadeOut {
  from {
    opacity: 0;
  }
  20% {
    opacity: 1;
  }
  70% { 
    opacity: 1;
  }
  80% { 
    opacity: 0;
  }
  to {  
    opacity: 0;
  }
}
  </style>
  
  
  

