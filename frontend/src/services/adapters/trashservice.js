export const TrashService = {
  // Returns all trashed items for a project
  GetTrashedItems: async (projectPath) => [],

  // Restores an item from trash
  RestoreItem: async (projectPath, itemId) => {},

  // Permanently deletes an item
  DeletePermanently: async (projectPath, itemId) => {},

  // Empties the trash for a project
  EmptyTrash: async (projectPath) => {},
};
