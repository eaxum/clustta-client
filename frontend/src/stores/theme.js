import { defineStore } from "pinia";
import {
  SettingsService,
} from "@/services";

let defaultTheme = "dark";

await SettingsService.GetTheme()
  .then((response) => {
    defaultTheme = response;
  })
  .catch((error) => console.log(error));

export const useThemeStore = defineStore("theme", {
  state: () => ({
    isDarkMode: true,
    selectedTheme: defaultTheme,
    themeInitialized: false,
    themes: [
      "light",
      "dark",
    //   "auto",
    ],
  }),
  actions: {
    toggleTheme() {
        this.isDarkMode = !this.isDarkMode;
        this.applyTheme();
    },
    
    applyTheme() {
        if (this.isDarkMode) {
            document.documentElement.setAttribute('data-theme', 'dark');
        } else {
            document.documentElement.removeAttribute('data-theme');
        }
    },
    
    async initializeTheme() {
      if (this.themeInitialized) return;
      await SettingsService.GetTheme().then((response) => {
        this.isDarkMode = response !== 'light';
      }).catch(() => {
        this.isDarkMode = true; // Default to dark mode
      })
        this.applyTheme();
        this.themeInitialized = true;
    }
},

getters: {
    currentTheme: (state) => state.isDarkMode ? 'dark' : 'light'
}
});
