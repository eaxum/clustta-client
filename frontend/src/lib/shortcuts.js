const MODIFIER_KEY = 'Mod';

export const SHORTCUTS = Object.freeze({
  cancel: ['Esc'],
  buildWithDependencies: [MODIFIER_KEY, 'Shift', 'F'],
  collapseAll: [MODIFIER_KEY, 'Shift', 'H'],
  confirm: ['Enter'],
  copy: [MODIFIER_KEY, 'C'],
  cut: [MODIFIER_KEY, 'X'],
  delete: ['Shift', 'Del'],
  duplicate: [MODIFIER_KEY, 'D'],
  edit: [MODIFIER_KEY, 'F2'],
  fetch: [MODIFIER_KEY, 'F'],
  freeUpSpace: ['Del'],
  navigateUp: ['Backspace'],
  newAsset: [MODIFIER_KEY, 'T'],
  newCollection: [MODIFIER_KEY, 'K'],
  newLink: [MODIFIER_KEY, 'L'],
  newProject: [MODIFIER_KEY, 'N'],
  refresh: ['F5'],
  removeProject: ['Del'],
  rename: ['F2'],
  sync: [MODIFIER_KEY, 'Alt', 'S'],
  toggleExtensions: [MODIFIER_KEY, 'E'],
  togglePathVisibility: [MODIFIER_KEY, 'P'],
  toggleUILock: [MODIFIER_KEY, 'U'],
});

export const getShortcutKeys = (shortcut, isMac = false) => {
  const keys = Array.isArray(shortcut) ? shortcut : SHORTCUTS[shortcut];
  if (!keys) return [];

  return keys.map(key => key === MODIFIER_KEY ? (isMac ? 'Cmd' : 'Ctrl') : key);
};
