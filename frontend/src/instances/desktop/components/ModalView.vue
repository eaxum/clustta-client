<template>
  <div class="modal-mask" :class="{ 'modal-mask-progress': progressRunning }" 
     v-esc="closeModals">
    <component v-for="modal in visibleModals" :key="modal.name" :is="modal.component" />
  </div>
</template>

<script setup>
// imports
import { computed, onBeforeUnmount, onMounted } from 'vue';

// components - browser
import AddDependencyPresetModal from '@/instances/desktop/modals/AddDependencyPresetModal.vue';
import AddProjectModal from '@/instances/desktop/modals/AddProjectModal.vue';
import AddWebLinkModal from '@/instances/desktop/modals/AddWebLinkModal.vue';
import AddWorkspaceModal from '@/instances/desktop/modals/AddWorkspaceModal.vue';
import CloneProjectModal from '@/instances/desktop/modals/CloneProjectModal.vue';
import CreateAssetModal from '@/instances/desktop/modals/CreateAssetModal.vue';
import CreateCheckpointModal from '@/instances/desktop/modals/CreateCheckpointModal.vue';
import CreateCollectionModal from '@/instances/desktop/modals/CreateCollectionModal.vue';
import CreateMultipleCheckpointsModal from '@/instances/desktop/modals/CreateMultipleCheckpointsModal.vue';
import EditAssetModal from '@/instances/desktop/modals/EditAssetModal.vue';
import EditCollectionModal from '@/instances/desktop/modals/EditCollectionModal.vue';
import EditProjectModal from '@/instances/desktop/modals/EditProjectModal.vue';
import ImportItemsModal from '@/instances/desktop/modals/ImportItemsModal.vue';
import ProjectDetailsModal from '@/instances/desktop/modals/ProjectDetailsModal.vue';
import SelectAppModal from '@/instances/desktop/modals/SelectAppModal.vue';
import UploadProjectModal from '@/instances/desktop/modals/UploadProjectModal.vue';

// components - general
import AppInfoModal from '@/instances/desktop/modals/AppInfoModal.vue';
import BackUpProjectModal from '@/instances/desktop/modals/BackUpProjectModal.vue';
import DirOnboardModal from '@/instances/desktop/modals/DirOnboardModal.vue';
import DirectoryConfigModal from '@/instances/desktop/modals/DirectoryConfigModal.vue';
import EulaModal from '@/instances/desktop/modals/EulaModal.vue';
import IgnoreConfigModal from '@/instances/desktop/modals/IgnoreConfigModal.vue';
import ImportProjectModal from '@/instances/desktop/modals/ImportProjectModal.vue';
import LoginModal from '@/instances/desktop/modals/LoginModal.vue';
import PopUpModal from '@/instances/desktop/modals/PopUpModal.vue';

// components - settings (project)
import AddAssetTypeModal from '@/instances/desktop/settings/modals/AddAssetTypeModal.vue';
import AddCollaboratorModal from '@/instances/desktop/settings/modals/AddCollaboratorModal.vue';
import AddCollectionTypeModal from '@/instances/desktop/settings/modals/AddCollectionTypeModal.vue';
import AddRoleModal from '@/instances/desktop/settings/modals/AddRoleModal.vue';
import AddTemplateModal from '@/instances/desktop/settings/modals/AddTemplateModal.vue';
import EditAssetTypeModal from '@/instances/desktop/settings/modals/EditAssetTypeModal.vue';
import EditCollectionTypeModal from '@/instances/desktop/settings/modals/EditCollectionTypeModal.vue';
import EditRoleModal from '@/instances/desktop/settings/modals/EditRoleModal.vue';
import EditTemplateModal from '@/instances/desktop/settings/modals/EditTemplateModal.vue';
import ManageCollaboratorModal from '@/instances/desktop/settings/modals/ManageCollaboratorModal.vue';

// components - settings (user)
import AddProjectTemplateModal from '@/instances/desktop/settings/modals/AddProjectTemplateModal.vue';
import AddUserAssetTypeModal from '@/instances/desktop/settings/modals/AddUserAssetTypeModal.vue';
import AddUserCollectionTypeModal from '@/instances/desktop/settings/modals/AddUserCollectionTypeModal.vue';
import AddUserTemplateModal from '@/instances/desktop/settings/modals/AddUserTemplateModal.vue';
import DuplicateProjectTemplateModal from '@/instances/desktop/settings/modals/DuplicateProjectTemplateModal.vue';
import EditProjectTemplateModal from '@/instances/desktop/settings/modals/EditProjectTemplateModal.vue';
import EditUserAssetTypeModal from '@/instances/desktop/settings/modals/EditUserAssetTypeModal.vue';
import EditUserCollectionTypeModal from '@/instances/desktop/settings/modals/EditUserCollectionTypeModal.vue';
import EditUserTemplateModal from '@/instances/desktop/settings/modals/EditUserTemplateModal.vue';

// components - studio
import ConfigClusttaCloudStudioModal from '@/instances/desktop/modals/ConfigClusttaCloudStudioModal.vue';
import ConfigSelfManagedStudioModal from '@/instances/desktop/modals/ConfigSelfManagedStudioModal.vue';
import SelectNewStudioTypeModal from '@/instances/desktop/modals/SelectNewStudioTypeModal.vue';
import UpdateStudioModal from '@/instances/desktop/modals/UpdateStudioModal.vue';

// components - sync
import SyncConflictModal from '@/instances/desktop/modals/SyncConflictModal.vue';

