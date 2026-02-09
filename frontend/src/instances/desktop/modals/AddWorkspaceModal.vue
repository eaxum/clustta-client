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
import { computed, onMounted, ref } from 'vue';

// components
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// services
import { SettingsService } from '@/services';

// stores
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useProjectStore } from '@/stores/projects';

const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const modals = useDesktopModalStore();
const projectStore = useProjectStore();

// refs
const isAwaitingResponse = ref(false);
const workspaceName = ref('');

// constants
const showSearch = false;
const title = 'Save Workspace';

// computed
// Returns whether the workspace name is valid and not already in use.
const isValueChanged = computed(() => {
  const workspaceNames = commonStore.workspaces.map((workspace) => workspace.name);
  const restrictedNames = ['', ...workspaceNames.map(name => name.toLowerCase())];
  return !restrictedNames.includes(workspaceName.value.toLowerCase());
});

// methods
// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Handles enter key press to submit form.
const handleEnterKey = (event) => {
  if (event.key === 'Enter' && isValueChanged.value) {
    saveWorkspace();
  }
};

// Saves the current workspace configuration.
const saveWorkspace = async () => {
  isAwaitingResponse.value = true;
  let collectionData = null;
  if (collectionStore.navigatedCollection) {
    collectionData = { ...collectionStore.navigatedCollection };
    delete collectionData.preview;
  }
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
    collection: collectionData,
    viewMode: commonStore.viewMode,
  };
  await SettingsService.AddProjectWorkspace(projectStore.getActiveProject.id, newWorkspace);
  commonStore.workspaces.push(newWorkspace);
  commonStore.setActiveWorkspace(newWorkspace);
  isAwaitingResponse.value = false;
  closeModal();
};

// lifecycle hooks
onMounted(() => {
  if (collectionStore.navigatedCollection) {
    workspaceName.value = collectionStore.navigatedCollection.name || '';
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.general-container {
  gap: .5rem;
}

.input-short {
  flex: 1;
  width: 100%;
}
</style>


