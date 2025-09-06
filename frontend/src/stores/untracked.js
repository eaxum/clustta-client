import { defineStore } from "pinia";

export const useUntrackedItemStore = defineStore("untrackedItem", {
  state: () => ({
    selectedUntrackedItem: null,
  }),
  getters: {},
  actions: {
    selectUntrackedItem(item) {
      this.selectedUntrackedItem = item;
    },
  },
});
