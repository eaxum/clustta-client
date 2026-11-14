import { defineStore } from 'pinia';
import { AssetService, FSService } from '@/services';
import { useStageStore } from '@/stores/stages';
import { canActOnAsset } from '@/lib/permissions';
import { useProjectStore } from '@/stores/projects';
import { useUserStore } from '@/stores/users';

const normalizedPath = (path) => path?.replace(/\\/g, '/').toLowerCase();
const pathsMatch = (left, right) => normalizedPath(left) === normalizedPath(right);

export const useCodeEditorStore = defineStore('codeEditor', {
  state: () => ({
    documents: [],
    activePath: '',
    searchQuery: '',
    searchResults: [],
    searchLoading: false,
    explorerVisible: true,
  }),
  getters: {
    activeDocument(state) {
      return state.documents.find((document) => document.path === state.activePath) || null;
    },
  },
  actions: {
    async openDocument(path, asset = null, preview = true) {
      const existing = this.documents.find((document) => pathsMatch(document.path, path));
      if (existing) {
        if (asset) existing.asset = asset;
        if (!preview) existing.preview = false;
        this.activateDocument(path);
        return existing;
      }

      const document = await FSService.ReadCodeFile(path);
      const trackedReadOnly = asset?.type !== 'untracked_asset' && asset && !canActOnAsset('update_asset', asset);
      const untrackedReadOnly = asset?.type === 'untracked_asset' && !useUserStore().canDo('update_asset');
      if (trackedReadOnly || untrackedReadOnly) {
        document.classification.editable = false;
        document.classification.read_only = true;
        document.classification.reason = 'permission_denied';
      }
      if (preview) {
        this.documents = this.documents.filter((openDocument) => !openDocument.preview || openDocument.dirty);
      }
      const openDocument = {
        ...document,
        asset,
        dirty: false,
        preview,
        savedContent: document.content,
        externalChanged: false,
        saveError: '',
        viewState: null,
      };
      this.documents.push(openDocument);
      this.activateDocument(openDocument.path);
      return openDocument;
    },

    activateDocument(path) {
      const document = this.documents.find((item) => pathsMatch(item.path, path));
      if (!document) return;
      this.activePath = document.path;
      useStageStore().setStageVisibility('codeEditor', true);
    },

    pinDocument(path = this.activePath) {
      const document = this.documents.find((item) => pathsMatch(item.path, path));
      if (document) document.preview = false;
    },

    updateContent(content) {
      const document = this.activeDocument;
      if (!document?.classification.editable) return;
      document.content = content;
      document.dirty = content !== document.savedContent;
      if (document.dirty) document.preview = false;
    },

    async refreshAssetState(document) {
      const asset = document?.asset;
      if (!asset?.id || asset.type === 'untracked_asset') return;
      try {
        asset.file_status = await AssetService.GetAssetState(
          useProjectStore().activeProject.uri,
          asset.id,
        );
      } catch (error) {
        console.error(error);
      }
    },

    async saveDocument(path = this.activePath) {
      const document = this.documents.find((item) => pathsMatch(item.path, path));
      if (!document || !document.dirty) return document;
      try {
        const saved = await FSService.WriteCodeFile(
          document.path,
          document.content,
          document.content_identity,
        );
        Object.assign(document, saved, {
          dirty: false,
          preview: false,
          savedContent: saved.content,
          externalChanged: false,
          saveError: '',
        });
        await this.refreshAssetState(document);
        return document;
      } catch (error) {
        document.saveError = error.message || String(error);
        throw error;
      }
    },

    async handleExternalChange(path) {
      const document = this.documents.find((item) => pathsMatch(item.path, path));
      if (!document) return;
      let changed;
      try {
        changed = await FSService.ReadCodeFile(path);
      } catch (error) {
        document.externalChanged = true;
        document.saveError = error.message || String(error);
        return;
      }
      if (changed.content_identity === document.content_identity) return;
      if (document.dirty) {
        document.externalChanged = true;
        document.externalIdentity = changed.content_identity;
        return;
      }
      Object.assign(document, changed, {
        dirty: false,
        externalChanged: false,
        saveError: '',
        externalIdentity: '',
        savedContent: changed.content,
      });
      await this.refreshAssetState(document);
    },

    async reloadDocument(path = this.activePath) {
      const index = this.documents.findIndex((item) => pathsMatch(item.path, path));
      if (index < 0) return;
      const changed = await FSService.ReadCodeFile(path);
      this.documents[index] = {
        ...this.documents[index],
        ...changed,
        dirty: false,
        externalChanged: false,
        saveError: '',
        externalIdentity: '',
        savedContent: changed.content,
      };
    },

    keepLocalVersion(path = this.activePath) {
      const document = this.documents.find((item) => pathsMatch(item.path, path));
      if (!document) return;
      document.content_identity = document.externalIdentity || document.content_identity;
      document.externalIdentity = '';
      document.externalChanged = false;
      document.saveError = '';
    },

    async saveAll() {
      for (const document of this.documents.filter((item) => item.dirty)) {
        await this.saveDocument(document.path);
      }
    },

    closeDocument(path) {
      const index = this.documents.findIndex((document) => pathsMatch(document.path, path));
      if (index < 0 || this.documents[index].dirty) return false;
      this.documents.splice(index, 1);
      if (!pathsMatch(this.activePath, path)) return true;
      const nextDocument = this.documents[Math.min(index, this.documents.length - 1)];
      this.activePath = nextDocument?.path || '';
      if (!nextDocument) useStageStore().setStageVisibility('browser', true);
      return true;
    },

    async searchFiles(projectRoot, query = this.searchQuery) {
      this.searchQuery = query;
      this.searchLoading = true;
      try {
        this.searchResults = await FSService.SearchCodeFiles(projectRoot, query, 200);
      } catch (error) {
        console.error(error);
        this.searchResults = [];
      } finally {
        this.searchLoading = false;
      }
    },
  },
});
