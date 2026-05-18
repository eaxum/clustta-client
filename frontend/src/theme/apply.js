// Applies a generated palette to the document root and keeps it in sync with
// the OS color-scheme preference when mode is "system".

import { buildPalette, resolveMode } from './palette';

let mql = null;
let systemChangeHandler = null;

// Writes every CSS variable from `palette` onto `<html>` and tags the root
// with data-mode / data-tint for any selector that needs to react.
function writeVars(palette, mode, tint) {
  const root = document.documentElement;
  for (const key in palette) {
    root.style.setProperty(key, palette[key]);
  }
  root.dataset.mode = mode;
  root.dataset.tint = tint;
  // Legacy attribute kept for any third-party CSS still keying off it.
  if (mode === 'dark') root.setAttribute('data-theme', 'dark');
  else root.removeAttribute('data-theme');
}

// Applies the theme for the given preference. If mode is "system" this also
// installs a listener so the palette follows OS-level scheme changes.
export function applyTheme({ mode = 'system', tint = 'neutral' } = {}) {
  const resolved = resolveMode(mode);
  writeVars(buildPalette({ mode: resolved, tint }), resolved, tint);

  if (typeof matchMedia !== 'function') return;
  if (mql && systemChangeHandler) {
    mql.removeEventListener('change', systemChangeHandler);
    systemChangeHandler = null;
  }
  if (mode === 'system') {
    mql = matchMedia('(prefers-color-scheme: dark)');
    systemChangeHandler = () => {
      const next = resolveMode('system');
      writeVars(buildPalette({ mode: next, tint }), next, tint);
    };
    mql.addEventListener('change', systemChangeHandler);
  }
}
