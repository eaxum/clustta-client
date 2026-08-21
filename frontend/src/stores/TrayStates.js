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
    queuedAssets: [],

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
    createMultipleCheckpointsCollectionPath: "",
    assignTo: "",

    trashTypes: [
      { name: "all", icon: "four-squares" },
      { name: "collections", icon: "folder" },
      { name: "asset", icon: "brush" },
      { name: "asset_checkpoint", icon: "layers" },
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
    popUpModalLoading: false,
    popUpModalInputValue: null,
    popUpModalPlaceholder: "",
    popUpModalButtons: ['Cancel', 'Confirm'],
    usePopUpModalInput: false,

    // ConfirmDangerousActionModal state
    dangerousActionTitle: "",
    dangerousActionMessage: "",
    dangerousActionIcon: "trash",
    dangerousActionConfirmLabel: "",
    dangerousActionConfirmText: "",
    dangerousActionFunction: null,
    dangerousActionShowInput: true,
    dangerousActionInputSecret: false,
    dangerousActionRequireExactInput: true,
    dangerousActionShowToggle: false,
    dangerousActionToggleLabel: "",
    dangerousActionToggleOffHint: "",
    dangerousActionToggleOnHint: "",

    activeModal: null,

    shareModalData: null,

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
    resetPopUpModal() {
      this.popUpModalTitle = "";
      this.popUpModalMessage = "";
      this.popUpModalIcon = "";
      this.popUpModalFunction = null;
      this.popUpModalLoading = false;
      this.popUpModalInputValue = null;
      this.popUpModalPlaceholder = "";
      this.popUpModalButtons = ["Cancel", "Confirm"];
      this.usePopUpModalInput = false;
    },
    async togglePin(user = false) {
      Events.Emit("pin-tray-window", !this.pin);
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
      await templateStore.reloadTemplates();
      await workflowStore.reloadWorkflows();
      await statusStore.reloadStatuses();
      await dependencyStore.reloadDependencyTypes();
      await tagStore.reloadTags();
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
