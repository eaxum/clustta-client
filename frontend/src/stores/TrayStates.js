// imports
import router from "@/router";
import { nextTick } from "vue";

// stores
import { defineStore } from "pinia";
import { useUserStore } from "@/stores/users";
import { useCollectionStore } from "./collections";
import { useAssetStore } from "@/stores/assets";
import { useTemplateStore } from "./template";
import { useDependencyStore } from "./dependency";
import { useStatusStore } from "./status";
import { useTagStore } from "@/stores/tags";
import { useProjectStore } from "@/stores/projects";
import { Events } from "@wailsio/runtime";
import { useWorkflowStore } from "./workflow";

export const useTrayStates = defineStore("useTrayStates", {
  state: () => ({
    pin: false,
    userPin: false,
    screenshot: null,
    previewFile: "",
    previewFullPath: "",

    itemTags: [],

    icons: {},
    queuedTasks: [],

    suggestedTags: [
      "todo",
      "wip",
      "done",
      "retake",
      "logo",
      "Animation",
      "bottles",
      "cars",
      "VFX",
      "Mixing",
    ],

    flyoutTop: 0,
    flyoutLeft: 0,
    flyoutWidth: 0,
    listItemsAnchor: 0,
    listItemsLeft: 0,
    listItemsWidth: 0,
    listItemMaxHeight: 0,
    listItemsBoundary: null,

    appsContainer: null,
    selectedApp: null,
    isHoveringText: false,
    isPopupOpen: false,
    pinSuggestions: false,
    showTags: false,

    popUpModalTitle: "",

    createMultipleCheckpoints: true,
    createMultipleCheckpointsEntityPath: "",
    assignTo: "",

    trashTypes: [
      { name: "all", icon: "four-squares" },
      { name: "collections", icon: "folder" },
      { name: "task", icon: "brush" },
      { name: "task_checkpoint", icon: "layers" },
      { name: "resource_checkpoint", icon: "layers" },
      { name: "resource", icon: "paperclip" },
      { name: "template", icon: "file" },
    ],
    trashables: [],

    checkpointsLoaded: false,
    showStatusOptions: false,

    tagSearchQuery: "",
    trashSearchQuery: "",

    autoStart: true,
    autoPaste: true,
    keepModalOpen: false,

    popUpModalTitle: "",
    popUpModalMessage: "",
    popUpModalIcon: "",
    popUpModalFunction: null,
    popUpModalInputValue: null,
    popUpModalPlaceholder: "",
    popUpModalButtons: ['Cancel', 'Confirm'],
    usePopUpModalInput: false,

    activeModal: null,

    showTraySearch: false,
    showMeta: false,
    undoItemId: "",
    undoMultipleItemIds: [],
    undoFunction: null,
  }),
  getters: {
    getUser: (state) => state.user,

    getAllProgress: (state) => state.progress,
  },
  actions: {
    async togglePin(user = false) {
      let eventData = new Events.WailsEvent("pin-tray-window", !this.pin);
      Events.Emit(eventData);
      this.pin = !this.pin;
      if (user) {
        this.userPin = !this.userPin;
      }
    },
    async refreshData() {
      const userStore = useUserStore();
      const collectionStore = useCollectionStore();
      const assetStore = useAssetStore();
      const templateStore = useTemplateStore();
      const workflowStore = useWorkflowStore();
      const dependencyStore = useDependencyStore();
      const statusStore = useStatusStore();
      const tagStore = useTagStore();
      const projectStore = useProjectStore();

      
      // await new Promise((r) => setTimeout(r, 5000));
      // console.time("loading_general_data");
      await userStore.reloadUsers();
      await collectionStore.reloadCollectionTypes();
      await assetStore.reloadAssetTypes();
      // await collectionStore.reloadCollections();
      await templateStore.reloadTemplates();
      await workflowStore.reloadWorkflows();
      await statusStore.reloadStatuses();
      await dependencyStore.reloadDependencyTypes();
      await tagStore.reloadTags();
      // await assetStore.reloadAssets();
      // await projectStore.reloadUntrackedItems();
      // console.timeEnd("loading_general_data");
    },

    navigateToPage(page) {
      router.push({ name: page });
      this.showTraySearch = false;
    },

    navigateBack() {
      router.go(-1);
    },

    changeSearchVisibility() {
      this.showTraySearch = !this.showTraySearch;
    },
    changeMetaVisibility() {
      this.showMeta = !this.showMeta;
    },

    handleHover(event) {
      let element = event.target;
      const elementChild = event.target.children[0];
      elementChild.style.overflow = "";
      elementChild.style.textOverflow = "";

      nextTick(() => {
        const isOverflowing = element.scrollWidth > element.offsetWidth;
        const scrollDist = element.scrollWidth - element.offsetWidth;
        if (isOverflowing) {
          //
          elementChild.style.transform = "translateX(" + -scrollDist + "px)";
          elementChild.style.transition = scrollDist / 12 + "s linear";
        }
      });
    },
    resetScroll(event) {
      let element = event.target;
      const elementChild = event.target.children[0];
      elementChild.style.transform = "translateX(0px)";
      elementChild.style.transition = 0 + "s linear";
      elementChild.style.overflow = "hidden";
      elementChild.style.textOverflow = "ellipsis";
    },

    togglePinSuggestions() {
      this.pinSuggestions = !this.pinSuggestions;
    },

    toggleShowTags() {
      this.showTags = !this.showTags;
    },

    removeTag(tag) {
      ////console.log(this.itemTags);
      // this.tags = this.tags.filter((t) => t !== tag);
    },

    toggleKeepModal() {
      this.keepModalOpen = !this.keepModalOpen;
    },
  },
});