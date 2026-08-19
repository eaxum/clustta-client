import { defineStore } from 'pinia';

import { DialogService, ExportService } from '@/services';
import { useBrowserTreeStore } from '@/stores/browserTree';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';

const defaultColumns = ['name', 'extension', 'parent', 'status', 'assignee', 'tags', 'asset_type'];

const formatOptions = {
  json: { filename: 'clustta-export.json', filterName: 'JSON', pattern: '*.json' },
  csv: { filename: 'clustta-export.csv', filterName: 'CSV', pattern: '*.csv' },
  txt: { filename: 'clustta-export.txt', filterName: 'Text', pattern: '*.txt' },
};

export const useExportStore = defineStore('exports', {
  state: () => ({
    scope: 'selection',
    assetIds: [],
    selectedColumns: [...defaultColumns],
    format: 'csv',
    preview: { columns: [], rows: [], total: 0, page: 1, page_size: 20 },
    loading: false,
    exporting: false,
  }),
  actions: {
    async open(scope = 'selection') {
      this.scope = scope;
      this.assetIds = scope === 'selection'
        ? useBrowserTreeStore().rootItems.filter((item) => item.type === 'asset').map((item) => item.id)
        : [];
      this.selectedColumns = [...defaultColumns];
      this.format = 'csv';
      useDesktopModalStore().setModalVisibility('exportModal', true);
      await this.loadPreview(1);
    },
    async loadPreview(page = 1) {
      const projectPath = useProjectStore().activeProject?.uri;
      if (!projectPath) return;
      this.loading = true;
      try {
        this.preview = await ExportService.Preview(projectPath, this.request(page));
      } catch (error) {
        useNotificationStore().errorNotification('Unable to preview export', error);
      } finally {
        this.loading = false;
      }
    },
    async toggleColumn(column) {
      if (column.required) return;
      if (this.selectedColumns.includes(column.key)) {
        this.selectedColumns = this.selectedColumns.filter((key) => key !== column.key);
      } else {
        this.selectedColumns = [...this.selectedColumns, column.key];
      }
      await this.loadPreview(1);
    },
    async save() {
      const projectPath = useProjectStore().activeProject?.uri;
      const option = formatOptions[this.format];
      if (!projectPath || !option || this.exporting) return;
      this.exporting = true;
      try {
        const destination = await DialogService.SaveFileDialog(
          'Export Clustta data',
          option.filename,
          option.filterName,
          option.pattern,
        );
        if (!destination) return;
        await ExportService.Export(projectPath, destination, this.format, this.request(1));
        useNotificationStore().addNotification('Export complete', '', 'success');
        useDesktopModalStore().setModalVisibility('exportModal', false);
      } catch (error) {
        useNotificationStore().errorNotification('Unable to export data', error);
      } finally {
        this.exporting = false;
      }
    },
    request(page) {
      return {
        asset_ids: this.assetIds,
        scope: this.scope,
        columns: this.selectedColumns,
        page,
      };
    },
  },
});
