const MODIFIER_KEY = 'Mod';

export const SHORTCUTS = Object.freeze({
  cancel: ['Esc'],
  confirm: ['Enter'],
  delete: ['Shift', 'Del'],
  duplicate: [MODIFIER_KEY, 'D'],
  edit: [MODIFIER_KEY, 'F2'],
  freeUpSpace: ['Del'],
  newAsset: [MODIFIER_KEY, 'T'],
  newCollection: [MODIFIER_KEY, 'K'],
  newLink: [MODIFIER_KEY, 'L'],
  refresh: ['F5'],
  rename: ['F2'],
});

export const getShortcutKeys = (shortcut, isMac = false) => {
  const keys = Array.isArray(shortcut) ? shortcut : SHORTCUTS[shortcut];
  if (!keys) return [];

  return keys.map(key => key === MODIFIER_KEY ? (isMac ? 'Cmd' : 'Ctrl') : key);
};
