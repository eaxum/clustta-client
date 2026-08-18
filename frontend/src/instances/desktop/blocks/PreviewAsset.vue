<template>
  <div class="preview-asset">
    <div class="asset-spacer">
      <img v-if="fileIcon" class="small-icons" :src="fileIcon" v-tooltip="fileExtension">
    </div>

    <div class="preview-content">
      <span class="preview-name">{{ asset.name }}</span>
      <span v-if="fileExtension" class="preview-extension">{{ fileExtension }}</span>
    </div>

    <div class="preview-meta">
      <span v-for="taskType in taskTypes" :key="taskType" class="preview-type-label">{{ taskType }}</span>
    </div>
    <span class="action-badge" :class="`action-${asset.action}`">{{ actionLabel }}</span>
    <CheckBox :modelValue="isSelected" :disabled="!isActionable"
      :ariaLabel="`${isSelected ? 'Exclude' : 'Include'} ${asset.name}`"
      @update:modelValue="emit('toggle-selection')" />
  </div>
</template>

<script setup>
// imports
import { computed, ref, onMounted } from 'vue';
import utils from '@/services/utils';
import CheckBox from '@/instances/common/components/CheckBox.vue';

// stores
import { useIconStore } from '@/stores/icons';

const iconStore = useIconStore();

// props
const props = defineProps({
  isSelected: { type: Boolean, default: false },
  asset: { type: Object, required: true },
});

// emits
const emit = defineEmits(['toggle-selection']);

// refs
const fileIcon = ref('');

// computed
// Returns the file extension from the template (e.g., ".blend").
const fileExtension = computed(() => {
  return props.asset.template_extension || null;
});

// Returns the type icon path.
const typeIcon = computed(() => {
  const iconName = props.asset.asset_type_icon || 'generic';
  return iconStore.getAppIcon(iconName);
});

// Returns the capitalized external type name.
const typeName = computed(() => {
  return utils.capitalizeStr(props.asset.external_type || 'Asset');
});

const taskTypes = computed(() => {
  const types = props.asset.task_types?.length ? props.asset.task_types : [typeName.value];
  return [...new Set(types.filter(Boolean).map(type => utils.capitalizeStr(type)))];
});

const isActionable = computed(() => props.asset.action === 'create' || props.asset.action === 'link');

const actionLabel = computed(() => {
  const labels = {
    create: 'New',
    link: 'Link',
    skip: 'No action',
  };
  return labels[props.asset.action] || 'No action';
});

// methods
// Loads the file icon based on extension.
const loadFileIcon = async () => {
  if (fileExtension.value) {
    // Remove leading dot for icon lookup
    const ext = fileExtension.value.startsWith('.') 
      ? fileExtension.value.substring(1) 
      : fileExtension.value;
    fileIcon.value = await iconStore.getIcon(ext) || '';
  }
};

// lifecycle
onMounted(() => {
  loadFileIcon();
});
</script>

<style scoped>
.preview-asset {
  display: flex;
  gap: 0.2rem;
  color: var(--text);
  align-items: center;
  padding: 0 0.5rem;
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  border-radius: var(--large-radius);
  background-color: var(--surface-2);
  outline: var(--transparent-line);
  outline-offset: -1px;
  transition: all 0.2s ease-out;
}

.preview-asset:hover {
  background-color: var(--surface-3);
  border-radius: var(--small-radius);
  outline: 1px solid var(--surface-4);
}

.asset-spacer {
  display: flex;
  align-items: center;
  justify-content: center;
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
  color: var(--surface-5);
  margin-left: 0.25rem;
  flex-shrink: 0;
}

.preview-meta {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
  margin-right: 0.25rem;
}

.preview-type-label {
  padding: 0.1rem 0.4rem;
  font-size: 11px;
  font-weight: 500;
  color: var(--text-tertiary);
  background-color: var(--surface-1);
  border-radius: var(--tiny-radius);
  text-transform: capitalize;
}

.action-badge {
  flex: 0 0 auto;
  padding: 2px 6px;
  border-radius: var(--tiny-radius);
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
}

.action-create {
  color: #4ade80;
  background-color: rgba(34, 197, 94, 0.15);
}

.action-link {
  color: #60a5fa;
  background-color: rgba(59, 130, 246, 0.15);
}

.action-skip {
  color: var(--text-muted);
  background-color: var(--surface-3);
}
</style>
