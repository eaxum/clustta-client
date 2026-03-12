<template>
  <div class="modal-container large-modal" v-esc="closeModal">
    <HeaderArea :title="'Task Type Templates'" :icon="'extension'" />

    <div class="general-container">
      <!-- Loading State -->
      <div v-if="isLoading" class="loading-state">
        <span>Loading...</span>
      </div>

      <!-- No Templates Warning -->
      <div v-else-if="templates.length === 0" class="empty-state">
        <img :src="getAppIcon('warning')" alt="" class="empty-icon" />
        <span class="empty-title">No Templates Found</span>
        <span class="empty-description">Create asset templates in your project before mapping task types.</span>
      </div>

      <div v-else class="mapping-content">
        <p class="section-description">
          Map each task type from Kitsu to an asset template. The template determines which file is created for each task.
        </p>

        <!-- Mapping Table -->
        <div class="mapping-table">
          <div class="table-header">
            <span class="col-task-type">Kitsu Task Type</span>
            <span class="col-template">Asset template</span>
          </div>

          <div class="table-body">
            <div v-for="taskType in externalTaskTypes" :key="taskType.id" class="mapping-row">
              <div class="col-task-type">
                <img :src="getTaskTypeIcon(taskType.id)" alt="" class="row-icon" />
                <span class="type-name">{{ taskType.name }}</span>
              </div>
              <div class="col-template">
                <DropDownBox :items="templateOptions" :selectedItem="getSelectedTemplateName(taskType.id)"
                  :onSelect="(val) => setMapping(taskType.id, val)" :placeHolder="'Select template'"
                  :useFilter="true" :fullWidth="true" />
              </div>
            </div>
          </div>
        </div>

        <!-- Unmapped Warning -->
        <div v-if="unmappedCount > 0" class="warning-banner">
          <img :src="getAppIcon('alert')" alt="" class="warning-icon" />
          <span>{{ unmappedCount }} task type{{ unmappedCount > 1 ? 's' : '' }} not mapped. Unmapped types won't create files during sync.</span>
        </div>
      </div>
    </div>

    <!-- Actions -->
    <div class="pop-up-actions">
      <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
      <GeneralButton :label="'Save'" :fullWidth="true" :buttonFunction="saveMapping" :isActive="true" :loading="isSaving" />
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useIntegrationStore } from '@/stores/integrations';
import { useNotificationStore } from '@/stores/notifications';
import { useTemplateStore } from '@/stores/template';

const { t } = useI18n();
const desktopModals = useDesktopModalStore();
const iconStore = useIconStore();
const integrationStore = useIntegrationStore();
const notificationStore = useNotificationStore();
const templateStore = useTemplateStore();

// refs
const isLoading = ref(false);
const isSaving = ref(false);
const mappings = ref({});

// computed
const externalTaskTypes = computed(() => {
  return integrationStore.externalTaskTypes || [];
});

const templates = computed(() => {
  return templateStore.getTemplates || [];
});

const templateOptions = computed(() => {
  return templates.value.map(t => ({
    id: t.id,
    name: `${t.name} (${t.extension})`,
  }));
});

const unmappedCount = computed(() => {
  return externalTaskTypes.value.filter(tt => !mappings.value[tt.id]).length;
});

// methods
// Closes the modal.
const closeModal = () => {
  desktopModals.setModalVisibility('taskTypeMappingModal', false);
};

// Returns the app icon path.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Gets the icon for a task type based on its mapped template extension.
const getTaskTypeIcon = (taskTypeId) => {
  const templateId = mappings.value[taskTypeId];
  if (!templateId) return getAppIcon('tag');
  const template = templates.value.find(t => t.id === templateId);
  if (!template?.extension) return getAppIcon('tag');
  const ext = template.extension.replace('.', '').toLowerCase();
  return `/file-icons/${ext}.svg`;
};

