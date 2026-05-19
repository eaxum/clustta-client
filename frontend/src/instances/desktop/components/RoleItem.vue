<template>
  <div class="role-item-main">
    <div class="role-item-spacer">
      <img class="role-item-icon small-icons" :src="getAppIcon('scale')" />
    </div>

    <div class="role-item-root">
      <div class="role-item-container">
        <div class="role-item-content">
          <div class="role-item-details">
            <div class="role-item-name">{{ role.name }}</div>
            <div class="role-item-summary">{{ permissionSummary }}</div>
          </div>
        </div>

        <div class="role-item-actions">
          <ActionButton v-if="canEdit" :icon="getAppIcon('edit')" @click="onEdit(role.id)" v-tooltip="$t('common.edit')" />
          <ActionButton v-if="canDelete" :icon="getAppIcon('trash')" @click="onDelete(role.id)" v-tooltip="$t('common.delete')" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
// imports
import { computed } from 'vue';
import { getPermissionSummary } from '@/lib/permissions';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// stores
import { useIconStore } from '@/stores/icons';

const iconStore = useIconStore();

// props
const props = defineProps({
  canDelete: { type: Boolean, default: true },
  canEdit: { type: Boolean, default: true },
  onDelete: { type: Function, default: () => {} },
  onEdit: { type: Function, default: () => {} },
  role: { type: Object, required: true },
});

// computed
// Returns a human-readable summary of active permissions.
const permissionSummary = computed(() => {
  return getPermissionSummary(props.role) || 'No permissions';
});

// methods
// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};
</script>

<style scoped>
.role-item-main {
  display: flex;
  gap: .2rem;
  color: var(--text);
  align-items: center;
  padding-left: .5rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: flex-end;
  align-items: flex-start;
  background-color: var(--surface-2);
  border-radius: var(--large-radius);
  overflow: hidden;
  padding-right: 0px;
  outline: var(--transparent-line);
  outline-offset: -1px;
  transition: all .2s ease-out;
}

.role-item-main:hover {
  background-color: var(--surface-3);
  border-radius: var(--small-radius);
  outline: 1px solid var(--surface-4);
}

.role-item-spacer {
  position: relative;
  width: 36px;
  height: 60px;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.role-item-icon {
  width: 20px;
  height: 20px;
  opacity: 0.6;
}

.role-item-root {
  display: flex;
  flex-direction: column;
  gap: .2rem;
  color: var(--text);
  align-items: center;
  padding: .3rem;
  box-sizing: border-box;
  width: 100%;
  height: min-content;
  justify-content: flex-end;
  border-radius: 10px;
  overflow: hidden;
  padding-right: 0px;
}

.role-item-container {
  display: flex;
  gap: .5rem;
  color: var(--text);
  align-items: center;
  padding: .2rem .4rem;
  box-sizing: border-box;
  width: 100%;
  height: 50px;
  justify-content: space-between;
  transition: all .3s ease-out;
}

.role-item-content {
  gap: .4rem;
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  width: 100%;
  overflow: hidden;
}

.role-item-details {
  padding: .2rem;
  flex-wrap: nowrap;
  overflow: hidden;
  box-sizing: border-box;
  align-items: center;
  justify-content: space-between;
  height: min-content;
  white-space: nowrap;
  text-overflow: ellipsis;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: .1rem;
}

.role-item-name {
  font-size: 14px;
  font-weight: 400;
}

.role-item-summary {
  font-size: 12px;
  font-weight: 300;
  color: var(--text);
  opacity: 0.5;
}

.role-item-actions {
  display: flex;
  box-sizing: border-box;
  align-items: center;
  justify-content: flex-end;
  width: min-content;
  min-width: max-content;
  gap: .5rem;
  height: 100%;
}
</style>
