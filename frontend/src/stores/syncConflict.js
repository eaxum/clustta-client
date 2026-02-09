import { defineStore } from "pinia";
import {
  buildConflictFullPath,
  filterTopLevelConflicts,
  findChildConflicts,
  getConflictIdsToRemove,
} from "@/lib/conflictUtils";

export const useSyncConflictStore = defineStore("syncConflict", {
  state: () => ({
    conflicts: [],
    projectPath: "",
    remoteURL: "",
  }),
  getters: {
    // Returns all conflicts regardless of parent-child relationships.
    allConflicts: (state) => state.conflicts,

    // Returns only entity type conflicts.
    entityConflicts: (state) => state.conflicts.filter(c => c.type === 'entity'),

    // Returns true if there are any conflicts remaining.
    hasConflicts: (state) => state.conflicts.length > 0,

    // Returns only task type conflicts.
    taskConflicts: (state) => state.conflicts.filter(c => c.type === 'task'),

    // Returns only top-level conflicts (items without a conflicted parent).
    topLevelConflicts: (state) => filterTopLevelConflicts(state.conflicts),
  },
  actions: {
    // Clears all conflict state.
    clearConflicts() {
      this.conflicts = [];
      this.projectPath = "";
      this.remoteURL = "";
    },

    // Finds all child conflicts of a parent entity.
    getChildConflicts(parent) {
      return findChildConflicts(parent, this.conflicts);
    },

    // Builds the full path for a conflict item.
    getConflictFullPath(conflict) {
      return buildConflictFullPath(conflict);
    },

    // Removes a single conflict by local_id.
    removeConflict(localId) {
      this.conflicts = this.conflicts.filter(c => c.local_id !== localId);
    },

    // Removes a conflict and all its children recursively.
    removeConflictWithChildren(conflict) {
      const idsToRemove = getConflictIdsToRemove(conflict, this.conflicts);
      this.conflicts = this.conflicts.filter(c => !idsToRemove.includes(c.local_id));
    },

    // Removes multiple conflicts by their local_ids.
    removeConflicts(localIds) {
      this.conflicts = this.conflicts.filter(c => !localIds.includes(c.local_id));
    },

    // Sets the initial conflict data.
    setConflicts(projectPath, remoteURL, conflicts) {
      this.projectPath = projectPath;
      this.remoteURL = remoteURL;
      this.conflicts = conflicts;
    },
  },
});
