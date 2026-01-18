import { defineStore } from "pinia";

export const useSyncConflictStore = defineStore("syncConflict", {
  state: () => ({
    conflicts: [],
    projectPath: "",
    remoteURL: "",
  }),
  getters: {
    hasConflicts: (state) => state.conflicts.length > 0,
    entityConflicts: (state) => state.conflicts.filter(c => c.type === 'entity'),
    taskConflicts: (state) => state.conflicts.filter(c => c.type === 'task'),
  },
  actions: {
    setConflicts(projectPath, remoteURL, conflicts) {
      this.projectPath = projectPath;
      this.remoteURL = remoteURL;
      this.conflicts = conflicts;
    },
    removeConflict(localId) {
      this.conflicts = this.conflicts.filter(c => c.local_id !== localId);
    },
    removeChildConflicts(parentPath) {
      // Remove all conflicts whose entity_path starts with the parent path
      this.conflicts = this.conflicts.filter(c => !c.entity_path?.startsWith(parentPath));
    },
    clearConflicts() {
      this.conflicts = [];
      this.projectPath = "";
      this.remoteURL = "";
    },
  },
});
