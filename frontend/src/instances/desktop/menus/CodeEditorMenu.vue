<template>
  <div class="filter-menu-container">
    <template v-for="item in menu.codeEditorMenuData" :key="item.id">
      <span v-if="item.divider" class="menu-divider"></span>
      <ActionButton
        v-else
        :icon="getAppIcon(item.icon)"
        :showLabel="true"
        :fullWidth="true"
        :label="item.label"
        :isDisabled="item.disabled"
        :buttonFunction="() => runAction(item)"
      />
    </template>
  </div>
</template>

<script setup>
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';

const iconStore = useIconStore();
const menu = useMenu();

const getAppIcon = (name) => iconStore.getAppIcon(name);
const runAction = async (item) => {
  try {
    if (!item.disabled) await item.action?.();
  } catch (error) {
    console.error(error);
  } finally {
    menu.hideContextMenu();
  }
};
</script>

<style scoped>
@import "@/assets/desktop.css";
@import "@/assets/menu.css";
</style>
