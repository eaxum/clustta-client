// Palette generator for Clustta theming.
// Builds a flat map of CSS custom properties from a tint config and mode.

// Lightness stops per mode. Index 0 = deepest background, 5 = highest elevation.
const ramps = {
  light: {
    surface:    [1.000, 0.985, 0.970, 0.950, 0.920, 0.880],
    text:       0.18,
    textMuted:  0.42,
    border:     0.90,
    hoverAlpha: 'oklch(0 0 0 / 0.05)',
    accentL:    0.45,
    accentFgL:  0.99,
    selectedAlpha:     0.70,
    selectedSoftAlpha: 0.22,
    iconInvert: 0,
  },
  dark: {
    surface:    [0.165, 0.200, 0.240, 0.285, 0.335, 0.400],
    text:       0.94,
    textMuted:  0.65,
    border:     0.28,
    hoverAlpha: 'oklch(1 0 0 / 0.05)',
    accentL:    0.72,
    accentFgL:  0.12,
    selectedAlpha:     0.70,
    selectedSoftAlpha: 0.22,
    iconInvert: 1,
  },
};

// Semantic colors that never retint (status indicators).
const fixed = {
  '--danger':  'oklch(0.58 0.22 25)',
  '--warning': 'oklch(0.75 0.16 75)',
  '--alert':   'oklch(0.70 0.15 60)',
  '--success': 'oklch(0.62 0.16 145)',
  '--online':  'oklch(0.72 0.20 145)',
  '--info':    'oklch(0.65 0.15 235)',
};

// Available tints. `chroma` tints surfaces (keep <= 0.025); `accentChroma`
// controls saturation of the accent. Optional `accent` overrides the ramp's
// accentL / accentFgL per mode.
export const tints = {
  default: {
    hue: 264, chroma: 0.000, accentHue: 255, accentChroma: 0.16,
    accent: {
      light: { l: 0.42, fgL: 0.99 },
      dark:  { l: 0.50, fgL: 0.99 },
    },
  },
  blue:    { hue: 240, chroma: 0.015, accentHue: 230, accentChroma: 0.20 },
  pink:    { hue: 350, chroma: 0.018, accentHue: 340, accentChroma: 0.20 },
};

export const TINT_NAMES = Object.keys(tints);
export const MODES = ['system', 'light', 'dark'];

// Backwards-compat re-export under the previous name.
export const TINTS = tints;

// Returns the effective mode, resolving "system" against prefers-color-scheme.
export function resolveMode(mode) {
  if (mode === 'system') {
    return typeof matchMedia === 'function'
      && matchMedia('(prefers-color-scheme: dark)').matches
      ? 'dark' : 'light';
  }
  return mode === 'dark' ? 'dark' : 'light';
}

// Builds the full set of CSS variables for a given (mode, tint) pair.
export function buildPalette({ mode = 'light', tint = 'default' } = {}) {
  const ramp = ramps[mode] || ramps.light;
  const t = tints[tint] || tints.default;
  const h = t.hue;
  const c = t.chroma;
  const s = (l) => `oklch(${l} ${c} ${h})`;
  const accentOverride = t.accent && t.accent[mode];
  const accentL = accentOverride?.l ?? ramp.accentL;
  const accentFgL = accentOverride?.fgL ?? ramp.accentFgL;

  const palette = {
    '--bg':           s(ramp.surface[0]),
    '--surface-1':    s(ramp.surface[1]),
    '--surface-2':    s(ramp.surface[2]),
    '--surface-3':    s(ramp.surface[3]),
    '--surface-4':    s(ramp.surface[4]),
    '--surface-5':    s(ramp.surface[5]),

    '--text':         `oklch(${ramp.text} ${c * 0.3} ${h})`,
    '--text-muted':   `oklch(${ramp.textMuted} ${c * 0.5} ${h})`,
    '--text-inverse': `oklch(${1 - ramp.text} ${c * 0.3} ${h})`,

    '--surface-inverse':       `oklch(${ramp.text} ${c * 0.3} ${h})`,
    '--surface-inverse-muted': `oklch(${ramp.textMuted} ${c * 0.5} ${h})`,

    '--skeleton':     `color-mix(in oklch, oklch(${ramp.text} ${c * 0.3} ${h}) 10%, transparent)`,

    '--border':        `oklch(${ramp.border} ${c} ${h})`,
    '--border-strong': `oklch(${mode === 'dark' ? 0.35 : 0.70} ${c} ${h})`,

    '--accent':       `oklch(${accentL} ${t.accentChroma} ${t.accentHue})`,
    '--accent-hover': `oklch(${accentL + (mode === 'dark' ? 0.05 : -0.05)} ${t.accentChroma} ${t.accentHue})`,
    '--accent-fg':    `oklch(${accentFgL} 0 0)`,

    '--selected':      `color-mix(in oklch, oklch(${accentL} ${t.accentChroma} ${t.accentHue}) ${ramp.selectedAlpha * 100}%, transparent)`,
    '--selected-soft': `color-mix(in oklch, oklch(${accentL} ${t.accentChroma} ${t.accentHue}) ${ramp.selectedSoftAlpha * 100}%, transparent)`,

    '--hover':        ramp.hoverAlpha,
    '--icon-invert':  String(ramp.iconInvert),
  };

  return { ...fixed, ...palette };
}
