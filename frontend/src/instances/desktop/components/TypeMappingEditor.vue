<template>
  <div class="type-mapping-editor">
    <div class="mapping-header">
      <CiCheckpointStone :size="20" class="header-icon" />
      <span class="header-title">Type Mappings</span>
      <span class="header-subtitle">Map {{ integrationName }} types to Clustta types</span>
    </div>

    <div class="mapping-sections">
      <div class="mapping-section">
        <div class="section-header">
          <span class="section-title">Collection Types</span>
          <span class="section-count">{{ collectionMappings.length }}</span>
        </div>

        <div class="mapping-rows">
          <div v-for="mapping in collectionMappings" :key="mapping.external_type" class="mapping-row">
            <div class="external-type">
              <span class="type-name">{{ mapping.external_name || mapping.external_type }}</span>
            </div>

            <CiArrowLeft :size="20" class="arrow-icon" />

            <div class="local-type">
              <DropDownBox :items="localCollectionTypes" :selectedItem="mapping.local_type" :onSelect="(value) => updateCollectionMapping(mapping.external_type, value)"
                :placeHolder="'Select type'" :useFilter="false" :fullWidth="true" />
            </div>
          </div>
        </div>
      </div>

      <div class="mapping-section">
        <div class="section-header">
          <span class="section-title">Asset Types</span>
          <span class="section-count">{{ assetMappings.length }}</span>
        </div>

        <div class="mapping-rows">
          <div v-for="mapping in assetMappings" :key="mapping.external_type" class="mapping-row">
            <div class="external-type">
              <span class="type-name">{{ mapping.external_name || mapping.external_type }}</span>
            </div>

            <CiArrowLeft :size="20" class="arrow-icon" />

            <div class="local-type">
              <DropDownBox :items="localAssetTypes" :selectedItem="mapping.local_type" :onSelect="(value) => updateAssetMapping(mapping.external_type, value)"
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
import { CiArrowLeft, CiCheckpointStone } from '@clustta/icons-vue';
import { resolveIcon } from '@/lib/icon-map';

// components
import DropDownBox from '@/instances/common/components/DropDownBox.vue';

// stores
import { useIconStore } from '@/stores/icons';

const iconStore = useIconStore();

// props
const props = defineProps({
  collectionMappings: { type: Array, default: () => [] },
  integrationName: { type: String, default: 'External' },
  localCollectionTypes: { type: Array, default: () => [] },
  localAssetTypes: { type: Array, default: () => [] },
  assetMappings: { type: Array, default: () => [] },
});

// emits
const emit = defineEmits(['update:collectionMappings', 'update:assetMappings']);

// methods
// Returns the app icon path.
const getAppIcon = (iconName) => {
  return iconStore.resolveIcon(iconName);
};

// Updates an collection type mapping.
const updateCollectionMapping = (externalType, localType) => {
  const updated = props.collectionMappings.map(mapping => {
    if (mapping.external_type === externalType) {
      return { ...mapping, local_type: localType };
    }
    return mapping;
  });
  emit('update:collectionMappings', updated);
};

// Updates a asset type mapping.
const updateAssetMapping = (externalType, localType) => {
  const updated = props.assetMappings.map(mapping => {
    if (mapping.external_type === externalType) {
      return { ...mapping, local_type: localType };
    }
    return mapping;
  });
  emit('update:assetMappings', updated);
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
  transform: scaleX(-1);
}

.local-type {
  flex: 1;
  min-width: 120px;
}
</style>
