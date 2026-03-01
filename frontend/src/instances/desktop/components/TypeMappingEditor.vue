<template>
  <div class="type-mapping-editor">
    <div class="mapping-header">
      <img :src="getAppIcon('layers')" alt="" class="header-icon" />
      <span class="header-title">Type Mappings</span>
      <span class="header-subtitle">Map {{ integrationName }} types to Clustta types</span>
    </div>

    <div class="mapping-sections">
      <div class="mapping-section">
        <div class="section-header">
          <span class="section-title">Entity Types</span>
          <span class="section-count">{{ entityMappings.length }}</span>
        </div>

        <div class="mapping-rows">
          <div v-for="mapping in entityMappings" :key="mapping.external_type" class="mapping-row">
            <div class="external-type">
              <span class="type-name">{{ mapping.external_name || mapping.external_type }}</span>
            </div>

            <img :src="getAppIcon('arrow-right')" alt="" class="arrow-icon" />

            <div class="local-type">
              <DropDownBox :items="localEntityTypes" :selectedItem="mapping.local_type" :onSelect="(value) => updateEntityMapping(mapping.external_type, value)"
                :placeHolder="'Select type'" :useFilter="false" :fullWidth="true" />
            </div>
          </div>
        </div>
      </div>

      <div class="mapping-section">
        <div class="section-header">
          <span class="section-title">Task Types</span>
          <span class="section-count">{{ taskMappings.length }}</span>
        </div>

        <div class="mapping-rows">
          <div v-for="mapping in taskMappings" :key="mapping.external_type" class="mapping-row">
            <div class="external-type">
              <span class="type-name">{{ mapping.external_name || mapping.external_type }}</span>
            </div>

            <img :src="getAppIcon('arrow-right')" alt="" class="arrow-icon" />

            <div class="local-type">
              <DropDownBox :items="localTaskTypes" :selectedItem="mapping.local_type" :onSelect="(value) => updateTaskMapping(mapping.external_type, value)"
                :placeHolder="'Select type'" :useFilter="false" :fullWidth="true" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed } from 'vue';

// components
import DropDownBox from '@/instances/common/components/DropDownBox.vue';

// stores
import { useIconStore } from '@/stores/icons';

const iconStore = useIconStore();

// props
const props = defineProps({
  entityMappings: { type: Array, default: () => [] },
  integrationName: { type: String, default: 'External' },
  localEntityTypes: { type: Array, default: () => [] },
  localTaskTypes: { type: Array, default: () => [] },
  taskMappings: { type: Array, default: () => [] },
});

// emits
const emit = defineEmits(['update:entityMappings', 'update:taskMappings']);

// methods
// Returns the app icon path.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Updates an entity type mapping.
const updateEntityMapping = (externalType, localType) => {
  const updated = props.entityMappings.map(mapping => {
    if (mapping.external_type === externalType) {
      return { ...mapping, local_type: localType };
    }
    return mapping;
  });
  emit('update:entityMappings', updated);
};

// Updates a task type mapping.
const updateTaskMapping = (externalType, localType) => {
  const updated = props.taskMappings.map(mapping => {
    if (mapping.external_type === externalType) {
      return { ...mapping, local_type: localType };
    }
    return mapping;
  });
  emit('update:taskMappings', updated);
};
</script>

<style scoped>
.type-mapping-editor {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 16px;
  background-color: var(--background-secondary);
  border-radius: var(--small-radius);
  background-color: royalblue;
  box-sizing: border-box;
}

.mapping-header {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.header-icon {
  width: 20px;
  height: 20px;
  opacity: 0.7;
}

.header-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.header-subtitle {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-left: auto;
}

.mapping-sections {
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow: hidden;
  overflow-y: scroll;
  box-sizing: border-box;
}

.mapping-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-title {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.section-count {
  font-size: 11px;
  color: var(--text-tertiary);
  background-color: var(--background-tertiary);
  padding: 2px 6px;
  border-radius: 10px;
}

.mapping-rows {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.mapping-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  background-color: var(--background-primary);
  border-radius: var(--small-radius);
}

.external-type {
  flex: 1;
  min-width: 0;
}

.type-name {
  font-size: 13px;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.arrow-icon {
  width: 16px;
  height: 16px;
  opacity: 0.4;
  flex-shrink: 0;
}

.local-type {
  flex: 1;
  min-width: 120px;
}
</style>
