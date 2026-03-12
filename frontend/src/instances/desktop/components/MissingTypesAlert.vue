<template>
  <div v-if="hasMissingTypes" class="missing-types-alert">
    <div class="alert-icon">
      <img :src="getAppIcon('warning')" alt="warning" />
    </div>

    <div class="alert-content">
      <div class="alert-title">New types will be created</div>
      <div class="alert-description">
        <span v-if="entityCount > 0">{{ entityCount }} collection type{{ entityCount > 1 ? 's' : '' }}</span>
        <span v-if="entityCount > 0 && taskCount > 0"> and </span>
        <span v-if="taskCount > 0">{{ taskCount }} task type{{ taskCount > 1 ? 's' : '' }}</span>
        <span> will be created to match {{ integrationName }}.</span>
      </div>
    </div>

    <div class="alert-actions">
      <button v-if="showDetails" class="toggle-details" @click="emit('toggle')">
        {{ expanded ? 'Hide' : 'Show' }} Details
      </button>
    </div>
  </div>

  <div v-if="expanded && hasMissingTypes" class="missing-types-list">
    <div v-for="typeItem in missingTypes" :key="typeItem.external_id" class="missing-type-item">
      <img :src="getTypeIcon(typeItem.suggested_icon)" alt="" class="type-icon" />
      <div class="type-info">
        <span class="type-name">{{ typeItem.external_name }}</span>
        <span class="type-category">{{ typeItem.type_category === 'entity' ? 'Collection Type' : 'Task Type' }}</span>
      </div>
      <span class="type-badge">{{ typeItem.suggested_name }}</span>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed } from 'vue';

// stores
import { useIconStore } from '@/stores/icons';

const iconStore = useIconStore();

// props
const props = defineProps({
  expanded: { type: Boolean, default: false },
  integrationName: { type: String, default: 'external system' },
  missingTypes: { type: Array, default: () => [] },
  showDetails: { type: Boolean, default: true },
});

// emits
const emit = defineEmits(['toggle']);

// computed
// Count of entity types missing.
const entityCount = computed(() => {
  return props.missingTypes.filter(t => t.type_category === 'entity').length;
});

// Whether there are any missing types.
const hasMissingTypes = computed(() => {
  return props.missingTypes.length > 0;
});

// Count of task types missing.
const taskCount = computed(() => {
  return props.missingTypes.filter(t => t.type_category === 'task').length;
});

// methods
// Returns the app icon path.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Returns the type icon path.
const getTypeIcon = (iconName) => {
  return '/types-icons/' + (iconName || 'other') + '.svg';
};
</script>

<style scoped>
.missing-types-alert {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: var(--color-warning-subtle);
  border: 1px solid var(--color-warning);
  border-radius: var(--small-radius);
}

.alert-icon {
  flex-shrink: 0;
}

.alert-icon img {
  width: 20px;
  height: 20px;
  opacity: 0.8;
}

.alert-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.alert-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.alert-description {
  font-size: 12px;
  color: var(--text-secondary);
}

.alert-actions {
  flex-shrink: 0;
}

.toggle-details {
  padding: 6px 12px;
  font-size: 12px;
  background: transparent;
  border: 1px solid var(--border-primary);
  border-radius: var(--small-radius);
  color: var(--text-primary);
  cursor: pointer;
  transition: background 0.15s;
}

.toggle-details:hover {
  background: var(--surface-secondary);
}

.missing-types-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-top: 8px;
  padding: 8px;
  background: var(--surface-primary);
  border-radius: var(--small-radius);
  max-height: 150px;
  overflow-y: auto;
}

.missing-types-list::-webkit-scrollbar {
  width: 4px;
}

.missing-types-list::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--light-steel);
}

.missing-type-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  background: var(--surface-secondary);
  border-radius: var(--small-radius);
}

.type-icon {
  width: 18px;
  height: 18px;
  opacity: 0.7;
}

.type-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.type-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
}

.type-category {
  font-size: 11px;
  color: var(--text-tertiary);
}

.type-badge {
  padding: 2px 8px;
  font-size: 11px;
  background: var(--accent-subtle);
  color: var(--accent-primary);
  border-radius: var(--small-radius);
}
</style>
