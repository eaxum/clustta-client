import { defineStore } from "pinia";
import fileIconIndex from "@/data/fileIconIndex.json";
import webIconIndex from "@/data/webIconIndex.json";
import utils from "@/services/utils";
import { useTemplateStore } from "./template";
import {
  FSService,
  SettingsService,
} from "@/../bindings/clustta/services/index";

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
      let iconPath = fileIconIndex[ext];
      if (!iconPath) {
        iconPath = this.icons[ext];
        if (!iconPath) {
          let fileExt = "." + ext;
          let iconStr = await FSService.GetFileIcon(fileExt);
          iconPath = "data:image/png;base64," + iconStr;
          this.icons[ext] = "data:image/png;base64," + iconStr;
        }
      }
      return iconPath;
    },
    async getWebIcon(link) {
      let domainName = utils.getDomainName(link);
      let iconPath = webIconIndex[domainName];
      if (!iconPath) {
        let iconStr = await FSService.GetFileIcon(".html");
        iconPath = "data:image/png;base64," + iconStr;
        this.icons["html"] = "data:image/png;base64," + iconStr;
      }
      return iconPath;
    },
    getAppIcon(name) {
      return `/icons/${this.selectedIconType}/${name}.svg`;
    },
  },
});
