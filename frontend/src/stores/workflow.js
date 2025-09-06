import { defineStore } from "pinia";
import { useIconStore } from "./icons";
import { WorkflowService } from "@/../bindings/clustta/services";
import { useProjectStore } from "./projects";

export const useWorkflowStore = defineStore("workflow", {
  state: () => ({
    workflows: [],
    lastUsedWorkflow: null,
    selectedWorkflow: null,
    selectedWorkflowName: null,
  }),
  getters: {
    getWorkflows: (state) => {
      const sortedWorkflows = state.workflows.sort((a, b) => {
        return a.name.localeCompare(b.name);
      });
      return sortedWorkflows;
    },
  },
  actions: {
    async reloadWorkflows() {
      const projectStore = useProjectStore();
      let workflows = await WorkflowService.GetWorkflows(
        projectStore.activeProject.uri
      );
      this.workflows = workflows;
    },
  },
});
