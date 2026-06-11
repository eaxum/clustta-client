import { useCollectionStore } from '@/stores/collections';
import { useUserStore } from '@/stores/users';

// True when the user can modify the given collection.
// Reads navigatedCollection.can_modify, which the backend stamps on every fetch.
export const canModifyCollection = (collectionId) => {
  if (!collectionId) return false;
  const collectionStore = useCollectionStore();
  const navigated = collectionStore.navigatedCollection;
  if (navigated?.id === collectionId) return !!navigated.can_modify;
  return false;
};

// Gate for actions on a tracked asset. Unrestricted users (view_asset role) are
// bound only by role permissions; scoped users must be in collaborator scope.
export const canActOnAsset = (action, asset) => {
  const userStore = useUserStore();
  if (userStore.canDo('view_asset')) return userStore.canDo(action);
  return canModifyCollection(asset?.collection_id);
};

// Gate for bulk/contextual actions in the currently navigated collection.
// At project root, only unrestricted users with the role permission pass.
export const canActInNavigatedCollection = (action) => {
  const collectionStore = useCollectionStore();
  const userStore = useUserStore();
  if (userStore.canDo('view_asset')) return userStore.canDo(action);
  return !!collectionStore.navigatedCollection?.can_modify;
};

// Gate for actions scoped to a specific collection object.
// Unrestricted users (view_asset) pass on role; scoped users need can_modify.
export const canActInCollection = (action, collection) => {
  const userStore = useUserStore();
  if (userStore.canDo('view_asset')) return userStore.canDo(action);
  return !!collection?.can_modify;
};

// Gate for creating a new tracked asset (new asset, untracked → asset, web link).
// Requires the create_asset role regardless of collaborator scope.
export const canCreateAssetHere = () => {
  const userStore = useUserStore();
  return userStore.canDo('create_asset') && canActInNavigatedCollection('create_asset');
};

// Same as canCreateAssetHere but scoped to a specific collection object.
export const canCreateAssetInCollection = (collection) => {
  const userStore = useUserStore();
  return userStore.canDo('create_asset') && canActInCollection('create_asset', collection);
};

// Gate for creating a new collection in the navigated collection.
// Requires the create_collection role regardless of collaborator scope.
export const canCreateCollectionHere = () => {
  const userStore = useUserStore();
  return userStore.canDo('create_collection') && canActInNavigatedCollection('create_collection');
};

export const permissionGroups = {
  assets: ['view_asset', 'create_asset', 'update_asset', 'delete_asset', 'manage_dependencies'],
  assignation: ['assign_asset', 'unassign_asset'],
  collections: ['view_collection', 'create_collection', 'update_collection', 'delete_collection'],
  users: ['add_user', 'remove_user', 'change_role'],
  status: ['view_done_asset', 'change_status', 'set_done_asset', 'set_retake_asset'],
  templates: ['view_template', 'create_template', 'update_template', 'delete_template'],
  checkpoints: ['view_checkpoint', 'create_checkpoint', 'delete_checkpoint', 'pull_chunk'],
  sharing: ['manage_share_links'],
};

// Formats a permission key to a display label.
export const formatLabel = (key) => {
  return key.replace(/_/g, ' ')
    .split(' ')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
};

// Returns a summary of active permission groups for a role.
export const getPermissionSummary = (role) => {
  const activeGroups = [];
  for (const [groupName, permissions] of Object.entries(permissionGroups)) {
    const activeCount = permissions.filter(p => role[p] === true).length;
    if (activeCount > 0) {
      activeGroups.push(`${formatLabel(groupName)} (${activeCount})`);
    }
  }
  return activeGroups.join(', ');
};
