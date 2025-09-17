<template>
  <div class="status-menu-list-parent">
    <div v-for="(status, index) in projectStatuses" :key="index" v-stop-propagation @click="selectStatus(status)"
      :style="{ backgroundColor: status.color }" class="status-menu-pill">
      <p class="status-menu-pill-text">{{ status.short_name }}</p>
    </div>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted } from 'vue';
import { useMenu } from '@/stores/menu';
import { useAssetStore } from '@/stores/assets';
import { useStatusStore } from '@/stores/status';
import { useUserStore } from '@/stores/users';
import { useStageStore } from '@/stores/stages';
import { useProjectStore } from '@/stores/projects';


// services
import { AssetService } from "@/../bindings/clustta/services";
import emitter from '@/lib/mitt';

const menu = useMenu();
const stage = useStageStore();
const assetStore = useAssetStore();
const statusStore = useStatusStore();
const userStore = useUserStore();
const projectStore = useProjectStore();


const emits = defineEmits(['statusSelected']);

const projectStatuses = computed(() => {
  const allStatuses = statusStore.statuses;
  if (!userStore.canDo('set_done_task')) {
    const limitedStatus = ['done', 'retake']
    return allStatuses.filter((item) => !limitedStatus.includes(item.short_name))
  } else {
    return allStatuses
  }
});

const selectStatus = async (fullStatus) => {
  let statusName = fullStatus.short_name;
  stage.operationActive = true;
  const projectPath = projectStore.activeProject.uri;
  const status = statusStore.statuses.find(item => item.short_name === statusName.toLowerCase());
  let task = assetStore.selectedAsset;
  
  await AssetService.ChangeStatus(projectPath, task.id, status.id)
    .then((data) => {
      task.status_short_name = status.short_name;
      task.status = status;
      
      emitTaskUpdates(task.id, [
        { property: 'status_short_name', value: status.short_name },
        { property: 'status', value: status }
      ]);
      
    })
    .catch((error) => {
      console.error('Error:', error);
    });
    
  stage.operationActive = false;
  emits('statusSelected')
  menu.hideContextMenu();
};

// Helper function to emit task data updates
const emitTaskUpdates = (taskId, updates) => {
  const updateData = { itemId: taskId, updates };
  
  // Emit to both Browser and VirtuaItem components
  emitter.emit('update-root-data', updateData);
  emitter.emit('update-children', updateData);
};


const handleClickOutside = (event) => {
  if (menu.showStatusOptions && !event.target.closest('.task-item-status')) {
    menu.hideContextMenu();;
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

.status-menu-list-parent {
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

.status-menu-list-parent {
  position: relative;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: max-content;
  /* padding: .35rem; */
  height: 100%;
  height: 30px;
  gap: .5rem;
  /* overflow: hidden; */
  /* border-radius: var(--normal-radius); */
  /* background-color: var(--steel); */
  /* background-color: firebrick; */
}

.status-menu-pill {
  display: flex;
  border-radius: var(--normal-radius);
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  width: min-content;
  width: 55px;
  padding: .4rem .4rem;
  height: 100%;
  background-color: firebrick;
  font-size: 12px;
  text-transform: uppercase;
  font-weight: 700;
  color: black;
  transition: all 0.2s ease-out;
}

.status-menu-pill:hover {
  background-color: rgb(206, 165, 51);
  border-radius: 10px;
  transform: scale(1.03);
}

.status-menu-pill-text {
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



