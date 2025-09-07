import { defineStore } from "pinia";
import { useIconStore } from "./icons";
import { useProjectStore } from "./projects";
import { useNotificationStore } from "./notifications";
import {
  SettingsService,
  ProjectService,
  TemplateService,
  TaskService,
  EntityService,
  SyncService,
  FSService,
} from "@/../bindings/clustta/services";

export const useProjectTemplateStore = defineStore("project_template", {
  state: () => ({
    projectTemplates: [],
    activeProjectTemplate: {},
    taskTemplates: [],
    assetTypes: [],
    collectionTypes: [],
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
    projectTemplateNames: (state) => {
      return state.projectTemplates.map((template) => template.name);
    },
    activeProjectTemplateName: (state) => {
      return state.activeProjectTemplate?.name;
    },
  },
  actions: {
    async loadProjectTemplates() {
      const notificationStore = useNotificationStore();
      await ProjectService.GetTemplates()
        .then(async (response) => {
          this.projectTemplates = response;
          if (this.projectTemplates.length > 0) {
            this.activeProjectTemplate = this.projectTemplates[0];
            await this.reloadProjectTemplate();
          }
        })
        .catch((error) => {
          notificationStore.errorNotification(
            "Error loading project templates",
            error
          );
        });
    },
    selectActiveProjectTemplate(templateName) {
      this.activeProjectTemplate = this.projectTemplates.find(
        (template) => template.name === templateName
      );
    },
    async reloadProjectTemplate() {
      const iconStore = useIconStore();
      let templates = await TemplateService.GetTemplates(
        this.activeProjectTemplate.uri
      );
      this.taskTemplates = [];
      for (let i = 0; i < templates.length; i++) {
        let template = templates[i];
        let extension = "";
        let iconPath = "";
        extension = templates[i].extension.toLowerCase().substring(1);
        iconPath = (await iconStore.getIcon(extension)) || "";
        template.icon = iconPath;
        this.taskTemplates.push(template);
      }

      let assetTypes = await TaskService.GetTaskTypes(
        this.activeProjectTemplate.uri
      );
      this.assetTypes = assetTypes;

      let collectionTypes = await EntityService.GetEntityTypes(
        this.activeProjectTemplate.uri
      );
      this.collectionTypes = collectionTypes;
    },
  },
});