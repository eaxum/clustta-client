<template>
  <div class="progress-bar-container">
    <div  v-if="taskProgress === 0" class="loaderBar"></div>
    <div v-if="taskProgress === 0" class="queued" :style="{ width: '100%' }"></div>
    <div v-else class="progress-bar" :style="{ width: taskProgress + '%' }"></div>
  </div>
</template>

<script setup>

const props = defineProps({

  taskProgress: {
    type: Number,
    required: true,
  }

})

</script>

<style scoped>
  @import "@/assets/tray.css";

  

.loaderBar { 
  position:absolute;
  top:0;
  right:100%;
  bottom:0;
  left:0;
  background: crimson; 
  width:0;
  animation:borealisBar 2s linear infinite;
  z-index: 1;
  background-color: rgb(255, 229, 201);
}

@keyframes borealisBar {
  0% {
    left:0%;
    right:100%;
    width:0%;
  }
  2.5% {
    left:0%;
    right:75%;
    width:40%;
  }
  22.5% {
    right:0%;
    left:75%;
    width:40%;
  }
  25% {
    left:100%;
    right:0%;
    width:0%;
  }
}

.progress-bar-container {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgb(24, 24, 24);
  /* opacity: .2; */
  cursor: not-allowed !important;
  border-radius: 10px;
  overflow: hidden;
}

.progress-bar {
  position: relative;
  height: 100%;
  background-color: rgb(67, 210, 67);
  opacity: 1;
  border-radius: 10px;
}

.queued {
  position: relative;
  height: 100%;
  background-color: rgb(205, 210, 67);
  opacity: 1;
  border-radius: 10px;
}

.progress-bar::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  height: 100%;
  width: 100%;
  background-color: rgba(255, 255, 255, 0.5);
  border-radius: 8px;
  opacity: 0;
  transition: opacity 0.5s ease-in-out;
  pointer-events: none;
}

.progress-bar.active::before {
  opacity: 1;
  pointer-events: none;
}
</style>


