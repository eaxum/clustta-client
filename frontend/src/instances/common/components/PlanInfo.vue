<template>
  <div v-if="visible" class="plan-info" :class="{ 'plan-free': !isPaid }" @click="handleClick" v-stop-propagation>
    <template v-if="!isPaid">
      <div class="plan-free-content">
        <img class="plan-free-icon" :src="getAppIcon('diamond')" />
        <span class="plan-upgrade">Upgrade</span>
      </div>
    </template>

    <template v-else>
      <svg class="storage-donut" viewBox="0 0 16 16">
        <circle class="storage-donut-track" cx="8" cy="8" r="6" />
        <circle class="storage-donut-fill" :class="storageBarClass" cx="8" cy="8" r="6" :style="{ strokeDashoffset: storageDashOffset }" />
      </svg>
      <!-- <span class="plan-label">{{ planDisplayName }}</span> -->
      <span class="plan-label">Storage</span>
      <span class="storage-text">{{ storageUsedFormatted }} / {{ storageLimitFormatted }}</span>
    </template>
  </div>
</template>

<script setup>
// imports
import { computed, watch } from 'vue';

// stores/services
import utils from '@/services/utils';
import { useAccountStore } from '@/stores/accounts';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useEntitlementStore } from '@/stores/entitlements';
import { useIconStore } from '@/stores/icons';
import { useProjectStore } from '@/stores/projects';
import { useUserStore } from '@/stores/users';

const accountStore = useAccountStore();
const entitlementStore = useEntitlementStore();
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const projectStore = useProjectStore();
const userStore = useUserStore();

// computed
// Returns whether the selected studio is a cloud studio.
const isCloudStudio = computed(() => {
  return projectStore.isCloudHosted;
});

// Returns whether the selected studio is Personal.
const isPersonal = computed(() => {
  return projectStore.selectedStudio?.name === 'Personal';
});

// Returns the studio entitlement bundle when in studio mode.
const studioBundle = computed(() => {
  if (!isCloudStudio.value) return null;
  return entitlementStore.studioEntitlements[projectStore.selectedStudio?.id] || null;
});

// Returns whether the component should be visible.
const visible = computed(() => {
  if (!userStore.isUserAuthenticated || accountStore.isOfflineMode) return false;
  if (isPersonal.value) return true;
  return isCloudStudio.value;
});

// Returns whether the current context has a paid plan.
const isPaid = computed(() => {
  if (isCloudStudio.value && studioBundle.value) return studioBundle.value.plan !== 'free';
  return entitlementStore.isPaidPlan;
});

// Returns the display name for the current plan.
const planDisplayName = computed(() => {
  const name = (isCloudStudio.value && studioBundle.value) ? (studioBundle.value.plan || 'free') : (entitlementStore.plan || 'free');
  return name.charAt(0).toUpperCase() + name.slice(1).replace(/_/g, ' ');
});

// Returns storage usage as a percentage.
const storagePercent = computed(() => {
  if (isCloudStudio.value && studioBundle.value) {
    const limit = studioBundle.value.limits?.storage_bytes || 0;
    if (limit <= 0) return 0;
    return Math.min((studioBundle.value.usage?.storage_bytes || 0) / limit * 100, 100);
  }
  return entitlementStore.storagePercent;
});

// Returns formatted storage used.
const storageUsedFormatted = computed(() => {
  if (isCloudStudio.value && studioBundle.value) return utils.formatBytes(studioBundle.value.usage?.storage_bytes || 0, 2);
  return entitlementStore.storageUsedFormatted;
});

// Returns formatted storage limit.
const storageLimitFormatted = computed(() => {
  if (isCloudStudio.value && studioBundle.value) return utils.formatBytes(studioBundle.value.limits?.storage_bytes || 0, 0);
  return entitlementStore.storageLimitFormatted;
});

// Returns the stroke-dashoffset for the donut ring.
const storageDashOffset = computed(() => {
  const circumference = 2 * Math.PI * 6;
  return circumference * (1 - storagePercent.value / 100);
});

// Returns the CSS class for the storage donut based on usage level.
const storageBarClass = computed(() => {
  if (isCloudStudio.value && studioBundle.value) {
    const limit = studioBundle.value.limits?.storage_bytes || 0;
    const used = studioBundle.value.usage?.storage_bytes || 0;
    if (limit > 0 && used >= limit) return 'storage-danger';
    if (limit > 0 && used / limit >= 0.7) return 'storage-warning';
    return '';
  }
  if (entitlementStore.isOverStorage) return 'storage-danger';
  if (storagePercent.value >= 70) return 'storage-warning';
  return '';
});

// methods
// Returns the app icon for the given icon name.
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

// Handles click on the plan info widget.
const handleClick = () => {
  modals.setModalVisibility('clusttaCloudModal', true);
};

// watchers
// Fetches studio entitlements when switching to a cloud studio.
watch(() => projectStore.selectedStudio, (studio) => {
  if (projectStore.isCloudHosted && studio?.id) {
    entitlementStore.getStudioEntitlements(studio.id);
  }
}, { immediate: true });
</script>

<style scoped>
.plan-info {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 3px 5px;
  padding-right: 10px;
  border-radius: var(--large-radius);
  cursor: pointer;
  font-size: 11px;
  color: hsl(var(--foreground));
  background-color: hsl(var(--accent));
  transition: background-color 0.15s ease;
  white-space: nowrap;
  height: 22px;
  margin-right: 6px;
}

.plan-info:hover {
  background-color: hsla(0, 0%, 100%, 0.15);
}

.plan-free {
  background-color: hsl(var(--primary));
}

.plan-free:hover {
  background-color: hsl(270, 50%, 38%);
}

.plan-free-content{
 color: hsl(var(--foreground));
 display: flex;
 align-items: center;
 gap: .3rem;
}

.plan-free-icon {
  width: 18px;
  height: 18px;
  /* filter: invert(100%); */
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

.storage-donut {
  width: 20px;
  height: 20px;
  transform: rotate(-90deg);
}

.storage-donut-track {
  fill: none;
  stroke: hsl(var(--primary) / 0.3);
  stroke: hsl(var(--primary));
  background-color: hsl(var(--muted));
  stroke-width: 4;
}

.storage-donut-fill {
  fill: none;
  stroke: hsl(var(--primary));
  stroke: hsl(var(--muted-foreground));
  stroke-width: 4;
  stroke-dasharray: 37.7;
  stroke-linecap: round;
  transition: stroke-dashoffset 0.3s ease, stroke 0.3s ease;
}

[data-theme="dark"] .storage-donut-fill  {
  stroke: hsl(var(--muted-foreground));
  stroke: hsl(var(--accent));
}

.storage-donut-fill.storage-warning {
  stroke: hsl(var(--destructive));
}

.storage-donut-fill.storage-danger {
  stroke: var(--danger);
}

.storage-text {
  font-weight: 400;
  opacity: 0.7;
  font-size: 10px;
}
</style>
