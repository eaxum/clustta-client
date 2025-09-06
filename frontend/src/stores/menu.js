import { defineStore } from "pinia";
import { nextTick } from "vue";

export const useMenu = defineStore("useMenu", {
  state: () => ({
    menuStates: {
      typeFilterMenu: false,
      assetTypeFilterMenu: false,
      resourceTypeFilterMenu: false,
      collectionTypeFilterMenu: false,
      statusFilterMenu: false,
      stateFilterMenu: false,
      extensionFilterMenu: false,
      tagsFilterMenu: false,
      assigneeFilterMenu: false,

      projectMenu: false,
      projectItemMenu: false,
      collectionMenu: false,
      untrackedItemMenu: false,
      assetMenu: false,
      resourceItemMenu: false,
      assignMenu: false,
      accountMenu: false,
    },

    activeMenu: null,
    nonFilterMenus: [
      'projectMenu',
      'projectItemMenu',
      'collectionMenu', 
      'assetMenu', 
      'untrackedItemMenu',
      'resourceItemMenu', 
      'assignMenu',
      'accountMenu'
    ],

    menuEl: null,
    position: { x: 0, y: 0 },
    anchorMenu: false,
    isAnimating: false,
    contextMenuVisible: false,
    taskPopUpMenuVisible: false,
    popupTrigger: null,

    popUpMenuTop: 0,
    popUpMenuLeft: 0,
    popUpMenuWidth: 0,
    popUpMenuHeight: 0,

    clickOutsideMask: null,
    listBoxExpanded: false,
    contextMenuBounds: null,
    contextMenu: null,

    popUpMenu: null,
    taskListContainer: null,

    flyoutTop: 0,
    flyoutLeft: 0,
    flyoutWidth: 0,
    showStatusOptions: false,
  }),

  getters: {
    getContextMenuBounds: (state) => {
      return state.contextMenuBounds || document.body;
    }
  },
  actions: {
    async showContextMenu(event, menuName, value, options = false) {
      this.setMenuVisibility(menuName, value);

      const targetRect = event.target.getBoundingClientRect();
      const targetTop = targetRect.y + targetRect.height + 15;
      const targetLeft = targetRect.x;
      const targetRight = targetRect.x + targetRect.width + 15;
      
      let x;
      let y;
      
      // Handle backward compatibility - if options is boolean (old anchor parameter)
      if (typeof options === 'boolean') {
        if (options) {
          this.anchorMenu = true;
          y = targetTop;
          x = targetLeft;
        } else {
          this.anchorMenu = false;
          y = event.clientY;
          x = event.clientX;
        }
      } else if (options && typeof options === 'object') {
        // New options object format
        this.anchorMenu = options.anchor || false;
        
        if (options.anchor) {
          switch (options.position) {
            case 'right':
              x = targetRight;
              y = targetRect.y; // Align to top of target
              break;
            case 'left':
              x = targetLeft - 15; // Add some spacing
              y = targetRect.y;
              break;
            case 'bottom':
              x = targetLeft;
              y = targetTop;
              break;
            case 'top':
              x = targetLeft;
              y = targetRect.y - 15; // Above the target
              break;
            default:
              // Default to bottom positioning (original behavior)
              x = targetLeft;
              y = targetTop;
          }
        } else {
          y = event.clientY;
          x = event.clientX;
        }
      } else {
        // No options provided, use default behavior
        this.anchorMenu = false;
        y = event.clientY;
        x = event.clientX;
      }

      const newPosition = { x, y };

      if (this.contextMenuVisible && !this.isAnimating) {
        this.contextMenuVisible = false;
        await nextTick();
        this.position = newPosition;
        this.contextMenuVisible = true;
      } else {
        this.position = newPosition;
        this.contextMenuVisible = true;
      }
    },

    hideContextMenu(event) {
      if (this.contextMenuVisible) {
        this.contextMenuVisible = false;
        this.disableAllMenus();
      }
    },

    setMenuVisibility(menuName, value) {
      if (this.menuStates.hasOwnProperty(menuName)) {
        // Check if the menu is already active
        if (value && this.activeMenu !== null && this.activeMenu !== menuName) {
          // Disable the currently active menu
          this.menuStates[this.activeMenu] = false;
        }

        this.menuStates[menuName] = value;
        this.activeMenu = value ? menuName : null;
      }
    },

    disableAllMenus() {
      for (const menuName in this.menuStates) {
        this.menuStates[menuName] = false;
      }
      this.activeMenu = null;
      this.contextMenuVisible = false;
    },

    isAnyMenuActive() {
      return Object.values(this.menuStates).some((isVisible) => isVisible);
    },

    triggerMenuItem(menu, page) {
      this.setModalVisibility(menu, false);
      this.showTraySearch = false;
      router.push({ name: page });
    },

    calculatePopUpPosition(event) {
      const targetTop = event.target.getBoundingClientRect().top;
      const targetHeight = event.target.getBoundingClientRect().height;
      const targetLeft = event.target.getBoundingClientRect().left;

      const container = this.contextMenuBounds.getBoundingClientRect();
      const contextMenu = this.contextMenu.getBoundingClientRect();

      const popUpMenuHeight = contextMenu.height;
      const top = container.top;
      const bottom = container.bottom;
      const spaceAbove = targetTop - top;
      const spaceBelow = bottom - targetTop;
      const halfHeight = popUpMenuHeight / 2;
      const offset = 10;

      if (spaceAbove < halfHeight) {
        this.popUpMenuTop = top + offset;
      } else if (spaceBelow < halfHeight) {
        this.popUpMenuTop = bottom - popUpMenuHeight - offset;
      } else {
        this.popUpMenuTop = targetTop - halfHeight + targetHeight / 2;
      }

      this.popUpMenuLeft = targetLeft - this.popUpMenuWidth - 10;
    },

    hidePopUpMenu() {
      this.taskPopUpMenuVisible = false;
    },

    // Force reposition of menu (useful when menu content changes height)
    forceRepositionMenu() {
      if (this.contextMenuVisible) {
        this.contextMenuVisible = false;
        nextTick(() => {
          this.contextMenuVisible = true;
        });
      }
    },
  },
});
