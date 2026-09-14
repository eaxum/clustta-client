import { defineStore } from 'pinia';
import { Events } from '@wailsio/runtime';
import * as ActivityService from '@/../bindings/clustta/services/activityservice.js';
import { useProjectStore } from './projects';
import { useBrowserTreeStore } from './browserTree';
import { applyActivitySnapshot, isActivityPending, sortActivityOperations } from '@/lib/activity';
import { formatError } from '@/lib/errors';
import emitter from '@/lib/mitt';

export const useActivityStore = defineStore('activity', {
  state: () => ({
    operations: [],
    revision: -1,
    expanded: false,
    batchIDs: [],
    errors: {},
    connectionError: '',
  }),
  getters: {
    sortedOperations: (state) => sortActivityOperations(state.operations),
    running: (state) => state.operations.filter((item) => item.status === 'running'),
    finished: (state) => state.operations.filter((item) => !isActivityPending(item)),
  },
  actions: {
    applySnapshot(snapshot, restore = false) {
      if (!snapshot || snapshot.revision <= this.revision) return;
      const finished = applyActivitySnapshot(this, snapshot, restore);
      const browser = useBrowserTreeStore();
      browser.updatePendingTransfers(this.operations);
      const project = useProjectStore().activeProject;
      for (const item of finished) {
        if (project?.id !== item.project_id || project?.uri !== item.project_uri) continue;
        if (browser.projectUri === project.uri) browser.markAssetsAvailable(item.restored_asset_ids || []);
        emitter.emit('refresh-browser');
      }
    },
    async initialize() {
      const unsubscribe = Events.On('operations-updated', ({ data }) => this.applySnapshot(data));
      try {
        this.applySnapshot(await ActivityService.List(), true);
        this.connectionError = '';
      } catch (error) {
        this.connectionError = formatError(error);
      }
      return unsubscribe;
    },
    open() {
      this.expanded = true;
    },
    async cancel(id) {
      try {
        await ActivityService.Cancel(id);
        delete this.errors[id];
      } catch (error) {
        this.errors[id] = formatError(error);
      }
    },
    async dismiss(id) {
      try {
        await ActivityService.Dismiss(id);
        delete this.errors[id];
      } catch (error) {
        this.errors[id] = formatError(error);
      }
    },
    async clearFinished() {
      try {
        await ActivityService.ClearFinished();
        this.errors = {};
        this.connectionError = '';
      } catch (error) {
        this.connectionError = formatError(error);
      }
    },
  },
});
