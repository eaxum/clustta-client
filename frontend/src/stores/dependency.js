import { defineStore } from "pinia";
import { DependencyTypeService } from "@/../bindings/clustta/services";
import { useProjectStore } from "./projects";

export const useDependencyStore = defineStore("dependency", {
  state: () => ({
    dependency_types: [],
  }),
  getters: {
    getDependencyTypes: (state) => {
      const filteredDependencyTypes = state.dependency_types.filter(
        (dependency_type) => {
          return !dependency_type.trashed;
        }
      );
      const sortedDependencyTypes = filteredDependencyTypes.sort((a, b) => {
        return a.name.localeCompare(b.name);
      });
      return sortedDependencyTypes;
    },
  },
  actions: {
    async reloadDependencyTypes() {
      const projectStore = useProjectStore();
      this.dependency_types = await DependencyTypeService.GetDependencyTypes(
        projectStore.activeProject.uri
      );
    },
  },
});
