import { defineStore } from 'pinia';
import { AgentService } from '@/services';

const referenceCacheKey = (parentId) => parentId || ':root';

export const useAgentReferencesStore = defineStore('agentReferences', {
  state: () => ({
    projectUri: '',
    childrenByParent: {},
    pendingByParent: {},
    generation: 0,
  }),

  actions: {
    setProject(projectUri) {
      const nextProjectUri = projectUri || '';
      if (this.projectUri === nextProjectUri) return;
      this.projectUri = nextProjectUri;
      this.childrenByParent = {};
      this.pendingByParent = {};
      this.generation++;
    },

    async getChildren(projectUri, parentId) {
      this.setProject(projectUri);
      const key = referenceCacheKey(parentId);
      if (this.childrenByParent[key]) return this.childrenByParent[key];
      if (this.pendingByParent[key]) return this.pendingByParent[key];

      const requestGeneration = this.generation;
      const request = AgentService.ListEntityReferenceChildren(projectUri, parentId || '')
        .then((references) => {
          if (this.projectUri === projectUri && this.generation === requestGeneration) {
            this.childrenByParent[key] = references || [];
          }
          return references || [];
        })
        .finally(() => {
          if (this.projectUri === projectUri && this.generation === requestGeneration) {
            delete this.pendingByParent[key];
          }
        });
      this.pendingByParent[key] = request;
      return request;
    },

    invalidate() {
      this.childrenByParent = {};
      this.pendingByParent = {};
      this.generation++;
    },
  },
});
