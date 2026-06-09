<template>
  <div class="grid-status-menu-overlay" v-stop-propagation>
    <div class="grid-status-menu-container">
      <div v-for="(status, index) in projectStatuses" :key="index" 
        v-stop-propagation 
        @click="selectStatus(status)"
        :style="{ backgroundColor: status.color }" 
        class="grid-status-menu-pill">
        <p class="grid-status-menu-pill-text">{{ status.short_name }}</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useAssetStore } from '@/stores/assets';
import { useStatusStore } from '@/stores/status';
import { useUserStore } from '@/stores/users';
import { useStageStore } from '@/stores/stages';
import { useProjectStore } from '@/stores/projects';
import { AssetService } from "@/services";
import emitter from '@/lib/mitt';

const stage = useStageStore();
const assetStore = useAssetStore();
const statusStore = useStatusStore();
const userStore = useUserStore();
const projectStore = useProjectStore();

const emits = defineEmits(['statusSelected', 'close']);

const projectStatuses = computed(() => {
  const allStatuses = statusStore.statuses;
  if (!userStore.canDo('set_done_asset')) {
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
  let asset = assetStore.selectedAsset;
  
  await AssetService.ChangeStatus(projectPath, asset.id, status.id)
    .then((data) => {
      asset.status_short_name = status.short_name;
      asset.status = status;
      
      emitAssetUpdates(asset.id, [
        { property: 'status_short_name', value: status.short_name },
        { property: 'status', value: status }
      ]);
      
    })
    .catch((error) => {
      console.error('Error:', error);
    });
    
  stage.operationActive = false;
  emits('statusSelected');
  emits('close');
};

// Helper function to emit asset data updates
const emitAssetUpdates = (assetId, updates) => {
  const updateData = { itemId: assetId, updates };
  
  // Emit to both Browser and VirtuaItem components
  emitter.emit('update-root-data', updateData);
  emitter.emit('update-children', updateData);
};
</script>

<style scoped>
@import "@/assets/desktop.css";

.grid-status-menu-overlay {
  position: absolute;
  bottom: 0px;
  left: 0px;
  z-index: 1000;
  opacity: 0;
  animation: fadeInScale 0.15s ease-out forwards;
  border-radius: var(--normal-radius);
  border: 1px solid hsl(var(--border));
  
  backdrop-filter: blur(35px);
}

@keyframes fadeInScale {
  from {
    opacity: 0;
    transform: translateY(10px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.grid-status-menu-container {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 0.5rem;
    padding: 0.75rem;
    border-radius: var(--normal-radius);
    transition: opacity 0.3s ease;

    box-sizing: border-box;
}

.grid-status-menu-pill {
  display: flex;
  border-radius: var(--normal-radius);
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  min-width: 70px;
  padding: 0.5rem 0.6rem;
  height: 32px;
  font-size: 12px;
  text-transform: uppercase;
  font-weight: 700;
  color: hsl(var(--background));
  transition: all 0.2s ease-out;
  cursor: pointer;
}

.grid-status-menu-pill:hover {
  filter: brightness(1.2);
  border-radius: var(--small-radius);
}

.grid-status-menu-pill-text {
  font-family: 'Inter', sans-serif;
  font-size: 12px;
  font-weight: 700;
  color: rgb(15, 15, 15);
  text-transform: uppercase;
  user-select: none;
}
</style>
