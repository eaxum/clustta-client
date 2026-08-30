<template>
  <div ref="menuRoot" class="compact-edit-menu" v-stop-propagation>
    <div v-if="menuData.title" class="compact-edit-title">{{ menuData.title }}</div>
    <div v-if="menuData.loading" class="compact-edit-state">Loading options...</div>
    <div v-else-if="!menuData.options.length" class="compact-edit-state">No options available</div>
    <template v-else>
      <button v-for="option in menuData.options" :key="option.id" class="compact-edit-option"
        :class="{ 'compact-edit-option-selected': option.id === menuData.selectedId }"
        :disabled="option.disabled || isSaving" type="button" @click="selectOption(option)">
        <img v-if="option.icon" class="small-icons compact-edit-option-icon" :src="option.icon">
        <span class="compact-edit-option-label">{{ option.label }}</span>
        <img v-if="option.id === menuData.selectedId" class="small-icons compact-edit-check" :src="getAppIcon('check')">
      </button>
    </template>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';

const iconStore = useIconStore();
const menu = useMenu();
const isSaving = ref(false);
const menuRoot = ref(null);

const menuData = computed(() => menu.compactEditMenuData);

const getAppIcon = (iconName) => iconStore.getAppIcon(iconName);

const closeOnOutsideClick = (event) => {
  if (menuRoot.value?.contains(event.target)) return;
  menu.hideContextMenu();
};

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

onMounted(() => document.addEventListener('pointerdown', closeOnOutsideClick, true));
onBeforeUnmount(() => document.removeEventListener('pointerdown', closeOnOutsideClick, true));
</script>

<style scoped>
.compact-edit-menu {
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
  background: var(--surface-2);
  background: color-mix(in srgb, var(--surface-2) 82%, transparent);
  backdrop-filter: blur(35px);
  box-shadow: 0 10px 28px rgba(0, 0, 0, .3);
  color: var(--text);
}

.compact-edit-title {
  color: var(--text-muted);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: .04em;
  text-transform: uppercase;
}

.compact-edit-title {
  padding: .35rem .5rem .45rem;
  border-bottom: var(--transparent-line);
}

.compact-edit-state {
  padding: .75rem;
  color: var(--text-muted);
  font-size: 13px;
  text-align: center;
}

.compact-edit-option {
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

.compact-edit-option:hover,
.compact-edit-option-selected {
  background: var(--surface-4);
}

.compact-edit-option:disabled {
  cursor: default;
  opacity: .55;
}

.compact-edit-option-label {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.compact-edit-option-label {
  font-size: 13px;
  font-weight: 400;
}

.compact-edit-option-icon,
.compact-edit-check {
  flex: 0 0 auto;
}

.compact-edit-menu::-webkit-scrollbar {
  width: 4px;
}

.compact-edit-menu::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--surface-4);
}

.compact-edit-menu::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}
</style>
