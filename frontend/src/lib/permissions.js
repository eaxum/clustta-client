import { useCollectionStore } from '@/stores/collections';
import { useDndStore } from '@/stores/dnd';
import { useUserStore } from '@/stores/users';

// True when the user can modify the given collection.
// Reads navigatedCollection.can_modify, which the backend stamps on every fetch.
export const canModifyCollection = (collectionId) => {
  if (!collectionId) return false;
  const collectionStore = useCollectionStore();
  const navigated = collectionStore.navigatedCollection;
  if (navigated?.id === collectionId) return !!navigated.can_modify;
  const selected = collectionStore.selectedCollection;
  if (selected?.id === collectionId) return !!selected.can_modify;
  const rendered = findRenderedCollection(collectionId);
  if (rendered?.id === collectionId) return !!rendered.can_modify;
  return false;
};

// Finds a currently rendered collection row by id.
export const findRenderedCollection = (collectionId) => {
  if (!collectionId) return null;
  const dndStore = useDndStore();
  return dndStore.allViewItems?.find((item) => {
    return item.id === collectionId && (item.type === 'collection' || item.collection_type_id);
  }) || null;
};

// Returns whether the current user is a project administrator.
export const isProjectAdmin = () => {
  const userStore = useUserStore();
  return userStore.user?.role?.name === 'admin';
};

// Returns whether checkpoint creation is blocked by another assignee.
export const isCheckpointBlockedByAssignment = (item) => {
  const userStore = useUserStore();
  return !!item?.assignee_id && item.assignee_id !== userStore.user?.id;
};

// Returns whether the current navigation context grants checkpoint scope.
const navigatedCollectionCanModify = () => {
  return !!useCollectionStore().navigatedCollection?.can_modify;
};

// Returns whether the item's parent collection grants checkpoint scope.
const itemParentCanModify = (item) => {
  const collectionId = item?.collection_id || item?.parent_id;
  return !!item?.can_modify
    || !!item?.parent?.can_modify
    || !!item?.collection?.can_modify
    || !!findRenderedCollection(collectionId)?.can_modify
    || !!useCollectionStore().findCollection(collectionId)?.can_modify
    || canModifyCollection(collectionId);
};

// Gate for creating checkpoints on one item. The create_checkpoint role is
// required; assignment locks and recursive collection assignment decide scope.
export const canCreateCheckpointForItem = (item) => {
  const userStore = useUserStore();
  if (!userStore.canDo('create_checkpoint')) return false;
  if (!item) return false;
  if (isCheckpointBlockedByAssignment(item)) return false;
  if (isProjectAdmin()) return true;

  if (item.type?.includes('untracked')) {
    return navigatedCollectionCanModify() || !!item.can_modify;
  }

  if (item.assignee_id === userStore.user?.id) return true;
  return itemParentCanModify(item);
};

// Gate for creating checkpoints from a collection-level action.
export const canCreateCheckpointInCollection = (collection) => {
  const userStore = useUserStore();
  const collectionStore = useCollectionStore();
  if (!userStore.canDo('create_checkpoint')) return false;
  if (isProjectAdmin()) return true;
  if (collection?.type?.includes('untracked')) {
    return !!collection.can_modify || !!collectionStore.navigatedCollection?.can_modify;
  }
  return !!collection?.can_modify;
};

// Gate for actions on a tracked asset. The role permission is mandatory;
// non-admin users additionally require direct assignment or parent collection scope.
export const canActOnAsset = (action, asset) => {
  const userStore = useUserStore();
  if (!userStore.canDo(action) || !asset) return false;
  if (isProjectAdmin()) return true;
  if (action !== 'create_asset' && asset.assignee_id === userStore.user?.id) return true;
  const collectionId = asset.collection_id || asset.parent_id;
  if (!collectionId) return true;
  return itemParentCanModify(asset);
};

// Bulk asset actions require permission and collection scope for every item.
export const canActOnAssets = (action, assets) => {
  return Array.isArray(assets)
    && assets.length > 0
    && assets.every(asset => canActOnAsset(action, asset));
};

// Gate for bulk/contextual actions in the currently navigated collection.
// Non-admin users require an actual collection with modify scope.
export const canActInNavigatedCollection = (action) => {
  const collectionStore = useCollectionStore();
  const userStore = useUserStore();
  if (!userStore.canDo(action)) return false;
  if (isProjectAdmin()) return true;
  if (!collectionStore.navigatedCollection) return false;
  return !!collectionStore.navigatedCollection?.can_modify;
};

// Gate for actions scoped to a specific collection object.
// The role permission is mandatory and scoped users also need can_modify.
export const canActInCollection = (action, collection) => {
  const userStore = useUserStore();
  if (!userStore.canDo(action) || !collection) return false;
  if (isProjectAdmin()) return true;
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
