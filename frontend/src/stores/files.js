import { defineStore } from "pinia";
import { useTrayStates } from "@/stores/TrayStates";

export const useFileStore = defineStore("files", {
  state: () => ({
    selectedFiles: [],
  }),
  getters: {},
  actions: {
    async selectFile({
      multiple = false,
      filters = [
        {
          name: "import files",
          extensions: ["*"],
        },
      ],
    }) {
      this.selectedFiles = [];
      const trayStates = useTrayStates();
      if (!trayStates.userPin) {
        await trayStates.togglePin();
      }

      // const result = await open({
      //   multiple: multiple,
      //   directory: false,
      //   filters: filters,
      // });
      if (result) {
        if (multiple) {
          for (let i = 0; i < results.length; i++) {
            let filePath = results[i].replace(/\\/g, "/");
            this.selectedFiles.push(filePath);
          }
        } else {
          let filePath = result.replace(/\\/g, "/");
          this.selectedFiles.push(filePath);
        }
      }

      if (!trayStates.userPin) {
        await trayStates.togglePin();
      }
    },
  },
});
