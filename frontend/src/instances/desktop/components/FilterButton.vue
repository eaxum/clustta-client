<template>
  <span v-stop-propagation @click="buttonFunction" :class="{'button-background': alert, 'full-width' : fullWidth, 
  'outline' : useOutline, 'icon-after' : iconAfter, 'centered' : centered }" 
  class="filter-button">
    <img class="small-icons no-cursor" :class="{ 'inverted-icon': alert }" :src="icon">
    <div v-if="showLabel && label" class="small-icons label-text no-cursor" :class="{ 'label-text-inverted': alert }">{{ label }}</div>
    <img class="small-icons chevron no-cursor" :class="{ 'inverted-icon': alert }" src="/icons/chevron_down_white.svg">
  </span>
</template>
  
<script setup>

const props = defineProps({
  icon: String,
  label: String,
  buttonFunction: Function,
  smallIcons: { type: Boolean, default: false},
  alert: { type: Boolean, default: false},
  showLabel: { type: Boolean, default: false},
  iconAfter: { type: Boolean, default: false},
  useOutline: { type: Boolean, default: false},
  centered: { type: Boolean, default: false},
  useBackground: { type: Boolean, default: false},
  fullWidth: { type: Boolean, default: false},

});
 
</script>
  
<style scoped>
  @import "@/assets/desktop.css";
.inverted-icon{
  filter: none;
}

[data-theme="dark"] .inverted-icon{
  filter: invert(1);
}

.filter-alert {
overflow: hidden;
width: 2px;
height: 2px;
background-color: #ecb603;
border-radius: 5px;
position: absolute;
display: flex;
align-items: center;
justify-content: center;
top: 2px;
right: 2px;
border-radius: 10px;
padding: 3px;
font-size: 12px;
color: var(--white);
}

.no-cursor{
  pointer-events: none;
}

.label-text{
    font-size: 14px;
    font-weight: 400;
}

.label-text-inverted{
    font-size: 14px;
    font-weight: 400;
}

.chevron{
  pointer-events: none;
  height: 10px;
  min-width: 10px;
  /* transform: rotate(-90deg); */
  transition: all .1s ease-in;
}
.filter-button{
  overflow: hidden;
  background-color: transparent;
  text-align: center;
  font-size: 14px;
  line-height: 14px;
  color: var(--white);
  position: relative;
  border-radius: 8px;
  border-radius: var(--normal-radius);
  box-sizing: border-box;
  cursor: pointer;
  display: flex;
  align-items: center;
  padding: .3rem;
  height: max-content;
  width: max-content;
  min-width: max-content;
  min-height: max-content;
  /* transition: all 0.3s ease; */
  outline-offset: -1px;
  background-color: var(--black-steel);
  padding-right: .5rem;
  gap: .2rem;


}

.filter-button:hover{
  /* background-color: var(--silver); */
  outline: var(--transparent-line);
}

.filter-button:active{
  background-color: #00000013;
}

.filter-button-pressed {
  box-sizing: border-box;
  background-color: rgba(0, 0, 0, 0.216);
  outline: solid 1px var(--white);
  outline-offset: -1px;
}



.full-width{
  width: 100%;
}

.icon-after{
  justify-content: space-between;
}

.centered{
  justify-content: space-around;
}

.button-background{
    background-color: var(--white);
    color: var(--black)
  /* outline: solid 1px #ecb603; */
  /* border-bottom: solid 2px white; */
}
</style>
  
  
  