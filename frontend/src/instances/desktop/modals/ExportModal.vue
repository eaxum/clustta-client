<template>
  <div class="modal-container export-modal" v-esc="closeModal" v-stop-propagation>
    <HeaderArea :title="$t('modals.export.title')" :icon="getAppIcon('data-upload')" />

    <div class="general-container export-container">
      <div class="export-toolbar">
        <span class="item-count">{{ $t('modals.export.itemCount', { count: exportStore.preview.total }) }}</span>
        <div class="toolbar-actions">
          <label class="format-control">
            <span>{{ $t('modals.export.format') }}</span>
            <select v-model="exportStore.format">
              <option value="json">JSON</option>
              <option value="csv">CSV</option>
              <option value="txt">{{ $t('modals.export.plainText') }}</option>
            </select>
          </label>
          <div class="columns-control">
            <button type="button" class="columns-button" @click="columnsOpen = !columnsOpen">
              {{ $t('modals.export.columns') }}
            </button>
            <div v-if="columnsOpen" class="columns-menu" v-stop-propagation>
              <label v-for="column in exportStore.preview.columns" :key="column.key" class="column-option">
                <CheckBox
                  :modelValue="exportStore.selectedColumns.includes(column.key)"
                  :disabled="column.required"
                  :ariaLabel="column.label"
                  @update:modelValue="exportStore.toggleColumn(column)"
                />
                <span>{{ column.label }}</span>
                <span v-if="column.required" class="required-label">{{ $t('modals.export.required') }}</span>
              </label>
            </div>
          </div>
        </div>
      </div>

      <div class="export-table">
        <div class="table-scroll">
          <table>
            <thead>
              <tr>
                <th v-for="column in visibleColumns" :key="column.key">{{ column.label }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="exportStore.loading">
                <td :colspan="visibleColumns.length">{{ $t('common.loading') }}</td>
              </tr>
              <tr v-else-if="!exportStore.preview.rows.length">
                <td :colspan="visibleColumns.length">{{ $t('modals.export.noItems') }}</td>
              </tr>
              <tr v-for="(row, index) in exportStore.preview.rows" v-else :key="`${exportStore.preview.page}-${index}`">
                <td v-for="column in visibleColumns" :key="column.key" :title="formatCell(row[column.key])">
                  {{ formatCell(row[column.key]) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

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
import { computed, ref } from 'vue';

import CheckBox from '@/instances/common/components/CheckBox.vue';
import GeneralButton from '@/instances/common/components/GeneralButton.vue';
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

import { useDesktopModalStore } from '@/stores/desktopModals';
import { useExportStore } from '@/stores/exports';
import { useIconStore } from '@/stores/icons';

const desktopModals = useDesktopModalStore();
const exportStore = useExportStore();
const iconStore = useIconStore();
const columnsOpen = ref(false);

const visibleColumns = computed(() => exportStore.preview.columns.filter((column) => (
  exportStore.selectedColumns.includes(column.key)
)));
const totalPages = computed(() => Math.max(1, Math.ceil(exportStore.preview.total / exportStore.preview.page_size)));

const closeModal = () => desktopModals.setModalVisibility('exportModal', false);
const getAppIcon = (name) => iconStore.getAppIcon(name);
const changePage = (page) => exportStore.loadPreview(page);
const formatCell = (value) => Array.isArray(value) ? value.join(', ') : String(value ?? '');
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
.format-control,
.column-option,
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

.format-control {
  gap: .5rem;
  color: var(--text-muted);
}

select,
.columns-button,
.pagination-controls button {
  color: var(--text);
  background: var(--surface-2);
  border: 1px solid var(--border);
  border-radius: var(--small-radius);
  padding: .45rem .7rem;
}

.columns-control {
  position: relative;
}

.columns-button,
.pagination-controls button {
  cursor: pointer;
}

.columns-menu {
  position: absolute;
  z-index: 4;
  top: calc(100% + .35rem);
  right: 0;
  width: 230px;
  max-height: 300px;
  overflow-y: auto;
  padding: .4rem;
  color: var(--text);
  background: var(--surface-1);
  border: 1px solid var(--border);
  border-radius: var(--normal-radius);
  box-shadow: 0 8px 24px rgba(0, 0, 0, .2);
}

.column-option {
  gap: .45rem;
  min-height: 34px;
  padding: .2rem .35rem;
}

.required-label {
  margin-left: auto;
  color: var(--text-muted);
  font-size: .75rem;
}

.export-table {
  width: 100%;
  height: 430px;
  overflow: hidden;
  border: 1px solid var(--border);
  border-radius: var(--normal-radius);
  background: var(--surface-2);
}

.table-scroll {
  width: 100%;
  height: 100%;
  overflow: auto;
}

table {
  width: 100%;
  min-width: 680px;
  border-collapse: collapse;
  table-layout: fixed;
}

th,
td {
  padding: .65rem .75rem;
  text-align: left;
  border-bottom: 1px solid var(--border);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

th {
  position: sticky;
  top: 0;
  z-index: 1;
  color: var(--text-muted);
  background: var(--surface-3);
  font-weight: 600;
}

tbody tr:hover {
  background: var(--surface-3);
}

.pagination-controls {
  justify-content: center;
  gap: .75rem;
  width: 100%;
}

.pagination-controls button:disabled {
  cursor: default;
  opacity: .45;
}

.table-scroll,
.columns-menu {
  scrollbar-color: var(--surface-5) transparent;
  scrollbar-width: thin;
}
</style>
