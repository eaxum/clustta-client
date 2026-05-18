// Palette generator for Clustta theming.
// Produces a flat map of CSS custom properties from a tint config and mode.
// Tint controls hue/chroma; mode controls the lightness ramp.
// All values use OKLCH for perceptual uniformity across tints.

// Lightness stops per mode. Index 0 = deepest background, 5 = highest elevation.
const RAMPS = {
  light: {
    surface:     [0.985, 0.97, 0.94, 0.90, 0.85, 0.78],
    text:        0.18,
    textMuted:   0.42,
    border:      0.86,
    hoverAlpha:  'oklch(0 0 0 / 0.06)',
    accentL:     0.45,   // darker so white fg gets >= 4.5:1
    accentFgL:   0.99,
    iconInvert:  0,
  },
  dark: {
    surface:     [0.04, 0.07, 0.11, 0.15, 0.19, 0.26],
    text:        0.98,
    textMuted:   0.68,
    border:      0.22,
    hoverAlpha:  'oklch(1 0 0 / 0.06)',
    accentL:     0.72,   // brighter so dark fg gets >= 4.5:1
    accentFgL:   0.12,   // near-black foreground on bright accent
    iconInvert:  1,
  },
};

// Fixed semantic colors that should never be retinted (status indicators).
const FIXED = {
  '--danger':  'oklch(0.58 0.22 25)',
  '--warning': 'oklch(0.75 0.16 75)',
  '--alert':   'oklch(0.70 0.15 60)',
  '--success': 'oklch(0.62 0.16 145)',
  '--online':  'oklch(0.72 0.20 145)',
  '--info':    'oklch(0.65 0.15 235)',
};

// Available tints. `chroma` controls how strongly the surface ramp picks up the hue;
// keep it small (<= 0.025) so backgrounds stay neutral-looking.
// `accentChroma` controls the saturation of the accent (selection / primary action).
export const TINTS = {
  neutral: { hue: 264, chroma: 0.000, accentHue: 215, accentChroma: 0.18 },
  blue:    { hue: 240, chroma: 0.015, accentHue: 230, accentChroma: 0.20 },
  pink:    { hue: 350, chroma: 0.018, accentHue: 340, accentChroma: 0.20 },
  orange:  { hue:  55, chroma: 0.020, accentHue:  35, accentChroma: 0.18 },
  green:   { hue: 150, chroma: 0.015, accentHue: 145, accentChroma: 0.18 },
  purple:  { hue: 290, chroma: 0.018, accentHue: 280, accentChroma: 0.20 },
};

export const TINT_NAMES = Object.keys(TINTS);
export const MODES = ['system', 'light', 'dark'];

// Returns the effective mode given a stored preference.
export function resolveMode(mode) {
  if (mode === 'system') {
    return typeof matchMedia === 'function'
      && matchMedia('(prefers-color-scheme: dark)').matches
      ? 'dark' : 'light';
  }
  return mode === 'dark' ? 'dark' : 'light';
}

// Builds the full set of CSS variables for a given (mode, tint) pair.
// Returns a plain object of { '--var-name': 'value' } pairs.
export function buildPalette({ mode = 'light', tint = 'neutral' } = {}) {
  const ramp = RAMPS[mode] || RAMPS.light;
  const t = TINTS[tint] || TINTS.neutral;
  const h = t.hue;
  const c = t.chroma;
  const s = (l) => `oklch(${l} ${c} ${h})`;

  const palette = {
    // Surface ramp (background -> highest elevation)
    '--bg':         s(ramp.surface[0]),
    '--surface-1':  s(ramp.surface[1]),
    '--surface-2':  s(ramp.surface[2]),
    '--surface-3':  s(ramp.surface[3]),
    '--surface-4':  s(ramp.surface[4]),
    '--surface-5':  s(ramp.surface[5]),

    // Foreground
    '--text':         `oklch(${ramp.text} ${c * 0.3} ${h})`,
    '--text-muted':   `oklch(${ramp.textMuted} ${c * 0.5} ${h})`,
    '--text-inverse': `oklch(${1 - ramp.text} ${c * 0.3} ${h})`,

    // Lines / focus
    '--border':       `oklch(${ramp.border} ${c} ${h})`,
    '--border-strong':`oklch(${mode === 'dark' ? 0.35 : 0.70} ${c} ${h})`,

    // Accent (selection, primary action, focus ring)
    '--accent':       `oklch(${ramp.accentL} ${t.accentChroma} ${t.accentHue})`,
    '--accent-hover': `oklch(${ramp.accentL + (mode === 'dark' ? 0.05 : -0.05)} ${t.accentChroma} ${t.accentHue})`,
    '--accent-fg':    `oklch(${ramp.accentFgL} 0 0)`,

    // Interaction
    '--hover':        ramp.hoverAlpha,
    '--icon-invert':  String(ramp.iconInvert),
  };

  return { ...FIXED, ...palette };
}
