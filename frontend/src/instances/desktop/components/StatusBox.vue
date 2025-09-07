<template>
  <div class="status-box-list-parent"
    :style="{ top: flyoutTop + 'px', right: flyoutLeft + 'px', width: flyoutWidth + 'px' }">
    <div v-for="(status, index) in statusStore.statuses" :key="index" @click="selectStatus(status)"
      :style="{ backgroundColor: status.color }" v-stop-propagation class="status-box-pill">
      <p class="status-box-pill-text">{{ status.short_name }}</p>
    </div>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted } from 'vue';
import { useTrayStates } from '@/stores/TrayStates';
import { useMenu } from '@/stores/menu';
import { useAssetStore } from '@/stores/assets';
import { useStatusStore } from '@/stores/status';

const menu = useMenu();
const assetStore = useAssetStore();
const statusStore = useStatusStore();

const flyoutTop = computed(() => menu.flyoutTop);
const flyoutLeft = computed(() => menu.flyoutLeft);
const flyoutWidth = computed(() => menu.flyoutWidth);

const selectStatus = (status) => {
  assetStore.setStatus(status);
  menu.showStatusOptions = false;
};

const handleClickOutside = (event) => {
  if (menu.showStatusOptions && !event.target.closest('.task-item-status')) {
    menu.showStatusOptions = false;
    //console.log('clicked outside');
  }
};

onMounted(() => {
  // //console.log('The document was loaded.');
  document.addEventListener('click', handleClickOutside);
});

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside);
});



</script>

<style scoped>
@import "@/assets/desktop.css";

.status-box-list-parent {
  opacity: 0;
  animation: fadeIn .1s ease-in-out forwards;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateX(10px);
  }

  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.status-box-list-parent {
  position: fixed;
  width: max-content;
  padding: 0.2rem;
  border-radius: 0.5rem;
  background-color: #fff;
  border: 1px solid #e0e0e0;
  z-index: 1000;
  display: flex;
  flex-direction: column;
  gap: 2px;
  align-items: center;
  justify-content: center;
  margin-left: auto;
  margin-right: auto;
  box-sizing: border-box;
  border-radius: var(--normal-radius);
}

.main-status-box-pill {
  box-sizing: border-box;
  background-color: #51e064;
  width: 4rem;
  height: 2rem;
  border-radius: 10px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.status-box-pill {
  box-sizing: border-box;
  background-color: rgb(81, 224, 100);
  width: 4rem;
  height: 2rem;
  width: 100%;
  border-radius: 10px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;



  display: flex;
  border-radius: var(--normal-radius);
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  padding: .2rem;
  background-color: firebrick;
}

.status-box-pill:hover {
  background-color: rgb(206, 165, 51);
}

.status-box-pill-text {
  font-family: 'Inter', sans-serif;
  font-size: 14px;
  font-weight: 400;
  color: rgb(15, 15, 15);

  font-size: 12px;
  text-transform: uppercase;
  font-weight: 700;
  color: black;
}
</style>



