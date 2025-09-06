import { defineStore } from "pinia";
import { useIconStore } from "./icons";
import { TemplateService } from "@/../bindings/clustta/services";
import { useProjectStore } from "./projects";

export const useTemplateStore = defineStore("template", {
  state: () => ({
    templates: [],
    lastUsedTemplate: null,
    selectedTemplate: {},
    selectedTemplateName: 0,
  }),
  getters: {
    getTemplates: (state) => {
      const filteredTemplates = state.templates.filter((template) => {
        return !template.trashed;
      });
      const sortedTemplates = filteredTemplates.sort((a, b) => {
        return a.name.localeCompare(b.name);
      });
      return sortedTemplates;
    },
  },
  actions: {
    markTemplateAsDeleted(templateId) {
      for (let i = 0; i < this.templates.length; i++) {
        if (this.templates[i].id === templateId) {
          this.templates[i].trashed = true;
          break;
        }
      }
    },
    unmarkTemplateAsDeleted(templateId) {
      for (let i = 0; i < this.templates.length; i++) {
        if (this.templates[i].id === templateId) {
          this.templates[i].trashed = false;
          break;
        }
      }
    },
    async reloadTemplates() {
      const projectStore = useProjectStore();
      const iconStore = useIconStore();
      let templates = await TemplateService.GetTemplates(
        projectStore.activeProject.uri
      );
      this.templates = [];

      for (let i = 0; i < templates.length; i++) {
        let template = templates[i];
        let extension = "";
        let iconPath = "";
        extension = templates[i].extension.toLowerCase().substring(1);
        iconPath = (await iconStore.getIcon(extension)) || "";
        template.icon = iconPath;
        this.templates.push(template);
      }
    },
    getTaskTypeIcon(templateId) {
      for (let i = 0; i < this.templates.length; i++) {
        if (this.templates[i].id === templateId) {
          return this.templates[i].icon;
        }
      }
      return "";
    },
  },
});
