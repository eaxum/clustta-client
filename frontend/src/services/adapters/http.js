// HTTP adapter - REST API implementations for web mode
// This file provides stub implementations that will be replaced with
// actual HTTP calls to your backend API

const API_BASE = import.meta.env.VITE_API_URL || 'https://api.clustta.com';

// Helper for API calls
async function apiCall(endpoint, method = 'GET', body = null) {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    method,
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
    },
    body: body ? JSON.stringify(body) : null,
    credentials: 'include',
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(error);
  }

  return response.json();
}

// Stub implementations - these return sensible defaults to allow the app to load
// Replace with actual HTTP implementations when connecting to your backend

export const SettingsService = {
  GetUseAltUrl: async () => false,
  IsProjectGridView: async () => true,
  IsShowUntrackedProjects: async () => true,
  GetLastStudio: async () => '',
  GetTheme: async () => 'dark',
  SetTheme: async (theme) => {},
  GetDefaultLocation: async () => ({ path: '' }),
  IsCompactView: async () => false,
  SetCompactView: async (compact) => {},
  IsShowHiddenFiles: async () => false,
  SetShowHiddenFiles: async (show) => {},
  GetProjectWorkspaces: async (projectId) => [],
  SetProjectWorkspaces: async (projectId, workspaces) => {},
  GetIgnoreList: async () => [],
  SetIgnoreList: async (list) => {},
};

export const AppService = {
  GetOS: async () => 'web',
  GetVersion: async () => '0.0.0-web',
  OpenURL: async (url) => window.open(url, '_blank'),
  GetAppDataDir: async () => '',
};

export const LogService = {
  LogInfo: async (message) => console.log('[INFO]', message),
  LogError: async (message) => console.error('[ERROR]', message),
  LogWarning: async (message) => console.warn('[WARN]', message),
  LogDebug: async (message) => console.debug('[DEBUG]', message),
};

export const AuthService = {
  Login: async (email, password) => apiCall('/auth/login', 'POST', { email, password }),
  Logout: async () => apiCall('/auth/logout', 'POST'),
  Register: async (data) => apiCall('/auth/register', 'POST', data),
  GetCurrentUser: async () => apiCall('/auth/me'),
  IsAuthenticated: async () => {
    try {
      await apiCall('/auth/me');
      return true;
    } catch {
      return false;
    }
  },
};

export const ProjectService = {
  GetProjects: async (studioName) => apiCall(`/studios/${studioName}/projects`),
  CreateProject: async (uri, studioName, workingDir, templateName) =>
    apiCall('/projects', 'POST', { uri, studioName, workingDir, templateName }),
  GetProject: async (projectId) => apiCall(`/projects/${projectId}`),
  DeleteProject: async (projectId) => apiCall(`/projects/${projectId}`, 'DELETE'),
  UpdateProject: async (projectId, data) => apiCall(`/projects/${projectId}`, 'PUT', data),
  GetProjectUsers: async (projectPath) => [],
  AddUser: async (projectPath, email, roleName) => ({}),
  RemoveUser: async (projectPath, userId) => {},
  ChangeRole: async (projectPath, userId, roleName) => {},
};

export const SyncService = {
  PullData: async (uri, url, force, options) => ({}),
  PushData: async (uri, url, options) => ({}),
  CloneProject: async (projectUri, studioName, workingDir, options) => ({}),
  CancelSync: async () => {},
};

export const StudioService = {
  GetStudios: async () => apiCall('/studios'),
  CreateStudio: async (name) => apiCall('/studios', 'POST', { name }),
  GetStudio: async (studioId) => apiCall(`/studios/${studioId}`),
  DeleteStudio: async (studioId) => apiCall(`/studios/${studioId}`, 'DELETE'),
  GetStudioUsers: async (studioName) => [],
  AddStudioUser: async (studioName, email, roleName) => ({}),
  RemoveStudioUser: async (studioName, userId) => {},
};