// Gets the currently selected template name for a task type.
const getSelectedTemplateName = (taskTypeId) => {
  const templateId = mappings.value[taskTypeId];
  if (!templateId) return null;
  const template = templates.value.find(t => t.id === templateId);
  return template ? `${template.name} (${template.extension})` : null;
};

// Sets a mapping between task type and template.
const setMapping = (taskTypeId, selectedName) => {
  // Find template by display name
  const template = templateOptions.value.find(t => t.name === selectedName);
  if (template) {
    mappings.value[taskTypeId] = template.id;
  } else {
    delete mappings.value[taskTypeId];
  }
};

// Saves the task type template mappings.
const saveMapping = async () => {
  isSaving.value = true;
  try {
    await integrationStore.saveTaskTypeTemplates(mappings.value);
    notificationStore.addNotification('Task type templates saved','', 'success');
    closeModal();
  } catch (error) {
    notificationStore.addNotification(error.message || 'Failed to save', '', 'error');
  } finally {
    isSaving.value = false;
  }
};

// lifecycle hooks
onMounted(async () => {
  isLoading.value = true;
  try {
    // Load templates
    await templateStore.reloadTemplates();

    // Ensure integration data is loaded
    await integrationStore.loadAvailableIntegrations();
    await integrationStore.loadLinkedIntegration();
    await integrationStore.loadTokens();

    // Load external task types if not already loaded
    if (integrationStore.externalTaskTypes.length === 0) {
      await integrationStore.getExternalTypes();
    }

    // Load existing mappings
    await integrationStore.loadTypeMappings();
    const existing = integrationStore.typeMappings?.task_type_templates || {};
    mappings.value = { ...existing };
  } catch (error) {
    console.error('Failed to load task type mappings:', error);
    notificationStore.addNotification(error.message || 'Failed to load task types', 'error');
  } finally {
    isLoading.value = false;
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.modal-container {
  max-height: 80vh;
  max-width: 500px;
}

.general-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1rem;
  overflow-y: auto;
  width: 500px;
  max-width: 500px;
}

.general-container::-webkit-scrollbar {
  width: 4px;
}

.general-container::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--bright-steel);
}

.general-container::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  color: var(--white);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 2rem;
  text-align: center;
}

.empty-icon {
  width: 48px;
  height: 48px;
  opacity: 0.5;
}

.empty-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--white);
}

.empty-description {
  font-size: 0.875rem;
  color: var(--bright-steel);
}

.mapping-content {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.section-description {
  font-size: 0.875rem;
  color: var(--bright-steel);
  margin: 0;
}

.mapping-table {
  display: flex;
  flex-direction: column;
  /* border: 1px solid var(--bright-steel); */
  outline: var(--transparent-line);
  outline-offset: -1px;
  border-radius: var(--large-radius);
  overflow: hidden;
}

.table-header {
  display: flex;
  padding: 0.75rem 1rem;
  background-color: var(--midnight-steel);
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--bright-steel);
  text-transform: uppercase;
}

.table-body {
  display: flex;
  flex-direction: column;
  max-height: 400px;
  overflow-y: auto;
}

.table-body::-webkit-scrollbar {
  width: 4px;
}

.table-body::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--bright-steel);
}

.mapping-row {
  display: flex;
  padding: 0.75rem 1rem;
  border-top: 1px solid var(--steel);
  align-items: center;
}

.mapping-row:hover {
  background-color: var(--steel);
}

.col-task-type {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.col-template {
  flex: 1;
}

.row-icon {
  width: 24px;
  height: 24px;
}

.type-name {
  font-size: 0.875rem;
  color: var(--white);
}

.warning-banner {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  background-color: rgba(255, 193, 7, 0.1);
  border: 1px solid var(--attention);
  border-radius: var(--small-radius);
  font-size: 0.875rem;
  color: var(--attention);
}

.warning-icon {
  width: 16px;
  height: 16px;
}

.pop-up-actions {
  display: flex;
  gap: 0.5rem;
  padding: 0.5rem 1rem 1rem 1rem;
  width: 700px;
  box-sizing: border-box;
}
</style>
