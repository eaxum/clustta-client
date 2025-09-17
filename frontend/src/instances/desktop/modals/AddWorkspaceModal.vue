<template>
  <div class="modal-container">
    <HeaderArea :title="title" :icon="'monitor-plus'" :showSearch="showSearch" />
    <div class="general-container">
      <div class="input-section">
        <div class="horizontal-flex">
          <input v-model="workspaceName" class="input-short" type="text" placeholder="Workspace Name" 
          @keydown.enter="handleEnterKey" v-focus />
        </div>
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="'Cancel'" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton :label="'Create'" :fullWidth="true" @click="saveWorkspace" :isActive="isValueChanged"
          :loading="isAwaitingResponse" />
      </div>
    </div>
  </div>
</template>

<script setup>

// imports
import { ref, onMounted, computed } from 'vue';
import { SettingsService } from "@/../bindings/clustta/services";

//stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useCommonStore } from '@/stores/common';
import { useProjectStore } from '@/stores/projects';

//components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';


//header vars
let title = 'Save Workspace';
let showSearch = false;

// stores/states
const modals = useDesktopModalStore();
const commonStore = useCommonStore();
const projectStore = useProjectStore();

//refs
const workspaceName = ref('');
const isAwaitingResponse = ref(false);

// computed props
const isValueChanged = computed(() => {
  const workspaceNames = commonStore.workspaces.map((workspace) => workspace.name);
  let restrictedNames = ['']
  for (let i = 0; i < workspaceNames.length; i++) {
    restrictedNames.push(workspaceNames[i].toLowerCase())
  }
  return !restrictedNames.includes(workspaceName.value.toLowerCase());
});

// methods
const closeModal = () => {
  modals.disableAllModals();
};

const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    saveWorkspace();
  }
};

const saveWorkspace = async () => {
  isAwaitingResponse.value = true;

  const newWorkspace = {
    name: workspaceName.value,
    filters: {
      taskFilters: commonStore.taskFilters,
      entityFilters: commonStore.entityFilters,
      resourceFilters: commonStore.resourceFilters,
      showEntities: commonStore.showEntities,
      showTasks: commonStore.showTasks,
      onlyAssets: commonStore.onlyAssets,
      showResources: commonStore.showResources,
      showChildEntities: commonStore.showChildEntities,
      showChildTasks: commonStore.showChildTasks,
      showChildResources: commonStore.showChildResources,
      showDependencies: commonStore.showDependencies,
      useDeep: commonStore.useDeep,
      hasAssignees: commonStore.hasAssignees,
      noAssignees: commonStore.noAssignees,
    },
    workspaceSearchQuery: commonStore.viewSearchQuery,
  };
  await SettingsService.AddProjectWorkspace(projectStore.getActiveProject.id, newWorkspace)
  commonStore.workspaces.push(newWorkspace);
  commonStore.setActiveWorkspace(newWorkspace);
  isAwaitingResponse.value = false;
  closeModal();

};

</script>

<style scoped>
@import "@/assets/desktop.css";

.general-container {
  gap: .5rem;
  /* background-color: firebrick; */
}

.modal-info {
  display: flex;
  flex-direction: column;
  max-width: 100%;
  justify-content: flex-start;
  align-self: stretch;
  width: 464px;
  align-items: flex-start;
  box-sizing: border-box;

}

.modal-text-container {
  display: flex;
  flex-direction: column;
  max-width: 100%;
  justify-content: flex-start;
  align-self: stretch;
  width: 464px;
  align-items: flex-start;
  /* margin-top: 20px; */
}

.modal-title {
  max-width: 100%;
  align-self: stretch;
  width: 464px;
  color: rgba(16, 24, 40, 1);
  color: white;
  font-size: 18px;
  line-height: 28px;
  letter-spacing: 0%;
  text-align: left;
}

.input-header {
  /* background-color: lightblue; */
  width: 100%;
  display: flex;
  align-items: center;
  margin: 10px 0px;
}

.input-count {
  background-color: none;
  font-size: 14px;
  color: white;
}

.modal-subtitle {
  /* background-color: beige; */
  /* max-width: 100%; */
  align-self: stretch;
  width: 464px;
  color: rgba(16, 24, 40, 1);
  color: white;
  font-size: 14px;
  /* line-height: 28px; */
  letter-spacing: 0%;
  text-align: left;
}



.modal-body {
  box-sizing: border-box;
  max-width: 100%;
  align-self: stretch;
  width: 464px;
  margin: 8px 0px;
  font-size: 14px;
  color: rgba(16, 24, 40, 1);
  line-height: 20px;
  letter-spacing: 0%;
  text-align: left;
}

.modal-actions {
  box-sizing: border-box;
  padding: 1rem 2rem;
  gap: 2rem;
  display: flex;
  flex-direction: row;
  max-width: 100%;
  align-self: stretch;
  align-items: center;
  justify-content: space-evenly;
  width: 464px;
  margin-top: 32px;
}

.div-10 {
  display: flex;
}

.task-options-container {
  position: relative;
  box-sizing: border-box;
  width: 100%;
  height: max-content;
  height: 40px;
  transition: all .2s ease-in-out;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  margin: 0;
}

.task-options-container-closed {
  height: 0px;
  padding: 0;
  margin-bottom: -1rem;
}

.input-short {
  flex: 1;
  width: 100%;
}

.listbox-short {

  flex: 1;
  width: 130px;
}

.input-label {
  font-family: Inter, sans-serif;
  color: white;
  font-size: 16px;
  white-space: nowrap;
  flex: 1;
}

.pop-up-prompt {
  gap: 10px;
  align-items: center;
  justify-content: center;
}
</style>


