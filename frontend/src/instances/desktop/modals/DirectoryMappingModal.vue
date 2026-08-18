<template>
  <div class="modal-container large-modal" v-esc="closeModal">
    <HeaderArea :title="'Directory Mapping'" :icon="'file-path'" />

    <div class="general-container">
      <!-- Loading State -->
      <div v-if="isLoading" class="loading-state">
        <span>Loading...</span>
      </div>

      <div v-else class="mapping-content">
        <!-- Preset and Style Row -->
        <div class="config-row">
          <div class="config-item">
            <span class="config-label">Preset</span>
            <DropDownBox :items="presetOptions" :selectedItem="selectedPreset" :onSelect="applyPreset"
              :placeHolder="'Select preset'" :useFilter="false" :fullWidth="true" />
          </div>

          <div class="config-item">
            <span class="config-label">Naming Style</span>
            <DropDownBox :items="styleOptions" :selectedItem="namingStyle" :onSelect="(val) => namingStyle = val"
              :placeHolder="'Select style'" :useFilter="false" :fullWidth="true" />
          </div>
        </div>

        <!-- Path Templates -->
        <div class="templates-section">
          <h3 class="section-title">Path Templates</h3>

          <!-- Dynamic Templates -->
          <div v-for="template in customTemplates" :key="template.id" class="template-item">
            <div class="template-header">
              <input type="text" class="template-name-input" v-model="template.name" placeholder="Template name" />
              <ActionButton :icon="getAppIcon('minus-circle')" v-tooltip="'Remove template'" :buttonFunction="() => removeTemplate(template.id)" />
            </div>
            <input :ref="(el) => setTemplateInputRef(template.id, el)" type="text" class="template-input"
              v-model="template.template" @focus="activeTemplateId = template.id"
              placeholder="<CollectionType>/<Asset>/<AssetType><TemplateExtension>" />
            <div class="template-preview">
              <span class="preview-label">Preview:</span>
              <span class="preview-path">{{ getTemplatePreview(template) }}</span>
            </div>

            <div class="task-output-section">
              <button type="button" class="task-output-toggle" @click="toggleTaskOutputs(template.id)">
                <img :src="getAppIcon(isTaskOutputExpanded(template.id) ? 'collapse-up' : 'collapse-down')" alt="" />
                <span>Task outputs</span>
                <span class="task-output-count">{{ configuredTaskOutputCount(template) }}/{{ getTemplateTaskTypes(template).length }}</span>
              </button>

              <div v-if="isTaskOutputExpanded(template.id)" class="task-output-rows">
                <div v-if="externalTaskTypes.length === 0" class="task-output-empty">
                  Connect to the integration to load task types.
                </div>
                <div v-else-if="getTemplateTaskTypes(template).length === 0" class="task-output-empty">
                  No task types match this template's task set.
                </div>
                <div v-for="taskType in getTemplateTaskTypes(template)" :key="`${template.id}-${taskType.id || taskType.name}`" class="task-output-row">
                  <CheckBox
                    :modelValue="isTaskOutputEnabled(template, taskType.name)"
                    :ariaLabel="`${isTaskOutputEnabled(template, taskType.name) ? 'Exclude' : 'Include'} ${taskType.name}`"
                    @update:modelValue="setTaskOutputEnabled(template, taskType.name, $event)"
                  />
                  <span class="task-output-name">{{ taskType.name }}</span>
                  <span class="task-output-arrow">-&gt;</span>
                  <input
                    type="text"
                    class="task-output-input"
                    v-model="template.task_outputs[taskType.name]"
                    :placeholder="defaultTaskOutput(taskType.name)"
                    :disabled="!isTaskOutputEnabled(template, taskType.name)"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- Add Template Button -->
          <ActionButton :icon="getAppIcon('plus-circle')" :label="'Add Template'" :showLabel="true" :useOutline="true"
            :buttonFunction="addTemplate" />
        </div>

        <!-- Placeholders -->
        <div class="placeholders-section">
          <h3 class="section-title">Available Placeholders</h3>
          <p class="section-description">Click to insert at cursor position</p>
          <div class="placeholders-grid">
            <Chip v-for="placeholder in placeholders" :key="placeholder.key" :icon="getAppIcon(placeholder.icon)"
              :label="placeholder.label" :readonly="true" @click="insertPlaceholder(placeholder.key)" />
          </div>
        </div>
      </div>
    </div>

    <!-- Actions (outside scrollable container) -->
    <div class="pop-up-actions">
      <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
      <GeneralButton :label="'Save'" :fullWidth="true" :buttonFunction="saveMapping" :isActive="isDirty" :loading="isSaving" />
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import CheckBox from '@/instances/common/components/CheckBox.vue';
import Chip from '@/instances/common/components/Chip.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// stores
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useIconStore } from '@/stores/icons';
import { useIntegrationStore } from '@/stores/integrations';
import { useNotificationStore } from '@/stores/notifications';

