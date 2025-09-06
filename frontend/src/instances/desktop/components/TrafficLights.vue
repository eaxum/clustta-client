<template>
  <div 
    class="titlebar"
    :class="{ 'titlebar--draggable': draggable }"
    @dblclick="handleDoubleClick"
  >
    <div class="titlebar__controls">
      <button 
        class="titlebar__button titlebar__button--close"
        @click="handleClose"
        :title="closeTitle"
      >
        <!-- <svg width="10" height="10" viewBox="0 0 10 10">
          <path d="M1 1l8 8M9 1l-8 8" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
        </svg> -->
      </button>
      
      <button 
        class="titlebar__button titlebar__button--minimize"
        @click="handleMinimize"
        :title="minimizeTitle"
      >
        <!-- <svg width="10" height="10" viewBox="0 0 10 10">
          <path d="M1 5h8" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
        </svg> -->
        
      </button>
      
      <button 
        class="titlebar__button titlebar__button--fullscreen"
        @click="handleFullscreen"
        :title="fullscreenTitle"
        @click.alt="handleMaximize"
      >
        <!-- <svg width="10" height="10" viewBox="0 0 10 10">
          <path d="M1 3v6h6M7 1v6M7 1H1" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" fill="none"/>
        </svg> -->

      </button>
    </div>
    
  </div>
</template>

<script setup>
import { ref } from 'vue'

// Props
const props = defineProps({
  title: {
    type: String,
    default: ''
  },
  draggable: {
    type: Boolean,
    default: true
  },
  style: {
    type: Boolean,
    default: true
  },
  closeTitle: {
    type: String,
    default: 'Close'
  },
  minimizeTitle: {
    type: String,
    default: 'Minimize'
  },
  fullscreenTitle: {
    type: String,
    default: 'Fullscreen'
  }
})

// Emits
const emit = defineEmits(['close', 'minimize', 'fullscreen', 'maximize'])

// Event handlers
const handleClose = (event) => {
  emit('close', event)
}

const handleMinimize = (event) => {
  emit('minimize', event)
}

const handleFullscreen = (event) => {
  emit('fullscreen', event)
}

const handleMaximize = (event) => {
  emit('maximize', event)
}

const handleDoubleClick = (event) => {
  emit('maximize', event)
}
</script>

<style scoped>
.titlebar {
  display: flex;
  align-items: center;
  height: 40px;
  height: 100%;
  box-sizing: border-box;
  border-bottom: 0px ;
  user-select: none;
  position: relative;
  /* background-color: royalblue; */
}

.titlebar--draggable {
  -webkit-app-region: drag;
}

.titlebar__controls {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
}

.titlebar__button {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  border: 0px solid rgba(0, 0, 0, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.15s ease;
  padding: 0;
  position: relative;
}

.titlebar__button svg {
  opacity: 0;
  transition: opacity 0.15s ease;
  color: rgba(0, 0, 0, 0.7);
}

.titlebar:hover .titlebar__button svg,
.titlebar__button:hover svg {
  opacity: 1;
}

.titlebar__button--close {
  background: linear-gradient(135deg, #ff6058 0%, #ff4f47 100%);
  border-color: #e33e32;
}

.titlebar__button--close:hover {
  background: linear-gradient(135deg, #ff5751 0%, #ff3d35 100%);
}

.titlebar__button--close:active {
  background: linear-gradient(135deg, #ff4a42 0%, #ff2d24 100%);
}

.titlebar__button--minimize {
  background: linear-gradient(135deg, #ffbd2e 0%, #ffab00 100%);
  border-color: #e09900;
}

.titlebar__button--minimize:hover {
  background: linear-gradient(135deg, #ffb524 0%, #ff9f00 100%);
}

.titlebar__button--minimize:active {
  background: linear-gradient(135deg, #ffac19 0%, #ff9200 100%);
}

.titlebar__button--fullscreen {
  background: linear-gradient(135deg, #28ca42 0%, #1aad34 100%);
  border-color: #128928;
}

.titlebar__button--fullscreen:hover {
  background: linear-gradient(135deg, #25c23e 0%, #17a330 100%);
}

.titlebar__button--fullscreen:active {
  background: linear-gradient(135deg, #22b83a 0%, #149929 100%);
}

.titlebar__content {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 0;
}


/* Dark mode support */
@media (prefers-color-scheme: dark) {
  .titlebar {
    /* background: linear-gradient(to bottom, #3c3c3c 0%, #2d2d2d 100%); */
    /* border-bottom-color: #1e1e1e; */
  }
  
  .titlebar__title {
    color: #fff;
  }
  
  .titlebar__button {
    /* border-color: rgba(255, 255, 255, 0.1); */
  }
  
  .titlebar__button svg {
    color: rgba(255, 255, 255, 0.8);
  }
}

/* Windows/Linux alternative styling */
.titlebar--windows {
  /* background: #f0f0f0; */
  /* border-bottom: 1px solid #ccc; */
}

.titlebar--windows .titlebar__button {
  border-radius: 0;
  width: 46px;
  height: 32px;
  background: transparent;
  border: none;
}

.titlebar--windows .titlebar__button:hover {
  /* background: rgba(0, 0, 0, 0.1); */
}

.titlebar--windows .titlebar__button--close:hover {
  background: #e81123;
  color: white;
}
</style>