export const AssetService = {
  GetAssets: async (projectPath, entityPath) => [],
  CreateAsset: async (projectPath, asset) => ({}),
  UpdateAsset: async (projectPath, asset) => ({}),
  DeleteAsset: async (projectPath, assetId) => {},
  GetAsset: async (projectPath, assetId) => ({}),
  MoveAsset: async (projectPath, assetId, newPath) => {},
  CopyAsset: async (projectPath, assetId, newPath) => {},
};

export const CollectionService = {
  GetCollections: async (projectPath) => [],
  CreateCollection: async (projectPath, collection) => ({}),
  UpdateCollection: async (projectPath, collection) => ({}),
  DeleteCollection: async (projectPath, collectionId) => {},
  GetCollection: async (projectPath, collectionId) => ({}),
};

export const CheckpointService = {
  GetCheckpoints: async (projectPath, taskId) => [],
  CreateCheckpoint: async (projectPath, checkpoint) => ({}),
  RestoreCheckpoint: async (projectPath, checkpointId) => {},
  DeleteCheckpoint: async (projectPath, checkpointId) => {},
};

export const TagService = {
  GetTags: async (projectPath) => [],
  CreateTag: async (projectPath, tag) => ({}),
  UpdateTag: async (projectPath, tag) => ({}),
  DeleteTag: async (projectPath, tagId) => {},
};

export const StatusService = {
  GetStatuses: async (projectPath) => [],
  CreateStatus: async (projectPath, status) => ({}),
  UpdateStatus: async (projectPath, status) => ({}),
  DeleteStatus: async (projectPath, statusId) => {},
};

export const TemplateService = {
  GetTemplates: async (projectPath) => [],
  CreateTemplate: async (projectPath, template) => ({}),
  UpdateTemplate: async (projectPath, template) => ({}),
  DeleteTemplate: async (projectPath, templateId) => {},
};

export const UserService = {
  GetUsers: async (projectPath) => [],
  GetUser: async (projectPath, userId) => ({}),
  UpdateUser: async (projectPath, user) => ({}),
};

export const WorkflowService = {
  GetWorkflows: async (projectPath) => [],
  CreateWorkflow: async (projectPath, workflow) => ({}),
  UpdateWorkflow: async (projectPath, workflow) => ({}),
  DeleteWorkflow: async (projectPath, workflowId) => {},
};

export const FSService = {
  WatchPath: async (path) => {},
  UnwatchPath: async (path) => {},
  GetFileInfo: async (path) => ({}),
  ReadDirectory: async (path) => [],
  OpenFile: async (path) => {},
  OpenInExplorer: async (path) => {},
};

export const DialogService = {
  OpenFile: async (options) => null,
  OpenDirectory: async (options) => null,
  SaveFile: async (options) => null,
  ShowMessage: async (options) => {},
};

export const ClipboardService = {
  WriteText: async (text) => navigator.clipboard.writeText(text),
  ReadText: async () => navigator.clipboard.readText(),
};

export const TrashService = {
  GetTrashedItems: async (projectPath) => [],
  RestoreItem: async (projectPath, itemId) => {},
  DeletePermanently: async (projectPath, itemId) => {},
  EmptyTrash: async (projectPath) => {},
};

export const ImportService = {
  ImportFiles: async (projectPath, files) => [],
  ImportFolder: async (projectPath, folderPath) => [],
};

export const ProfileService = {
  GetProfile: async () => ({}),
  UpdateProfile: async (profile) => ({}),
  UploadPhoto: async (photo) => '',
};

export const AccountService = {
  GetAccount: async () => ({}),
  UpdateAccount: async (account) => ({}),
  DeleteAccount: async () => {},
};

export const DependencyTypeService = {
  GetDependencyTypes: async (projectPath) => [],
  CreateDependencyType: async (projectPath, type) => ({}),
  UpdateDependencyType: async (projectPath, type) => ({}),
  DeleteDependencyType: async (projectPath, typeId) => {},
};

export const DeploymentService = {
  Deploy: async (projectPath, options) => ({}),
  GetDeployments: async (projectPath) => [],
};
