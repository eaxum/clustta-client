<template>
  <div class="preview-asset" :class="{ 'preview-asset-selected': isSelected }">
    <div class="preview-checkbox" @click.stop="toggleSelection">
      <img class="tiny-icons" :src="checkboxIcon">
    </div>

    <div class="asset-spacer" ></div>

    <div class="preview-type-icon">
      <img class="small-icons" :src="typeIcon" v-tooltip="typeName">
    </div>

    <div class="preview-content" @click="console.log(task)" >
      <span class="preview-name">{{ task.name }}</span>
      <span v-if="templateExtension" class="preview-extension">{{ templateExtension }}</span>
    </div>

    <div class="preview-meta">
      <span class="preview-type-label">{{ typeName }}</span>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed } from 'vue';
import utils from '@/services/utils';

// stores
import { useIconStore } from '@/stores/icons';
import { useIntegrationStore } from '@/stores/integrations';
import { useTemplateStore } from '@/stores/template';

const iconStore = useIconStore();
const integrationStore = useIntegrationStore();
const templateStore = useTemplateStore();

// props
const props = defineProps({
  isSelected: { type: Boolean, default: false },
  task: { type: Object, required: true },
});

// emits
const emit = defineEmits(['toggle-selection']);

// computed
// Returns the checkbox icon based on selection state.
const checkboxIcon = computed(() => {
  return props.isSelected 
    ? iconStore.getAppIcon('checkbox-selected') 
    : iconStore.getAppIcon('checkbox-unselected');
});

// Returns the template extension for this task type.
const templateExtension = computed(() => {
  const taskTypeTemplates = integrationStore.typeMappings?.task_type_templates || {};
  const templateId = taskTypeTemplates[props.task.external_type_id];
  if (!templateId) return null;
  
  const template = templateStore.templates.find(t => t.id === templateId);
  return template?.extension || null;
});

// Returns the type icon path.
const typeIcon = computed(() => {
  const iconName = props.task.task_type_icon || 'generic';
  return iconStore.getAppIcon(iconName);
});

// Returns the capitalized external type name.
const typeName = computed(() => {
  return utils.capitalizeStr(props.task.external_type || 'Task');
});

// methods
// Toggles the selection state.
const toggleSelection = () => {
  emit('toggle-selection', props.task.id);
};
</script>

<style scoped>
.preview-asset {
  display: flex;
  gap: 0.2rem;
  color: var(--white);
  align-items: center;
  padding: 0 0.5rem;
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  border-radius: var(--large-radius);
  background-color: var(--dark-steel);
  outline: var(--transparent-line);
  outline-offset: -1px;
  transition: all 0.2s ease-out;
}

.preview-asset:hover {
  background-color: var(--steel);
  border-radius: var(--small-radius);
  outline: 1px solid var(--light-steel);
}

.preview-asset-selected {
  background-color: var(--blue-steel);
}

.preview-asset-selected:hover {
  background-color: var(--solid-blue-steel);
}

.preview-checkbox {
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  padding: 0.2rem;
  border-radius: var(--tiny-radius);
}

.preview-checkbox:hover {
  background-color: var(--light-steel);
}

.asset-spacer {
  min-width: 24px;
}

.preview-type-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: min-content;
}

.preview-content {
  display: flex;
  align-items: center;
  flex: 1;
  overflow: hidden;
}

.preview-name {
  font-size: 13px;
  font-weight: 400;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.preview-extension {
  font-size: 11px;
  font-weight: 400;
  color: var(--bright-steel);
  margin-left: 0.25rem;
  flex-shrink: 0;
}

.preview-meta {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.1rem 0.4rem;
  background-color: var(--black-steel);
  border-radius: var(--tiny-radius);
  margin-right: 0.25rem;
}

.preview-type-label {
  font-size: 11px;
  font-weight: 500;
  color: var(--text-tertiary);
  text-transform: capitalize;
}
</style>