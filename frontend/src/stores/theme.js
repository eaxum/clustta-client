import { defineStore } from "pinia";
import { SettingsService } from "@/services";
import { applyTheme } from "@/theme/apply";
import { MODES, TINT_NAMES, resolveMode } from "@/theme/palette";

// Theme store: tracks `mode` (system|light|dark) and `tint` (neutral|blue|...)
// and applies the resulting palette to the document root.
// Backward-compatible getters/actions (isDarkMode, currentTheme, themes,
// selectedTheme, toggleTheme, applyTheme, initializeTheme) are preserved so
// existing components do not need to change.
export const useThemeStore = defineStore("theme", {
  state: () => ({
    mode: "system",
    tint: "neutral",
    initialized: false,
  }),

  getters: {
    // Resolved mode after taking OS preference into account.
    resolvedMode: (state) => resolveMode(state.mode),
    availableModes: () => MODES,
    availableTints: () => TINT_NAMES,

    // --- Backward-compatible getters (deprecated) ---
    isDarkMode: (state) => resolveMode(state.mode) === "dark",
    currentTheme: (state) => state.mode,
    selectedTheme: (state) => state.mode,
    themes: () => MODES,
  },

  actions: {
    // Re-applies the current palette to <html>.
    apply() {
      applyTheme({ mode: this.mode, tint: this.tint });
    },

    // Sets and persists the theme mode.
    async setMode(mode) {
      if (!MODES.includes(mode)) return;
      this.mode = mode;
      this.apply();
      try { await SettingsService.SetTheme(mode); } catch (e) { console.error(e); }
    },

    // Sets and persists the theme tint.
    async setTint(tint) {
      if (!TINT_NAMES.includes(tint)) return;
      this.tint = tint;
      this.apply();
      try { await SettingsService.SetThemeTint(tint); } catch (e) { console.error(e); }
    },

    // Loads stored preferences and applies them. Safe to call multiple times.
    async initialize() {
      if (this.initialized) return;
      try {
        const storedMode = await SettingsService.GetTheme();
        if (storedMode && MODES.includes(storedMode)) this.mode = storedMode;
      } catch (e) { /* keep default */ }
      try {
        if (typeof SettingsService.GetThemeTint === "function") {
          const storedTint = await SettingsService.GetThemeTint();
          if (storedTint && TINT_NAMES.includes(storedTint)) this.tint = storedTint;
        }
      } catch (e) { /* keep default */ }
      this.apply();
      this.initialized = true;
    },

    // --- Backward-compatible actions (deprecated) ---
    applyTheme() { this.apply(); },
    async initializeTheme() { await this.initialize(); },
    toggleTheme() {
      const next = this.resolvedMode === "dark" ? "light" : "dark";
      this.setMode(next);
    },
  },
});