const { t } = useI18n();
const desktopModals = useDesktopModalStore();
const iconStore = useIconStore();
const integrationStore = useIntegrationStore();
const notificationStore = useNotificationStore();

// refs
const activeTemplateId = ref(null);
const customTemplates = ref([
  { id: 'asset', name: 'Assets', icon: 'package', template: 'Assets/<CollectionType>/<Asset>/<OutputName><TemplateExtension>', task_outputs: {}, task_output_enabled: {} },
  { id: 'shot', name: 'Shots', icon: 'clapperboard', template: 'Episodes/<Episode>/<Sequence>/<Shot>/<OutputName><TemplateExtension>', task_outputs: {}, task_output_enabled: {} },
]);
const expandedTaskOutputs = ref({});
const isLoading = ref(false);
const isSaving = ref(false);
const namingStyle = ref('lowercase');
const originalSnapshot = ref('');
const selectedPreset = ref('3d-animation');
const templateInputRefs = ref({});

// computed

// Serializes the editable state so it can be compared for changes.
const serializeState = () => {
  return JSON.stringify({
    preset: selectedPreset.value,
    style: namingStyle.value,
    templates: customTemplates.value.map(t => ({
      id: t.id,
      name: t.name,
      icon: t.icon,
      template: t.template,
      task_outputs: t.task_outputs || {},
      task_output_enabled: t.task_output_enabled || {},
    })),
  });
};

// Returns whether the editable state differs from the originally loaded state.
const isDirty = computed(() => {
  return serializeState() !== originalSnapshot.value;
});

const placeholders = computed(() => [
  { key: '<Episode>', label: 'Episode', icon: 'film-reel' },
  { key: '<Sequence>', label: 'Sequence', icon: 'film-strip' },
  { key: '<Shot>', label: 'Shot', icon: 'clapboard' },
  { key: '<Asset>', label: 'Asset', icon: 'cube' },
  { key: '<CollectionType>', label: 'CollectionType', icon: 'folder' },
  { key: '<AssetType>', label: 'AssetType', icon: 'file' },
  { key: '<OutputName>', label: 'OutputName', icon: 'file-name' },
  { key: '<TemplateExtension>', label: 'TemplateExtension', icon: 'extension' },
]);

const presetOptions = computed(() => ['3d-animation', 'custom']);

const styleOptions = computed(() => ['lowercase', 'uppercase', 'capitalize', 'kebab-case']);

const externalTaskTypes = computed(() => {
  return (integrationStore.externalAssetTypes || [])
    .map(type => ({
      id: type?.id || type?.ID || '',
      name: type?.name || type?.Name || type,
      for_entity: type?.for_entity || type?.ForEntity || '',
    }))
    .filter(type => type.name);
});

// methods
// Adds a new blank template.
const addTemplate = () => {
  const id = `template_${Date.now()}`;
  customTemplates.value.push({
    id,
    name: 'New Template',
    icon: 'folder',
    template: '<CollectionType>/<Asset>/<OutputName><TemplateExtension>',
    task_outputs: {},
    task_output_enabled: {},
  });
};

// Applies a preset template configuration.
const applyPreset = (preset) => {
  selectedPreset.value = preset;
  if (preset === '3d-animation') {
    // Reset to default templates
    customTemplates.value = [
      { id: 'asset', name: 'Assets', icon: 'package', template: 'Assets/<CollectionType>/<Asset>/<OutputName><TemplateExtension>', task_outputs: {}, task_output_enabled: {} },
      { id: 'shot', name: 'Shots', icon: 'clapperboard', template: 'Episodes/<Episode>/<Sequence>/<Shot>/<OutputName><TemplateExtension>', task_outputs: {}, task_output_enabled: {} },
    ];
  }
};

// Closes the modal.
const closeModal = () => {
  desktopModals.setModalVisibility('directoryMappingModal', false);
};

