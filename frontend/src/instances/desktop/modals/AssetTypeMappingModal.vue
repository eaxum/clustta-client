<template>
  <div class="modal-container large-modal" v-esc="closeModal">
    <HeaderArea :title="'Asset Type Mapping'" :icon="'extension'" />

    <div class="general-container">
      <!-- Loading State -->
      <div v-if="isLoading" class="loading-state">
        <span>Loading...</span>
      </div>

      <!-- No Templates Warning -->
      <div v-else-if="!templates.length" class="page-state-container">
        <PageState :message="'No templates found. Create asset templates in your project before mapping asset types'" :illustration="'/page-states/template.png'" />
      </div>

      <div v-else class="mapping-content">
        <div class="section-header">
          <p class="section-description">
            Map each asset type from Kitsu to an asset template. The template determines which file is created for each asset.
          </p>
          <ActionButton :icon="getAppIcon('sparkles')" :label="'Auto'" :buttonFunction="autoAssign" :showLabel="true" :useBackground="true" />
        </div>

        <!-- Mapping Table -->
        <div class="mapping-table">
          <div class="table-header">
            <span class="col-asset-type">Kitsu Asset Type</span>
            <span class="col-template">Asset template</span>
          </div>

          <div class="table-body">
            <div v-for="assetType in externalAssetTypes" :key="assetType.id" class="mapping-row">
              <div class="col-asset-type">
                <img :src="getAssetTypeIcon(assetType.id)" alt="" class="row-icon small-icons" :class="{ 'no-filter': mappings[assetType.id] }" />
                <span class="type-name">{{ assetType.name }}</span>
              </div>
              <div class="col-template">
                <DropDownBox :items="templateOptions" :selectedItem="getSelectedTemplateName(assetType.id)"
                  :onSelect="(val) => setMapping(assetType.id, val)" :placeHolder="'Select template'"
                  :useFilter="true" :fullWidth="true" />
              </div>
            </div>
          </div>
        </div>

        <!-- Unmapped Warning -->
        <div v-if="unmappedCount > 0" class="warning-banner">
          <img :src="getAppIcon('alert')" alt="" class="warning-icon" />
          <span>{{ unmappedCount }} asset type{{ unmappedCount > 1 ? 's' : '' }} not mapped. Unmapped types won't create files during sync.</span>
        </div>
        
      </div>
    </div>

    <!-- Actions -->
    <div class="pop-up-actions">
      <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
      <GeneralButton :label="'Save'" :fullWidth="true" :buttonFunction="saveMapping" :isActive="isDirty && templates.length > 0" :loading="isSaving" />
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import PageState from '@/instances/common/components/PageState.vue';

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
const originalMappings = ref({});

