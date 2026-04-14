/**
 * Utilities for handling sync conflict resolution.
 * Provides helpers for parent-child conflict relationships and recursive operations.
 */

/**
 * Builds the full path for a conflict item (collection_path + name + trailing slash for collections).
 * @param {Object} conflict - The conflict object with collection_path, name, and type
 * @returns {string} The full path of the conflict item
 */
export function buildConflictFullPath(conflict) {
  const basePath = conflict.collection_path || '';
  const name = conflict.name || '';
  
  if (conflict.type === 'collection') {
    // Collections need trailing slash since they can contain children
    return `${basePath}${name}/`;
  }
  // Assets don't need trailing slash
  return `${basePath}${name}`;
}

/**
 * Checks if a conflict item is a child of a potential parent conflict.
 * @param {Object} child - The potential child conflict
 * @param {Object} parent - The potential parent conflict (must be collection type)
 * @returns {boolean} True if child is nested under parent
 */
export function isChildOfConflict(child, parent) {
  if (parent.type !== 'collection') return false;
  if (child.local_id === parent.local_id) return false;
  
  const parentFullPath = buildConflictFullPath(parent);
  const childPath = child.collection_path || '';
  
  return childPath.startsWith(parentFullPath);
}

/**
 * Finds all child conflicts of a given parent collection conflict.
 * @param {Object} parent - The parent collection conflict
 * @param {Array} allConflicts - Array of all conflicts to search
 * @returns {Array} Array of child conflicts
 */
export function findChildConflicts(parent, allConflicts) {
  if (parent.type !== 'collection') return [];
  
  return allConflicts.filter(conflict => isChildOfConflict(conflict, parent));
}

/**
 * Filters conflicts to only include top-level items (those without a conflicted parent).
 * @param {Array} conflicts - Array of all conflicts
 * @returns {Array} Array of top-level conflicts only
 */
export function filterTopLevelConflicts(conflicts) {
  const collectionConflicts = conflicts.filter(c => c.type === 'collection');
  
  return conflicts.filter(conflict => {
    // Check if any collection conflict is a parent of this conflict
    const hasConflictedParent = collectionConflicts.some(
      collectionConflict => isChildOfConflict(conflict, collectionConflict)
    );
    return !hasConflictedParent;
  });
}

/**
 * Gets all conflict IDs that should be removed when a parent is resolved.
 * Includes the parent itself and all its children.
 * @param {Object} parent - The parent conflict being resolved
 * @param {Array} allConflicts - Array of all conflicts
 * @returns {Array} Array of local_ids to remove
 */
export function getConflictIdsToRemove(parent, allConflicts) {
  const childConflicts = findChildConflicts(parent, allConflicts);
  return [parent.local_id, ...childConflicts.map(c => c.local_id)];
}

/**
 * Prepares conflicts for batch merge operation (parent + all children).
 * @param {Object} parent - The parent conflict to merge
 * @param {Array} allConflicts - Array of all conflicts
 * @returns {Array} Array of conflicts to merge (parent + children)
 */
export function prepareRecursiveMergeConflicts(parent, allConflicts) {
  if (parent.type !== 'collection') {
    return [parent];
  }
  
  const childConflicts = findChildConflicts(parent, allConflicts);
  return [parent, ...childConflicts];
}

/**
 * Counts total items affected by a conflict resolution (parent + children).
 * @param {Object} conflict - The conflict being resolved
 * @param {Array} allConflicts - Array of all conflicts
 * @returns {number} Total count of affected items
 */
export function countAffectedItems(conflict, allConflicts) {
  if (conflict.type !== 'collection') return 1;
  
  const childConflicts = findChildConflicts(conflict, allConflicts);
  return 1 + childConflicts.length;
}

/**
 * Gets a summary description for a conflict resolution.
 * @param {Object} conflict - The conflict being resolved
 * @param {Array} allConflicts - Array of all conflicts
 * @param {string} action - 'rename' or 'merge'
 * @returns {string} Human-readable summary
 */
export function getResolutionSummary(conflict, allConflicts, action) {
  const total = countAffectedItems(conflict, allConflicts);
  const typeName = conflict.type === 'collection' ? 'Collection' : 'Asset';
  
  if (total === 1) {
    return `${typeName} "${conflict.name}" ${action === 'rename' ? 'renamed' : 'merged'}`;
  }
  
  const childCount = total - 1;
  const childWord = childCount === 1 ? 'child item' : 'child items';
  return `${typeName} "${conflict.name}" and ${childCount} ${childWord} ${action === 'rename' ? 'resolved' : 'merged'}`;
}
