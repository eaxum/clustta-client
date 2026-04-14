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
