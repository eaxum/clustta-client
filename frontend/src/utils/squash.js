/**
 * Squash utility functions for combining multiple untracked files into a single asset.
 * Handles pattern detection, ordering, and validation.
 */

// Common version/numbering patterns (order matters - more specific first)
const PATTERNS = [
  { regex: /[_\-\s]?v(\d+)$/i,       label: 'version suffix (v1, v2)' },
  { regex: /^v(\d+)[_\-\s]/i,        label: 'version prefix (v1_, v2_)' },
  { regex: /[_\-\s]rev(\d+)$/i,      label: 'revision suffix (_rev1)' },
  { regex: /[_\-\s]r(\d+)$/i,        label: 'revision short (_r1)' },
  { regex: /[_\-\s](\d+)$/,          label: 'trailing number (_01)' },
  { regex: /^(\d+)[_\-\s]/,          label: 'leading number (01_)' },
  { regex: /[_\-\s](\d+)[_\-\s]?$/,  label: 'number suffix' },
  { regex: /^(\d+)$/,                label: 'bare number (001, 002)' },
];

/**
 * Formats a file name into a readable checkpoint comment.
 * Replaces underscores and hyphens with spaces, then applies sentence case.
 */
function formatNameAsComment(name) {
  const spaced = name.replace(/[_-]/g, ' ').replace(/\s+/g, ' ').trim();
  if (spaced.length === 0) return '';
  return spaced.charAt(0).toUpperCase() + spaced.slice(1).toLowerCase();
}

/**
 * Extracts version/sequence numbers from file names using common patterns.
 * Returns an object with the detected pattern, common base name, and ordered items.
 */
export function analyzeFileNames(items) {
  if (!items || items.length === 0) return { commonName: '', orderedItems: [], patternFound: false };

  const names = items.map(item => item.name || '');

  // Try each pattern against all names
  for (const pattern of PATTERNS) {
    const matches = names.map(name => pattern.regex.exec(name));
    const allMatch = matches.every(m => m !== null);

    if (allMatch) {
      // Extract the common base name by removing the matched portion
      const baseNames = names.map((name, i) => {
        const match = matches[i];
        return name.slice(0, match.index) + name.slice(match.index + match[0].length);
      });

      // Check if all base names are the same
      const uniqueBaseNames = [...new Set(baseNames.map(b => b.trim().replace(/[_\-\s]+$/, '').replace(/^[_\-\s]+/, '')))];

      if (uniqueBaseNames.length === 1) {
        const commonName = uniqueBaseNames[0];
        const numbers = matches.map(m => parseInt(m[1], 10));

        // Pair items with their extracted numbers and sort
        const paired = items.map((item, i) => ({ item, number: numbers[i] }));
        paired.sort((a, b) => a.number - b.number);

        return {
          commonName,
          orderedItems: paired.map(p => p.item),
          patternFound: true,
        };
      }
    }
  }

  // No consistent naming pattern found - order by file modified date (oldest first)
  const byDate = [...items].sort((a, b) => {
    const dateA = a.modified_at || a.created_at || 0;
    const dateB = b.modified_at || b.created_at || 0;
    return dateA - dateB;
  });

  return {
    commonName: '',
    orderedItems: byDate,
    patternFound: false,
  };
}

/**
 * Validates whether the selected items can be squashed.
 * Returns { valid: boolean, reason: string }.
 */
export function canSquash(selectedItems) {
  if (!selectedItems || selectedItems.length < 2) {
    return { valid: false, reason: 'Select at least 2 untracked files to squash.' };
  }

  if (selectedItems.length > 99) {
    return { valid: false, reason: 'Cannot squash more than 99 files.' };
  }

  // All must be untracked_task
  const allUntracked = selectedItems.every(item => item.type === 'untracked_task');
  if (!allUntracked) {
    return { valid: false, reason: 'All selected items must be untracked files.' };
  }

  // All must have the same extension
  const extensions = new Set(selectedItems.map(item => (item.extension || '').toLowerCase()));
  if (extensions.size !== 1) {
    return { valid: false, reason: 'All selected files must have the same extension.' };
  }

  // All must be siblings (same entity_id)
  const entityIds = new Set(selectedItems.map(item => item.entity_id || ''));
  if (entityIds.size !== 1) {
    return { valid: false, reason: 'All selected files must be in the same collection.' };
  }

  return { valid: true, reason: '' };
}

/**
 * Generates checkpoint labels for the ordered items.
 * When a pattern was found, labels are sequential (v01, v02...).
 * When no pattern was found, labels are formatted file names.
 * Returns an array of objects with { item, label, index }.
 */
export function generateCheckpointLabels(orderedItems, patternFound) {
  return orderedItems.map((item, i) => ({
    item,
    label: patternFound
      ? `v${String(i + 1).padStart(2, '0')}`
      : formatNameAsComment(item.name || ''),
    index: i,
  }));
}
