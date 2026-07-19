<template>
  <button class="general-button" :class="{ 'item-inactive' : !isActive, 'colored' : colored, 'full-width' : fullWidth, 'loading' : loading }"
    @click="buttonFunction" v-stop-propagation>
    <div v-if="!loading" class="general-button-text">
      <img v-if="icon" class="general-button-icon small-icons" :src="icon" alt="" aria-hidden="true" />
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
  min-width: 30%;
}

.default{
  background-color: var(--surface-inverse);
  color: var(--text);
}

.colored{
  background-color: rgb(44, 117, 226);
  color: var(--text);
}

.inverted-icon{
  filter: invert(100%);
}
.loading{
  width: 35px;
  min-width: 35px;
  padding: 0;
  border-radius: 18px;
}

.general-button:hover{
  /* transform: scale(1.02); */
  border-radius: 10px ;
}

.general-button.loading:hover{
  border-radius: 50%;
}

.general-button:active{
  
  transform: scale(.98);
}

.general-button-loading-icon{
  height: 100%;
  aspect-ratio: 1/1;
  animation: loadingRotate .5s linear infinite;
}

.general-button-text{
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 16px;
}

.general-button-icon {
  width: 14px;
  height: 14px;
  object-fit: contain;
  flex-shrink: 0;
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
  
  
  

