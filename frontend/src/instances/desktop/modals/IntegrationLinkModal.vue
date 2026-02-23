<template>
  <div class="modal-container" v-esc="closeModal">
    <HeaderArea :title="title" :icon="'link'" />

    <div class="general-container">
      <!-- Linked Integration Info -->
      <div v-if="linkedIntegration" class="linked-info">
        <div class="linked-header">
          <img :src="getIntegrationIcon(linkedIntegration.integration_id)" class="integration-icon" />
          <div class="linked-details">
            <span class="project-name">{{ linkedIntegration.external_project_name }}</span>
            <span class="integration-name">{{ linkedIntegration.integration_id }}</span>
          </div>
        </div>
        <div class="linked-actions">
          <ActionButton :icon="getAppIcon('sync')" :label="'Sync Now'" :buttonFunction="openSyncModal" />
          <ActionButton :icon="getAppIcon('disconnect')" :label="'Unlink'" :buttonFunction="unlinkProject" />
        </div>
      </div>

      <!-- Integration Selection -->
      <div v-else-if="!selectedIntegration" class="integration-selection">
        <p class="section-label">Select an integration to link this project:</p>
        <div class="integration-list">
          <div v-for="integration in authenticatedIntegrations" :key="integration.id" class="integration-item"
            @click="selectIntegration(integration)">
            <img :src="getIntegrationIcon(integration.icon)" class="integration-icon" />
            <span class="integration-name">{{ integration.name }}</span>
          </div>
        </div>
        <div v-if="authenticatedIntegrations.length === 0" class="empty-state">
          <p>No integrations connected.</p>
          <GeneralButton :label="'Connect Integration'" :buttonFunction="openAuthModal" />
        </div>
      </div>

      <!-- Project Selection -->
      <div v-else class="project-selection">
        <div class="selection-header">
          <img :src="getIntegrationIcon(selectedIntegration.icon)" class="integration-icon" />
          <span>{{ selectedIntegration.name }}</span>
        </div>

        <div v-if="isLoadingProjects" class="loading-state">
          <span>Loading projects...</span>
        </div>

        <div v-else-if="externalProjects.length > 0" class="project-list">
          <div v-for="project in externalProjects" :key="project.id" class="project-item"
            :class="{ 'selected': selectedProject?.id === project.id }" @click="selectProject(project)">
            <span class="project-name">{{ project.name }}</span>
            <span v-if="project.description" class="project-desc">{{ project.description }}</span>
          </div>
        </div>

        <div v-else class="empty-state">
          <p>No projects found in {{ selectedIntegration.name }}</p>
        </div>

        <div class="pop-up-actions">
          <GeneralButton :label="$t('common.back')" :fullWidth="true" :buttonFunction="clearSelection" :colored="false" />
          <GeneralButton :label="'Link Project'" :fullWidth="true" @click="linkProject" :isActive="!!selectedProject"
            :loading="isLinking" />
        </div>
      </div>

      <!-- Close Button -->
      <div v-if="linkedIntegration || !selectedIntegration" class="pop-up-actions">
        <GeneralButton :label="$t('common.close')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useIntegrationStore } from '@/stores/integrations';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';

const { t } = useI18n();
const iconStore = useIconStore();
const integrationStore = useIntegrationStore();
const modals = useDesktopModalStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();

// refs
const externalProjects = ref([]);
const isLinking = ref(false);
const isLoadingProjects = ref(false);
const selectedIntegration = ref(null);
const selectedProject = ref(null);

// constants
const title = 'Link Integration';

// computed
// Returns integrations the user has authenticated with.
const authenticatedIntegrations = computed(() => {
  return integrationStore.availableIntegrations.filter(i => integrationStore.isAuthenticated(i.id));
});

// Returns the linked integration for current project.
const linkedIntegration = computed(() => integrationStore.linkedIntegration);

// methods
// Clears the selection state.
const clearSelection = () => {
  selectedIntegration.value = null;
  selectedProject.value = null;
  externalProjects.value = [];
};

// Closes the modal.
const closeModal = () => {
  modals.disableAllModals();
};

// Returns the app icon path.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Returns the integration icon path.
const getIntegrationIcon = (iconName) => {
  return `/icons/${iconName}.svg`;
};

// Links the current project to the selected external project.
const linkProject = async () => {
  if (!selectedProject.value) return;

  isLinking.value = true;
  try {
    await integrationStore.linkProject(
      selectedIntegration.value.id,
      selectedProject.value.id,
      selectedProject.value.name
    );
    clearSelection();
  } catch (error) {
    // Error handled by store
  } finally {
    isLinking.value = false;
  }
};

// Loads external projects for the selected integration.
const loadExternalProjects = async () => {
  if (!selectedIntegration.value) return;

  isLoadingProjects.value = true;
  externalProjects.value = [];

  try {
    externalProjects.value = await integrationStore.getExternalProjects(selectedIntegration.value.id);
  } catch (error) {
    notificationStore.sendNotification('Failed to load projects: ' + error.message, 'error');
  } finally {
    isLoadingProjects.value = false;
  }
};

// Opens the authentication modal.
const openAuthModal = () => {
  modals.setModalVisibility('integrationAuthModal', true);
};

// Opens the sync modal.
const openSyncModal = () => {
  modals.setModalVisibility('integrationSyncModal', true);
};

// Selects an integration to browse projects.
const selectIntegration = (integration) => {
  selectedIntegration.value = integration;
  loadExternalProjects();
};

// Selects an external project to link.
const selectProject = (project) => {
  selectedProject.value = project;
};

// Unlinks the current project from its integration.
const unlinkProject = async () => {
  try {
    await integrationStore.unlinkProject();
  } catch (error) {
    // Error handled by store
  }
};

// watchers
watch(() => projectStore.activeProject, async () => {
  await integrationStore.loadLinkedIntegration();
});

// lifecycle
onMounted(async () => {
  await integrationStore.loadAvailableIntegrations();
  await integrationStore.loadLinkedIntegration();
});
</script>

<style scoped>
.linked-info {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 16px;
  background: var(--surface-primary);
  border-radius: var(--small-radius);
}

.linked-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.linked-details {
  display: flex;
  flex-direction: column;
}

.project-name {
  font-weight: 500;
  color: var(--text-primary);
}

.integration-name {
  font-size: 12px;
  color: var(--text-secondary);
  text-transform: capitalize;
}

.linked-actions {
  display: flex;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid var(--border-primary);
}

.section-label {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 12px;
}

.integration-list,
.project-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 300px;
  overflow-y: auto;
}

.integration-list::-webkit-scrollbar,
.project-list::-webkit-scrollbar {
  width: 4px;
}

.integration-list::-webkit-scrollbar-thumb,
.project-list::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--light-steel);
}

.integration-item,
.project-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: var(--small-radius);
  background: var(--surface-primary);
  cursor: pointer;
  transition: background 0.15s;
}

.integration-item:hover,
.project-item:hover {
  background: var(--surface-secondary);
}

.project-item {
  flex-direction: column;
  align-items: flex-start;
}

.project-item.selected {
  border: 1px solid var(--accent-primary);
  background: var(--accent-primary-subtle);
}

.project-desc {
  font-size: 12px;
  color: var(--text-secondary);
}

.integration-icon {
  width: 32px;
  height: 32px;
  object-fit: contain;
}

.selection-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-bottom: 16px;
  margin-bottom: 16px;
  border-bottom: 1px solid var(--border-primary);
  font-weight: 500;
}

.loading-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 32px;
  color: var(--text-secondary);
  text-align: center;
  gap: 12px;
}
</style>
