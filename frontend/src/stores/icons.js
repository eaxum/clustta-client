import { defineStore } from "pinia";
import fileIconIndex from "@/data/fileIconIndex.json";
import webIconIndex from "@/data/webIconIndex.json";
import utils from "@/services/utils";
import { useTemplateStore } from "./template";
import { usePlatformStore } from "./platform";
import {
  FSService,
  SettingsService,
} from "@/services";

let defaultIconScheme = "solid";
await SettingsService.GetIconScheme()
  .then((response) => {
    defaultIconScheme = response;
  })
  .catch((error) => console.log(error));

export const useIconStore = defineStore("icons", {
  state: () => ({
    icons: {},
    appsWithIcons: [],
    selectedIconType: defaultIconScheme,
    iconTypes: [
      "solid",
      // 'monochrome',
      "outline",
    ],
  }),
  getters: {},
  actions: {
    async fetchAppIcons() {
      const templateStore = useTemplateStore();
      const iconDirectory = "/file-icons/";
      const apps = templateStore.getTemplates;
      // Clear existing icons before fetching new ones (optional)
      this.appsWithIcons.splice(0, this.appsWithIcons.length);

      for (const app of apps) {
        const extension = app.extension.toLowerCase().substring(1);
        const iconPath = (await this.getIcon(extension)) || "";

        this.appsWithIcons.push({
          ...app,
          id: app.id,
          icon: iconPath ? `${iconPath}` : "",
        });
      }
    },

    async getIcon(ext) {
      const platformStore = usePlatformStore();
      let iconPath = fileIconIndex[ext];
      if (!iconPath) {
        iconPath = this.icons[ext];
        if (!iconPath) {
          // In web mode, FSService isn't available - use default icon
          if (platformStore.isWeb) {
            const videoExts = ['mp4', 'mov', 'avi', 'mkv', 'wmv', 'flv', 'webm', 'm4v'];
            const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'svg', 'webp', 'ico', 'tiff', 'tif'];
            const audioExts = ['mp3', 'wav', 'flac', 'aac', 'ogg', 'wma', 'm4a', 'aiff'];
            
            if (videoExts.includes(ext.toLowerCase())) {
              iconPath = '/file-icons/video.svg';
            } else if (imageExts.includes(ext.toLowerCase())) {
              iconPath = '/file-icons/image.svg';
            } else if (audioExts.includes(ext.toLowerCase())) {
              iconPath = '/file-icons/music.svg';
            } else {
              iconPath = '/file-icons/default.svg';
            }
            this.icons[ext] = iconPath;
          } else {
            let fileExt = "." + ext;
            let iconStr = await FSService.GetFileIcon(fileExt);
            iconPath = "data:image/png;base64," + iconStr;
            this.icons[ext] = "data:image/png;base64," + iconStr;
          }
        }
      }
      return iconPath;
    },
    async getWebIcon(link) {
      let domainName = utils.getDomainName(link);
      let iconPath = webIconIndex[domainName];
      if (!iconPath) {
        iconPath = '/file-icons/default.svg';
        this.icons["html"] = iconPath;
      }
      return iconPath;
    },
    getAppIcon(name) {
      return `/icons/${this.selectedIconType}/${name}.svg`;
    },

    // Reload icon scheme from user settings (used on account switch)
    async reloadIconScheme() {
      try {
        const scheme = await SettingsService.GetIconScheme();
        this.selectedIconType = scheme || 'solid';
      } catch (error) {
        console.error('Failed to reload icon scheme:', error);
        this.selectedIconType = 'solid';
      }
    },
  },
});
