// Wails adapter - re-exports all services from Wails bindings
// This is used in desktop mode (default)

export * from '@/../bindings/clustta/services';

import { getActivePinia } from 'pinia';
import * as bindings from '@/../bindings/clustta/services';

// Legacy callers must not apply completion effects to a project selected during a transfer.
async function finishInProject(projectPath, request) {
  const result = await request;
  const project = getActivePinia()?.state.value.projects?.activeProject;
  if (project?.uri !== projectPath) throw new Error('cancelled');
  return result;
}

export const SyncService = {
  ...bindings.SyncService,
  DownloadCheckpoint: (projectPath, ...args) => finishInProject(projectPath, bindings.SyncService.DownloadCheckpoint(projectPath, ...args)),
};

export const CheckpointService = {
  ...bindings.CheckpointService,
  Revert: (projectPath, ...args) => finishInProject(projectPath, bindings.CheckpointService.Revert(projectPath, ...args)),
};

export const CollectionService = {
  ...bindings.CollectionService,
  Fetch: (projectPath, ...args) => finishInProject(projectPath, bindings.CollectionService.Fetch(projectPath, ...args)),
};