// Returns the app icon path.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Returns preview for a template.
const getTemplatePreview = (template) => {
  return resolvePath(template.template, {
    Episode: 'ep01',
    Sequence: 'seq001',
    Shot: 'shot010',
    CollectionType: 'character',
    Asset: 'hero',
    AssetType: 'modeling',
    OutputName: 'hero-modeling',
    TemplateExtension: '.ext',
  });
};

const configuredTaskOutputCount = (template) => {
  return getTemplateTaskTypes(template)
    .filter(taskType => isTaskOutputEnabled(template, taskType.name))
    .length;
};

const defaultTaskOutput = (taskType) => {
  return `<Asset>-${String(taskType || '').toLowerCase()}`;
};

const isTaskOutputEnabled = (template, taskType) => {
  return template.task_output_enabled?.[taskType] !== false;
};

const setTaskOutputEnabled = (template, taskType, enabled) => {
  template.task_output_enabled = {
    ...(template.task_output_enabled || {}),
    [taskType]: enabled,
  };
};

const isTaskOutputExpanded = (templateId) => {
  return !!expandedTaskOutputs.value[templateId];
};

const toggleTaskOutputs = (templateId) => {
  expandedTaskOutputs.value = {
    ...expandedTaskOutputs.value,
    [templateId]: !expandedTaskOutputs.value[templateId],
  };
};

const inferTemplateEntity = (id, template = '') => {
  const value = `${id || ''} ${template || ''}`.toLowerCase();
  if (value.includes('<shot>') || value.includes('shot')) return 'shot';
  if (value.includes('<asset>') || value.includes('asset')) return 'asset';
  return '';
};

const getTemplateTaskTypes = (template) => {
	const entity = inferTemplateEntity(template.id, template.template);
	if (!entity) return externalTaskTypes.value;
	if (!externalTaskTypes.value.some(taskType => taskType.for_entity)) return externalTaskTypes.value;

	return externalTaskTypes.value.filter(taskType => {
		const forEntity = String(taskType.for_entity || '').toLowerCase();
    return forEntity === entity;
  });
};

// Inserts a placeholder at cursor position in the active input.
const insertPlaceholder = (placeholder) => {
  if (!activeTemplateId.value) return;

  const inputRef = templateInputRefs.value[activeTemplateId.value];
  const template = customTemplates.value.find(t => t.id === activeTemplateId.value);

  if (inputRef && template) {
    const start = inputRef.selectionStart;
    const end = inputRef.selectionEnd;
    const value = template.template;
    template.template = value.substring(0, start) + placeholder + value.substring(end);

    // Restore focus and cursor position
    setTimeout(() => {
      inputRef.focus();
      inputRef.setSelectionRange(start + placeholder.length, start + placeholder.length);
    }, 0);
  }
};

// Removes a custom template.
const removeTemplate = (id) => {
  customTemplates.value = customTemplates.value.filter(t => t.id !== id);
};

// Resolves a path template with sample values.
const resolvePath = (template, values) => {
  let result = template;
  for (const [key, value] of Object.entries(values)) {
    const placeholder = `<${key}>`;
    let resolvedValue = value;

    // Apply naming style
    switch (namingStyle.value) {
      case 'lowercase':
        resolvedValue = value.toLowerCase();
        break;
      case 'uppercase':
        resolvedValue = value.toUpperCase();
        break;
      case 'capitalize':
        resolvedValue = value.charAt(0).toUpperCase() + value.slice(1).toLowerCase();
        break;
      case 'kebab-case':
        resolvedValue = value.toLowerCase().replace(/\s+/g, '-');
        break;
    }

    result = result.replace(new RegExp(placeholder, 'g'), resolvedValue);
  }
  return result;
};

// Sets the input ref for a template.
const setTemplateInputRef = (id, el) => {
  if (el) {
    templateInputRefs.value[id] = el;
  }
};

