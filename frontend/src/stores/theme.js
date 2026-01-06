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
      await SettingsService.GetTheme().then((response) => {
        this.isDarkMode = response === 'dark';
      })
        this.applyTheme();
    }
},

getters: {
    currentTheme: (state) => state.isDarkMode ? 'dark' : 'light'
}
});
