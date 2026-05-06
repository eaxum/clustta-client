<template>
  <div class="preview-asset" :class="{ 'preview-asset-selected': isSelected }">
    <div class="asset-spacer">
      <img v-if="fileIcon" class="small-icons" :src="fileIcon" v-tooltip="fileExtension">
    </div>

    <div class="preview-type-icon">
      <img class="small-icons" :src="typeIcon" v-tooltip="typeName">
    </div>

    <div class="preview-content" @click="console.log(asset)" >
      <span class="preview-name">{{ asset.name }}</span>
      <span v-if="fileExtension" class="preview-extension">{{ fileExtension }}</span>
    </div>

    <div class="preview-meta">
      <span class="preview-type-label">{{ typeName }}</span>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed, ref, onMounted } from 'vue';
import utils from '@/services/utils';

// stores
import { useIconStore } from '@/stores/icons';

const iconStore = useIconStore();

// props
const props = defineProps({
  isSelected: { type: Boolean, default: false },
  asset: { type: Object, required: true },
});

// emits
const emit = defineEmits([]);

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
  return iconStore.resolveIcon(iconName);
});

// Returns the capitalized external type name.
const typeName = computed(() => {
  return utils.capitalizeStr(props.asset.external_type || 'Asset');
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