// Saves the directory mapping configuration.
const saveMapping = async () => {
  isSaving.value = true;
  try {
    // Build paths object from custom templates
    const paths = {};
    customTemplates.value.forEach(t => {
      paths[t.id] = {
        name: t.name,
        icon: t.icon,
        template: t.template,
        task_outputs: t.task_outputs || {},
        task_output_enabled: t.task_output_enabled || {},
      };
    });

    await integrationStore.saveDirectoryStructure({
      preset: selectedPreset.value,
      style: namingStyle.value,
      paths,
    });
    originalSnapshot.value = serializeState();
    notificationStore.addNotification('Directory mapping saved','', 'success');
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
    await integrationStore.loadTypeMappings();
    try {
      await integrationStore.getExternalTypes();
    } catch (error) {
      console.warn('Failed to load external task types:', error);
    }
    const structure = integrationStore.typeMappings?.directory_structure;
    if (structure) {
      selectedPreset.value = structure.preset || '3d-animation';
      namingStyle.value = structure.style || 'lowercase';
      
      // Load custom templates from paths
      if (structure.paths && Object.keys(structure.paths).length > 0) {
        customTemplates.value = Object.entries(structure.paths).map(([id, data]) => ({
          id,
          name: data.name || id,
          icon: data.icon || 'folder',
          template: data.template || '',
          task_outputs: data.task_outputs || {},
          task_output_enabled: data.task_output_enabled || {},
        }));
      }
    }
    originalSnapshot.value = serializeState();
  } catch (error) {
    console.error('Failed to load directory structure:', error);
  } finally {
    isLoading.value = false;
  }
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.modal-container {
  max-height: 85vh;
  max-width: 800px;
}

.general-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1rem;
  overflow-y: auto;
  width: 800px;
  max-width: 800px;
}

.general-container::-webkit-scrollbar {
  width: 4px;
}

.general-container::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--surface-inverse);
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

.mapping-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.config-row {
  display: flex;
  gap: 1rem;
}

.config-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.config-label {
  font-size: 0.875rem;
  color: var(--text);
}

.templates-section,
.placeholders-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.section-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text);
  margin: 0;
}

.section-description {
  font-size: 0.75rem;
  color: var(--text);
  margin: 0;
}

.template-item {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 0.75rem;
  background-color: var(--surface-3);
  border-radius: var(--small-radius);
}

.template-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.template-icon {
  width: 16px;
  height: 16px;
}

.template-name-input {
  flex: 1;
  padding: 0.25rem 0.5rem;
  background-color: transparent;
  border: 1px solid transparent;
  border-radius: var(--small-radius);
  color: var(--text);
  font-size: 0.875rem;
  font-weight: 500;
}

.template-name-input:hover {
  background-color: var(--bg);
}

.template-name-input:focus {
  outline: none;
  border-color: var(--border-strong);
  background-color: var(--bg);
}

.template-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text);
}

.template-input {
  width: 100%;
  padding: 0.5rem 0.75rem;
  background-color: var(--bg);
  border: 1px solid var(--border-strong);
  border-radius: var(--small-radius);
  color: var(--text);
  font-family: monospace;
  font-size: 0.875rem;
  box-sizing: border-box;
}

.template-input:focus {
  outline: none;
  border-color: var(--border-strong);
}

.template-preview {
  display: flex;
  gap: 0.5rem;
  font-size: 0.75rem;
}

.preview-label {
  color: var(--text);
}

.preview-path {
  color: var(--text);
  font-family: monospace;
}

.task-output-section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.task-output-toggle {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.35rem 0;
  border: 0;
  background: transparent;
  color: var(--text);
  cursor: pointer;
  font-size: 0.8rem;
  text-align: left;
}

.task-output-toggle img {
  width: 14px;
  height: 14px;
}

.task-output-count {
  margin-left: auto;
  color: var(--text);
  opacity: 0.6;
  font-size: 0.75rem;
}

.task-output-rows {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.35rem;
  padding: 0.5rem;
  background-color: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--small-radius);
}

.task-output-row {
  display: grid;
  grid-template-columns: 24px minmax(120px, 1fr) auto minmax(220px, 2fr);
  align-items: center;
  gap: 0.5rem;
}

.task-output-name,
.task-output-arrow,
.task-output-empty {
  color: var(--text);
  font-size: 0.75rem;
}

.task-output-arrow {
  opacity: 0.6;
}

.task-output-input {
  min-width: 0;
  padding: 0.35rem 0.5rem;
  background-color: var(--surface-3);
  border: 1px solid var(--border-strong);
  border-radius: var(--small-radius);
  color: var(--text);
  font-family: monospace;
  font-size: 0.75rem;
}

.task-output-input:focus {
  outline: none;
  border-color: var(--border-strong);
}

.task-output-input:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.placeholders-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.placeholders-grid .chip {
  cursor: pointer;
}

.pop-up-actions {
  display: flex;
  gap: 0.5rem;
  padding: 0.5rem 1rem 1rem 1rem;
  width: 800px;
  box-sizing: border-box;
}
</style>