// computed
const externalAssetTypes = computed(() => {
  return integrationStore.externalAssetTypes || [];
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

const isDirty = computed(() => {
  const current = mappings.value;
  const original = originalMappings.value;
  const currentKeys = Object.keys(current);
  const originalKeys = Object.keys(original);
  if (currentKeys.length !== originalKeys.length) return true;
  return currentKeys.some(key => current[key] !== original[key]);
});

const unmappedCount = computed(() => {
  return externalAssetTypes.value.filter(tt => !mappings.value[tt.id]).length;
});

// methods
// Software keywords associated with common creative task types.
const SOFTWARE_HINTS = {
  modelling: ['blender', 'maya', 'zbrush', '3ds', 'max', 'modo', 'houdini', 'cinema'],
  modeling: ['blender', 'maya', 'zbrush', '3ds', 'max', 'modo', 'houdini', 'cinema'],
  rigging: ['blender', 'maya', 'houdini'],
  animation: ['blender', 'maya', 'harmony', 'toon boom', 'animate', 'houdini', 'ase', 'spine'],
  layout: ['blender', 'maya', 'houdini', 'cinema'],
  lighting: ['blender', 'maya', 'houdini', 'cinema', 'katana'],
  rendering: ['blender', 'maya', 'houdini', 'cinema', 'katana'],
  compositing: ['nuke', 'after effects', 'fusion', 'natron'],
  'fx': ['houdini', 'blender', 'maya', 'embergen'],
  'effects': ['houdini', 'blender', 'maya', 'embergen'],
  concept: ['photoshop', 'clip studio', 'krita', 'procreate', 'sai', 'sketchbook'],
  texture: ['substance', 'photoshop', 'quixel', 'mari', 'krita'],
  lookdev: ['substance', 'blender', 'maya', 'katana'],
  storyboard: ['storyboard pro', 'photoshop', 'clip studio', 'krita', 'toon boom'],
  previz: ['blender', 'maya', 'unreal', 'unity'],
  edit: ['premiere', 'davinci', 'resolve', 'final cut', 'avid'],
  editing: ['premiere', 'davinci', 'resolve', 'final cut', 'avid'],
  sound: ['audition', 'audacity', 'pro tools', 'reaper', 'logic'],
  audio: ['audition', 'audacity', 'pro tools', 'reaper', 'logic'],
};

// Attempts to auto-assign templates to external asset types by matching names.
const autoAssign = () => {
  for (const assetType of externalAssetTypes.value) {
    if (mappings.value[assetType.id]) continue;

    const externalName = assetType.name.toLowerCase().trim();

    // Pass 1: exact or substring match on template name
    const nameMatch = templates.value.find(t =>
      t.name.toLowerCase().includes(externalName) || externalName.includes(t.name.toLowerCase())
    );
    if (nameMatch) {
      mappings.value[assetType.id] = nameMatch.id;
      continue;
    }

    // Pass 2: match using software hints for the task type
    const hintKey = Object.keys(SOFTWARE_HINTS).find(key => externalName.includes(key));
    if (hintKey) {
      const softwareKeywords = SOFTWARE_HINTS[hintKey];
      const softwareMatch = templates.value.find(t => {
        const tName = t.name.toLowerCase();
        const tExt = (t.extension || '').toLowerCase();
        return softwareKeywords.some(sw => tName.includes(sw) || tExt.includes(sw));
      });
      if (softwareMatch) {
        mappings.value[assetType.id] = softwareMatch.id;
        continue;
      }
    }
  }
};

// Closes the modal.
const closeModal = () => {
  desktopModals.setModalVisibility('assetTypeMappingModal', false);
};

// Returns the app icon path.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Gets the icon for a asset type based on its mapped template extension.
const getAssetTypeIcon = (assetTypeId) => {
  const templateId = mappings.value[assetTypeId];
  if (!templateId) return getAppIcon('file');
  const template = templates.value.find(t => t.id === templateId);
  if (!template?.extension) return getAppIcon('file');
  const ext = template.extension.replace('.', '').toLowerCase();
  return `/file-icons/${ext}.svg`;
};

// Gets the currently selected template name for a asset type.
const getSelectedTemplateName = (assetTypeId) => {
  const templateId = mappings.value[assetTypeId];
  if (!templateId) return null;
  const template = templates.value.find(t => t.id === templateId);
  return template ? `${template.name} (${template.extension})` : null;
};

// Sets a mapping between asset type and template.
const setMapping = (assetTypeId, selectedName) => {
  // Find template by display name
  const template = templateOptions.value.find(t => t.name === selectedName);
  if (template) {
    mappings.value[assetTypeId] = template.id;
  } else {
    delete mappings.value[assetTypeId];
  }
};

// Saves the asset type template mappings.
const saveMapping = async () => {
  isSaving.value = true;
  try {
    await integrationStore.saveAssetTypeTemplates(mappings.value);
    originalMappings.value = { ...mappings.value };
    notificationStore.addNotification('Asset type templates saved','', 'success');
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

    // Load external asset types if not already loaded
    if (integrationStore.externalAssetTypes.length === 0) {
      await integrationStore.getExternalTypes();
    }

    // Load existing mappings
    await integrationStore.loadTypeMappings();
    const existing = integrationStore.typeMappings?.asset_type_templates || {};
    mappings.value = { ...existing };
    originalMappings.value = { ...existing };
  } catch (error) {
    console.error('Failed to load asset type mappings:', error);
    notificationStore.addNotification(error.message || 'Failed to load asset types', 'error');
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
  overflow: hidden;
  width: 500px;
  max-width: 500px;
}

.general-container::-webkit-scrollbar {
  width: 4px;
}

.general-container::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--surface-5);
}

.general-container::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  color: var(--text);
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
  color: var(--text);
}

.empty-description {
  font-size: 0.875rem;
  color: var(--surface-5);
}

.mapping-content {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.section-description {
  font-size: 0.875rem;
  margin: 0;
  color: var(--text);
}

.mapping-table {
  display: flex;
  flex-direction: column;
  outline: var(--transparent-line);
  outline-offset: -1px;
  border-radius: var(--large-radius);
  overflow: hidden;
}

.table-header {
  display: flex;
  padding: 0.75rem 1rem;
  background-color: var(--bg);
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--surface-5);
  text-transform: uppercase;
}

.table-body {
  display: flex;
  flex-direction: column;
  max-height: 300px;
  overflow-y: auto;
}

.table-body::-webkit-scrollbar {
  width: 4px;
}

.table-body::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--surface-5);
}

.mapping-row {
  display: flex;
  padding: 0.75rem 1rem;
  border-top: 1px solid var(--surface-3);
  align-items: center;
}

.mapping-row:hover {
  background-color: var(--surface-3);
}

.col-asset-type {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--text);
}

.col-template {
  flex: 1;
  color: var(--text);
}

.row-icon {
  width: 24px;
  height: 24px;
}

.type-name {
  font-size: 0.875rem;
  color: var(--text);
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
