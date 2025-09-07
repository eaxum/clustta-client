<template>
  <div class="modal-mask" v-esc="closeModals">
    <component v-for="modal in visibleModals" :key="modal.name" :is="modal.component" />
  </div>
</template>

<script setup>
import { computed, onMounted, onBeforeUnmount } from 'vue';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useDndStore } from '@/stores/dnd';

// General
import PopUpModal from '@/instances/desktop/modals/PopUpModal.vue';
import LoginModal from '@/instances/desktop/modals/LoginModal.vue';

import IgnoreConfigModal from '@/instances/desktop/modals/IgnoreConfigModal.vue';
import DirectoryConfigModal from '@/instances/desktop/modals/DirectoryConfigModal.vue';
import AppInfoModal from '@/instances/desktop/modals/AppInfoModal.vue';
import EulaModal from '@/instances/desktop/modals/EulaModal.vue';
import DirOnboardModal from '@/instances/desktop/modals/DirOnboardModal.vue';

// Browser
import AddProjectModal from '@/instances/desktop/modals/AddProjectModal.vue';
import CloneProjectModal from '@/instances/desktop/modals/CloneProjectModal.vue';
import EditProjectModal from '@/instances/desktop/modals/EditProjectModal.vue';

import CreateCollectionModal from '@/instances/desktop/modals/CreateCollectionModal.vue';
import EditCollectionModal from '@/instances/desktop/modals/EditCollectionModal.vue';

import CreateAssetModal from '@/instances/desktop/modals/CreateAssetModal.vue';
import SelectAppModal from '@/instances/desktop/modals/SelectAppModal.vue';
import EditAssetModal from '@/instances/desktop/modals/EditAssetModal.vue';

import AddWebLinkModal from '@/instances/desktop/modals/AddWebLinkModal.vue';

import CreateCheckpointModal from '@/instances/desktop/modals/CreateCheckpointModal.vue';
import CreateMultipleCheckpointsModal from '@/instances/desktop/modals/CreateMultipleCheckpointsModal.vue';

import ImportItemsModal from '@/instances/desktop/modals/ImportItemsModal.vue';
import TimelineModal from '@/instances/desktop/modals/TimelineModal.vue';
import AddWorkspaceModal from '@/instances/desktop/modals/AddWorkspaceModal.vue';

import ConfigWorkflowModal from '@/instances/desktop/modals/ConfigWorkflowModal.vue';
import SelectWorkflowModal from '@/instances/desktop/modals/SelectWorkflowModal.vue';

// Settings
import AddTemplateModal from '@/instances/desktop/settings/modals/AddTemplateModal.vue';
import EditTemplateModal from '@/instances/desktop/settings/modals/EditTemplateModal.vue';

import AddAssetTypeModal from '@/instances/desktop/settings/modals/AddAssetTypeModal.vue';
import EditAssetTypeModal from '@/instances/desktop/settings/modals/EditAssetTypeModal.vue';

import AddCollectionTypeModal from '@/instances/desktop/settings/modals/AddCollectionTypeModal.vue';
import EditCollectionTypeModal from '@/instances/desktop/settings/modals/EditCollectionTypeModal.vue';

import EditCollaboratorModal from '@/instances/desktop/settings/modals/EditCollaboratorModal.vue';
import ManageCollaboratorModal from '@/instances/desktop/settings/modals/ManageCollaboratorModal.vue';
import AddCollaboratorModal from '@/instances/desktop/settings/modals/AddCollaboratorModal.vue';

import AddRoleModal from '@/instances/desktop/settings/modals/AddRoleModal.vue';
import EditRoleModal from '@/instances/desktop/settings/modals/EditRoleModal.vue';

import ComposeWorkflowModal from '@/instances/desktop/modals/ComposeWorkflowModal.vue';

// User
import AddProjectTemplateModal from '@/instances/desktop/settings/modals/AddProjectTemplateModal.vue';
import EditProjectTemplateModal from '@/instances/desktop/settings/modals/EditProjectTemplateModal.vue';
import DuplicateProjectTemplateModal from '@/instances/desktop/settings/modals/DuplicateProjectTemplateModal.vue';


import AddUserTemplateModal from '@/instances/desktop/settings/modals/AddUserTemplateModal.vue';
import EditUserTemplateModal from '@/instances/desktop/settings/modals/EditUserTemplateModal.vue';

import AddUserAssetTypeModal from '@/instances/desktop/settings/modals/AddUserAssetTypeModal.vue';
import EditUserAssetTypeModal from '@/instances/desktop/settings/modals/EditUserAssetTypeModal.vue';

import AddUserCollectionTypeModal from '@/instances/desktop/settings/modals/AddUserCollectionTypeModal.vue';
import EditUserCollectionTypeModal from '@/instances/desktop/settings/modals/EditUserCollectionTypeModal.vue';

