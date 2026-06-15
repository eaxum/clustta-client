import { defineStore } from "pinia";
import { AppService, UpdateService } from "@/services";
import { Browser } from "@wailsio/runtime";

// How often to re-check for updates while the app stays open (24 hours).
const CHECK_INTERVAL_MS = 24 * 60 * 60 * 1000;

export const useUpdateStore = defineStore("updates", {
  state: () => ({
    currentVersion: "",
    channel: "direct",
    updateInfo: null,
    checked: false,
    isChecking: false,
    intervalId: null,
  }),

  getters: {
    isUpdateAvailable: (state) => !!state.updateInfo?.available,
    isUpdateRequired: (state) => !!state.updateInfo?.required,
    latestVersion: (state) => state.updateInfo?.latest_version || "",
  },

  actions: {
    // Loads the running application version
    async initialize() {
      try {
        this.currentVersion = await AppService.GetVersion();
      } catch (error) {
        console.warn("Failed to get app version:", error);
      }
    },

    // Fetches the release manifest and stores the resulting update info
    async checkForUpdate() {
      this.isChecking = true;
      try {
        const info = await UpdateService.CheckForUpdate();
        this.updateInfo = info;
        if (info?.channel) this.channel = info.channel;
        this.checked = true;
        return info;
      } catch (error) {
        console.warn("Update check failed:", error);
        return null;
      } finally {
        this.isChecking = false;
      }
    },

    // Runs an immediate check then re-checks every 24 hours
    startAutoCheck() {
      if (this.intervalId) return;
      this.checkForUpdate();
      this.intervalId = setInterval(() => this.checkForUpdate(), CHECK_INTERVAL_MS);
    },

    // Stops the periodic update check
    stopAutoCheck() {
      if (!this.intervalId) return;
      clearInterval(this.intervalId);
      this.intervalId = null;
    },

    // Opens the update destination for the current channel
    async handleUpdateClick() {
      const info = this.updateInfo;
      if (!info || !info.available) return;
      if (info.target_url) {
        Browser.OpenURL(info.target_url);
      }
    },
  },
});
