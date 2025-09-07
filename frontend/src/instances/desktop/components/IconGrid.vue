<template>
  <div class="apps-container-full">
    <div class="apps-container">

      <div ref="scrollableElement" class="apps-grid">
        <div v-for="icon in icons" class="apps-grid-item-container"
          @click="selectIcon(icon)">
          <div class="apps-grid-item" :class="{ 'apps-grid-item-selected' : selectedIcon === icon}" >
            <img class="small-icons icon-picker-icon" :src="getAppIcon(icon)">
          </div>
        </div>
      </div>

    </div>
    </div>
</template>
  
<script setup>
import { ref, computed, onMounted } from 'vue';
import { useIconStore } from '@/stores/icons';

const props = defineProps({
  icons: Array
});

const emit = defineEmits([
  'icon-selected', 
]);

const iconStore = useIconStore();

const selectedIcon = ref('generic');
const selectedIconType = ref(iconStore.iconTypes[0]);

const getAppIcon = (iconName) => {
    const icon = iconStore.getAppIcon(iconName);
    return icon
};

const selectIcon = (icon) => {
  selectedIcon.value = icon;
  emit('icon-selected', icon);
};


onMounted(async () => {


});

</script>

<style scoped>

.icon-picker-icon{
  box-sizing: border-box;
  padding: 2px;
  height: 32px;
  width: 32px;
  overflow: hidden;
  display: flex;
  align-items: center;
}

.apps-container-full {
  box-sizing: border-box;
  flex-direction: column;
  gap: .5rem;
  width: 100%;
  display: flex;
  position: relative;
  display: flex;
  justify-content: space-between;
  align-items: center;
  /* max-width: 300px; */
  max-height: 260px;
  background-color: crimson;
  background-color: var(--black-steel);
  border-radius: var(--large-radius);
  padding: .5rem;
  outline: var(--transparent-line);
  outline-offset: -1px;
}

.apps-container {
  box-sizing: border-box;
  overflow: hidden;
  /* background-color: chartreuse; */
  width: 100%;
  padding: .2rem;
  flex: 1;
  display: flex;
  position: relative;
  display: flex;
  justify-content: space-between;
  align-items: center;
  align-items: baseline;
  overflow: hidden;
  overflow-y: scroll;
  /* max-height: 400px; */
}

.apps-container::-webkit-scrollbar {
  width: 4px;
}

.apps-container::-webkit-scrollbar-thumb {
  border-radius: 10px;
  background-color: rgba(255, 255, 255, 0.295);
}

.apps-container::-webkit-scrollbar-track {
  border-radius: 10px;
  /* background-color: rgba(0, 0, 0, 0.295); */
}

.apps-grid {
  /* background-color: deeppink; */
  width: 100%;
  display: grid;
  gap: 10px;
  grid-template-columns: auto auto auto auto auto;
  box-sizing: border-box;
}


.apps-grid-item-container {
  box-sizing: border-box;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  overflow: hidden;
  /* gap: 10px; */
  /* padding: .1rem; */
}


.apps-grid-item{
  border-radius: 8px;
  padding: .2rem;
  
}
.apps-grid-item:hover {
  background-color: rgb(121, 121, 121);
  background-color: #ffffff15;
  outline: var(--transparent-line);
  outline-offset: -1px;
}

.apps-grid-item:active {
  background-color: rgb(70, 70, 70);
  background-color: #00000013;
}

.apps-grid-item-selected {
  box-sizing: border-box;
  background-color: rgba(0, 0, 0, 0.216);
  outline: solid 1px white;
  outline-offset: -1px;

}

.apps-grid-item-selected:hover {
  background-color: rgba(0, 0, 0, 0.216);

}
</style>

