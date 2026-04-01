<template>
  <div v-if="visible" class="plan-info" :class="{ 'plan-free': !entitlementStore.isPaidPlan }" @click="handleClick" v-stop-propagation>
    <template v-if="!entitlementStore.isPaidPlan">
      <span class="plan-label">Free</span>
      <span class="plan-separator">·</span>
      <span class="plan-upgrade">Upgrade</span>
    </template>

    <template v-else>
      <span class="plan-label">{{ planDisplayName }}</span>
      <div class="storage-bar-container">
        <div class="storage-bar-fill" :class="storageBarClass" :style="{ width: entitlementStore.storagePercent + '%' }"></div>
      </div>
      <span class="storage-text">{{ entitlementStore.storageUsedFormatted }} / {{ entitlementStore.storageLimitFormatted }}</span>
    </template>
  </div>
</template>

<script setup>
// imports
import { computed } from 'vue';

// stores
import { useAccountStore } from '@/stores/accounts';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useEntitlementStore } from '@/stores/entitlements';
import { useProjectStore } from '@/stores/projects';
import { useUserStore } from '@/stores/users';

const accountStore = useAccountStore();
const entitlementStore = useEntitlementStore();
const modals = useDesktopModalStore();
const projectStore = useProjectStore();
const userStore = useUserStore();

// computed
// Returns whether the component should be visible.
const visible = computed(() => {
  return userStore.isUserAuthenticated && projectStore.selectedStudio?.name === 'Personal' && !accountStore.isOfflineMode;
});

// Returns the display name for the current plan.
const planDisplayName = computed(() => {
  const name = entitlementStore.plan || 'free';
  return name.charAt(0).toUpperCase() + name.slice(1).replace(/_/g, ' ');
});

// Returns the CSS class for the storage bar based on usage level.
const storageBarClass = computed(() => {
  if (entitlementStore.isOverStorage) return 'storage-danger';
  if (entitlementStore.isNearQuota) return 'storage-warning';
  return '';
});

// methods
// Handles click on the plan info widget.
const handleClick = () => {
  modals.setModalVisibility('clusttaCloudModal', true);
};
</script>

<style scoped>
.plan-info {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 3px 10px;
  border-radius: var(--normal-radius);
  cursor: pointer;
  font-size: 11px;
  color: var(--white);
  background-color: hsla(0, 0%, 100%, 0.08);
  transition: background-color 0.15s ease;
  white-space: nowrap;
  height: 22px;
  margin-right: 6px;
}

.plan-info:hover {
  background-color: hsla(0, 0%, 100%, 0.15);
}

.plan-free {
  background-color: var(--grape);
}

.plan-free:hover {
  background-color: hsl(270, 50%, 38%);
}

.plan-label {
  font-weight: 500;
}

.plan-separator {
  opacity: 0.5;
}

.plan-upgrade {
  font-weight: 400;
  opacity: 0.9;
}

.storage-bar-container {
  width: 48px;
  height: 4px;
  border-radius: 2px;
  background-color: hsla(0, 0%, 100%, 0.15);
  overflow: hidden;
}

.storage-bar-fill {
  height: 100%;
  border-radius: 2px;
  background-color: var(--grape);
  transition: width 0.3s ease;
}

.storage-bar-fill.storage-warning {
  background-color: var(--attention);
}

.storage-bar-fill.storage-danger {
  background-color: var(--danger);
}

.storage-text {
  font-weight: 300;
  opacity: 0.7;
  font-size: 10px;
}
</style>