// components - integrations
import IntegrationAuthModal from '@/instances/desktop/modals/IntegrationAuthModal.vue';
import IntegrationLinkModal from '@/instances/desktop/modals/IntegrationLinkModal.vue';
import IntegrationSyncModal from '@/instances/desktop/modals/IntegrationSyncModal.vue';

// components - diagnostics
import SubmitDiagnosticsModal from '@/instances/desktop/modals/SubmitDiagnosticsModal.vue';

// components - workflow
import ComposeWorkflowModal from '@/instances/desktop/modals/ComposeWorkflowModal.vue';
import ConfigWorkflowModal from '@/instances/desktop/modals/ConfigWorkflowModal.vue';
import SelectWorkflowModal from '@/instances/desktop/modals/SelectWorkflowModal.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useDndStore } from '@/stores/dnd';
import { useNotificationStore } from '@/stores/notifications';

const dndStore = useDndStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();

// props
const props = defineProps({
  component: String,
  show: Boolean,
});

// modal components mapping
const modalComponents = {
  // browser
  addDependencyPresetModal: AddDependencyPresetModal,
  addProjectModal: AddProjectModal,
  addWebLinkModal: AddWebLinkModal,
  addWorkspaceModal: AddWorkspaceModal,
  cloneProjectModal: CloneProjectModal,
  createAssetModal: CreateAssetModal,
  createCheckpointModal: CreateCheckpointModal,
  createCollectionModal: CreateCollectionModal,
  createMultipleCheckpointsModal: CreateMultipleCheckpointsModal,
  editAssetModal: EditAssetModal,
  editCollectionModal: EditCollectionModal,
  editProjectModal: EditProjectModal,
  importItemsModal: ImportItemsModal,
  projectDetailsModal: ProjectDetailsModal,
  selectAppModal: SelectAppModal,
  uploadProjectModal: UploadProjectModal,

  // general
  appInfoModal: AppInfoModal,
  backUpProjectModal: BackUpProjectModal,
  dirOnboardModal: DirOnboardModal,
  directoryConfigModal: DirectoryConfigModal,
  eulaModal: EulaModal,
  ignoreConfigModal: IgnoreConfigModal,
  importProjectModal: ImportProjectModal,
  loginModal: LoginModal,
  popUpModal: PopUpModal,

  // settings (project)
  addAssetTypeModal: AddAssetTypeModal,
  addCollaboratorModal: AddCollaboratorModal,
  addCollectionTypeModal: AddCollectionTypeModal,
  addRoleModal: AddRoleModal,
  addTemplateModal: AddTemplateModal,
  editAssetTypeModal: EditAssetTypeModal,
  editCollectionTypeModal: EditCollectionTypeModal,
  editRoleModal: EditRoleModal,
  editTemplateModal: EditTemplateModal,
  manageCollaboratorModal: ManageCollaboratorModal,

  // settings (user)
  addProjectTemplateModal: AddProjectTemplateModal,
  addUserAssetTypeModal: AddUserAssetTypeModal,
  addUserCollectionTypeModal: AddUserCollectionTypeModal,
  addUserTemplateModal: AddUserTemplateModal,
  duplicateProjectTemplateModal: DuplicateProjectTemplateModal,
  editProjectTemplateModal: EditProjectTemplateModal,
  editUserAssetTypeModal: EditUserAssetTypeModal,
  editUserCollectionTypeModal: EditUserCollectionTypeModal,
  editUserTemplateModal: EditUserTemplateModal,

  // studio
  configClusttaCloudStudioModal: ConfigClusttaCloudStudioModal,
  configSelfManagedStudioModal: ConfigSelfManagedStudioModal,
  selectNewStudioTypeModal: SelectNewStudioTypeModal,
  updateStudioModal: UpdateStudioModal,

  // sync
  syncConflictModal: SyncConflictModal,

  // integrations
  integrationAuthModal: IntegrationAuthModal,
  integrationLinkModal: IntegrationLinkModal,
  integrationSyncModal: IntegrationSyncModal,

  // diagnostics
  submitDiagnosticsModal: SubmitDiagnosticsModal,

  // workflow
  composeWorkflowModal: ComposeWorkflowModal,
  configWorkflowModal: ConfigWorkflowModal,
  selectWorkflowModal: SelectWorkflowModal,
};

// computed
// Checks if a progress operation is running.
const progressRunning = computed(() => notificationStore.getProgress.running);

// Returns list of currently visible modal components.
const visibleModals = computed(() => {
  return Object.entries(modals.modalStates)
    .filter(([name, isVisible]) => isVisible)
    .map(([name]) => ({
      name,
      component: modalComponents[name],
    }));
});

// methods
// Closes all modals unless they are restricted.
const closeModals = () => {
  const restrictedModals = ['dirOnboardModal', 'eulaModal'];
  if (restrictedModals.includes(modals.activeModal)) return;
  modals.disableAllModals();
};

// Handles clicks outside the modal to close it.
const handleClickOutside = (event) => {
  const restrictedModals = ['dirOnboardModal', 'eulaModal'];
  if (restrictedModals.includes(modals.activeModal)) return;
  if (modals.activeModal) {
    if (event && !event.target.closest('.modal-container')) {
      dndStore.resetValues();
      modals.disableAllModals();
    }
  }
};

// lifecycle hooks
onMounted(() => {
  document.addEventListener('click', handleClickOutside);
});

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside);
});
</script>


<style scoped>
@import "@/assets/desktop.css";

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

.modal-mask-progress {
  cursor: wait;
}

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

