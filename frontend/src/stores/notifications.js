import { defineStore } from "pinia";
import { Events } from "@wailsio/runtime";
import { LogService } from "@/services";
import { useProjectStore } from "./projects";

export const useNotificationStore = defineStore("notifications", {
  state: () => ({
    notifications: [],
    progress: {
      running: false,
      title: "",
      message: "",
      percentage: 0,
      current: 0,
      total: 0,
      extra_message: "",
      isMinimized: false,
      operationType: null, // 'read' or 'write'
    },
    activeWriteOperation: null,
    cancleFunction: null,
    canCancel: false,
  }),
  getters: {
    getNotifications: (state) => state.notifications,
    getProgress: (state) => state.progress,
    canStartWriteOperation: (state) => {
      // Can start write operation if no write operation is currently active
      return state.activeWriteOperation === null;
    },
  },
  actions: {
    addNotification(message, longMessage, type, hasUndo = false) {
      let notification = {
        message: message,
        longMessage: longMessage,
        type: type,
        hasUndo: hasUndo,
        read: false,
      };
      this.notifications.unshift(notification);
      if (type === "error") {
        this.resetProgress();
        LogService.LogError(longMessage);
      }
      Events.Emit("add_message", notification);

      const projectStore = useProjectStore();
      projectStore.refreshActiveProject();
    },
    errorNotification(defaultMessage, error) {
      // remove "Error calling method:" from beginning of error message if present
      let errorMesage;
      if (!error) {
        errorMesage = defaultMessage || 'An unknown error occurred';
      } else if (typeof error === "string" || error instanceof String) {
        errorMesage = error;
      } else {
        errorMesage = error.message || error.toString() || defaultMessage || 'An unknown error occurred';
      }

      if (errorMesage.startsWith("Error calling method:")) {
        errorMesage = errorMesage.replace("Error calling method: ", "");
      }
      this.resetProgress();
      console.log(errorMesage);
      console.log(errorMesage === "cancelled");
      if (errorMesage === "cancelled") {
        return;
      }
      if (errorMesage === "database is locked") {
        errorMesage = "project is busy"
      }

      if (errorMesage.length < 100) {
        this.addNotification(errorMesage, errorMesage, "error");
      } else {
        this.addNotification(defaultMessage, errorMesage, "error");
      }

      const projectStore = useProjectStore();
      projectStore.refreshActiveProject();
      LogService.LogError(errorMesage.long_message);
    },
    updateProgress(progressData) {
      let current = progressData.current;
      let total = progressData.total;
      let message = progressData.message;
      let title = progressData.title;
      let extra_message = progressData.extra_message;
      let operation_type = progressData.operation_type;

      let percentage = progressData.percentage;
      if (total === 0 || (current === total && percentage == 100)) {
        this.resetProgress();
      } else {
        this.progress.running = true;
        this.progress.title = title;
        this.progress.message = message;
        this.progress.percentage = percentage;
        this.progress.current = current;
        this.progress.total = total;
        this.progress.extra_message = extra_message;
        this.progress.operationType = operation_type;
        
        // Track active write operations
        if (operation_type === 'write') {
          this.activeWriteOperation = title;
        }
      }
    },
    minimizeProgress() {
      this.progress.isMinimized = true;
    },
    restoreProgress() {
      this.progress.isMinimized = false;
    },
    resetProgress() {
      this.progress = {
        running: false,
        title: "",
        message: "",
        percentage: 0,
        current: 0,
        total: 0,
        extra_message: "",
        isMinimized: false,
        operationType: null,
      };
      this.activeWriteOperation = null;

      const projectStore = useProjectStore();
      projectStore.refreshActiveProject();
    },
  },
});
