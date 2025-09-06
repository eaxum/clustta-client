<template>
  <button class="general-button" :class="{ 'item-inactive' : !isActive, 'colored' : colored, 'full-width' : fullWidth }"
    @click="buttonFunction" v-stop-propagation>
    <div v-if="!loading" class="general-button-text">
      <div v-if="label">{{ label }}</div>
    </div>
    <div v-else class="general-button-loading-icon" :class="{ 'inverted-icon' : !colored }" >
      <img src="/icons/loading.svg"/>
    </div>
  </button>
</template>
  
<script setup>
const props = defineProps({
  icon: String,
  buttonFunction: { type:Function, default: () => {  } },
  label: { type: String, default: 'Button'},
  colored: { type: Boolean, default: true},
  loading: { type: Boolean, default: false},
  isActive: { type: Boolean, default: true},
  showLabel: { type: Boolean, default: false},
  fullWidth: { type: Boolean, default: true},
});
 
</script>
  
<style scoped>

/* @import "@/assets/desktop.css"; */
.general-button {
  display: flex;
  flex-direction: row;
  justify-content: center;
  align-items: center;
  border-radius: 16px;
  padding: 10px 18px;
  border: 0px;
  font-size: 16px;
  cursor: pointer;
  transition: all 0.3s ease;
  height: 35px;
  overflow: hidden;
  width: max-content;
  min-width: 100px;
  outline-offset: -1px;
  outline: var(--transparent-line);
}

.full-width{
  width: 30%;
}

.default{
  background-color: var(--white);
  color: var(--white);
}

.colored{
  background-color: rgb(44, 117, 226);
  color: var(--white);
}

.general-button:hover{
  transform: scale(1.02);
}

.general-button:active{
  
  transform: scale(.98);
}

.general-button-loading-icon{
  height: 150%;
  aspect-ratio: 1/1;
  animation: loadingRotate .5s linear infinite;
}

.general-button-text{
  font-size: 16px;
}

@keyframes loadingRotate {
  from {
      transform: rotate(0deg);
  }
  to {
      transform: rotate(360deg);
  }
}
</style>
  
  
  