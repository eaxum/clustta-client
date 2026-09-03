<template>
  <div ref="columnsMenu" class="filter-menu-container" v-stop-propagation>
    <span
      v-for="column in exportStore.preview.columns"
      :key="column.key"
      class="filter-menu-item"
      :class="{ disabled: column.required }"
      @click="exportStore.toggleColumn(column)"
    >
      <img class="small-icons" :src="getAppIcon(columnIcon(column.key))" />
      <div class="horizontal-flex">
        <div class="menu-item-text">{{ column.label }}</div>
        <CheckBox :modelValue="exportStore.selectedColumns.includes(column.key)" :disabled="column.required"
          :ariaLabel="`Include ${column.label}`" @click.stop @change="exportStore.toggleColumn(column)" />
      </div>
    </span>
  </div>
</template>

<script setup>
import { onBeforeUnmount, onMounted, ref } from 'vue';

import CheckBox from '@/instances/common/components/CheckBox.vue';

import { useExportStore } from '@/stores/exports';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';

const exportStore = useExportStore();
const iconStore = useIconStore();
const menu = useMenu();
const columnsMenu = ref(null);

const iconsByColumn = {
  name: 'file-name',
  extension: 'extension',
  parent: 'folder',
  status: 'clock',
  assignee: 'person',
  tags: 'tag',
  asset_type: 'file',
  created_at: 'clock',
  updated_at: 'clock',
  description: 'list',
  kind: 'shapes',
  path: 'file-path',
};

const columnIcon = (key) => iconsByColumn[key] || 'file';
const getAppIcon = (name) => iconStore.getAppIcon(name);

onMounted(() => {
  menu.assetMenuWidth = columnsMenu.value.getBoundingClientRect().width;
  menu.collectionMenu = columnsMenu.value;
});

onBeforeUnmount(() => {
  menu.assetMenuWidth = columnsMenu.value?.getBoundingClientRect().width || 0;
  menu.assetMenuHeight = columnsMenu.value?.getBoundingClientRect().height || 0;
});
</script>

<style scoped>
@import "@/assets/desktop.css";
@import "@/assets/menu.css";

.horizontal-flex {
  min-width: 160px;
}
</style>
