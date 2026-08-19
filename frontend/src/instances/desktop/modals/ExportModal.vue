<template>
  <div class="modal-container export-modal" v-esc="closeModal" v-stop-propagation>
    <HeaderArea :title="$t('modals.export.title')" :icon="getAppIcon('data-upload')" />

    <div class="general-container export-container">
      <div class="export-toolbar">
        <span class="item-count">{{ $t('modals.export.itemCount', { count: exportStore.preview.total }) }}</span>
        <div class="toolbar-actions">
          <label class="toolbar-control">
            <span>{{ $t('modals.export.format') }}</span>
            <DropDownBox
              :items="formatOptions"
              :selectedItem="selectedFormatLabel"
              :onSelect="selectFormat"
              :useFilter="false"
              :fullWidth="false"
              :fixedWidth="true"
            />
          </label>
          <label class="toolbar-control name-format-control">
            <span>{{ $t('modals.export.nameFormat') }}</span>
            <DropDownBox
              :items="nameFormatOptions"
              :selectedItem="selectedNameFormatLabel"
              :onSelect="selectNameFormat"
              :useFilter="false"
              :fullWidth="false"
              :fixedWidth="true"
            />
          </label>
          <FilterButton
            :icon="getAppIcon('list')"
            :label="$t('modals.export.columns')"
            :showLabel="true"
            :buttonFunction="openColumnsMenu"
          />
        </div>
      </div>

      <DataTable
        :columns="visibleColumns"
        :rows="exportStore.preview.rows"
        :rowKey="previewRowKey"
        :loading="exportStore.loading"
        :loadingText="$t('common.loading')"
        :emptyText="$t('modals.export.noItems')"
        maxHeight="430px"
        minWidth="680px"
      />

      <div v-if="totalPages > 1" class="pagination-controls">
        <button type="button" :disabled="exportStore.preview.page === 1" @click="changePage(exportStore.preview.page - 1)">
          {{ $t('modals.agentApproval.previous') }}
        </button>
        <span>{{ $t('modals.agentApproval.page', { current: exportStore.preview.page, total: totalPages }) }}</span>
        <button type="button" :disabled="exportStore.preview.page === totalPages" @click="changePage(exportStore.preview.page + 1)">
          {{ $t('modals.agentApproval.next') }}
        </button>
      </div>

      <div class="pop-up-actions">
        <GeneralButton :label="$t('common.cancel')" :fullWidth="true" :buttonFunction="closeModal" :colored="false" />
        <GeneralButton
          :label="$t('modals.export.export')"
          :fullWidth="true"
          :buttonFunction="exportStore.save"
          :isActive="exportStore.preview.total > 0"
          :loading="exportStore.exporting"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import DataTable from '@/instances/common/components/DataTable.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import FilterButton from '@/instances/desktop/components/FilterButton.vue';

import { useDesktopModalStore } from '@/stores/desktopModals';
import { useExportStore } from '@/stores/exports';
import { useIconStore } from '@/stores/icons';
import { useMenu } from '@/stores/menu';

const desktopModals = useDesktopModalStore();
const exportStore = useExportStore();
const iconStore = useIconStore();
const menu = useMenu();
const { t } = useI18n();

const formatOptions = computed(() => [
  { name: 'JSON', value: 'json' },
  { name: 'CSV', value: 'csv' },
  { name: t('modals.export.plainText'), value: 'txt' },
]);
const nameFormatOptions = computed(() => [
  { name: t('modals.export.original'), value: 'original' },
  { name: 'kebab-case', value: 'kebab' },
  { name: 'snake_case', value: 'snake' },
  { name: 'camelCase', value: 'camel' },
  { name: 'PascalCase', value: 'pascal' },
  { name: 'UPPERCASE', value: 'uppercase' },
  { name: 'lowercase', value: 'lowercase' },
  { name: 'Title Case', value: 'title' },
]);
const selectedFormatLabel = computed(() => formatOptions.value.find((option) => option.value === exportStore.format)?.name || 'CSV');
const selectedNameFormatLabel = computed(() => nameFormatOptions.value.find((option) => option.value === exportStore.nameFormat)?.name || t('modals.export.original'));
const visibleColumns = computed(() => exportStore.preview.columns.filter((column) => exportStore.selectedColumns.includes(column.key)));
const totalPages = computed(() => Math.max(1, Math.ceil(exportStore.preview.total / exportStore.preview.page_size)));

const closeModal = () => {
  menu.disableAllMenus();
  desktopModals.setModalVisibility('exportModal', false);
};
const getAppIcon = (name) => iconStore.getAppIcon(name);
const changePage = (page) => exportStore.loadPreview(page);
const previewRowKey = (_, index) => `${exportStore.preview.page}-${index}`;
const selectFormat = (label) => exportStore.format = formatOptions.value.find((option) => option.name === label)?.value || 'csv';
const selectNameFormat = (label) => {
  const nameFormat = nameFormatOptions.value.find((option) => option.name === label)?.value || 'original';
  exportStore.setNameFormat(nameFormat);
};
const openColumnsMenu = (event) => menu.showContextMenu(event, 'exportColumnsMenu', true, true);
</script>

<style scoped>
@import "@/assets/desktop.css";

.modal-container {
  max-height: 80vh;
  max-width: 90vw;
  width: min(900px, calc(100vw - 4rem));
  font-size: 14px;
}

.general-container {
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-width: 600px;
  width: 1200px;
  max-width: 96%;
  box-sizing: border-box;
}

.export-toolbar,
.toolbar-actions,
.toolbar-control,
.pagination-controls {
  display: flex;
  align-items: center;
}

.export-toolbar {
  width: 100%;
  justify-content: space-between;
  gap: 1rem;
}

.item-count {
  font-weight: 600;
}

.toolbar-actions {
  gap: .75rem;
}

.toolbar-control {
  gap: .5rem;
  color: var(--text-muted);
}

.name-format-control :deep(.list-box-container) {
  min-width: 150px;
}

.pagination-controls {
  justify-content: center;
  gap: .75rem;
  width: 100%;
}

.pagination-controls button {
  cursor: pointer;
  color: var(--text);
  background: var(--surface-2);
  border: 1px solid var(--border);
  border-radius: var(--small-radius);
  padding: .45rem .7rem;
}

.pagination-controls button:disabled {
  cursor: default;
  opacity: .45;
}
</style>