const props = defineProps({
  component: String,
  show: Boolean,
});

const modals = useDesktopModalStore();
const dndStore = useDndStore();

const modalComponents = {

  // user
  addProjectTemplateModal: AddProjectTemplateModal,
  editProjectTemplateModal: EditProjectTemplateModal,
  duplicateProjectTemplateModal: DuplicateProjectTemplateModal,

  addUserTemplateModal: AddUserTemplateModal,
  editUserTemplateModal: EditUserTemplateModal,

  addUserAssetTypeModal: AddUserAssetTypeModal,
  editUserAssetTypeModal: EditUserAssetTypeModal,
  addUserCollectionTypeModal: AddUserCollectionTypeModal,
  editUserCollectionTypeModal: EditUserCollectionTypeModal,

  addWebLinkModal: AddWebLinkModal,

  addAssetTypeModal: AddAssetTypeModal,
  editAssetTypeModal: EditAssetTypeModal,

  addCollectionTypeModal: AddCollectionTypeModal,
  editCollectionTypeModal: EditCollectionTypeModal,

  addRoleModal: AddRoleModal,
  editRoleModal: EditRoleModal,

  composeWorkflowModal: ComposeWorkflowModal,

  ignoreConfigModal: IgnoreConfigModal,
  directoryConfigModal: DirectoryConfigModal,
  importItemsModal: ImportItemsModal,
  timelineModal: TimelineModal,
  addWorkspaceModal: AddWorkspaceModal,
  appInfoModal: AppInfoModal,
  eulaModal: EulaModal,
  dirOnboardModal: DirOnboardModal,
  
  editProjectModal: EditProjectModal,
  editAssetModal: EditAssetModal,
  editCollectionModal: EditCollectionModal,
  createMultipleCheckpointsModal: CreateMultipleCheckpointsModal,
  editCollaboratorModal: EditCollaboratorModal,
  editTemplateModal: EditTemplateModal,
  addTemplateModal: AddTemplateModal,
  manageCollaboratorModal: ManageCollaboratorModal,
  addCollaboratorModal: AddCollaboratorModal,
  createCheckpointModal: CreateCheckpointModal,
  addProjectModal: AddProjectModal,
  cloneProjectModal: CloneProjectModal,
  popUpModal: PopUpModal,
  loginModal: LoginModal,

  configWorkflowModal: ConfigWorkflowModal,
  selectWorkflowModal: SelectWorkflowModal,

  createAssetModal: CreateAssetModal,
  createCollectionModal: CreateCollectionModal,
  selectAppModal: SelectAppModal,
};

const closeModals = () => {
  const restrictedModals = ['dirOnboardModal', 'eulaModal']
  if(restrictedModals.includes(modals.activeModal)) return
  modals.disableAllModals();
};

const visibleModals = computed(() => {
  return Object.entries(modals.modalStates)
    .filter(([name, isVisible]) => isVisible)
    .map(([name]) => ({
      name,
      component: modalComponents[name],
    }));
});

const handleClickOutside = (event) => {
  const restrictedModals = ['dirOnboardModal', 'eulaModal']
  if(restrictedModals.includes(modals.activeModal)) return
  if (modals.activeModal) {
    if (event && !event.target.closest('.modal-container')) {
      console.log('clicked outside');
      dndStore.resetValues();
      modals.disableAllModals();
    }

  }
};

onMounted(() => {
  document.addEventListener('click', handleClickOutside);
});

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside);
});
</script>


<style scoped>
@import "@/assets/desktop.css";

.list-item {

  cursor: pointer;
  display: flex;
  gap: 10px;
  align-items: center;
  width: 100%;
  padding: 10px 15px;

}


.icons {
  height: 24px;
  width: 24px;

}

.profile-menu-list {

  width: 240px;
  display: flex;
  align-items: center;
  flex-direction: column;
  align-items: flex-start;
  padding: 0px px;
  /* gap: 20px */
}

.modal-mask {
  position: absolute;
  z-index: 2;
  width: 100%;
  height: 100%;
  display: flex;
  transition: opacity 0.3s ease;
  background-color: rgba(0, 0, 0, 0.5);
  align-items: center;
  justify-content: center;
  backdrop-filter: blur(3px);
  box-sizing: border-box;
}


.modal-header h3 {
  margin-top: 0;
  color: #42b983;
}

.modal-body {
  margin: 20px 0;
}

.modal-default-button {
  float: right;
}

/*
 * The following styles are auto-applied to elements with
 * transition="modal" when their visibility is toggled
 * by Vue.js.
 *
 * You can easily play with the modal transition by editing
 * these styles.
 */

.modal-enter-from {
  opacity: 0;
}

.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .modal-container,
.modal-leave-to .modal-container {
  -webkit-transform: scale(1.1);
  transform: scale(1.1);
}
</style>

