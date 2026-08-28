import { defineStore } from "pinia";
import { AppService } from "@/services";
import utils from "@/services/utils";

export const usePlatformStore = defineStore("platform", {
  state: () => ({
    platform: import.meta.env.VITE_PLATFORM || 'desktop',
    os: '',
    channel: '',
    version: '',
    initialized: false,
  }),

  getters: {
    // Returns true if running in web mode
    isWeb: (state) => state.platform === 'web',

    // Returns true if running in desktop mode
    isDesktop: (state) => state.platform !== 'web',

    // Returns true if running on macOS
    isMac: (state) => state.os === 'darwin',

    // Returns true if running in the sandboxed Mac App Store build
    isMacAppStore: (state) => state.os === 'darwin' && state.channel === 'mas',

    // Returns true if running on Windows
    isWindows: (state) => state.os === 'windows',

    // Returns true if running on Linux
    isLinux: (state) => state.os === 'linux',
  },

  actions: {
    // Initializes the platform store with OS and version info
    async initialize() {
      if (this.initialized) return;

      try {
        this.os = await AppService.GetOS();
        this.version = await utils.getClusttaVersion();
        this.channel = this.isDesktop ? await AppService.GetChannel() : 'web';
      } catch (error) {
        console.warn('Failed to get platform info:', error);
      }

      this.initialized = true;
    },
  },
});
