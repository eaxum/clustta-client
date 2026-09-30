<template>
  <div class="role-selection-menu" v-stop-propagation>
    <div class="role-selection-title">{{ menuData.title }}</div>
    <button v-for="option in menuData.options" :key="option.id" class="role-selection-option"
      :class="{ 'role-selection-option-selected': option.id === menuData.selectedId }"
      :disabled="option.disabled || isSaving" type="button" @click="selectOption(option)">
      <img v-if="option.icon" class="small-icons role-selection-icon" :src="option.icon">
      <span class="role-selection-label">{{ option.label }}</span>
      <img v-if="option.id === menuData.selectedId" class="small-icons role-selection-icon"
        :src="getAppIcon('check')">
    </button>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';

const iconStore = useIconStore();
const menu = useMenu();
const isSaving = ref(false);
const menuData = computed(() => menu.roleSelectionMenuData);

const getAppIcon = iconName => iconStore.getAppIcon(iconName);

const selectOption = async (option) => {
  if (option.disabled || isSaving.value || !menuData.value.onSelect) return;
  if (option.id === menuData.value.selectedId) {
    menu.hideContextMenu();
    return;
  }
  isSaving.value = true;
  try {
    const shouldClose = await menuData.value.onSelect(option);
    if (shouldClose !== false) menu.hideContextMenu();
  } finally {
    isSaving.value = false;
  }
};
</script>

<style scoped>
.role-selection-menu {
  display: flex;
  width: 260px;
  max-height: min(420px, 70vh);
  overflow-y: auto;
  flex-direction: column;
  gap: .15rem;
  padding: .4rem;
  box-sizing: border-box;
  border: 1px solid var(--border-color);
  border-radius: var(--large-radius);
  background: color-mix(in srgb, var(--surface-2) 82%, transparent);
  backdrop-filter: blur(35px);
  box-shadow: 0 10px 28px rgba(0, 0, 0, .3);
  color: var(--text);
}

.role-selection-title {
  padding: .35rem .5rem .45rem;
  border-bottom: var(--transparent-line);
  color: var(--text-muted);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: .04em;
  text-transform: uppercase;
}

.role-selection-option {
  display: flex;
  align-items: center;
  gap: .5rem;
  width: 100%;
  min-height: 36px;
  padding: .35rem .5rem;
  border: 0;
  border-radius: var(--normal-radius);
  background: transparent;
  color: var(--text);
  cursor: pointer;
  text-align: left;
}

.role-selection-option:hover,
.role-selection-option-selected {
  background: var(--surface-4);
}

.role-selection-option:disabled {
  cursor: default;
  opacity: .55;
}

.role-selection-label {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
  font-weight: 400;
}

.role-selection-icon {
  flex: 0 0 auto;
}
</style>
