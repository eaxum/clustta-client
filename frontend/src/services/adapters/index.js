// =============================================================================
// HTTP ADAPTER - INDEX
// =============================================================================
// Re-exports all services for convenient importing
// Usage: import { AuthService, ProjectService } from '@/services/adapters';

// Core utilities (for advanced use cases)
export { GLOBAL_API, CLUSTTA_AGENT, STORAGE_KEYS, isDev } from './config.js';
export { globalApiCall, studioApiCall, getActiveStudioUrl, setActiveStudioUrl } from './http-client.js';
export {
  getSettings,
  setSetting,
  getSetting,
  getMultiAccountToken,
  setMultiAccountToken,
  addAccountToStorage,
  removeAccountFromStorage,
  switchActiveAccount,
  clearUserSpecificData,
  clearAllUserData,
  migrateToMultiAccount,
} from './storage.js';

// Services
export { AuthService } from './auth.service.js';
export { AccountService } from './account.service.js';
export { SettingsService } from './settings.service.js';
export { StudioService } from './studio.service.js';
export { ProjectService } from './project.service.js';
export { SyncService } from './sync.service.js';
export { ProfileService } from './profile.service.js';
export { UserService } from './user.service.js';
export { AssetService } from './asset.service.js';
export { CollectionService } from './collection.service.js';
export { CheckpointService } from './checkpoint.service.js';
export { TagService } from './tag.service.js';
export { StatusService } from './status.service.js';
export { TemplateService } from './template.service.js';
export { WorkflowService } from './workflow.service.js';
export { DependencyTypeService } from './dependency-type.service.js';
export { FSService } from './fs.service.js';
export { DialogService } from './dialog.service.js';
export { ClipboardService } from './clipboard.service.js';
export { TrashService } from './trash.service.js';
export { ImportService } from './import.service.js';
export { DeploymentService } from './deployment.service.js';
export { AppService } from './app.service.js';
export { LogService } from './log.service.js';
